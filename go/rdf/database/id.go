package database

import "math/rand"

type TempID struct {
	inst       int64
	isAssigned bool
	ID         uint64
}

func Fresh() *TempID {
	return &TempID{inst: rand.Int63()}
}
