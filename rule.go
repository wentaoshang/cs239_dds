package main

//import "fmt"
import "strings"
import "strconv"

type Atom struct {
	name string
	args []string
}

func createAtom(fact string) *Atom {
	var a Atom
	ss := strings.Split(fact, "(")
	a.name = ss[0]
	rest := strings.Trim(ss[1], ")")
	args := strings.Split(rest, ",")
	for i, arg := range args {
		args[i] = strings.Trim(arg, " \"")
	}
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

func (self *Atom) getName() string {
	var s string
	s = self.name + "/" + strconv.Itoa(len(self.args))
	return s
}

// Generate canonical string for the query
// Variables are replaced by placeholders
func (self *Atom) toCString() string {
	var s string
	s = self.name + "-"
	for i, arg := range self.args {
		if strings.HasPrefix(arg, "?") {
			s += "?" + string(i) + "-"
		} else {
			s += arg + "-"
		}
	}
	return s
}

type Rule struct {
	head *Atom
	body [](*Atom)
}

func createRule(rule string) *Rule {
	var r Rule
	ss := strings.Split(rule, "<-")
	head := strings.Trim(ss[0], ". ")
	r.head = createAtom(head)

	if len(ss) > 1 {
		body := strings.Trim(ss[1], ". ")
		args := strings.Split(body, "),")  //XXX: hack!
		var atoms [](*Atom)
		for _, arg := range args {
			atoms = append(atoms, createAtom(strings.Trim(arg, "\" ")))
		}
		r.body = atoms
	}
	return &r
}

func (self *Rule) toString() string {
	var s string
	s = self.head.toString()
	if self.body != nil {
		s += " <- "
		for i, a := range self.body {
			s += a.toString()
			if i != len(self.body) - 1 {
				s += ", "
			}
		}
	}
	s += "."
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
