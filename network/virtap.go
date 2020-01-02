package network

type VirTap struct {
	isTAP bool
	name  string
}

func (t *VirTap) IsTUN() bool {
	return !t.isTAP
}

func (t *VirTap) IsTAP() bool {
	return t.isTAP
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
