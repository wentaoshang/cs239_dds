package main

import "fmt"
import "time"
import "strconv"
import "strings"

// Abstraction for distributed Datalog Solver

type Interface struct {
	in chan *Packet
	out chan *Packet
}

type PendingRequest struct {
	query *Atom
	solve *Atom
	depend map[string]int  // Dependency list of rules
	results map[string]string  // Partial unification result
	origin string  // For now, assume only one requester for each query
}

type Solver struct {
	id string
	rc [](*Rule)  // Rule cache
	fib map[string](*Route)  // FIB
	prt map[string](*PendingRequest)  // Pending request table
	ift map[string](*Interface)  // Interface table
}

func createSolver(id string) *Solver {
	var s Solver
	s.id = id
	s.fib = make(map[string](*Route))
	s.prt = make(map[string](*PendingRequest))
	s.ift = make(map[string](*Interface))
	return &s
}

func (self *Solver) addInterface(id string, in chan *Packet, out chan *Packet) {
	var f Interface
	f.in = in
	f.out = out
	self.ift[id] = &f
}

func (self *Solver) addSource(src *Source) {
	self.addInterface(src.id, src.out, src.in)
	self.fib[src.cname] = createRoute(src.id, 0)  // Attached source has a routing metric of 0
	fmt.Println(self.id + ": add data source " + src.cname)
}

func (self *Solver) addRule(r string) {
	rule := createRule(r)
	self.rc = append(self.rc, rule)
	fmt.Println(self.id + ": add rule " + rule.toString())
}

func (self *Solver) addForwardingEntry(cname string, nexthop *Solver, metric int) {
	if _, ok := self.ift[nexthop.id]; !ok {
		// Don't add this fib entry if the nexthop is not in link table
		return
	}
	self.fib[cname] = createRoute(nexthop.id, metric)
	fmt.Println(self.id + ": add forwarding entry " + cname + " -> " + nexthop.id + ", metric=" + strconv.Itoa(metric))
}

func (self *Solver) addRequest(req *Request) {
	self.addInterface(req.id, req.out, req.in)
}

func solveAndForward(s *Solver, origin string, query *Atom, solve *Atom) {
	// Lookup rule cache
	foundRule := false
	for _, rule := range s.rc {
		res, match := rule.unify(query)
		if match {
			foundRule = true
			fmt.Println(s.id + ": ?" + query.toString() + " matches rule " + rule.toString() + " with " + mapToString(res))
			// Record query in prt with dependencies
			var pr PendingRequest
			pr.query = query
			pr.solve = solve
			pr.depend = make(map[string]int)
			for i, a := range rule.body {
				pr.depend[a.toString()] = i
			}
			pr.results = res
			pr.origin = origin
			s.prt[query.toString()] = &pr

			for _, a := range rule.body {
				solveAndForward(s, "", a, query)
			}
		}
	}

	if foundRule {
		return
	}

	// No match in rule table
	fmt.Println(s.id + ": no rule for ?" + query.toString())

	// Record query in prt, without dependencies
	var pr PendingRequest
	pr.query = query
	pr.solve = solve
	pr.origin = origin
	s.prt[query.toString()] = &pr

	// Lookup fib and forward
	route, ok := s.fib[query.getName()]
	if ok {
		nexthop := route.nexthop
		fmt.Println(s.id + ": nexthop for ?" + query.toString() + " is " + nexthop)
		of := s.ift[nexthop]
		var pkt Packet
		pkt.query = query
		of.out <- &pkt
	}	
}

// Consume 'query' using 'result' of 'from' 
func consumePendingRequest(s *Solver, query *Atom, result map[string]string, from *Atom) {
	// Lookup prt
	pr, ok := s.prt[query.toString()]
	if ok {
		if query.toString() == from.toString() {
			// Consume the query using its own result
			fmt.Println(s.id + ": consume ?" + query.toString() + " with " + mapToString(result))

			// Delete query from prt
			delete(s.prt, query.toString())

			if pr.solve != nil {
				// This query is a dependency of another query
				fmt.Println(s.id + ": ?" + query.toString() + " satisfies a dependency of ?" + pr.solve.toString())
				// Pass the result up
				consumePendingRequest(s, pr.solve, result, query)
			} else {
				// This query doesn't solve other queries
				of := s.ift[pr.origin]  // ASSERT: pr.origin should not be empty if pr.solve is nil
				// Send result back
				var pkt Packet
				pkt.query = query
				pkt.result = result
				fmt.Println(s.id + ": send " + pkt.toString() + " back to " + pr.origin)
				of.out <- &pkt
			}
		} else {
			// A dependency is resolved
			// Copy the result and clear the dependency
			if pr.results == nil {
				pr.results = result
			} else {
				for key, val := range pr.results {
					if val2, ok := result[val]; ok {
						pr.results[key] = val2
					}
				}
			}
			delete(pr.depend, from.toString())

			if len(pr.depend) == 0 {
				fmt.Println(s.id + ": all dependencies resolved for ?" + query.toString())
				// Consume itself
				consumePendingRequest(s, pr.query, pr.results, pr.query)
			}
		}
	}
}

func (self *Solver) waitForSolver(id string, f *Interface) {
	for {  // Infinite loop
		//fmt.Println(self.id + ": loop on interface to " + id)
		select {
		case pkt := <-f.in:
			if pkt.query.name == "fib" {
				// This is a FIB announcement
				self.processFIB(pkt, id)
			} else if pkt.result == nil { // This is a request packet
				fmt.Println(self.id + ": ?" + pkt.query.toString() + " from " + id)

				// Recursively solve and forward
				solveAndForward(self, id, pkt.query, nil)
			} else { // This is a response packet
				fmt.Println(self.id + ": " + pkt.toString() + " from " + id)

				// Recursively consume the requests
				consumePendingRequest(self, pkt.query, pkt.result, pkt.query)
			}
		case <-time.After(time.Second * 2):
			for addr, route := range self.fib {
				if route.nexthop == id {
					// Split-horizon
					continue;
				}

				// Add 1 to the metric when making annoucement
				self.announceFIB(id, addr, route.metric + 1)
			}
		}
	}
}

func (self *Solver) waitForNonSolver(id string, f *Interface) {
	// For requester and data source, just do blocking recv
	for {  // Infinite loop
		pkt := <-f.in

		if pkt.result == nil { // This is a request packet
			fmt.Println(self.id + ": ?" + pkt.query.toString() + " from " + id)

			// Recursively solve and forward
			solveAndForward(self, id, pkt.query, nil)
		} else { // This is a response packet
			fmt.Println(self.id + ": " + pkt.toString() + " from " + id)

			// Recursively consume the requests
			consumePendingRequest(self, pkt.query, pkt.result, pkt.query)
		}
	}
}

func (self *Solver) run() {
	for id, f := range self.ift {
		if strings.HasPrefix(id, "s") {
			// This is interface to a solver
			go self.waitForSolver(id, f)
		} else {
			go self.waitForNonSolver(id, f)
		}
	}
}
