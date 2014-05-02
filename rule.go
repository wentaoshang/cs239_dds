package main

//import "fmt"
import "strings"

type Atom struct {
	name string
	args []string
}

func createAtom(name string, args []string) *Atom {
	var a Atom
	a.name = name
	a.args = args
	return &a
}

func (self *Atom) toString() string {
	var s string
	s = self.name + "("
	for i, arg := range self.args {
		s += "\"" + arg + "\""
		if i != len(self.args) - 1 {
			s += ", "
		}
	}
	s += ")"
	return s
}

type Rule struct {
	head *Atom
	body [](*Atom)
}

func createRule(atoms [](*Atom)) *Rule {
	var r Rule
	r.head = atoms[0]
	r.body = atoms[1:]
	return &r
}

func (self *Rule) toString() string {
	var s string
	s = self.head.toString() + " <- "
	for i, a := range self.body {
		s += a.toString()
		if i == len(self.body) - 1 {
			s += "."
		} else {
			s += ", "
		}
	}
	return s
}

func (self *Rule) unify(query *Atom) (res map[string]string, ok bool) {
	head := self.head
	if query.name != head.name {
		ok = false
		return
	}

	if len(head.args) != len(query.args) {
		ok = false
		return
	}

	ok = true
	res = make(map[string]string)
	for i, arg := range query.args {
		if strings.HasPrefix(arg, "?") {
			// Is a variable
			res[arg] = head.args[i]
		} else if arg != head.args[i] {
			// Constant must be exact match
			ok = false
			res = nil
			return
		}
	}
	return
}
