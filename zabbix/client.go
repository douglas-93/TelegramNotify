package zabbix

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

type Client struct {
	URL   string
	Token string
}

type request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Auth    string      `json:"auth,omitempty"`
}

type response struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient() *Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var env map[string]string
	env, e := godotenv.Read()

	if e != nil {
		log.Fatal(e)
	}

	url := env["ZABBIX_API_URL"]
	token := env["ZABBIX_API_TOKEN"]

	return &Client{
		URL:   url,
		Token: token,
	}
}

func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	req := request{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
		Auth:    c.Token,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(c.URL, "application/json", bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r response
	json.NewDecoder(resp.Body).Decode(&r)

	if r.Error != nil {
		return nil, errors.New(r.Error.Message)
	}

	return r.Result, nil
}
