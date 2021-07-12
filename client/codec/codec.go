package codec

import "io"

//客户端发来的请求头
type Header struct {
	ServiceMethod string // 服务名.方法名(Service.Method)
	Seq           uint64 // 请求序号
	Error         string // 错误信息
}

//定义一个接口，进行消息的编解码
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}