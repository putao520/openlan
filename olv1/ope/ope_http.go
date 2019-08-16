package olv1ope

import (
	"fmt"
	"html"
	"log"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

type User struct {
	Name string `json:"name"`
	Token string `json:"token"`
	Password string `json:"password"`
}

type OpeHttp struct {
	wroker *OpeWroker
	listen string
	Users map[string]User
	adminToken string
}

func NewOpeHttp(wroker *OpeWroker, listen string, token string)(this *OpeHttp) {
	this = &OpeHttp {
		wroker: wroker,
		listen: listen,
		adminToken: token,
		Users: make(map[string]User),
	}

	//TODO save token to default files.

	http.HandleFunc("/", this.Index)
	http.HandleFunc("/hi", this.Hi)
	http.HandleFunc("/user", this.User)
	return 
}

func (this *OpeHttp) GoStart() error {
	log.Printf("NewHttp on %s", this.listen)
    if err := http.ListenAndServe(this.listen, nil); err != nil {
		log.Printf("Error| OpeHttp.GoStart on %s: %s", this.listen, err)
		return err
	}
	return nil
}

func (this *OpeHttp) Hi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, %s %q", r.Method, html.EscapeString(r.URL.Path))
}

func (this *OpeHttp) Index(w http.ResponseWriter, r *http.Request) {
	switch (r.Method) {
	case "GET":  
		body := "remoteaddr, localdevice, receipt, transmission, error\n"
		for client, ifce := range this.wroker.Clients {
			body += fmt.Sprintf("%s, %s, %d, %d, %d\n", 
								client.GetAddr(), ifce.Name(),
								client.RxOkay, client.TxOkay, client.TxError)
		}
		fmt.Fprintf(w, body)
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 500)
		return 
	}

	
}

func (this *OpeHttp) User(w http.ResponseWriter, r *http.Request) {
	//TODO authority by adminToken.

	switch (r.Method) {
	case "GET":
		pagesJson, err := json.Marshal(this.Users)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("Error| OpeHttp.User: %s", err))
			return
		}

		fmt.Fprintf(w, string(pagesJson))
	case "POST":
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error| OpeHttp.User: %s", err), 500)
			return
		}

		var user User
		if err := json.Unmarshal([]byte(body), &user); err != nil {
			http.Error(w, fmt.Sprintf("Error| OpeHttp.User: %s", err), 500)
			return
		}

		if user.Name != "" {
			this.Users[user.Name] = user
		} else if (user.Token != "") {
			this.Users[user.Token] = user
		}

		fmt.Fprintf(w, "Saved it.")
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 500)
	}
}