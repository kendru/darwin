package testutil

func GenericInt64Slice(ns ...int64) []interface{} {
	out := make([]interface{}, len(ns))
	for i, n := range ns {
		out[i] = n
	}
	return out
}

func GenericUInt64Slice(ns ...uint64) []interface{} {
	out := make([]interface{}, len(ns))
	for i, n := range ns {
		out[i] = n
	}
	return out
}

func GenericStringSlice(ns ...string) []interface{} {
	out := make([]interface{}, len(ns))
	for i, n := range ns {
		out[i] = n
	}
	return out
}
