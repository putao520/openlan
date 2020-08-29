package libol

import "time"

type Promise struct {
	Count  int
	MaxTry int
	First  time.Duration // the delay time.
	MinInt time.Duration // the normal time.
	MaxInt time.Duration // the max delay time.
}

func (p *Promise) Done(call func() error) {
	for {
		p.Count++
		if p.MaxTry > 0 && p.Count > p.MaxTry {
			return
		}
		if err := call(); err == nil {
			return
		}
		time.Sleep(p.First)
		if p.First < p.MaxInt {
			p.First += p.MinInt
		}
	}
}
