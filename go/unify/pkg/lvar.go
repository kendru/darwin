package pkg

type LVar struct {
	id int
}

func (lv *LVar) Eq(other *LVar) bool {
	return lv.id == other.id
}
