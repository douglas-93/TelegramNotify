package zabbix

import "encoding/json"

type Host struct {
	Hostid    string `json:"hostid"`
	Host      string `json:"host"`
	Status    string `json:"status"` /* 0 - Ativo 1 - Inativo */
	Lastvalue string `json:"lastvalue"`
	Prevvalue string `json:"prevvalue"`
	Error     bool
}

func (c *Client) GetHosts() ([]Host, error) {
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]string{
			"status": "0",
		},
	}

	resp, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	var hosts []Host
	json.Unmarshal(resp, &hosts)
	return hosts, nil
}
