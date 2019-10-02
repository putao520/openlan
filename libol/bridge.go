package libol 

import (
    "fmt"
    "os"
)

func brSysPath(name string, fun string) string {
    file := fmt.Sprintf("/sys/devices/virtual/net/%s/bridge/%s", name, fun)
    return file
}

func BrCtlStp(name string, on bool) error {
    file := brSysPath(name, "stp_state")
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
