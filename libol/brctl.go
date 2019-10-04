package libol

import (
	"fmt"
	"os"
)

type BrCtl struct {
	Name string
	Path string
}

func NewBrCtl(name string) (this *BrCtl) {
	this = &BrCtl{
		Name: name,
	}
	return
}

func (this *BrCtl) SysPath(fun string) string {
	if this.Path == "" {
		this.Path = fmt.Sprintf("/sys/devices/virtual/net/%s/bridge", this.Name)
	}

	return fmt.Sprintf("%s/%s", this.Path, fun)
}

func (this *BrCtl) Stp(on bool) error {
	file := this.SysPath("stp_state")
	fp, err := os.OpenFile(file, os.O_RDWR, 0600)
	defer fp.Close()
	if err != nil {
		return err
	}

	if on {
		if _, err := fp.Write([]byte("1")); err != nil {
			return err
		}
	} else {
		if _, err := fp.Write([]byte("0")); err != nil {
			return err
		}
	}
	return nil
}
