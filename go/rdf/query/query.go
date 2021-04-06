package query

import (
	"fmt"

	"github.com/kendru/darwin/go/rdf/database"
	"github.com/kendru/darwin/go/rdf/dataflow"
	"github.com/kendru/darwin/go/rdf/index"
	"github.com/kendru/darwin/go/rdf/tuple"
)

func Fresh() *Variable {
	return &Variable{}
}

func NamedVar(name string) *Variable {
	return &Variable{name: name}
}

type Variable struct {
	// name is optional
	name string
}

func (v *Variable) String() string {
	var name string
	if v.name == "" {
		name = "?"
	} else {
		name = v.name
	}

	return fmt.Sprintf("<%s@%p>", name, v)
}

// NewRule creates a rule that acts as a query clause. Any of subject,
// predicate, and object may be values or Variables.
func NewRule(s interface{}, p interface{}, o interface{}) *Rule {
	switch sub := s.(type) {
	case *Variable, uint64:
		// ok - do nothing
	case int:
		s = uint64(sub)
	default:
		panic(fmt.Errorf("Unsupported subject type: %T", s))
	}

	return &Rule{
		Subject:   s,
		Predicate: p,
		Object:    o,
	}
}

type Rule struct {
	Subject, Predicate, Object interface{}
}

func NewPattern(vars ...*Variable) *Pattern {
	return &Pattern{vars}
}

// TODO: Make pattern more flexible and allow projection/computation.
type Pattern struct {
	vars []*Variable
}

func NewQuery(p *Pattern, r ...*Rule) *Query {
	return &Query{
		pattern: p,
		rules:   r,
	}
}

type Query struct {
	pattern *Pattern
	rules   []*Rule
}

func (q *Query) Execute(db database.Database) (res QueryResult, err error) {
	type nodeWithProjectedVars struct {
		node       *LogicalPlanNode
		varAliases []string
	}
	scans := make([]nodeWithProjectedVars, len(q.rules))
	// Each rule turns into a scan
	for i, rule := range q.rules {
		subVar, subIsVar := rule.Subject.(*Variable)
		predVar, predIsVar := rule.Predicate.(*Variable)
		objVar, objIsVar := rule.Object.(*Variable)

		var index string
		var prefix *tuple.Tuple

		var projections []dataflow.Projection
		var projectedVarAliases []string
		addToProjection := func(src string, dest *Variable) {
			destAlias := dest.String()
			projections = append(projections, dataflow.MakeProjection(src, destAlias))
			projectedVarAliases = append(projectedVarAliases, destAlias)
		}

		switch {
		case !subIsVar && !predIsVar && !objIsVar:
			// A fully-specified pattern adds nothing to the output and does not
			// restrict the selection.
			continue

		case subIsVar && !predIsVar:
			addToProjection("subject", subVar)
			if objIsVar {
				addToProjection("object", objVar)
				index = "pso"
				prefix = tuple.New(rule.Predicate)
			} else {
				// TODO: ensure that Object is indexable type.
				index = "pos"
				prefix = tuple.New(rule.Predicate, rule.Object)
			}

		case predIsVar && !subIsVar:
			addToProjection("predicate", predVar)
			if objIsVar {
				addToProjection("object", objVar)
			}
			// TODO: Filter on the results if the object is specified.
			index = "spo"
			prefix = tuple.New(rule.Subject)

		case objIsVar && !subIsVar:
			addToProjection("object", objVar)
			index = "spo"
			if predIsVar {
				addToProjection("predicate", predVar)
				prefix = tuple.New(rule.Subject)
			} else {
				prefix = tuple.New(rule.Subject, rule.Predicate)
			}

		default:
			return QueryResult{}, fmt.Errorf("Unsupported rule: %s", rule)
		}

		scans[i] = nodeWithProjectedVars{
			node: NewLogicalPlanNode(NodeProject, map[string]interface{}{
				"projections": projections,
				"source": NewLogicalPlanNode(NodeScan, map[string]interface{}{
					"index":  index,
					"prefix": prefix.Serialize(),
				}),
			}),
			varAliases: projectedVarAliases,
		}
	}

	// Join All Scans:
	// All projections have a variable name as the destination alias.
	// In this naive implementation, we iterate through each scan and look through the set of
	// nodes. If we do not find anything to attach it to, we move on to the next element. We
	// repeat this process until there is only 1 node left that represents all scans joined
	// together (success case) or we iterate through the set of nodes without being able to
	// make any changes (failure case).

	// Iterate scans until fixpoint found.
	var changed bool
fixpoint:
	for {
		changed = false

		// TODO: Rename scan, since it is just some table-yielding node.
		for i, scan := range scans {
			for j, otherScan := range scans {
				// Do not join a node to itself
				if i == j {
					continue
				}

				for thisElementIdx, alias := range scan.varAliases {
					for otherElementIdx, otherAlias := range otherScan.varAliases {
						if alias == otherAlias {
							changed = true
							// Replace `scan` with join(scan, other scan)
							scans[i] = nodeWithProjectedVars{
								node: NewLogicalPlanNode(NodeJoin, map[string]interface{}{
									"leftSrc":  scan.node,
									"rightSrc": otherScan.node,
									// Since the only fields in the scans are the variables, we can rely on the fact that
									// the nth alias is the alias for the nth element of the tuple.
									"leftIdx":  thisElementIdx,
									"rightIdx": otherElementIdx,
								}),
								// TODO: Do we need to exclude the join attribute from the righthand side and project to remove it?
								// XXX: Check me first when things break.
								varAliases: append(scan.varAliases, otherScan.varAliases...),
							}

							// Remove otherScan
							// TODO: remove in-place and avoid allocation.
							scans = append(scans[:j], scans[j+1:]...)
							continue fixpoint
						}
					}
				}
			}
		}

		if !changed {
			break
		}
	}

	// If there are more than one elements in scans, that means that there were multiple join trees that could not be connected,
	// implying that there are two or more independent result sets, and joining them would yield a cartesian product.
	if len(scans) > 1 {
		panic("Queries with cartesian joins not permitted")
	}

	var planRoot *LogicalPlanNode
	planRoot = scans[0].node

	outProjections := make([]dataflow.Projection, len(q.pattern.vars))
	for i, v := range q.pattern.vars {
		alias := v.String()
		outProjections[i] = dataflow.MakeProjection(alias, alias)
	}

	planRoot = NewLogicalPlanNode(NodeProject, map[string]interface{}{
		"projections": outProjections,
		"source":      planRoot,
	})

	return executePlan(db, planRoot)
}

func executePlan(db database.Database, p *LogicalPlanNode) (QueryResult, error) {
	physicalPlan, err := optimizePlan(db, p)
	if err != nil {
		return QueryResult{}, fmt.Errorf("error optimizing logical plan: %w", err)
	}

	res, err := runPlan(physicalPlan)
	if err != nil {
		return QueryResult{}, fmt.Errorf("error running physical plan: %w", err)
	}

	return res, nil
}

func optimizePlan(db database.Database, p *LogicalPlanNode) (dataflow.Node, error) {
	switch p.typ {
	case NodeScan:
		indexName := p.slots["index"].(string)
		prefix := p.slots["prefix"].([]byte)

		s := dataflow.MakeElementDescriptor(tuple.TypeUint64, "subject")
		p := dataflow.MakeElementDescriptor(tuple.TypeUnicode, "predicate")
		o := dataflow.MakeElementDescriptor(tuple.TypeUnknown, "object")

		var idx index.Scanner
		var schema *dataflow.RowSchema
		switch indexName {
		case "spo":
			idx = db.SPO()
			schema = dataflow.NewRowSchema(s, p, o)
		case "pso":
			idx = db.PSO()
			schema = dataflow.NewRowSchema(p, s, o)
		case "pos":
			idx = db.POS()
			schema = dataflow.NewRowSchema(p, o, s)
		}

		return dataflow.NewIndexScan(schema, idx, index.ScanArgs{
			Prefix: prefix,
		}), nil

	case NodeProject:
		projections := p.slots["projections"].([]dataflow.Projection)
		sourcePlan := p.slots["source"].(*LogicalPlanNode)

		source, err := optimizePlan(db, sourcePlan)
		if err != nil {
			return nil, err
		}

		return dataflow.NewProjectRenameNode(source, projections...), nil

	case NodeJoin:
		leftIdx := p.slots["leftIdx"].(int)
		leftSrcPlan := p.slots["leftSrc"].(*LogicalPlanNode)
		leftSrc, err := optimizePlan(db, leftSrcPlan)
		if err != nil {
			return nil, err
		}

		rightIdx := p.slots["rightIdx"].(int)
		rightSrcPlan := p.slots["rightSrc"].(*LogicalPlanNode)
		rightSrc, err := optimizePlan(db, rightSrcPlan)
		if err != nil {
			return nil, err
		}

		return dataflow.NewInnerJoinNode(leftSrc, leftIdx, rightSrc, rightIdx), nil

	default:
		return nil, fmt.Errorf("Unknown plan node: %d", p.typ)
	}
}

func runPlan(queryRoot dataflow.Node) (QueryResult, error) {
	rows, err := dataflow.Collect(queryRoot)
	if err != nil {
		return QueryResult{}, fmt.Errorf("error getting rows: %w", err)
	}

	return QueryResult{
		Rows: rows,
	}, nil
}

type QueryResult struct {
	Rows []*tuple.Tuple
}
