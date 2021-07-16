package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

//定义GobCodec结构体
type GobCodec struct {
	conn io.ReadWriteCloser //链接
	buf  *bufio.Writer      //buffer
	dec  *gob.Decoder       //解码
	enc  *gob.Encoder       //解码
}

var _ Codec = (*GobCodec)(nil)

//实例化gob
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

//读取头
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

//读取结构体
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

//写方法
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	//写结束关闭连接
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.conn.Close()
		}
	}()
	//编码
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

//关闭连接
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
