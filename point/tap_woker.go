package point

import (
    "github.com/songgao/water"
    "github.com/lightstar-dev/openlan-go/libol"
)

type TapWroker struct {
    ifce *water.Interface
    writechan chan []byte
    ifmtu int
    verbose int
}

func NewTapWoker(ifce *water.Interface, c *Config) (this *TapWroker) {
    this = &TapWroker {
        ifce: ifce,
        writechan: make(chan []byte, 1024*10),
        ifmtu: c.Ifmtu, //1514
        verbose: c.Verbose,
    }
    return
}

func (this *TapWroker) GoRecv(dorecv func([]byte)(error)) {
    defer this.ifce.Close()
    for {
        data := make([]byte, this.ifmtu)
        n, err := this.ifce.Read(data)
        if err != nil {
            libol.Error("TapWroker.GoRev: %s", err)
            break
        }
        if this.IsVerbose() {
            libol.Debug("TapWroker.GoRev: % x\n", data[:n])
        }

        dorecv(data[:n])
    }
}

func (this *TapWroker) DoSend(data []byte) error {
    if this.IsVerbose() {
        libol.Debug("TapWroker.DoSend: % x\n", data)
    }

    this.writechan <- data
    return nil
}

func (this *TapWroker) GoLoop() error {
    defer this.ifce.Close()
    for {
        select {
        case wdata := <- this.writechan:
            if _, err := this.ifce.Write(wdata); err != nil {
                libol.Error("TapWroker.GoLoop: %s", err)   
            }
        }
    }
}

func (this *TapWroker) IsVerbose() bool {
    return this.verbose != 0
}

func (this *TapWroker) Close() {
    this.ifce.Close()
}