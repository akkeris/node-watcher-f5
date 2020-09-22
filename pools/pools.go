package pools

import (
	structs "github.com/akkeris/node-watcher-f5/structs"
	utils "github.com/akkeris/node-watcher-f5/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ResetUnipool(partition string, nodes []structs.Node, unipool string, monitor string, unipoolport string) {
	poolname := "/" + partition + "/" + unipool
	fixed := strings.Replace(poolname, "/", "~", -1)
	var memberlist structs.Memberlist
	var memberlistmembers []structs.Memberlistmember
	for _, element := range nodes {
		var mlm structs.Memberlistmember
		mlm.Name = element.Name + ":" + unipoolport
		mlm.Monitor = monitor
		memberlistmembers = append(memberlistmembers, mlm)
	}
	memberlist.Members = memberlistmembers
	_, err := json.MarshalIndent(&memberlist, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println("Updating " + poolname)
	urlStr := utils.F5url + "/mgmt/tm/ltm/pool/" + fixed
	str, err := json.Marshal(memberlist)
	if err != nil {
		panic(err)
	}
	jsonStr := []byte(string(str))
	req, _ := http.NewRequest("PUT", urlStr, bytes.NewBuffer(jsonStr))
	req.Header.Add("X-F5-Auth-Token", utils.F5token)
	req.Header.Add("Content-Type", "application/json")
	
	if _, err = utils.F5Client.Do(req); err != nil {
		panic(err)
	}

}
