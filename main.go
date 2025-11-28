package main

import (
	"TelegramNotify/monitor"
	"TelegramNotify/zabbix"
	"fmt"
	"log"
)

func main() {
	c := zabbix.NewClient()
	s, e := monitor.CheckOfflineHosts(c)
	if e != nil {
		log.Fatal(e)
	}

	for _, status := range s {
		fmt.Println(status)
	}
}
