package gorpc

import (
	"errors"
	"gorpc/codec"
	"io"
	"net"
	"sync"
)

// Call 活跃的rpc
type Call struct {
	Seq           uint64      // 序列号
	ServiceMethod string      // 格式 "<service>.<method>"
	Args          interface{} // 参数列表
	Reply         interface{} // 方法的返回信息
	Error         error       // error信息处理
	Done          chan *Call  // call结束执行done
}

func (call *Call) done() {
	call.Done <- call
}

// Client RPC的客户端结构体
// 一个客户端会有多个call
type Client struct {
	cc       codec.Codec      //编码格式
	opt      *Option          //header头选项
	sending  sync.Mutex       //发送时的锁
	header   codec.Header     //header头
	mu       sync.Mutex       //锁
	seq      uint64           //序列号
	pending  map[uint64]*Call //可能有多个call
	closing  bool             //client 关闭状态
	shutdown bool             //client 停止状态
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

//关闭服务端链接
func (client *Client) Close() error {

}

//判断服务端是否可以得到
func (client *Client) IsAvailable() bool {

}

//
func (client *Client) registerCall(call *Call) (uint64, error) {

}

func (client *Client) removeCall(seq uint64) *Call {

}

func (client *Client) terminateCall(err error) {

}

//
func (client *Client) recevice() {

}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {

}

func newClientCodec(cc codec.Codec, opt *Option) *Client {

}

func parseOptions(opts ...*Option) (*Option, error) {

}

func Dial(network, address string, opts ...*Option) (client *Client, err error) {

}

func (client *Client) send(call *Call) {

}

func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {

}

func (client *Client) Call(serviceMethod string, args, reply interface{}) *Call {

}
