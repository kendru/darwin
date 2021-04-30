package internal

type Database interface {
	Find(name string) (*Command, error)
}

