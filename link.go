package main

// Abstraction for network link

type Link struct {
	id string
	c1 chan *Packet
	c2 chan *Packet
}

func createLink(id string) *Link {
	var l Link
	l.id = id
	l.c1 = make(chan *Packet)
	l.c2 = make(chan *Packet)
	return &l
}

func (self *Link) connect(s1 *Solver, s2 *Solver) {
	s1.addInterface(s2.id, self.c1, self.c2)
	s2.addInterface(s1.id, self.c2, self.c1)
}
