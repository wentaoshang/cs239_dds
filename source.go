package main

import "fmt"

// Abstraction for data source

type Source struct {
	id string
	cname string
	data [](*Rule)
	in chan *Packet
	out chan *Packet
}

func createSource(id string, data [](*Rule)) *Source {
	var s Source
	s.id = id
	s.data = data
	s.cname = data[0].head.getName()
	s.in = make(chan *Packet)
	s.out = make(chan *Packet)
	return &s
}

func (self *Source) run() {
	pkt := <-self.in
	if pkt.query.getName() != self.cname {
		return
	}

	fmt.Println(self.id + ": search for ?" + pkt.query.toString())

	pkt.result = nil
	for _, d := range self.data {
		res, ok := d.unify(pkt.query)
		if ok {
			pkt.result = res
			break;
		}
	}
	fmt.Println(self.id + ": " + pkt.resultToString())
	self.out <- pkt
}
