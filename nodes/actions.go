package nodes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	utils "github.com/akkeris/node-watcher-f5/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Node struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	Address   string `json:"address"`
	Labels    struct {
		NodeRoleKubernetesIoWorker string `json:"node-role.kubernetes.io/worker"`
	}
}

type NodeList struct {
	Items []Node `json:"items"`
}

type Members struct {
	Kind     string `json:"kind"`
	SelfLink string `json:"selfLink"`
	Items    []struct {
		Kind            string `json:"kind"`
		Name            string `json:"name"`
		Partition       string `json:"partition"`
		FullPath        string `json:"fullPath"`
		Generation      int    `json:"generation"`
		SelfLink        string `json:"selfLink"`
		Address         string `json:"address"`
		ConnectionLimit int    `json:"connectionLimit"`
		DynamicRatio    int    `json:"dynamicRatio"`
		Ephemeral       string `json:"ephemeral"`
		Fqdn            struct {
			Autopopulate string `json:"autopopulate"`
		} `json:"fqdn"`
		InheritProfile string `json:"inheritProfile"`
		Logging        string `json:"logging"`
		Monitor        string `json:"monitor"`
		PriorityGroup  int    `json:"priorityGroup"`
		RateLimit      string `json:"rateLimit"`
		Ratio          int    `json:"ratio"`
		Session        string `json:"session"`
		State          string `json:"state"`
	} `json:"items"`
}

type Memberlist struct {
	Members []Memberlistmember `json:"members"`
}

type Memberlistmember struct {
	Name    string `json:"name"`
	Monitor string `json:"monitor"`
}

func getNodePrefix() string {
	prefix, ok := os.LookupEnv("NODE_PREFIX")
	if !ok || prefix == "" {
		return "uid"
	} else {
		return prefix
	}
}

func IsWorkerNode(kubeNode *corev1.Node) bool {
	if kubeNode == nil || kubeNode.Labels == nil {
		return false
	}
	for name, value := range kubeNode.Labels {
		if name == "node-role.kubernetes.io/worker" && value == "true" {
			return true
		}
	}
	return false
}

func findIpFromKubernetesNode(kubeNode *corev1.Node) (*string, error) {
	for _, address := range kubeNode.Status.Addresses {
		if address.Type == "InternalIP" {
			return &address.Address, nil
		}
	}
	return nil, errors.New("Unable to find IP address")
}

func RemoveNodeFromF5(clientset *kubernetes.Interface, kubeNode *corev1.Node, partition string, poolName string, monitorName string, monitorPort string) {
	uidparts := strings.Split(fmt.Sprintf("%v", kubeNode.ObjectMeta.UID), "-")
	nodeName := getNodePrefix() + uidparts[0]
	fmt.Printf("[actions] Received request to remove node: /%s/%s\n", partition, nodeName)
	ResyncNodes(clientset, partition, poolName, monitorName, monitorPort)
}

func AddNodeToF5(clientset *kubernetes.Interface, kubeNode *corev1.Node, partition string, poolName string, monitorName string, monitorPort string) {
	uidparts := strings.Split(fmt.Sprintf("%v", kubeNode.ObjectMeta.UID), "-")
	nodeName := getNodePrefix() + uidparts[0]
	fmt.Printf("[actions] Received request to add node: /%s/%s\n", partition, nodeName)
	ResyncNodes(clientset, partition, poolName, monitorName, monitorPort)
}

func ResyncNodes(clientset *kubernetes.Interface, partition string, poolName string, monitorName string, monitorPort string) {
	f5Nodes := GetNodesFromF5(partition)
	kubeNodes := GetNodesFromKubernetes(clientset, partition)
	f5NodesToRemove := make([]Node, 0)
	newF5Nodes := make([]Node, 0)
	var dirty = false
	// Find nodes that need to be removed
	for _, f5Node := range f5Nodes.Items {
		var found = false
		for _, kubeNode := range kubeNodes {
			if kubeNode.Name == f5Node.Name {
				found = true
				newF5Nodes = append(newF5Nodes, f5Node)
			}
		}
		if found == false {
			// We'll need to update the pools at a later time then
			// we can try and remove the node.
			dirty = true
			f5NodesToRemove = append(f5NodesToRemove, f5Node)
		}
	}
	// Find nodes that need to be added from kube
	for _, kubeNode := range kubeNodes {
		var found = false
		for _, f5Node := range f5Nodes.Items {
			if kubeNode.Name == f5Node.Name {
				found = true
			}
		}
		if found == false {
			dirty = true
			fmt.Printf("[actions] Adding node: %s (%s)\n", kubeNode.Name, kubeNode.Address)
			CreateNodeOnF5(kubeNode)
			newF5Nodes = append(newF5Nodes, kubeNode)
		}
	}

	if dirty == true {
		UpdatePool(partition, newF5Nodes, poolName, monitorName, monitorPort)
	}

	if len(f5NodesToRemove) != 0 {
		for _, f5Node := range f5NodesToRemove {
			fmt.Printf("[actions] Removing node: /%s/%s (%s)\n", partition, f5Node.Name, f5Node.Address)
			DeleteNodeOnF5(partition, f5Node.Name)
		}
	}
}

func UpdatePool(partition string, nodes []Node, pool string, monitor string, poolPort string) {
	fmt.Printf("[actions] Updating pool %s on partition %s with monitor %s and port %s and members: %#+v\n", pool, partition, monitor, poolPort, nodes)
	var memberList Memberlist
	memberList.Members = make([]Memberlistmember, 0)
	for _, node := range nodes {
		memberList.Members = append(memberList.Members, Memberlistmember{
			Name:    node.Name + ":" + poolPort,
			Monitor: monitor,
		})
	}
	payload, err := json.Marshal(memberList)
	if err != nil {
		panic(err)
	}
	url := utils.F5url + "/mgmt/tm/ltm/pool/" + strings.Replace("/"+partition+"/"+pool, "/", "~", -1)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode > 299 {
		panic(errors.New(string(body)))
	}
}

func GetNodesFromKubernetes(clientset *kubernetes.Interface, partition string) []Node {
	var nodesids []Node
	nodesresp := (*clientset).CoreV1().RESTClient().Get().Resource("nodes").Do()
	nodes, err := nodesresp.Raw()
	if err != nil {
		panic(err)
	}
	var nodelist corev1.NodeList
	if err := json.Unmarshal(nodes, &nodelist); err != nil {
		panic(err)
	}
	for _, element := range nodelist.Items {
		if !element.Spec.Unschedulable {
			for name, value := range element.Labels {
				if name == "node-role.kubernetes.io/worker" && value == "true" {
					ip := element.Status.Addresses[0].Address
					uidparts := strings.Split(fmt.Sprintf("%v", element.ObjectMeta.UID), "-")
					nodesids = append(nodesids, Node{
						Name:      getNodePrefix() + uidparts[0],
						Partition: partition,
						Address:   ip,
					})
				}
			}
		}
	}
	return nodesids
}

func GetNodesFromF5(partition string) NodeList {
	req, err := http.NewRequest("GET", utils.F5url+"/mgmt/tm/ltm/node?$filter=partition+eq+"+partition, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	var nodes NodeList
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &nodes); err != nil {
		panic(errors.New(err.Error() + ": " + string(body) + ": " + resp.Status))
	}
	if resp.StatusCode > 299 {
		panic(errors.New(string(body)))
	}
	newNodes := make([]Node, 0)
	for _, node := range nodes.Items {
		if strings.HasPrefix(node.Name, getNodePrefix()) {
			newNodes = append(newNodes, node)
		}
	}
	nodes.Items = newNodes
	return nodes
}

func CreateNodeOnF5(node Node) {
	payload, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", utils.F5url+"/mgmt/tm/ltm/node", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	var rnode Node
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &rnode); err != nil {
		panic(errors.New(err.Error() + ": " + string(body) + ": " + resp.Status))
	}
	if resp.StatusCode > 299 {
		panic(errors.New(string(body)))
	}
	fmt.Printf("Created node %#+v received response: %s\n", node, resp.Status)
}

func DeleteNodeOnF5(partition string, node string) {
	req, err := http.NewRequest("DELETE", utils.F5url+"/mgmt/tm/ltm/node/~"+partition+"~"+node, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := utils.F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(errors.New(string(body)))
	}
	fmt.Printf("Deleted node %#+v received response: %s\n", node, resp.Status)
}
