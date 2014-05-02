package main

// Abstraction for data source

type Source struct {
	id string
	name string
	data string
	in chan *Packet
	out chan *Packet
}

func createSource(id string, name string, data string) *Source {
	var s Source
	s.id = id
	s.name = name
	s.data = data
	s.in = make(chan *Packet)
	s.out = make(chan *Packet)
	return &s
}

func (self *Source) run() {
	pkt := <-self.in
	if pkt.name == self.name {
		pkt.data = self.data
	} else {
		pkt.data = "NULL"
	}
	self.out <- pkt
}
