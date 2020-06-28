package libol

import (
	"fmt"
	"os"
	"strconv"
)

type BrCtl struct {
	Name string
	Path string
}

func NewBrCtl(name string) (b *BrCtl) {
	b = &BrCtl{
		Name: name,
	}
	return
}

func (b *BrCtl) SysPath(fun string) string {
	if b.Path == "" {
		b.Path = fmt.Sprintf("/sys/devices/virtual/net/%s/bridge", b.Name)
	}
	return fmt.Sprintf("%s/%s", b.Path, fun)
}

func (b *BrCtl) Stp(on bool) error {
	file := b.SysPath("stp_state")
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

func (b *BrCtl) Delay(delay int) error { // by second
	file := b.SysPath("forward_delay")
	fp, err := os.OpenFile(file, os.O_RDWR, 0600)
	defer fp.Close()
	if err != nil {
		return err
	}
	if _, err := fp.Write([]byte(strconv.Itoa(delay * 100))); err != nil {
		return err
	}
	return nil
}
