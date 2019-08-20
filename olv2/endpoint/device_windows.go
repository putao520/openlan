package endpoint

import (
    "log"

    "github.com/songgao/water"
)

type Device struct {
    ifce *water.Interface
    ifmtu int
    ethmtu int
    verbose bool
}

const ETHLEN = 14

func NewDevice(c *Config) (this *Device) {
    this = &Device {
        ifce: nil,
        ifmtu: c.Ifmtu,
        verbose: c.Verbose,
        ethmtu: c.Ifmtu+ETHLEN, //6+6+2+2
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
    //TODO
    return nil
}

func (this *Device) GoRecv(doRecv func([]byte)(error)) {
    log.Printf("Info| Device.GoRecv")

    defer this.ifce.Close()
    for {
        data := make([]byte, this.ethmtu)
        n, err := this.ifce.Read(data)
        if err != nil {
            log.Printf("Error|Device.GoRev: %s", err)
            break
        }
        if n > 0 {
            if this.verbose {
                log.Printf("Debug| Device.GoRev: %d: % x...% x\n", n, data[:16], data[n-16:n])
            }
            if err := doRecv(data[:n]); err != nil {
                log.Printf("Error| Device.GoRev.doRecv: %s\n", err)
            }
        }
    }
}

func (this *Device) DoSend(data []byte) error {
    if this.verbose {
        n := len(data)
        log.Printf("Debug| Device.DoSend: %d:% x...% x\n", n, data[:16], data[n-16:])
    }

    if _, err := this.ifce.Write(data); err != nil {
        log.Printf("Error|Device.DoSend: %s", err)  
        return err
    }

    return nil
}