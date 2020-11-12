package database

import (
	"fmt"

	"github.com/kendru/darwin/go/rdf/tuple"
)

type Fact struct {
	Subject   uint64
	Predicate string
	Object    interface{}
}

func (f Fact) predicateSubjectKey() []byte {
	return tuple.New(f.Predicate, f.Subject).Serialize()
}

func (f Fact) subjectPredicateKey() []byte {
	return tuple.New(f.Subject, f.Predicate).Serialize()
}

func (f Fact) String() string {
	return fmt.Sprintf("%d %s %v", f.Subject, f.Predicate, f.Object)
}
