package main

import "ShadowO/server"

func main() {
	s := server.NewServer()
	s.Run()
	// client.Run()
	// config1.Init("s.yaml")
	// config.Init("c.yaml")
	// c1 := config1.Get()
	// c2 := config.Get()
	// fmt.Println("c1: ", c1.Secret)
	// fmt.Println("c2: ", c2.Secret)
	// fmt.Println(c1.Secret == c2.Secret)
	// sh := sha256.Sum256([]byte(c1.Secret))
	// fmt.Println(fmt.Sprintf("%x", sh))
	// sh1 := sha256.Sum256([]byte(c2.Secret))
	// fmt.Println(fmt.Sprintf("%x", sh1))
	// fmt.Println(len(sh))
	// fmt.Println(bytes.Equal(sh[:], sh1[:]))
}

// The client code is not included in this version.
