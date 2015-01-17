package main

import "log"

func main() {
	log.Println("The server is running...")
	c := make(chan int)
	<-c
}
