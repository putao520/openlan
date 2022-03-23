package libol

import (
	"bufio"
	"bytes"
	"encoding/base64"
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
	"strconv"
	"strings"
	"syscall"
	"time"
)

const LeaseTime = "2006-01-02T15"

func GenRandom(n int) string {
	letters := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	buffer := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range buffer {
		buffer[i] = letters[rand.Int63()%int64(len(letters))]
	}
	buffer[0] = letters[rand.Int63()%26+10]
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

func GenUint32() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

func GenInt32() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
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
		return NewErr("%s %s", file, err)
	}
	contents, err := LoadWithoutAnn(file)
	if err != nil {
		return NewErr("%s %s", file, err)
	}
	if err := json.Unmarshal(contents, v); err != nil {
		return NewErr("%s", err)
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

func IPNetmask(ipAddr string) (string, error) {
	if i, n, err := net.ParseCIDR(ipAddr); err == nil {
		return i.String() + "/" + net.IP(n.Mask).String(), nil
	} else {
		return "", err
	}
}

func IPNetwork(ipAddr string) (string, error) {
	if _, n, err := net.ParseCIDR(ipAddr); err == nil {
		return n.IP.String() + "/" + net.IP(n.Mask).String(), nil
	} else {
		return ipAddr, err
	}
}

func PrettyTime(t int64) string {
	s := ""
	if t < 0 {
		s = "-"
		t = -t
	}
	min := t / 60
	if min < 60 {
		return fmt.Sprintf("%s%dm%ds", s, min, t%60)
	}
	hours := min / 60
	if hours < 24 {
		return fmt.Sprintf("%s%dh%dm", s, hours, min%60)
	}
	days := hours / 24
	return fmt.Sprintf("%s%dd%dh", s, days, hours%24)
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
	_addr, _ := GetHostPort(addr)
	return _addr
}

func GetHostPort(addr string) (string, string) {
	values := strings.SplitN(addr, ":", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return values[0], ""
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

func OpenTrunk(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
}

func OpenWrite(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

func OpenRead(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, os.ModePerm)
}

func CreateFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}

func ParseNet(addr string) (*net.IPNet, error) {
	if ip, ipNet, err := net.ParseCIDR(addr); err != nil {
		return nil, err
	} else {
		ipNet.IP = ip
		return ipNet, nil
	}
}

func Uint2S(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

func IfName(name string) string {
	size := len(name)
	if size < 15 {
		return name
	}
	return name[size-15 : size]
}

func GetLocalTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

func Base64Decode(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func Base64Encode(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func GetPrefix(value string, index int) string {
	if len(value) >= index {
		return value[:index]
	}
	return ""
}

func GetSuffix(value string, index int) string {
	if len(value) >= index {
		return value[index:]
	}
	return ""
}
