package main

import (
	"fmt"
	nodes "github.com/akkeris/node-watcher-f5/nodes"
	utils "github.com/akkeris/node-watcher-f5/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"os"
	"path/filepath"
	"time"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func kubernetesConfig() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// An explicit instruction must be given to allow us to use the local kube config.
		if os.Getenv("USE_LOCAL_KUBE_CONTEXT") == "true" {
			fmt.Println("[process] Using local kubernetes current context!")
			config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homeDir(), ".kube", "config"))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func reportNodeStatusToInflux(oldobj interface{}, newobj interface{}) {
	if os.Getenv("INFLUX_LINE_IP") != "" {
		var conn net.Conn
		conn, err := net.Dial("udp", os.Getenv("INFLUX_LINE_IP"))
		if err != nil {
			fmt.Printf("Unable to connect to influx to report changes to node status: %s\n", err.Error())
		} else {
			defer conn.Close()
			name := newobj.(*corev1.Node).ObjectMeta.Name
			conditions := newobj.(*corev1.Node).Status.Conditions
			for _, element := range conditions {
				if element.Status == "True" {
					fmt.Fprintf(conn, "kubernetes.node.condition."+fmt.Sprintf("%v", element.Type)+",cluster="+os.Getenv("CLUSTER")+",node="+name+" value=1")
				} else {
					fmt.Fprintf(conn, "kubernetes.node.condition."+fmt.Sprintf("%v", element.Type)+",cluster="+os.Getenv("CLUSTER")+",node="+name+" value=0")
				}
			}
		}
	}
}

var clientset kubernetes.Interface

func main() {
	utils.Variableinit()
	utils.Startclient()

	var err error
	clientset, err = kubernetesConfig()
	rest := clientset.CoreV1().RESTClient()
	if err != nil {
		panic(err.Error())
	}
	listWatch := cache.NewListWatchFromClient(rest, "nodes", "", fields.Everything())

	listWatch.ListFunc = func(options api.ListOptions) (runtime.Object, error) {
		return rest.Get().Resource("nodes").Do().Get()
	}
	listWatch.WatchFunc = func(options api.ListOptions) (watch.Interface, error) {
		return clientset.CoreV1().Nodes().Watch(v1.ListOptions{})
	}

	_, controller := cache.NewInformer(
		listWatch, 
		&corev1.Node{},
		time.Second*0, 
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				// Ingore any node "adds" where the node is more than five minutes old.
				if (v1.Now().Unix() - obj.(*corev1.Node).ObjectMeta.CreationTimestamp.Unix()) <= 300 {
					nodes.AddNodeToF5(&clientset, obj.(*corev1.Node), utils.Partition, utils.Unipool, utils.Monitor, utils.Defaultmonitorport)
					nodes.AddNodeToF5(&clientset, obj.(*corev1.Node), utils.InsidePartition, utils.UnipoolInside, utils.InsideMonitor, utils.Insidemonitorport)
				}
			},
			DeleteFunc: func(obj interface{}) {
				nodes.RemoveNodeFromF5(&clientset, obj.(*corev1.Node), utils.Partition, utils.Unipool, utils.Monitor, utils.Defaultmonitorport)
				nodes.RemoveNodeFromF5(&clientset, obj.(*corev1.Node), utils.InsidePartition, utils.UnipoolInside, utils.InsideMonitor, utils.Insidemonitorport)
			},
			UpdateFunc: reportNodeStatusToInflux,
		},
	)
	fmt.Println("[process] Syncing changes between F5 and kubernetes...")
	nodes.ResyncNodes(&clientset, utils.Partition, utils.Unipool, utils.Monitor, utils.Defaultmonitorport)
	nodes.ResyncNodes(&clientset, utils.InsidePartition, utils.UnipoolInside, utils.InsideMonitor, utils.Insidemonitorport)
	fmt.Println("[process] Watching for changes in nodes...")
	controller.Run(wait.NeverStop)

}
