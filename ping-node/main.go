package main

import "log"

func main() {
	log.Println("The server is up...")

	c := make(chan int)
	<-c
}
