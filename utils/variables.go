package utils

import (
	"fmt"
	"os"
)

var Partition string
var InsidePartition string

var Monitor string
var InsideMonitor string

var Defaultmonitorport string
var Insidemonitorport string

var Unipool string
var UnipoolInside string

func Variableinit() {
	fmt.Println("[variables] Setting Partition to " + os.Getenv("F5_PARTITION"))
	Partition = os.Getenv("F5_PARTITION")

	fmt.Println("[variables] Setting InsidePartition to " + os.Getenv("F5_INSIDE_PARTITION"))
	InsidePartition = os.Getenv("F5_INSIDE_PARTITION")

	fmt.Println("[variables] Setting Monitor to " + os.Getenv("F5_MONITOR"))
	Monitor = os.Getenv("F5_MONITOR")

	fmt.Println("[variables] Setting InsideMonitor to " + os.Getenv("F5_INSIDE_MONITOR"))
	InsideMonitor = os.Getenv("F5_INSIDE_MONITOR")

	fmt.Println("[variables] Setting Defaultmonitorport to " + os.Getenv("DEFAULT_MONITOR_PORT"))
	Defaultmonitorport = os.Getenv("DEFAULT_MONITOR_PORT")

	fmt.Println("[variables] Setting Insidemonitorport to " + os.Getenv("INSIDE_MONITOR_PORT"))
	Insidemonitorport = os.Getenv("INSIDE_MONITOR_PORT")

	fmt.Println("[variables] Setting unipool to " + os.Getenv("UNIPOOL"))
	Unipool = os.Getenv("UNIPOOL")

	fmt.Println("[variables] Setting unipool inside to" + os.Getenv("UNIPOOL_INSIDE"))
	UnipoolInside = os.Getenv("UNIPOOL_INSIDE")
}
