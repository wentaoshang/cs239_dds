package main

//import "fmt"
//import "time"

func main() {
	// Create links
	l1 := createLink("l1")

	// Create solvers
	s1 := createSolver("s1")
	s2 := createSolver("s2")

	// Connect solvers by links
	l1.connect(s1, s2)

	// Create data sources
	rules := []string{"location(Westwood).",
		"location(Wilshire).",
		"location(National).",
		"location(Sepulveda).",
	}
	d1 := createSource("d1", rules)

	// Add data sources to solvers
	s2.addSource(d1)

	// Configure FIB in solvers
	s1.addForwardingEntry(d1.cname, s2)

	s1.addRule("loc(?X) <- loc2(?X).")
	s1.addRule("loc2(?L) <- location(?L).")

	// Create request
	r1 := createRequest("r1", "loc(?X)")

	// Add request to a solver
	s1.addRequest(r1)

	// Start goroutines
	go d1.run()
	go s1.run()
	go s2.run()
	go r1.run()

	// Wait for r1 to finish
	<- r1.done
}
