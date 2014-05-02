package main

// Abstraction for packet format

type Packet struct {
	query *Atom
	result string
}

func (self *Packet) toString() string {
	var s string
	s = self.query.toString() + " --> " + self.result
	return s
}
