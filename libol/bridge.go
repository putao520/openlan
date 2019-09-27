package libol 

import (
    "fmt"
    "os"
)

func brSysPath(name string, fun string) string {
    return fmt.Sprintf("/sys/devices/virtual/net/%s/bridge/%s", name, fun)
}

func BrCtlStp(name string, on bool) error {
    f, err := os.OpenFile(brSysPath(name, "stp_state"), os.O_RDWR, 0600)
    defer f.Close()
    if err != nil {
        return err
    }

    if on {
        if _, err := f.Write([]byte("1")); err != nil {
            return err
        }
    } else {
        if _, err := f.Write([]byte("0")); err != nil {
            return err
        }
    }

    return nil
}
