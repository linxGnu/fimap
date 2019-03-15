package fimap

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/valyala/fastrand"
)

func TestArraySize(t *testing.T) {
	if arraySize(0, 12) != 2 {
		t.Fatal()
	}

	if arraySize(math.MaxUint32, 1) != math.MaxUint32 {
		t.Fatal()
	}
}

func TestNewFIMap(t *testing.T) {
	if _, err := New(120, 1); err == nil {
		t.Fatal()
	}
	if _, err := New(120, 0); err == nil {
		t.Fatal()
	}
	if _, err := New(120, 2); err == nil {
		t.Fatal()
	}
	if _, err := New(120, -1); err == nil {
		t.Fatal()
	}
	if _, err := New(0, 0.9); err != nil {
		t.Fatal()
	}
	if _, err := New(1, 0.9); err != nil {
		t.Fatal()
	}

	// Test basic ops
	s, _ := New(1000, 0.5)
	s.Set(123, struct{}{})
	s.Set(456, 789)
	s.Set(1, uint64(128))

	if v, ok := s.Get(123); !ok {
		t.Fatal()
	} else if _, ok = v.(struct{}); !ok {
		t.Fatal(reflect.TypeOf(v))
	}

	if v, ok := s.Get(456); !ok {
		t.Fatal()
	} else if _v, ok := v.(int); !ok || _v != 789 {
		t.Fatal(reflect.TypeOf(v))
	}

	if v, ok := s.Get(1); !ok {
		t.Fatal()
	} else if _v, ok := v.(uint64); !ok || _v != 128 {
		t.Fatal(reflect.TypeOf(v))
	}
}

func TestFIMapOps(t *testing.T) {
	for i := 0; i < 10; i++ {
		m := initTestData()
		testFIMap(t, m)
	}
}

func initTestData() (r []uint64) {
	m := make(map[uint64]struct{})
	for i := 0; i < 500000; i++ {
		m[124+uint64(fastrand.Uint32n(2000000000))] = struct{}{}
	}
	m[0] = struct{}{}

	r = make([]uint64, 0, len(m))
	for k := range m {
		r = append(r, k)
	}

	return
}

func testFIMap(t *testing.T, m []uint64) {
	s, _ := New(1000, 0.5)

	removeNotExist := func(x keyType) {
		if _, ok := s.Get(x); ok {
			t.Fatal()
		}

		oldSize := s.size

		s.Remove(x)
		if s.size != oldSize {
			t.Fatal()
		}
	}

	removeNotExist(0)
	removeNotExist(122)
	removeNotExist(123)

	// try to put
	for _, k := range m {
		s.Set(k, k)
		s.Set(k, k+1) // duplicate put
	}

	// check size
	if s.Size() != len(m) {
		t.Fatal()
	}

	for _, k := range m {
		if x, ok := s.Get(k); !ok || x != k+1 {
			t.Fatal(k)
		}

		if _, ok := s.Get(k + 2000000000); ok {
			t.Fatal(k + 2000000000)
		}
	}

	// try to clone
	cl := s.Clone()
	if !reflect.DeepEqual(cl, s) {
		t.Fatal()
	}

	// try to remove on clone
	testRemoveOneByOne(t, cl, m)

	// iterate
	s.Set(0, uint64(0)) // set free key
	s.Iterate(func(k uint64, v interface{}) error {
		if _v, ok := v.(uint64); !ok || (k != 0 && _v != k+1) || (k == 0 && _v != 0) {
			t.Log(_v, k)
			t.Fatal()
		}
		return nil
	})

	// iterate but stop with fake error
	fakeErr := fmt.Errorf("fake error")
	if s.Iterate(func(k uint64, v interface{}) error { return fakeErr }) != fakeErr {
		t.Fatal()
	}

	// reset
	oldLen := len(s.keys)
	if s.Reset(); len(s.keys) != oldLen || len(s.values) != oldLen || s.size != 0 || s.hasFreeKey || s.freeVal != nil {
		t.Fatal()
	} else {
		for i := range s.keys {
			if s.keys[i] != freeKey || s.values[i] != nil {
				t.Fatal()
			}
		}
	}
}

func testRemoveOneByOne(t *testing.T, s *Map, m []uint64) {
	// remove all on clone
	for _, k := range m {
		if x, ok := s.Get(k); !ok || x != k+1 {
			t.Fatal(k)
		}

		oldSize := s.size
		s.Remove(k)
		if s.size != oldSize-1 {
			t.Fatal()
		}

		if _, ok := s.Get(k); ok {
			t.Fatal(k)
		}
	}

	// check again on clone
	for i := range s.keys {
		if s.keys[i] != freeKey || s.values[i] != nilValue {
			t.Fatal()
		}
	}

	if len(s.keys) != 2 {
		t.Fatal()
	}
}
