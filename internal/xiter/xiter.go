package xiter

import (
	"iter"
)

func Enum[V any](it iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		for value := range it {
			if !yield(i, value) {
				break
			}
			i++
		}
	}
}

func Enum2[K, V any](it iter.Seq2[K, V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		for _, value := range it {
			if !yield(i, value) {
				break
			}
			i++
		}
	}
}
