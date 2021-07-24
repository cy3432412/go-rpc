package gorpc

import (
	"debug/dwarf"
	"errors"
	"fmt"
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
	sending  sync.Mutex       //发送时的锁，防止出现多个请求报文混淆
	header   codec.Header     //header头
	mu       sync.Mutex       //锁
	seq      uint64           //发送的序列号
	pending  map[uint64]*Call //可能有多个call需要处理
	closing  bool             //client 关闭状态
	shutdown bool             //client 停止状态
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

//Close 关闭服务端链接
func (client *Client) Close() error {
	//加锁
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

//IsAvailable 判断服务端是否可以得到
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Lock()
	//没有被关闭也不是正在关闭
	return !client.shutdown && !client.closing
}

//registerCall 注册call
//return call编号，错误
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	// 异常处理,同上
	if client.shutdown || client.closing {
		return 0, ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil

}

//removeCall 根据seq删除call
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()

	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

//terminateCall 错误调用时，将错误信息通知给所有call
func (client *Client) terminateCall(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

//receive 接收call，并响应
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		//响应有三种情况
		//1. call不存在
		//2. call存在，但是server端处理出错
		//3. call存在，从body读取reply
		call := client.removeCall(h.Seq)
		if call == nil {
			err = client.cc.ReadBody(nil)
		} else if h.Error != "" {
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		} else {
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body" + err.Error())
			}
			call.done()
		}
	}
	//发生错误，中止pending队列中的call
	client.terminateCall(err)
}

//NewClient 新建client实例
func NewClient(conn net.Conn, opt *Option) (*Client, error) {

}

//newClientCodec 新建消息的编解码格式，本项目中使用gob
func newClientCodec(cc codec.Codec, opt *Option) *Client {

}

//parseOptions 解析options
func parseOptions(opts ...*Option) (*Option, error) {

}

//Dial dial方法连接RPC的服务端
func Dial(network, address string, opts ...*Option) (client *Client, err error) {

}

//send 客户端发送请求
func (client *Client) send(call *Call) {

}

//Go 服务调用接口 返回一个call实例
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {

}

//Call 对go的封装
func (client *Client) Call(serviceMethod string, args, reply interface{}) *Call {

}
