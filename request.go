package main

import "fmt"

// Abstraction for user request

type Request struct {
	id string
	query *Atom
	in chan *Packet
	out chan *Packet
	done chan int
}

func createRequest(id string, query string) *Request {
	var r Request
	r.id = id
	r.query = createAtom(query)
	r.in = make(chan *Packet, 2)
	r.out = make(chan *Packet, 2)
	r.done = make(chan int)
	return &r
}

func (self *Request) run() {
	var pkt Packet
	pkt.query = self.query
	fmt.Println(self.id + ": ?" + pkt.queryToString())
	self.out <- &pkt
	ans := <-self.in
	fmt.Println(self.id + ": " + ans.resultToString())
	self.done <- 0
}
