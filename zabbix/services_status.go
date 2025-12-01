package zabbix

import "encoding/json"

type ServiceStatus struct {
	Hostid    string `json:"hostid"`
	Name      string `json:"name"`
	Status    string `json:"status"` /* 0 - Ativo 1 - Inativo */
	Lastvalue string `json:"lastvalue"`
	Prevvalue string `json:"prevvalue"`
}

func (c *Client) GetProtheusServiceStatus() ([]ServiceStatus, error) {
	params := map[string]interface{}{
		"output":  "extend",
		"groupid": "25",
		"search": map[string]string{
			"key_": "TOTVS",
		},
	}
	resp, err := c.Call("item.get", params)
	if err != nil {
		return nil, err
	}

	var services []ServiceStatus
	json.Unmarshal(resp, &services)
	return services, nil
}
