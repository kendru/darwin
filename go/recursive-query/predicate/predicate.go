package predicate

// TODO: Evaluate predicates inside an environment, and allow predicates
// to reference variables from the environment. The filter machinery could
// provide these variables from the row under evaluation.

type Predicate interface {
	Test() bool
}

type StringEquals struct {
	Lh string
	Rh string
}

func (p StringEquals) Test() bool {
	return p.Lh == p.Rh
}

type And struct {
	Predicates []Predicate
}

func (p And) Test() bool {
	for _, pred := range p.Predicates {
		if !pred.Test() {
			return false
		}
	}
	return true
}

type Or struct {
	Predicates []Predicate
}

func (p Or) Test() bool {
	for _, pred := range p.Predicates {
		if pred.Test() {
			return true
		}
	}
	return false
}

type constantFalse struct {}

func (p constantFalse) Test() bool {
	return false
}

var ConstantFalse = constantFalse{}

