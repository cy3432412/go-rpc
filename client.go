package gorpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorpc/codec"
	"io"
	"log"
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
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error: ", err)
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil

}

//newClientCodec 新建消息的编解码格式，本项目中使用gob
func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1,
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

//parseOptions 解析options
func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of  options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

//Dial dial方法连接RPC的服务端
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

//send 客户端发送请求
func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	//定义请求头
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	//发送请求
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)

		if call != nil {
			call.Error = err
			call.done()
		}

	}

}

//Go 服务调用接口 返回一个call实例
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client : done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

//Call 对go的封装
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
