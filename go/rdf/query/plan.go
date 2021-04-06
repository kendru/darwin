package query

type NodeType int

const (
	NodeScan NodeType = iota
	NodeProject
	NodeJoin
)

func NewLogicalPlanNode(typ NodeType, slots map[string]interface{}) *LogicalPlanNode {
	return &LogicalPlanNode{
		typ:   typ,
		slots: slots,
	}
}

// LogicalPlanNode is a dynamically typed node in a query plan.
// The advantage of dynamically typing is that it allows query transformations to be
// more easily applied in a generic way. Once we have a better understanding of the
// types of transformations that we need to perform, we should switch this out for
// statically typed nodes and code generation for building and transforming a plan
// tree.
type LogicalPlanNode struct {
	typ   NodeType
	slots map[string]interface{}
	cost  int
}

/*
Misc NOTES:
- A Scan always yields the triple in the order specified by the index.
  - Later we can optimize to project zero or more fields (zero fields could be useful for counts).

*/
