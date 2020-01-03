package main

import (
	"fmt"
	"github.com/golang-collections/go-datastructures/queue"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	rb := queue.NewRingBuffer(3)
	for i := 0; i < 4; i++ {
		rb.Put(i)
	}

	wg.Add(1)
	go func () {
		for {
			if rb.Len() == 0 {
				break
			}
			if d, err := rb.Get(); err == nil {
				fmt.Println(d)
			}
		}
		wg.Done()
	}()
	wg.Wait()

	q := queue.New(99)
	for i := 0; i < 4; i++ {
		q.Put(i)
	}

	wg.Add(1)
	go func () {
		for {
			if q.Len() == 0 {
				break
			}
			if d, err := q.Get(1); err == nil {
				fmt.Println(d)
			}
		}
		wg.Done()
	}()
	wg.Wait()
}
