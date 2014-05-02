package main

import "fmt"

// Abstraction for user request

type Request struct {
	id string
	in chan *Packet
	out chan *Packet
	done chan int
}

func createRequest(id string) *Request {
	var r Request
	r.id = id
	r.in = make(chan *Packet)
	r.out = make(chan *Packet)
	r.done = make(chan int)
	return &r
}

func (self *Request) run() {
	var pkt Packet
	pkt.name = "test"
	self.out <- &pkt
	ans := <-self.in
	fmt.Println("Answer: name=" + ans.name + ", data=" + ans.data)
	self.done <- 0
}
