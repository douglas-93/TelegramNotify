package monitor

import (
	"LapaTelegramBot/zabbix"
	"encoding/json"
	"fmt"
	"sync"
)

func CheckHostsStatus(z *zabbix.Client) ([]string, error) {
	hosts, err := z.GetHosts()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for i := range hosts {
		wg.Add(1)
		go getStatusItemValue(z, &hosts[i], &wg)
	}

	func() {
		wg.Wait()
	}()

	var onlineHosts []string
	var offlineHosts []string
	for _, host := range hosts {
		if host.Lastvalue == "1" && host.Prevvalue == "1" {
			onlineHosts = append(onlineHosts, fmt.Sprintf("✅ %s", host.Host))
		} else {
			offlineHosts = append(offlineHosts, fmt.Sprintf("❌ %s", host.Host))
		}
	}
	return append(onlineHosts, offlineHosts...), nil
}

// CheckHostsStatusExcludingGroups checa hosts, mas exclui hosts que pertençam
// aos grupos listados em excludeNames.
func CheckHostsStatusExcludingGroups(z *zabbix.Client, excludeNames []string) ([]string, error) {
	hosts, err := z.GetHostsExcludingGroups(excludeNames)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for i := range hosts {
		wg.Add(1)
		go getStatusItemValue(z, &hosts[i], &wg)
	}

	func() {
		wg.Wait()
	}()

	var onlineHosts []string
	var offlineHosts []string
	for _, host := range hosts {
		if host.Lastvalue == "1" && host.Prevvalue == "1" {
			onlineHosts = append(onlineHosts, fmt.Sprintf("✅ %s", host.Host))
		} else {
			offlineHosts = append(offlineHosts, fmt.Sprintf("❌ %s", host.Host))
		}
	}
	return append(onlineHosts, offlineHosts...), nil
}

func getStatusItemValue(z *zabbix.Client, host *zabbix.Host, wg *sync.WaitGroup) {
	defer wg.Done()

	params := map[string]interface{}{
		"output":  "extend",
		"hostids": host.Hostid,
		"search": map[string]string{
			"key_": "icmpping",
		},
	}

	resp, err := z.Call("item.get", params)
	if err != nil {
		host.Error = true
		return
	}

	var items []struct {
		Itemid    string `json:"itemid"`
		Lastvalue string `json:"lastvalue"`
		Prevvalue string `json:"prevvalue"`
	}

	json.Unmarshal(resp, &items)

	if len(items) > 0 {
		host.Lastvalue = items[0].Lastvalue
		host.Prevvalue = items[0].Prevvalue
	}
}
