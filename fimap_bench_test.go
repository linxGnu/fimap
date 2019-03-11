package fimap

import (
	"testing"

	"github.com/valyala/fastrand"
)

var smallSet, mediumSet, largeSet []keyType

func initializeSet(size int) []keyType {
	v := make([]keyType, size)
	for i := range v {
		v[i] = 1 + keyType(fastrand.Uint32n(2000000000))
	}
	return v
}

func init() {
	smallSet = initializeSet(1024)
	mediumSet = initializeSet(2000000) // 2 millions
	largeSet = initializeSet(10000000) // 10 millions
}

func BenchmarkMapSmall(b *testing.B) {
	benchmarkMap(b, smallSet)
}

func BenchmarkMapMedium(b *testing.B) {
	benchmarkMap(b, mediumSet)
}

func BenchmarkMapLarge(b *testing.B) {
	benchmarkMap(b, largeSet)
}

func BenchmarkFIMapSmall(b *testing.B) {
	benchmarkFIMap(b, smallSet)
}

func BenchmarkFIMapMedium(b *testing.B) {
	benchmarkFIMap(b, mediumSet)
}

func BenchmarkFIMapLarge(b *testing.B) {
	benchmarkFIMap(b, largeSet)
}

func benchmarkFIMap(b *testing.B, set []keyType) {
	for i := 0; i < b.N; i++ {
		s, _ := New(0, 0.75)

		var exist bool
		for _, v := range set {
			if _, exist = s.Get(v); !exist {
				s.Set(v, struct{}{})
			}
		}
	}
}

func benchmarkMap(b *testing.B, set []keyType) {
	for i := 0; i < b.N; i++ {
		s := make(map[keyType]struct{})

		var exist bool
		for _, v := range set {
			if _, exist = s[v]; !exist {
				s[v] = struct{}{}
			}
		}
	}
}
