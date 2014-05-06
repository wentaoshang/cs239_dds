package main

import "fmt"
import "strings"
import "strconv"

type Route struct {
	nexthop string
	metric int
}

func createRoute(nexthop string, metric int) *Route {
	var r Route
	r.nexthop = nexthop
	r.metric = metric
	return &r
}

// Implement a minimized RIP routing protocol

func (self *Solver) processFIB(pkt *Packet) {
	addr := pkt.query.args[0]
	nexthop := pkt.query.args[1]
	metric, _ := strconv.Atoi(pkt.query.args[2])
	fmt.Println(self.id + ": receive FIB " + addr + " -> " + nexthop + ", metric=" + strconv.Itoa(metric))
	route, ok := self.fib[addr]
	if ok {
		// Use the new route if metric is smaller or nexthop solver id is smaller
		if route.metric > metric || (route.metric == metric && route.nexthop > nexthop) {
			self.fib[addr] = createRoute(nexthop, metric)
			fmt.Println(self.id + ": update FIB " + addr + " -> " + nexthop + ", metric=" + strconv.Itoa(metric))
		}
	} else {
		// Add the new route
		self.fib[addr] = createRoute(nexthop, metric)
		fmt.Println(self.id + ": insert FIB " + addr + " -> " + nexthop + ", metric=" + strconv.Itoa(metric))
		// Triggered update
		for i, _ := range self.ift {
			if i != nexthop && strings.HasPrefix(i, "s") {
				self.announceFIB(i, addr, metric)
			}
		}
	}
}

func (self *Solver) announceFIB(id string, addr string, metric int) {
	fmt.Println(self.id + ": announce FIB " + addr + " -> " + self.id + ", metric=" + strconv.Itoa(metric) + " to " + id)
	var pkt Packet
	var fib Atom
	fib.name = "fib"
	fib.args = append(fib.args, addr)
	fib.args = append(fib.args, self.id)
	fib.args = append(fib.args, strconv.Itoa(metric))
	pkt.query = &fib
	of := self.ift[id]
	of.out <- &pkt
	// Note: we can safely do this periodic annoucement without 
	//       worrying about deadlock because we use buffered
	//       channel to handle asynchronous send/receive.
	//       The size of the buffer is determined by the size
	//       of the network. For now it is set to 10.
}
