package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"
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

type Cmd struct {
}

func (c Cmd) Output(data interface{}, format string, tmpl string) error {
	switch format {
	case "json":
		if out, err := libol.Marshal(data, true); err == nil {
			fmt.Println(string(out))
		} else {
			return err
		}
	case "yaml":
		if out, err := yaml.Marshal(data); err == nil {
			fmt.Println(string(out))
		} else {
			return err
		}
	default:
		funcMap := template.FuncMap{
			"ps": func(space int, args ...interface{}) string {
				format := "%" + strconv.Itoa(space) + "s"
				if space < 0 {
					format = "%-" + strconv.Itoa(space) + "s"
				}
				return fmt.Sprintf(format, args...)
			},
		}
		if tmpl, err := template.New("main").Funcs(funcMap).Parse(tmpl); err != nil {
			return err
		} else {
			if err := tmpl.Execute(os.Stdout, data); err != nil {
				return err
			}
		}
	}
	return nil
}
