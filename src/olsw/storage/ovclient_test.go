package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestOvClient_ListStatus(t *testing.T) {
	fmt.Println(time.Now().Unix())
	for v := range OvClient.List("yunex") {
		if v == nil {
			break
		}
		fmt.Println(v)
	}
	for v := range OvClient.List("guest") {
		if v == nil {
			break
		}
		fmt.Println(v)
	}
}
