package xclient

import (
	"context"
	. "gorpc"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	mu      sync.Mutex // protect following
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*Client)}
}
func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		// I have no idea how to deal with error, just ignore it.
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}
func (xc *XClient) dial(rpcAddr string) (*Client, error) {

}
func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {

}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
// xc will choose a proper server.
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {

}

// Broadcast invokes the named function for every server registered in discovery
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {

}
