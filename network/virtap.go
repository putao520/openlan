package network

type VirTap struct {
	isTap bool
	name  string
}

func NewVirTap(isTap bool, name string) (*VirTap, error) {
	tap := &VirTap{
		isTap: isTap,
		name:  name,
	}

	return tap, nil
}

func (t *VirTap) IsTUN() bool {
	return !t.isTap
}

func (t *VirTap) IsTAP() bool {
	return t.isTap
}

func (t *VirTap) Name() string {
	return t.name
}

func (t *VirTap) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (t *VirTap) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (t *VirTap) Close() error {
	return nil
}
