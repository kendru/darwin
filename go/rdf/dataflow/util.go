package dataflow

import "github.com/kendru/darwin/go/rdf/tuple"

func Collect(n Node) ([]*tuple.Tuple, error) {
	var results []*tuple.Tuple
	for {
		elem, err := n.Next()
		if err != nil {
			return nil, err
		}
		if elem == nil {
			break
		}
		results = append(results, elem)
	}

	return results, nil
}
