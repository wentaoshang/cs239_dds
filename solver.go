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
	depend map[string]int  // Dependency list of rules, the int value is not important
	results map[string]string  // Temporary storage for dependency results
	compose *Atom  // Signature for the composition rule, in the form of "@(?X, ?A, ?B, ?C, ...)"
	unify map[string]string  // Unification result
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
	// Triggered update
	for i, _ := range self.ift {
		if strings.HasPrefix(i, "s") {
			self.announceFIB(i, src.cname, 0)
		}
	}
	go self.waitForNonSolver(src.id, self.ift[src.id])
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
	go self.waitForNonSolver(req.id, self.ift[req.id])
}

// Magic function to compose a single piece of data from multiple answers.
// sig is the signature of the compose atom defined in the rules.
// It will take the form of "@(?X, ?A, ?B, ?C)", which means ?X is composed
// by concatenating answers for ?A, ?B, and ?C.
// We only define concatenation semantics for simplicity. Ideally this could 
// be any user-defined functions as long as they have the same function type.
func compose(sig *Atom, results map[string]string) (v string, ans string) {
	v = sig.args[0]  // Variable to be composed is listed in front
	ans = "{"
	for i := 1; i < len(sig.args); i++ {
		vdeps := sig.args[i]  // Dependent variable
		ans += results[vdeps]  // It is an error if vdeps doesn't appear in results
		// Remove vdeps from results
		delete(results, vdeps)
		if i != len(sig.args) - 1 {
			ans += " | "  // Composition separator
		}
	}
	ans += "}"
	return
}

func (self *Solver) solveAndForward(origin string, query *Atom, solve *Atom) {
	// Lookup rule cache
	foundRule := false
	for _, rule := range self.rc {
		res, match := rule.unify(query)
		if match {
			foundRule = true
			fmt.Println(self.id + ": ?" + query.toString() + " matches rule " + rule.toString() + " with " + mapToString(res))
			// Record query in prt with dependencies
			var pr PendingRequest
			pr.query = query
			pr.solve = solve
			pr.depend = make(map[string]int)
			pr.results = make(map[string]string)
			for i, a := range rule.body {
				if a.name == "@" {
					// For now only one composition rule is allowed
					pr.compose = a
				} else {
					pr.depend[a.toString()] = i
				}
			}
			pr.unify = res
			pr.origin = origin
			self.prt[query.toString()] = &pr

			for _, a := range rule.body {
				self.solveAndForward("", a, query)
			}
		}
	}

	if foundRule {
		return
	}

	// No match in rule table
	fmt.Println(self.id + ": no rule for ?" + query.toString())

	// Record query in prt, without dependencies
	var pr PendingRequest
	pr.query = query
	pr.solve = solve
	pr.origin = origin
	self.prt[query.toString()] = &pr

	// Lookup fib and forward
	route, ok := self.fib[query.getName()]
	if ok {
		nexthop := route.nexthop
		fmt.Println(self.id + ": nexthop for ?" + query.toString() + " is " + nexthop)
		of := self.ift[nexthop]
		var pkt Packet
		pkt.query = query
		of.out <- &pkt
	}
}

// Consume 'query' using 'result' of 'from' 
func (self *Solver) consumePendingRequest(query *Atom, result map[string]string, from *Atom) {
	// Lookup prt
	pr, ok := self.prt[query.toString()]
	if ok {
		if query.toString() == from.toString() {
			// Consume the query using its own result
			fmt.Println(self.id + ": consume ?" + query.toString() + " with " + mapToString(result))

			// Delete query from prt
			delete(self.prt, query.toString())

			if pr.compose != nil {
				// Need to compose multiple results
				v, res := compose(pr.compose, result)
				// Store the composed result
				result[v] = res
			}

			if pr.solve != nil {
				// This query is a dependency of another query
				//fmt.Println(self.id + ": " + query.toString() + " -> " + mapToString(result) + " satisfies a dependency of ?" + pr.solve.toString())
				// Pass the result up
				self.consumePendingRequest(pr.solve, result, query)
			} else {
				// This query doesn't solve other queries
				if pr.unify == nil {
					// Directly send back the result
					pr.unify = result
				} else {
					// Unify the original query with the result
					for key, val := range pr.unify {
						if val2, ok := result[val]; ok {
							pr.unify[key] = val2
						}
					}
				}

				// Send result back
				of := self.ift[pr.origin]  // ASSERT: pr.origin should not be empty if pr.solve is nil
				var pkt Packet
				pkt.query = query
				pkt.result = pr.unify
				fmt.Println(self.id + ": send " + pkt.toString() + " back to " + pr.origin)
				of.out <- &pkt
			}
		} else {
			// A dependency is resolved
			fmt.Println(self.id + ": " + from.toString() + " -> " + mapToString(result) + " satisfies a dependency of ?" + pr.query.toString())
			// Copy the result and clear the dependency
			for key, val := range result {
				pr.results[key] = val
			}
			delete(pr.depend, from.toString())

			fmt.Println(self.id + ": dependency results are " + mapToString(pr.results))

			if len(pr.depend) == 0 {
				fmt.Println(self.id + ": all dependencies resolved ?" + query.toString() + " with " + mapToString(pr.results))
				// Consume itself
				self.consumePendingRequest(pr.query, pr.results, pr.query)
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
				self.processFIB(pkt)
			} else if pkt.result == nil { // This is a request packet
				fmt.Println(self.id + ": ?" + pkt.query.toString() + " from " + id)

				// Recursively solve and forward
				self.solveAndForward(id, pkt.query, nil)
			} else { // This is a response packet
				fmt.Println(self.id + ": " + pkt.toString() + " from " + id)

				// Recursively consume the requests
				self.consumePendingRequest(pkt.query, pkt.result, pkt.query)
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
			self.solveAndForward(id, pkt.query, nil)
		} else { // This is a response packet
			fmt.Println(self.id + ": " + pkt.toString() + " from " + id)

			// Recursively consume the requests
			self.consumePendingRequest(pkt.query, pkt.result, pkt.query)
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
