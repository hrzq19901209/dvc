package main

import (
	"bughunter.com/dvc/config"
	"fmt"
)

func main() {
	config.LoadConfig("test.conf")
	conf := config.GetConfig()
	fmt.Println(conf.IP)
	fmt.Println(conf.Hostname)
	fmt.Println(conf.Cluster)
	fmt.Println(conf.EtcdEndpoints)
}
