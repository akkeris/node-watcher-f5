package utils

import (
	"fmt"
	"k8s.io/client-go/rest"
	"os"
)

var Partition string
var InsidePartition string

var Monitor string
var InsideMonitor string

var Virtual string
var InsideVirtual string

var Defaultdomain string
var InsideDomain string

var Defaultmonitorport string
var Insidemonitorport string

var Unipool string
var UnipoolInside string

var Client rest.Interface

func Variableinit() {

	fmt.Println("setting Partition to " + os.Getenv("F5_PARTITION"))
	Partition = os.Getenv("F5_PARTITION")

	fmt.Println("setting InsidePartition to " + os.Getenv("F5_INSIDE_PARTITION"))
	InsidePartition = os.Getenv("F5_INSIDE_PARTITION")

	fmt.Println("setting Monitor to " + os.Getenv("F5_MONITOR"))
	Monitor = os.Getenv("F5_MONITOR")

	fmt.Println("setting InsideMonitor to " + os.Getenv("F5_INSIDE_MONITOR"))
	InsideMonitor = os.Getenv("F5_INSIDE_MONITOR")

	fmt.Println("setting f5virtual to " + os.Getenv("F5_VIRTUAL"))
	f5virtual := os.Getenv("F5_VIRTUAL")

	fmt.Println("setting f5insidevirtual to " + os.Getenv("F5_INSIDE_VIRTUAL"))
	f5insidevirtual := os.Getenv("F5_INSIDE_VIRTUAL")

	fmt.Println("setting Virtual to " + "~" + Partition + "~" + f5virtual)
	Virtual = "~" + Partition + "~" + f5virtual
	Virtual = f5virtual

	fmt.Println("setting InsideVirtual to " + "~" + InsidePartition + "~" + f5insidevirtual)
	InsideVirtual = "~" + InsidePartition + "~" + f5insidevirtual
	InsideVirtual = f5insidevirtual

	fmt.Println("setting Defaultdomain to " + os.Getenv("DEFAULT_DOMAIN"))
	Defaultdomain = os.Getenv("DEFAULT_DOMAIN")

	fmt.Println("setting InsideDomain to " + os.Getenv("INSIDE_DOMAIN"))
	InsideDomain = os.Getenv("INSIDE_DOMAIN")

        fmt.Println("setting Defaultmonitorport to "+os.Getenv("DEFAULT_MONITOR_PORT"))
        Defaultmonitorport = os.Getenv("DEFAULT_MONITOR_PORT")

        fmt.Println("setting Insidemonitorport to "+os.Getenv("INSIDE_MONITOR_PORT"))
        Insidemonitorport = os.Getenv("INSIDE_MONITOR_PORT")

	Unipool = os.Getenv("UNIPOOL")
	UnipoolInside = os.Getenv("UNIPOOL_INSIDE")
}
