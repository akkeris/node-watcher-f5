package main

import (
	k8sconfig "github.com/akkeris/node-watcher-f5/k8sconfig"
	nodes "github.com/akkeris/node-watcher-f5/nodes"
	utils "github.com/akkeris/node-watcher-f5/utils"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

var cnt int

func main() {
	cnt = 0
	utils.Variableinit()
	utils.Startclient()
	
	k8sconfig.CreateConfig()
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	utils.Client = clientset.CoreV1().RESTClient()

	if os.Getenv("RESET_NODEPOOL") == "true" {
		nodes.AddToUnipool()
	} else {
		listWatch := cache.NewListWatchFromClient(
			utils.Client, "nodes", "",
			fields.Everything())

		listWatch.ListFunc = func(options api.ListOptions) (runtime.Object, error) {
			return utils.Client.Get().Resource("nodes").Do().Get()
		}
		listWatch.WatchFunc = func(options api.ListOptions) (watch.Interface, error) {
			return clientset.CoreV1().Nodes().Watch(v1.ListOptions{})
		}

		store, controller := cache.NewInformer(
			listWatch, &corev1.Node{},
			time.Second*0, cache.ResourceEventHandlerFuncs{
				AddFunc:    printNodeAdd,
				DeleteFunc: printNodeDelete,
				UpdateFunc: printNodeUpdate,
			},
		)
		fmt.Println(store.ListKeys())

		fmt.Println("Watching for changes in Nodes....")

		controller.Run(wait.NeverStop)
	}

}

func printNodeUpdate(oldobj interface{}, newobj interface{}) {
	name := newobj.(*corev1.Node).ObjectMeta.Name
	conditions := newobj.(*corev1.Node).Status.Conditions
	var conn net.Conn
	conn, err := net.Dial("udp", os.Getenv("INFLUX_LINE_IP"))
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	for _, element := range conditions {
		var value string
		if element.Status == "True" {
			value = "1"
		}
		if element.Status == "False" {
			value = "0"
		}
		conditiontype := fmt.Sprintf("%v", element.Type)
		line := "kubernetes.node.condition." + conditiontype + ",cluster=" + os.Getenv("CLUSTER") + ",node=" + name + " value=" + value
		fmt.Fprintf(conn, line)
	}
	conn.Close()

}

func printNodeAdd(obj interface{}) {
	created := obj.(*corev1.Node).ObjectMeta.CreationTimestamp.Unix()
	now := v1.Now().Unix()
	diff := now - created
	fmt.Println(diff)
	//     if cnt == 0  {
	if diff < 301 {

		fmt.Println("ADD")
		var jsn []byte
		var err error

		jsn, err = json.Marshal(obj)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsn))
		cnt = 1
		time.Sleep(60 * time.Second)
		nodes.AddToUnipool()
	}
}

func printNodeDelete(obj interface{}) {
	fmt.Println("DELETE")

	var jsn []byte
	var err error

	jsn, err = json.Marshal(obj)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsn))
	nodes.RemoveFromUnipool(obj.(*corev1.Node))

}


