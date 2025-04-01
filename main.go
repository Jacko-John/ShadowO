package main

import "fmt"

func main() {
	// s := server.NewServer("secret")
	// go s.Listen()
	// go s.ServeContent()
	// time.Sleep(1 * time.Second)
	// c := client.NewClient("secret", "wss://localhost:8081/tunnel", "localhost:80")
	// c.Run()

	for range 10 {
		fmt.Println("Hello, world!")
	}
}

// The client code is not included in this version.
