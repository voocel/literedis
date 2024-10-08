package tcp

import (
	"fmt"
	"literedis/pkg/network"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient("127.0.0.1:8899")

	c.OnConnect(func(conn network.Conn) {
		fmt.Println("connect")
	})

	c.OnReceive(func(conn network.Conn, msg []byte) {
		fmt.Println("msg: ", string(msg))
	})

	c.OnDisconnect(func(conn network.Conn, err error) {
		fmt.Println("disconnect: ", err)
	})

	conn, err := c.Dial()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		b := []byte("test\n")
		if err = conn.Send(b); err != nil {
			fmt.Println("send err:", err)
		}
	}
}
