package math

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T, rest ...T) T {
	m := a
	if b > a {
		m = b
	}
	for _, val := range rest {
		if val > m {
			m = val
		}
	}
	return m
}

func Min[T constraints.Ordered](a, b T, rest ...T) T {
	m := a
	if b < a {
		m = b
	}
	for _, val := range rest {
		if val < m {
			m = val
		}
	}
	return m
}
