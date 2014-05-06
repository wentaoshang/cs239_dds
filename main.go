package main

//import "fmt"
import "time"

func main() {
	// Create links
	l1 := createLink("l1")
	l2 := createLink("l2")

	// Create solvers
	s1 := createSolver("s1")
	s2 := createSolver("s2")
	s3 := createSolver("s3")

	// Connect solvers by links
	l1.connect(s1, s2)
	l2.connect(s2, s3)

	// Create data sources
	rules := []string{"location(Westwood).",
		"location(Wilshire).",
		"location(National).",
		"location(Sepulveda).",
	}
	d1 := createSource("d1", rules)

	rules2 := []string{"price($5)."}
	d2 := createSource("d2", rules2)

	// Configure FIB in solvers
	//s1.addForwardingEntry(d1.cname, s2)

	//s1.addRule("loc(?X) <- loc2(?X).")
	//s1.addRule("loc2(?L) <- location(?L).")

	s1.addRule("item(?X) <- price(?P), location(?L), @(?X, ?P, ?L).")

	// Create request
	r1 := createRequest("r1", "item(?X)")

	// Start goroutines
	go d1.run()
	go d2.run()
	go s1.run()
	go s2.run()
	go s3.run()

	time.Sleep(time.Second * 2)
	// Add data sources to solvers
	s3.addSource(d1)
	s2.addSource(d2)

	time.Sleep(time.Second * 5)
	// Add request to a solver
	s1.addRequest(r1)
	go r1.run()

	// Wait for r1 to finish
	<- r1.done
}
