package rotaryphone

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type Binder interface {
	// For net.Listener usage
	Accept() (net.Conn, error)
	Addr() net.Addr
	Close() error
	// For http.Client
	Client() *http.Client
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type binder struct {
	connLock  sync.Mutex
	connIndex map[net.Conn]bool
	accept    chan net.Conn

	closed bool
}

func New() Binder {
	return &binder{
		connIndex: map[net.Conn]bool{},
		accept:    make(chan net.Conn),
	}
}

func (b *binder) Accept() (net.Conn, error) {
	conn := <-b.accept
	if conn == nil {
		return nil, net.ErrClosed
	}

	return conn, nil
}

func (b *binder) Close() error {
	b.connLock.Lock()
	defer b.connLock.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	close(b.accept)

	for conn := range b.connIndex {
		conn.Close()
	}

	return nil
}

func (b *binder) Client() *http.Client {
	trans := http.DefaultTransport.(*http.Transport).Clone()
	trans.DialContext = b.DialContext
	return &http.Client{
		Transport: trans,
	}
}

func (b *binder) Dial(network, address string) (net.Conn, error) {
	return b.DialContext(context.Background(), network, address)
}

func (b *binder) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	client, server := net.Pipe()

	b.connLock.Lock()
	defer b.connLock.Unlock()

	if b.closed {
		return nil, net.ErrClosed
	}

	b.accept <- b.bind(server)

	return b.bind(client), nil
}

func (b *binder) Network() string { return "pipe" }
func (n *binder) String() string  { return "rotaryphone" }
func (b *binder) Addr() net.Addr  { return b }

func (b *binder) bind(conn net.Conn) net.Conn {
	return &boundClose{Conn: conn, close: func() {
		b.connLock.Lock()
		defer b.connLock.Unlock()

		delete(b.connIndex, conn)
	}}
}

type boundClose struct {
	net.Conn
	close func()
}

func (bc *boundClose) Close() error {
	bc.close()
	return bc.Conn.Close()
}
