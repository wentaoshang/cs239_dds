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
	d1 := createSource("d1", "test", "test data")

	// Add data sources to solvers
	s2.addSource(d1)

	// Configure FIB in solvers
	s1.addForwardingEntry("test", s2)

	// Create request
	r1 := createRequest("r1", "test")

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
