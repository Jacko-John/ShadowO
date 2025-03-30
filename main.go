package main

import (
	"test/client"
	"test/server"
	"time"
)

func main() {
	s := server.NewServer("secret")
	go s.Listen()
	go s.ServeContent()
	time.Sleep(1 * time.Second)
	c := client.NewClient("secret", "wss://localhost:8081/tunnel", "localhost:80")
	c.Run()
	// conn, err := net.Dial("tcp", "localhost:80")
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close()
	// str := "GET / HTTP/2\r\nHost: localhost\r\nUser-Agent: curl/7.88.1\r\nAccept: */*\r\n\r\n"
	// for i := 0; i < 10; i++ {
	// 	time.Sleep(1 * time.Second)
	// 	_, err := conn.Write([]byte(str))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	buf := make([]byte, 2048)
	// 	n, err := conn.Read(buf)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	println(string(buf[:n]))
	// }

}

// The client code is not included in this version.
