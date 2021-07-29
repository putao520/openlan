package cmd

import (
	"bytes"
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

func (cl Client) GetBody(url string) ([]byte, error) {
	client := cl.NewRequest(url)
	r, err := client.Do()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, libol.NewErr(r.Status)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cl Client) GetJSON(url string, v interface{}) error {
	out := cl.Log()
	client := cl.NewRequest(url)
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
	out.Debug("client.GetJSON %s", body)
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}

func (cl Client) SetJSON(client *libol.HttpClient, v interface{}) error {
	out := cl.Log()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	out.Debug("Client.SetJSON %s %s %s", client.Url, client.Method, string(data))
	client.Payload = bytes.NewReader(data)
	if r, err := client.Do(); err != nil {
		return err
	} else if r.StatusCode != http.StatusOK {
		return libol.NewErr(r.Status)
	}
	return nil
}

func (cl Client) PostJSON(url string, v interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "POST"
	return cl.SetJSON(client, v)
}

func (cl Client) PutJSON(url string, v interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "PUT"
	return cl.SetJSON(client, v)
}

func (cl Client) DeleteJSON(url string, v interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "DELETE"
	return cl.SetJSON(client, v)
}

func (cl Client) Log() *libol.SubLogger {
	return libol.NewSubLogger("cli")
}

type Cmd struct {
}

func (c Cmd) NewHttp(token string) Client {
	client := Client{
		Auth: libol.Auth{
			Username: token,
		},
	}
	return client
}

func (c Cmd) Url(prefix, name string) string {
	return ""
}

func (c Cmd) Tmpl() string {
	return ""
}

func (c Cmd) OutJson(data interface{}) error {
	if out, err := libol.Marshal(data, true); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func (c Cmd) OutYaml(data interface{}) error {
	if out, err := yaml.Marshal(data); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func (c Cmd) OutTable(data interface{}, tmpl string) error {
	funcMap := template.FuncMap{
		"ps": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "s"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "s"
			}
			return fmt.Sprintf(format, args...)
		},
		"pi": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "d"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "d"
			}
			return fmt.Sprintf(format, args...)
		},
		"pu": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "u"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "u"
			}
			return fmt.Sprintf(format, args...)
		},
		"pt": func(value int64) string {
			return libol.PrettyTime(value)
		},
		"p2": func(space int, format, key1, key2 string) string {
			value := fmt.Sprintf(format, key1, key2)
			format = "%" + strconv.Itoa(space) + "s"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "s"
			}
			return fmt.Sprintf(format, value)
		},
	}
	if tmpl, err := template.New("main").Funcs(funcMap).Parse(tmpl); err != nil {
		return err
	} else {
		if err := tmpl.Execute(os.Stdout, data); err != nil {
			return err
		}
	}
	return nil
}

func (c Cmd) Out(data interface{}, format string, tmpl string) error {
	switch format {
	case "json":
		return c.OutJson(data)
	case "yaml":
		return c.OutYaml(data)
	default:
		return c.OutTable(data, tmpl)
	}
}

func (c Cmd) Log() *libol.SubLogger {
	return libol.NewSubLogger("cli")
}
