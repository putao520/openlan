package models

type TapDevice struct {
	isTAP bool
	name  string
}

func (t *TapDevice) IsTUN() bool {
	return !t.isTAP
}

func (t *TapDevice) IsTAP() bool {
	return t.isTAP
}

func (t *TapDevice) Name() string {
	return t.name
}

func (t *TapDevice) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (t *TapDevice) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (t *TapDevice) Close() error {
	return nil
}
