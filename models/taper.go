package models

type Taper interface {
	IsTUN() bool
	IsTAP() bool
	Name() string
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}