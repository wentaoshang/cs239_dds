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

type Interface struct {
	in chan *Packet
	out chan *Packet
}

func (self *Link) connect(s1 *Solver, s2 *Solver) {
	var s1_if Interface
	s1_if.in = self.c1
	s1_if.out = self.c2
	s1.ift[s2.id] = &s1_if

	var s2_if Interface
	s2_if.in = self.c2
	s2_if.out = self.c1
	s2.ift[s1.id] = &s2_if
}
