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

// GetHostsExcludingGroups retorna hosts ativos (status=0) incluindo informações de grupo
// e exclui quaisquer hosts que pertençam a grupos cujo nome esteja na lista excludeNames.
func (c *Client) GetHostsExcludingGroups(excludeNames []string) ([]Host, error) {
	params := map[string]interface{}{
		"output":       "extend",
		"filter":       map[string]string{"status": "0"},
		"selectGroups": "extend",
	}

	resp, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	// Tipo auxiliar para leitura com grupos
	var rawHosts []struct {
		Hostid string `json:"hostid"`
		Host   string `json:"host"`
		Status string `json:"status"`
		Groups []struct {
			Name string `json:"name"`
		} `json:"groups"`
	}

	json.Unmarshal(resp, &rawHosts)

	// Cria mapa de exclusão para checagem rápida
	excludeMap := make(map[string]bool)
	for _, n := range excludeNames {
		excludeMap[n] = true
	}

	var hosts []Host
	for _, rh := range rawHosts {
		skip := false
		for _, g := range rh.Groups {
			if excludeMap[g.Name] {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		hosts = append(hosts, Host{Hostid: rh.Hostid, Host: rh.Host, Status: rh.Status})
	}

	return hosts, nil
}
