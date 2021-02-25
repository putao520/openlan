package cmd

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Auth libol.Auth
	Host string
}

func (cl Client) NewRequest(url string) *libol.HttpClient {
	client := &libol.HttpClient{
		Auth: libol.Auth{
			Type:     "basic",
			Username: cl.Auth.Username,
			Password: cl.Auth.Password,
		},
		Url: url,
	}
	return client
}

func (cl Client) GetJSON(client *libol.HttpClient, v interface{}) error {
	r, err := client.Do()
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return libol.NewErr(r.Status)
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	libol.Debug("client.GetJSON %s", body)
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}
