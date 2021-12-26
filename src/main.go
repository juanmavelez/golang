package main

import "fmt"

func main(){
	fmt.Printf("Connection open on http://localhost:3000/")
	server := NewServer(":3000")
	server.Handler("/",HandleRoot)
	server.Handler("/api",HandleHome)
	server.Listen()
 }
