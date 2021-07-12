package main

import (
	"encoding/json"
	"fmt"
	"gorpc"
	"gorpc/codec"
	"log"
	"net"
	"time"
)

//启动服务
func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	//错误处理
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("Start gorpc server on", l.Addr())
	addr <- l.Addr().String()
	gorpc.Accept(l)
}

//实现了一个极简client用于测试服务端
func main() {
	addr := make(chan string)
	// 先启动服务端，确保服务端监听正常
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer func() {
		_ = conn.Close()
	}()
	time.Sleep(time.Second)

	//序列化并且简单发送接收消息
	_ = json.NewEncoder(conn).Encode(gorpc.DefaultOption)
	cc := codec.NewGobCodec(conn)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("gorpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadHeader(h)
		log.Println("reply: ", reply)
	}

}
