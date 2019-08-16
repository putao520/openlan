package olv1ope

import (
	"os"
	"bufio"
	"log"
	"strings"
	"fmt"
	"errors"
	"sort"
	"encoding/json"

	"github.com/songgao/water"
	"github.com/milosgajdos83/tenus"
	"github.com/danieldin95/openlan-go/olv1/olv1"
)

type OpeWroker struct {
	//Public variable
	Server *TcpServer
	Clients map[*olv1.TcpClient]*water.Interface
	Users map[string]*User

	//Private variable
	verbose int
	br tenus.Bridger
	keys []int
	hooks map[int]func(*olv1.TcpClient, *olv1.Frame) error
	ifmtu int
}

func NewOpeWroker(server *TcpServer, c *Config) (this *OpeWroker) {
	this = &OpeWroker {
		Server: server,
		Clients: make(map[*olv1.TcpClient]*water.Interface, 1024),
		Users: make(map[string]*User, 1024),
		verbose: c.Verbose,
		br: nil,
		ifmtu: 1514,
		hooks: make(map[int]func(*olv1.TcpClient, *olv1.Frame) error),
		keys: make([]int, 0, 1024),
	}

	this.newBr(c.Brname)
	this.setHook(0x10, this.checkAuth)
	this.setHook(0x11, this.handleReq)
	this.showHook()
	this.loadUsers(c.Password)

	return 
}

func (this *OpeWroker) loadUsers(path string) error {
	file, err := os.Open(path)
    if err != nil {
        return err
	}

	defer file.Close()
    reader := bufio.NewReader(file)

    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            break
		}
		
		values := strings.Split(line, ":")
		if len(values) == 2 {
			user := &User{Name: values[0], Password: strings.TrimSpace(values[1])}
			this.Users[user.Name] = user
		}
    }

	return nil
}

func (this *OpeWroker) newBr(brname string) {
	addrs := strings.Split(this.Server.GetAddr(), ":")
	if len(addrs) != 2 {
		log.Printf("Error|OpeWroker.newBr: address: %s", this.Server.GetAddr())
		return
	}

	var err error
	var br tenus.Bridger

	if (brname == "") {
		brname = fmt.Sprintf("brol-%s", addrs[1])
		br, err = tenus.BridgeFromName(brname)
		if err != nil {
			br, err = tenus.NewBridgeWithName(brname)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		br, err = tenus.BridgeFromName(brname)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err = br.SetLinkUp(); err != nil {
		log.Printf("Error|OpeWroker.newBr: %s", err)
	}

	log.Printf("OpeWroker.newBr %s", brname)

	this.br = br
}

func (this *OpeWroker) newTap() (*water.Interface, error) {
	log.Printf("OpeWroker.newTap")	
	ifce, err := water.New(water.Config {
        DeviceType: water.TAP,
    })
    if err != nil {
		log.Printf("Error|OpeWroker.newTap: %s", err)
		return nil, err
	}
	
	link, err := tenus.NewLinkFrom(ifce.Name())
	if err != nil {
		log.Printf("Error|OpeWroker.newTap: Get ifce %s: %s", ifce.Name(), err)
		return nil, err
	}
	
	if err := link.SetLinkUp(); err != nil {
		log.Printf("Error|OpeWroker.newTap: ", err)
	}

	if err := this.br.AddSlaveIfc(link.NetInterface()); err != nil {
		log.Printf("Error|OpeWroker.newTap: Switch ifce %s: %s", ifce.Name(), err)
		return nil, err
	}

	log.Printf("OpeWroker.newTap %s", ifce.Name())	

	return ifce, nil
}

func (this *OpeWroker) Start() {
    go this.Server.GoAccept()
    go this.Server.GoLoop(this.onClient, this.onRecv, this.onClose)
}

func (this *OpeWroker) showHook() {
	for _, k := range this.keys {
		log.Printf("k:%d func: %p", k, this.hooks[k])
	}
} 

func (this *OpeWroker) setHook(index int, hook func(*olv1.TcpClient, *olv1.Frame) error) {
	this.hooks[index] = hook
	this.keys = append(this.keys, index)
	sort.Ints(this.keys)
}

func (this *OpeWroker) onHook(client *olv1.TcpClient, data []byte) error {
	frame := olv1.NewFrame(data)

	for _, k := range this.keys {
		if this.IsVerbose() {
			log.Printf("Debug| OpeWroker.onHook k:%d", k)
		}
		if f, ok := this.hooks[k]; ok {
			if err := f(client, frame); err != nil {
				return err
			}
		}	
	}

	return nil
}

func (this *OpeWroker) checkAuth(client *olv1.TcpClient, frame *olv1.Frame) error {
	if this.IsVerbose() {
		log.Printf("Debug| OpeWroker.checkAuth % x.", frame.Data)
	}

	if olv1.IsInst(frame.Data) {
		action := olv1.DecAction(frame.Data)
		log.Printf("Debug| OpeWroker.checkAuth.action: %s", action)

		if action == "logi=" {
			if err := this.handlelogin(client, olv1.DecBody(frame.Data)); err != nil {
				log.Printf("Error| OpeWroker.checkAuth: %s", err)
				client.SendResp("login", err.Error())
				client.Close()
				return err
			}
			client.SendResp("login", "okay.")
		}

		return nil
	}

	if client.Status != olv1.CL_AUTHED {
		client.Droped++
		if this.IsVerbose() {
			log.Printf("Debug|OpeWroker.onRecv: %s unauth", client.GetAddr())
		}
		return errors.New("Unauthed client.")
	}

	return nil
}

func  (this *OpeWroker) handlelogin(client *olv1.TcpClient, data string) error {
	if this.IsVerbose() {
		log.Printf("Debug| OpeWroker.handlelogin: %s", data)
	}
	user := &User {}
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return errors.New("Invalid json data.")
	}

	name := user.Name
	if user.Token != "" {
	 	name = user.Token
	}

	if _user, ok := this.Users[name]; ok {
		if _user.Password == user.Password {
			client.Status = olv1.CL_AUTHED
			log.Printf("Info| OpeWroker.handlelogin: %s Authed", client.GetAddr())
			this.onAuth(client)
			return nil
		}

		client.Status = olv1.CL_UNAUTH
	}

	return errors.New("Auth failed.")
}

func (this *OpeWroker) handleReq(client *olv1.TcpClient, frame *olv1.Frame) error {
	return nil
}

func (this *OpeWroker) onClient(client *olv1.TcpClient) error {
	client.Status = olv1.CL_CONNECTED
	log.Printf("Info|OpeWroker.onClient: %s", client.GetAddr())	

	return nil
}

func (this *OpeWroker) onAuth(client *olv1.TcpClient) error {
	if client.Status != olv1.CL_AUTHED {
		return errors.New("not authed.")
	}

	log.Printf("Info|OpeWroker.onAuth: %s", client.GetAddr())	

	ifce, err := this.newTap()
	if err != nil {
		return err
	}

	this.Clients[client] = ifce
	go this.GoRecv(ifce, client.SendMsg)

	return nil
}

func (this *OpeWroker) onRecv(client *olv1.TcpClient, data []byte) error {
	//TODO Hook packets such as ARP Learning.
	if this.IsVerbose() {
		log.Printf("Debug|OpeWroker.onRecv: %s % x", client.GetAddr(), data)	
	}

	if err := this.onHook(client, data); err != nil {
		if this.IsVerbose() {
			log.Printf("Debug|OpeWroker.onRecv: %s dropping by %s", client.GetAddr(), err)
			return err
		}
	}

	ifce := this.Clients[client]
	if ifce == nil {
		return errors.New("Tap devices is nil")
	}

	if _, err := ifce.Write(data); err != nil {
		log.Printf("Error|OpeWroker.onRecv: %s", err)
	}

	return nil
}

func (this *OpeWroker) onClose(client *olv1.TcpClient) error {
	log.Printf("Info|OpeWroker.onClose: %s", client.GetAddr())
	if ifce := this.Clients[client]; ifce != nil {
		ifce.Close()
		delete(this.Clients, client)
	} 
	return nil
}

func (this *OpeWroker) Close() {
	this.Server.Close()
}

func (this *OpeWroker) GoRecv(ifce *water.Interface, dorecv func([]byte)(error)) {
	log.Printf("Info|OpeWroker.GoRecv: %s", ifce.Name())	
	defer ifce.Close()
	for {
		data := make([]byte, this.ifmtu)
        n, err := ifce.Read(data)
        if err != nil {
			log.Printf("Error|OpeWroker.GoRev: %s", err)
			break
        }
		if this.IsVerbose() {
			log.Printf("OpeWroker.GoRev: % x\n", data[:n])
		}

		if err := dorecv(data[:n]); err != nil {
			log.Printf("Error|OpeWroker.GoRev: do-recv %s %s", ifce.Name(), err)
		}
	}
}

func (this *OpeWroker) IsVerbose() bool {
	return this.verbose != 0
}