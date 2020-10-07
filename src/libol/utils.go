package libol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func GenToken(n int) string {
	letters := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	buffer := make([]byte, n)

	size := len(letters)
	rand.Seed(time.Now().UnixNano())
	for i := range buffer {
		buffer[i] = letters[rand.Int63()%int64(size)]
	}
	return string(buffer)
}

func GenEthAddr(n int) []byte {
	if n == 0 {
		n = 6
	}
	data := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range data {
		data[i] = byte(rand.Uint32() & 0xFF)
	}
	data[0] &= 0xfe
	return data
}

func Marshal(v interface{}, pretty bool) ([]byte, error) {
	str, err := json.Marshal(v)
	if err != nil {
		Error("Marshal error: %s", err)
		return nil, err
	}
	if !pretty {
		return str, nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, str, "", "  "); err != nil {
		return str, nil
	}
	return out.Bytes(), nil
}

func MarshalSave(v interface{}, file string, pretty bool) error {
	f, err := CreateFile(file)
	if err != nil {
		Error("MarshalSave: %s", err)
		return err
	}
	defer f.Close()
	str, err := Marshal(v, true)
	if err != nil {
		Error("MarshalSave error: %s", err)
		return err
	}
	if _, err := f.Write(str); err != nil {
		Error("MarshalSave: %s", err)
		return err
	}
	return nil
}

func FileExist(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}
	return nil
}

func ScanAnn(r io.Reader) ([]byte, error) {
	data := make([]byte, 0, 1024)
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		bs := scan.Bytes()
		dis := false
		for i, b := range bs {
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
				continue
			}
			if b == '/' && len(bs) > i+1 && bs[i+1] == '/' {
				dis = true // if start with //, need discard it.
			}
			break
		}
		if !dis {
			data = append(data, bs...)
		}
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

func LoadWithoutAnn(file string) ([]byte, error) {
	fp, err := OpenRead(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return ScanAnn(fp)
}

func UnmarshalLoad(v interface{}, file string) error {
	if err := FileExist(file); err != nil {
		return NewErr("UnmarshalLoad: %s %s", file, err)
	}
	contents, err := LoadWithoutAnn(file)
	if err != nil {
		return NewErr("UnmarshalLoad: %s %s", file, err)
	}
	if err := json.Unmarshal(contents, v); err != nil {
		return NewErr("UnmarshalLoad: %s", err)
	}
	return nil
}

func FunName(i interface{}) string {
	ptr := reflect.ValueOf(i).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	return path.Base(name)
}

func Netmask2Len(s string) int {
	mask := net.IPMask(net.ParseIP(s).To4())
	prefixSize, _ := mask.Size()
	return prefixSize
}

func IpAddrFormat(ipAddr string) string {
	if ipAddr == "" {
		return ""
	}
	address := ipAddr
	netmask := "255.255.255.255"
	s := strings.SplitN(ipAddr, "/", 2)
	if len(s) == 2 {
		address = s[0]
		_, n, err := net.ParseCIDR(ipAddr)
		if err == nil {
			netmask = net.IP(n.Mask).String()
		} else {
			netmask = s[1]
		}
	}
	return address + "/" + netmask
}

func PrettyTime(t int64) string {
	min := t / 60
	if min < 60 {
		return fmt.Sprintf("%dm%ds", min, t%60)
	}
	hours := min / 60
	if hours < 24 {
		return fmt.Sprintf("%dh%dm", hours, min%60)
	}
	days := hours / 24
	return fmt.Sprintf("%dd%dh", days, hours%24)
}

func PrettyBytes(b int64) string {
	split := func(_v int64, _m int64) (i int64, d int) {
		_d := float64(_v%_m) / float64(_m)
		return _v / _m, int(_d * 100) //move two decimal to integer
	}
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	k, d := split(b, 1024)
	if k < 1024 {
		return fmt.Sprintf("%d.%02dK", k, d)
	}
	m, d := split(k, 1024)
	if m < 1024 {
		return fmt.Sprintf("%d.%02dM", m, d)
	}
	g, d := split(m, 1024)
	return fmt.Sprintf("%d.%02dG", g, d)
}

func GetIPAddr(addr string) string {
	return strings.Split(addr, ":")[0]
}

func Wait() {
	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGKILL)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	Info("Wait: ...")
	n := <-x
	Warn("Wait: ... Signal %d received ...", n)
}

func OpenWrite(file string) (*os.File, error) {
	return os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

func OpenRead(file string) (*os.File, error) {
	return os.OpenFile(file, os.O_RDONLY, os.ModePerm)
}

func CreateFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}
