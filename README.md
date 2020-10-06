# Node Watcher F5

Watches for odes in kubernetes and creates corresponding F5 pools and rules. 

## Settings

** Required **

* `F5_URL` - F5 Url `https://my.f5.com`
* `F5_USERNAME` - F5 login
* `F5_PASSWORD` - F5 password
* `F5_INSIDE_MONITOR` and `F5_MONITOR` - In the format of `/PARTITION/monitor` to use for monitoring node ports for the INSIDE (private) apps, and public apps
* `F5_INSIDE_PARTITION` and `F5_PARTITION` - The inside and outside partition name (e.g., `DEV_INSIDE`, `DEV_OUTSIDE`)
* `UNIPOOL_INSIDE` and `UNIPOOL` - The name of the unipool to use to add nodes (and remove them from).
* `INSIDE_MONITOR_PORT` and `DEFAULT_MONITOR_PORT` - The port on all kubernetes nodes to use to monitor if the node is healthy or not.  The inside should be used to tell whether the inside is healthy, the default should check whether the outside is healthy.

** Optional **
* `CLUSTER` - The cluster name
* `INFLUX_LINE_IP` - The ip address of influx to use to report changes to node statuses. 
* `USE_LOCAL_KUBE_CONTEXT` - Set to "true" to use the local kubernetes config, otherwise this will try to look for a mounted kubernetes service account. This service account must have the ability to WATCH/LIST/GET for nodes.

## Building

```
./build.sh
```

## Running

```
./start.sh
```

## Running (locally)

```
USE_LOCAL_KUBE_CONTEXT=true ./start.sh
```