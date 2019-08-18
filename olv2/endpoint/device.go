package endpoint

import (
	"log"

	"github.com/songgao/water"
	"github.com/milosgajdos83/tenus"
)

type Device struct {
	ifce *water.Interface
	ifmtu int
	verbose bool
}

func NewDevice(c *Config) (this *Device) {
	this = &Device {
		ifce: nil,
		ifmtu: c.Ifmtu, //1514
		verbose: c.Verbose,
	}

	ifce, err := water.New(water.Config {DeviceType: water.TAP})
    if err != nil {
        log.Fatal(err)
	}

	this.ifce = ifce
	this.UpLink(this.ifce.Name())

	return
}

func (this *Device) UpLink(name string) error {
	log.Printf("Info| Device.Uplink %s", name)
    link, err := tenus.NewLinkFrom(name)
	if err != nil {
		log.Printf("Error|Device.UpLink: Get ifce %s: %s", name, err)
		return err
	}
	
	if err := link.SetLinkUp(); err != nil {
        log.Printf("Error|Device.UpLink: %s : %s", name, err)
        return err
    }
    
    return nil
}


func (this *Device) GoRecv(doRecv func([]byte)(error)) {
	log.Printf("Info| Device.GoRecv")

	defer this.ifce.Close()
	for {
		data := make([]byte, this.ifmtu)
        n, err := this.ifce.Read(data)
        if err != nil {
			log.Printf("Error|Device.GoRev: %s", err)
			break
		}
		if n > 0 {
			if this.verbose {
				log.Printf("Debug| Device.GoRev: % x\n", data[:n])
			}
			if err := doRecv(data[:n]); err != nil {
				log.Printf("Error| Device.GoRev.doRecv: %s\n", err)
			}
		}
	}
}

func (this *Device) DoSend(data []byte) error {
	if this.verbose {
		log.Printf("Debug| Device.DoSend: % x\n", data)
	}

	if _, err := this.ifce.Write(data); err != nil {
		log.Printf("Error|Device.DoSend: %s", err)	
		return err
	}

	return nil
}