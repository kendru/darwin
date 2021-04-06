package dataflow

import (
	"fmt"

	"github.com/kendru/darwin/go/rdf/index"
	"github.com/kendru/darwin/go/rdf/tuple"
)

func NewRowSchema(elements ...ElementDescriptor) *RowSchema {
	return &RowSchema{elements}
}

type RowSchema struct {
	elements []ElementDescriptor
}

func MakeElementDescriptor(typ tuple.ElemType, alias string) ElementDescriptor {
	return ElementDescriptor{
		typ:   typ,
		alias: alias,
	}
}

type ElementDescriptor struct {
	typ   tuple.ElemType
	alias string
}

type Node interface {
	Next() (*tuple.Tuple, error)
	Close() error
	Schema() *RowSchema
}

// BEGIN SliceGenerator

func NewSliceGenerator(schema *RowSchema, slice []*tuple.Tuple) *SliceGenerator {
	return &SliceGenerator{
		schema: schema,
		slice:  slice,
	}
}

type SliceGenerator struct {
	schema *RowSchema
	slice  []*tuple.Tuple
	i      int
}

func (n *SliceGenerator) Next() (res *tuple.Tuple, err error) {
	if n.i < len(n.slice) {
		res = n.slice[n.i]
		n.i++
	}
	return
}

func (n *SliceGenerator) Schema() *RowSchema {
	return n.schema
}

func (n *SliceGenerator) Close() error {
	n.slice = nil
	return nil
}

// END SliceGenerator

// BEGIN IndexScan

type indexScanRowResult struct {
	tup *tuple.Tuple
	err error
}

func NewIndexScan(schema *RowSchema, scanner index.Scanner, args index.ScanArgs) *IndexScan {
	return &IndexScan{
		schema:  schema,
		scanner: scanner,
		args:    args,
	}
}

type IndexScan struct {
	schema   *RowSchema
	scanner  index.Scanner
	args     index.ScanArgs
	elem     chan indexScanRowResult
	isClosed int32
}

func (n *IndexScan) Next() (res *tuple.Tuple, err error) {
	if n.elem == nil {
		n.elem = make(chan indexScanRowResult, 100)
		go func() {
			// TODO: recover panic if we write to a closed channel.
			if err := index.ScanAll(n.scanner, n.args, func(entry *index.IndexEntry) (bool, error) {
				key, err := tuple.Deserialize(entry.Key)
				if err != nil {
					return false, fmt.Errorf("error deserializing key as tuple: %w", err)
				}

				for _, valBytes := range entry.Values {
					val, err := tuple.Deserialize(valBytes)
					if err != nil {
						return false, fmt.Errorf("error deserializing value as tuple: %w", err)
					}
					n.elem <- indexScanRowResult{tup: key.Concat(val)}
				}

				return true, nil
			}); err != nil {
				n.elem <- indexScanRowResult{err: err}
			}
			close(n.elem)
		}()
	}

	for row := range n.elem {
		return row.tup, nil
	}

	return nil, nil
}

func (n *IndexScan) Schema() *RowSchema {
	return n.schema
}

func (n *IndexScan) Close() error {
	close(n.elem)
	return nil
}

// END IndexScan

// BEGIN FilterNode

func NewFilterNode(source Node, pred func(*tuple.Tuple, *RowSchema) (bool, error)) *FilterNode {
	return &FilterNode{
		source: source,
		pred:   pred,
	}
}

type FilterNode struct {
	source Node
	pred   func(*tuple.Tuple, *RowSchema) (bool, error)
}

func (n *FilterNode) Next() (res *tuple.Tuple, err error) {
	schema := n.source.Schema()
	for {
		res, err := n.source.Next()
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, nil
		}
		if ok, err := n.pred(res, schema); err != nil {
			return nil, fmt.Errorf("error evaluating predicate: %w", err)
		} else if ok {
			return res, nil
		}
	}
}

func (n *FilterNode) Schema() *RowSchema {
	return n.source.Schema()
}

func (n *FilterNode) Close() error {
	return n.source.Close()
}

// END FilterNode

// BEGIN LimitNode

func NewLimitNode(source Node, limit int) *LimitNode {
	return &LimitNode{
		source: source,
		limit:  limit,
	}
}

type LimitNode struct {
	source Node
	limit  int
}

func (n *LimitNode) Next() (*tuple.Tuple, error) {
	if n.limit == 0 {
		return nil, nil
	}
	n.limit--
	return n.source.Next()
}

func (n *LimitNode) Schema() *RowSchema {
	return n.source.Schema()
}

func (n *LimitNode) Close() error {
	return n.source.Close()
}

// END LimitNode

// BEGIN ProjectRenameNode

func MakeProjection(src, dest string) Projection {
	return Projection{
		Src:  src,
		Dest: dest,
	}
}

type Projection struct {
	// TODO: make dest potentially an expression of src.
	Src, Dest string
}

func NewProjectRenameNode(source Node, projections ...Projection) *ProjectRenameNode {
	// Schema should reflect rename and projection.
	outElemCount := len(projections)
	descriptors := make([]ElementDescriptor, outElemCount)
	idxMapping := make([]int, outElemCount)

	origElems := source.Schema().elements
	sourceElementIdxByAlias := make(map[string]int)
	for i, schemaElem := range origElems {
		sourceElementIdxByAlias[schemaElem.alias] = i
	}

	for outIdx, projection := range projections {
		// TODO: Validate that the source exists
		origIdx := sourceElementIdxByAlias[projection.Src]
		oridElem := origElems[origIdx]
		idxMapping[outIdx] = origIdx
		descriptors[outIdx] = MakeElementDescriptor(oridElem.typ, projection.Dest)
	}

	return &ProjectRenameNode{
		source:     source,
		schema:     NewRowSchema(descriptors...),
		idxMapping: idxMapping,
	}
}

type ProjectRenameNode struct {
	source     Node
	schema     *RowSchema
	idxMapping []int // [outputIdx]->inputIdx
}

func (n *ProjectRenameNode) Next() (*tuple.Tuple, error) {
	in, err := n.source.Next()
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
	}

	elements := make([]interface{}, len(n.idxMapping))
	for idxOut, idxIn := range n.idxMapping {
		if elements[idxOut], err = in.GetUntyped(idxIn); err != nil {
			return nil, err
		}
	}

	return tuple.New(elements...), nil
}

func (n *ProjectRenameNode) Schema() *RowSchema {
	return n.schema
}

func (n *ProjectRenameNode) Close() error {
	return n.source.Close()
}

// END ProjectRenameNode

// BEGIN IntoDocumentNode

func NewIntoDocumentNode(source Node) *IntoDocumentNode {
	return &IntoDocumentNode{
		source: source,
		// TODO: expand the type system for tuples such that they can hold "structs".
		schema: NewRowSchema(),
	}
}

type IntoDocumentNode struct {
	source Node
	schema *RowSchema
}

func (n *IntoDocumentNode) Next() (*tuple.Tuple, error) {
	tup, err := n.source.Next()
	if err != nil {
		return nil, err
	}
	if tup == nil {
		return nil, nil
	}

	doc := make(map[string]interface{})
	for i, descriptor := range n.source.Schema().elements {
		name := descriptor.alias
		value, err := tup.GetUntyped(i)
		if err != nil {
			return nil, err
		}
		doc[name] = value
	}

	return tuple.New(doc), nil
}

func (n *IntoDocumentNode) Schema() *RowSchema {
	return n.schema
}

func (n *IntoDocumentNode) Close() error {
	return n.source.Close()
}

// END IntoDocumentNode

// BEGIN InnerJoinNode

func NewInnerJoinNode(leftSrc Node, leftIdx int, rightSrc Node, rightIdx int) *InnerJoinNode {
	leftSchema := leftSrc.Schema()
	rightSchema := rightSrc.Schema()
	schema := NewRowSchema(append(append([]ElementDescriptor{}, leftSchema.elements...), rightSchema.elements...)...)

	return &InnerJoinNode{
		schema:   schema,
		leftSrc:  leftSrc,
		leftIdx:  leftIdx,
		rightSrc: rightSrc,
		rightIdx: rightIdx,
	}
}

type InnerJoinNode struct {
	schema            *RowSchema
	leftSrc, rightSrc Node
	leftIdx, rightIdx int // nth tuple elements to join on for each side
	leftRel, rightRel []*tuple.Tuple
	fetched           bool
	i, j              int
}

func (n *InnerJoinNode) Next() (*tuple.Tuple, error) {
	if !n.fetched {
		var err error
		n.leftRel, err = Collect(n.leftSrc)
		if err != nil {
			return nil, err
		}
		n.rightRel, err = Collect(n.rightSrc)
		if err != nil {
			return nil, err
		}

		if len(n.rightRel) == 0 || len(n.leftRel) == 0 {
			return nil, nil
		}

		n.fetched = true
	}

	for {
		if n.i >= len(n.leftRel) {
			return nil, nil
		}
		leftTup := n.leftRel[n.i]
		keyLeft, err := leftTup.GetUntyped(n.leftIdx)
		if err != nil {
			return nil, err
		}

		for {
			if n.j >= len(n.rightRel) {
				n.j = 0
				break
			}
			rightTup := n.rightRel[n.j]
			n.j++

			keyRight, err := rightTup.GetUntyped(n.rightIdx)
			if err != nil {
				return nil, err
			}

			if keyLeft == keyRight {
				return leftTup.Concat(rightTup), nil
			}
		}

		n.i++
	}
}

func (n *InnerJoinNode) Schema() *RowSchema {
	return n.schema
}

func (n *InnerJoinNode) Close() error {
	return coalesceErrors(
		n.leftSrc.Close(),
		n.rightSrc.Close(),
	)
}

// END InnerJoinNode

func coalesceErrors(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
