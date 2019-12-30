package vswitch

import (
	"github.com/songgao/water"
	"sync"
	"testing"
)

func TestBridgeWriteAndReadByTap(t *testing.T) {
	var wg sync.WaitGroup

	//open bridge.
	br := NewBridger("br-test", 1500)
	br.Open("")

	//open tap device
	cfg := water.Config{DeviceType: water.TAP}
	dev01, err := water.New(cfg)
	if err != nil {
		t.Errorf("Tap.Open %s", err)
		return
	}

	dev02, err := water.New(cfg)
	if err != nil {
		t.Errorf("Tap.Open %s", err)
		return
	}

	br.AddSlave(dev01.Name())
	br.AddSlave(dev02.Name())

	wg.Add(1)
	go func() {
		t.Logf("Tap.write: %s\n", dev01.Name())

		frame := make([]byte, 65)
		for i := 0; i < 64; i++ {
			frame[i] = uint8(i)
		}
		t.Logf("Tap.write: %x", frame)
		n, err := dev01.Write(frame)
		if err != nil {
			t.Errorf("Tap.write: %s", err)
		}
		if n != len(frame) {
			t.Errorf("Tap.write: %d", n)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		frame := make([]byte, 65)
		t.Logf("Tap.read: %s\n", dev02.Name())

		n, err := dev02.Read(frame)
		if err != nil {
			t.Errorf("Tap.read: %s", err)
		}
		if n != len(frame) {
			t.Errorf("Tap.read: %d", n)
		}
		wg.Done()
	}()

	wg.Wait()
}