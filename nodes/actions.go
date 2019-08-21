package nodes

import (
	pools "github.com/akkeris/node-watcher-f5/pools"
	structs "github.com/akkeris/node-watcher-f5/structs"
	utils "github.com/akkeris/node-watcher-f5/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
)

func RemoveFromUnipool(node *corev1.Node) {
	uidparts := strings.Split(fmt.Sprintf("%v", node.ObjectMeta.UID), "-")
	nodeid := "uid" + uidparts[0]

	utils.NewToken()
	nodes, err := getNodes(utils.Partition)
	if err != nil {
		fmt.Println(err)
	}
	for _, element := range nodes {
		createnode(element)
	}

	pools.ResetUnipool(utils.Partition, nodes, utils.Monitor, utils.Defaultmonitorport)
	deleteNode(utils.Partition, nodeid)

	nodesi, err := getNodes(utils.InsidePartition)
	if err != nil {
		fmt.Println(err)
	}
	for _, element := range nodesi {
		createnode(element)
	}
	pools.ResetUnipool(utils.InsidePartition, nodesi, utils.InsideMonitor, utils.Insidemonitorport)
	deleteNode(utils.InsidePartition, nodeid)
}

func AddToUnipool(node *corev1.Node) {
	utils.NewToken()
	nodes, err := getNodes(utils.Partition)
	if err != nil {
		fmt.Println(err)
	}
	for _, element := range nodes {
		createnode(element)
	}

	pools.ResetUnipool(utils.Partition, nodes, utils.Monitor, utils.Defaultmonitorport)

	nodesi, err := getNodes(utils.InsidePartition)
	if err != nil {
		fmt.Println(err)
	}
	for _, element := range nodesi {
		createnode(element)
	}
	pools.ResetUnipool(utils.InsidePartition, nodesi, utils.InsideMonitor, utils.Insidemonitorport)
}

func getNodes(p string) (n []structs.Node, e error) {
	var nodesids []structs.Node
	nodesresp := utils.Client.Get().Resource("nodes").Do()
	nodes, err := nodesresp.Raw()
	if err != nil {
		fmt.Println(err)
		return nodesids, err
	}
	var nodelist corev1.NodeList
	err = json.Unmarshal(nodes, &nodelist)
	if err != nil {
		fmt.Println(err)
		return nodesids, err
	}
	for _, element := range nodelist.Items {
		if !element.Spec.Unschedulable {
			for name, value := range element.Labels {
				if name == "node-role.kubernetes.io/worker" && value == "true" {
					ip := element.Status.Addresses[0].Address
					uidparts := strings.Split(fmt.Sprintf("%v", element.ObjectMeta.UID), "-")
					nodeid := "uid" + uidparts[0]
					var curnode structs.Node
					curnode.Name = nodeid
					curnode.Partition = p
					curnode.Address = ip
					nodesids = append(nodesids, curnode)
				}
			}
		}

	}
	return nodesids, nil
}

func createnode(node structs.Node) {
	urlStr := utils.F5url + "/mgmt/tm/ltm/node"
	str, err := json.Marshal(node)
	if err != nil {
		fmt.Println("Error preparing request")
	}
	jsonStr := []byte(string(str))
	req, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonStr))
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	fmt.Println(resp.Status)
	if err != nil {
		fmt.Println(err)
	}
	var rnode structs.Node

	defer resp.Body.Close()
	bb, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bb))
	_ = json.Unmarshal(bb, &rnode)
}

func deleteNode(partition string, node string) {
	newnodename := "~" + partition + "~" + node
	fmt.Println(newnodename)
	urlStr := utils.F5url + "/mgmt/tm/ltm/node/" + newnodename
	req, _ := http.NewRequest("DELETE", urlStr, nil)
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	bodybytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bodybytes[:]))
	if resp.StatusCode == 200 {
		fmt.Println("Node deleted")
	}

}