package p2p

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, ":3000")

	assert.Nil(t, tr.ListenAndAccept())
	t.Cleanup(func() { tr.Close() })
}

func TestTCPTransport_Dial(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Nil(t, tr.ListenAndAccept())
	t.Cleanup(func() { tr.Close() })

	assert.Nil(t, tr.Dial(":3000"))
}

func TestTCPTransport_Close(t *testing.T)	{
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Nil(t, tr.ListenAndAccept())
	t.Cleanup(func() { tr.Close() })

	assert.Nil(t, tr.Close())
}

func TestTCPTransport_Consume(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Nil(t, tr.ListenAndAccept())
	t.Cleanup(func() { tr.Close() })

	ch := tr.Consume()
	assert.NotNil(t, ch)
}

func Test_TCPTransport_handleConn(t *testing.T){
	opts:= TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Nil(t, tr.ListenAndAccept())
	t.Cleanup(func() { tr.Close() })

	conn, err := net.Dial("tcp", ":3000")
	assert.Nil(t, err)
	conn.Write([]byte{0x1, 'h', 'e', 'l', 'l', 'o'})
	go tr.handleConn(conn, true)

	select {
		case rpc := <-tr.Consume():
		assert.Equal(t, []byte("hello"), rpc.Payload)
		case <-time.After(time.Second * 2):
		t.Fail()
	}		
 
}