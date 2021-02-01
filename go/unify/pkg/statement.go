package pkg

import "fmt"

type LogicStatement interface {
	Run(db *DB) error
}

func Unify(lhs, rhs interface{}) *UnifyStatement {
	return &UnifyStatement{
		lhs: lhs,
		rhs: rhs,
	}
}

type UnifyStatement struct {
	lhs, rhs interface{}
}

func (stmt *UnifyStatement) Run(db *DB) error {
	lhVal := db.Walk(stmt.lhs)
	rhVal := db.Walk(stmt.rhs)

	if l, ok := lhVal.(*LVar); ok {
		if r, ok := rhVal.(*LVar); ok && l.Eq(r) {
			return nil
		}
		db.Extend(l, rhVal)
		return nil
	}

	if r, ok := rhVal.(*LVar); ok {
		db.Extend(r, lhVal)
		return nil
	}

	if lhVal == rhVal {
		return nil
	}

	return fmt.Errorf("cannot unify %v to %v", lhVal, rhVal)
}

func RunAll(db *DB, stmts ...LogicStatement) error {
	for _, stmt := range stmts {
		if err := stmt.Run(db); err != nil {
			return err
		}
	}
	return nil
}
