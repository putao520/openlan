package libol

type KcpServer struct {
	socketServer
}

func NewKcpServer() *KcpServer {
	return &KcpServer{}
}
func (k *KcpServer) Listen() (err error) {
	return nil
}

func (k *KcpServer) Close() {

}

func (k *KcpServer) Accept() {

}

func (k *KcpServer) OffClient(client SocketClient) {

}

func (k *KcpServer) Loop(call ServerListener) {

}

func (k *KcpServer) Read(client SocketClient, ReadAt ReadClient) {

}

// Client Implement

type KcpClient struct {
	socketClient
}

func NewKcpClient() *KcpClient {
	return &KcpClient{}
}

func (c *KcpClient) LocalAddr() string {
	return ""
}

func (c *KcpClient) Connect() (err error) {
	return nil
}

func (c *KcpClient) Close() {

}

func (c *KcpClient) WriteMsg(data []byte) error {
	return nil
}

func (c *KcpClient) ReadMsg(data []byte) (int, error) {
	return 0, nil
}

func (c *KcpClient) WriteReq(action string, body string) error {
	return nil
}

func (c *KcpClient) WriteResp(action string, body string) error {
	return nil
}

func (c *KcpClient) Terminal() {
	return
}

func (c *KcpClient) SetStatus(v uint8) {
	return
}

func (c *KcpClient) IsOk() bool {
	return false
}
