package main

import "fmt"

// Abstraction for distributed Datalog Solver

type Solver struct {
	id string
	rule map[string]([]string)
	fib map[string]string
	prt map[string]string
	ift map[string](*Interface)
}

func createSolver(id string) *Solver {
	var s Solver
	s.id = id
	s.fib = make(map[string]string)
	s.prt = make(map[string]string)
	s.ift = make(map[string](*Interface))
	return &s
}

type Interface struct {
	in chan *Packet
	out chan *Packet
}

func (self *Solver) addInterface(id string, in chan *Packet, out chan *Packet) {
	var f Interface
	f.in = in
	f.out = out
	self.ift[id] = &f
}

func (self *Solver) addSource(src *Source) {
	self.addInterface(src.id, src.out, src.in)
	self.fib[src.name] = src.id
}

func (self *Solver) addRule(s string, r []string) {
	self.rule[s] = r
}

func (self *Solver) addForwardingEntry(data string, nexthop *Solver) {
	if _, ok := self.ift[nexthop.id]; !ok {
		// Don't add this fib entry if the nexthop is not in link table
		return
	}
	self.fib[data] = nexthop.id
}

func (self *Solver) addRequest(req *Request) {
	self.addInterface(req.id, req.out, req.in)
}

func (self *Solver) run() {
	for id, f := range self.ift {
		go func(id string, f *Interface) {
			pkt := <-f.in
			hint := pkt.query.name

			if pkt.result == "" { // This is request packet
				fmt.Println(self.id + ": ?" + pkt.query.toString() + " from " + id)
				//TODO: Lookup rules
				
				// Lookup fib
				nexthop, ok := self.fib[hint]
				if ok {
					fmt.Println(self.id + ": nexthop is " + nexthop)
					of := self.ift[nexthop]
					of.out <- pkt
					
					// Record in prt
					self.prt[hint] = id
				}
			} else { // This is response packet
				fmt.Println(self.id + ": " + pkt.toString() + " from " + id)
				// Lookup prt
				nexthop, ok := self.prt[hint]
				if ok {
					fmt.Println(self.id + ": consume request from " + nexthop)
					of := self.ift[nexthop]
					of.out <- pkt

					// Clear prt
					delete(self.prt, hint)
				}
			}
		}(id, f)
	}
}
