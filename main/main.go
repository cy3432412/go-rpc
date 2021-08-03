package main

import (
	"context"
	"gorpc"
	"log"
	"net"
	"sync"
	"time"
)

//启动服务
func startServer(addr chan string) {
	var foo gorpc.Foo
	//注册foo
	if err := gorpc.Register(&foo); err != nil {
		log.Fatal("register error: ", err)
	}
	//随机选择一个空闲port
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
	log.SetFlags(0)
	addr := make(chan string)
	// 先启动服务端，确保服务端监听正常
	go startServer(addr)

	client, _ := gorpc.Dial("tcp", <-addr)
	defer func() {
		_ = client.Close()
	}()
	time.Sleep(time.Second)

	//序列化并且简单发送接收消息
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &gorpc.Args{
				Num1: i,
				Num2: i * i,
			}

			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo Sum error : ", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)

	}
	wg.Wait()

}
