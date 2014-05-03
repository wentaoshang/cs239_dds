package main

// Abstraction for packet format

type Packet struct {
	query *Atom
	result map[string]string
}

func mapToString(m map[string]string) string {
	var s string
	for key, val := range m {
		s += key + " = " + val + ", "
	}
	return s
}

func (self *Packet) queryToString() string {
	return self.query.toString()
}

func (self *Packet) resultToString() string {
	return mapToString(self.result)
}

func (self *Packet) toString() string {
	var s string
	s = self.queryToString() + " --> " + self.resultToString()
	return s
}
