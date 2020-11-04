package plan

type Property interface {
	Type() PropertyType
}

type PropertyType int

const (
	PropertySorted PropertyType = iota
)

type Sorted struct {
	sortField     string
	sortDirection SortDirection
}

func NewSorted(sortField string, sortDirection SortDirection) *Sorted {
	return &Sorted{sortField, sortDirection}
}

func (p *Sorted) Type() PropertyType {
	return PropertySorted
}

type SortDirection int

const (
	SortDirectionAsc SortDirection = iota
	SortDirectionDesc
)
