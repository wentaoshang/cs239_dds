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

type Channel struct {
	in chan *Packet
	out chan *Packet
}

func (self *Link) connect(s1 *Solver, s2 *Solver) {
	var s1_c Channel
	s1_c.in = self.c1
	s1_c.out = self.c2
	s1.link[s2.id] = &s1_c

	var s2_c Channel
	s2_c.in = self.c2
	s2_c.out = self.c1
	s2.link[s1.id] = &s2_c
}
