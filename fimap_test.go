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
		testFIMap(m, t)
	}
}

func initTestData() (m map[uint64]struct{}) {
	m = make(map[uint64]struct{})
	for i := 0; i < 500000; i++ {
		m[1+uint64(fastrand.Uint32n(2000000000))] = struct{}{}
	}
	m[0] = struct{}{}
	return
}

func testFIMap(m map[uint64]struct{}, t *testing.T) {
	s, _ := New(1000, 0.5)

	if _, ok := s.Get(0); ok {
		t.Fatal()
	}

	if _, ok := s.Get(123); ok {
		t.Fatal()
	}

	for k := range m {
		s.Set(k, k) // try duplicate put
		s.Set(k, k)
	}

	for k := range m {
		if _, ok := s.Get(k); !ok {
			t.Fatal()
		}
		if _, ok := s.Get(k + 2000000000); ok {
			t.Fatal(k + 2000000000)
		}
	}

	if s.Size() != len(m) {
		t.Fatal()
	}

	if cl := s.Clone(); !reflect.DeepEqual(cl, s) {
		t.Fatal()
	}

	// try iterate
	s.Set(0, uint64(0)) // set free key
	s.Iterate(func(k uint64, v interface{}) error {
		if _v, ok := v.(uint64); !ok || _v != k {
			t.Fatal()
		}
		return nil
	})

	// try iterate but stop with fake error
	fakeErr := fmt.Errorf("fake error")
	if s.Iterate(func(k uint64, v interface{}) error { return fakeErr }) != fakeErr {
		t.Fatal()
	}
}
