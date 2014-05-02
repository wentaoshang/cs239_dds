package main

import "fmt"

// Abstraction for distributed Datalog Solver

type Solver struct {
	id string
	rule map[string]([]string)
	fib map[string]string
	prt map[string]string
	link map[string](*Channel)
}

func createSolver(id string) *Solver {
	var s Solver
	s.id = id
	s.fib = make(map[string]string)
	s.prt = make(map[string]string)
	s.link = make(map[string](*Channel))
	return &s
}

func (self *Solver) addSource(src *Source) {
	var c Channel
	c.in = src.out
	c.out = src.in
	self.link[src.id] = &c
	self.fib[src.name] = src.id
}

func (self *Solver) addRule(s string, r []string) {
	self.rule[s] = r
}

func (self *Solver) addForwardingEntry(data string, nexthop *Solver) {
	if _, ok := self.link[nexthop.id]; !ok {
		// Don't add this fib entry if the nexthop is not in link table
		return
	}
	self.fib[data] = nexthop.id
}

func (self *Solver) addRequest(req *Request) {
	var c Channel
	c.in = req.out
	c.out = req.in
	self.link[req.id] = &c
}

func (self *Solver) run() {
	for id, c := range self.link {
		go func(id string, c *Channel) {
			pkt := <-c.in
			fmt.Println(self.id + ": name=" + pkt.name + ", data=" + pkt.data + " from " + id)

			if pkt.data == "" { // This is request packet
				//TODO: Lookup rules
				
				// Lookup fib
				nexthop, ok := self.fib[pkt.name]
				if ok {
					fmt.Println(self.id + ": nexthop is " + nexthop)
					oc := self.link[nexthop]
					oc.out <- pkt
					
					// Record in prt
					self.prt[pkt.name] = id
				}
			} else { // This is response packet
				// Lookup prt
				nexthop, ok := self.prt[pkt.name]
				if ok {
					fmt.Println(self.id + ": consume request from " + nexthop)
					oc := self.link[nexthop]
					oc.out <- pkt

					// Clear prt
					delete(self.prt, pkt.name)
				}
			}
		}(id, c)
	}
}
