package fimap

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fastrand"
)

func TestArraySize(t *testing.T) {
	require.EqualValues(t, 2, arraySize(0, 12))
	require.EqualValues(t, math.MaxUint32, arraySize(math.MaxUint32, 1))
}

func TestNewFIMap(t *testing.T) {
	_, err := New(120, 1)
	require.NotNil(t, err)

	_, err = New(120, 0)
	require.NotNil(t, err)

	_, err = New(120, 2)
	require.NotNil(t, err)

	_, err = New(120, -1)
	require.NotNil(t, err)

	_, err = New(1, 0.9)
	require.Nil(t, err)

	_, err = New(0, 0.9)
	require.Nil(t, err)

	// Test basic ops
	s, err := New(1000, 0.5)
	require.Nil(t, err)

	s.Set(123, struct{}{})
	s.Set(456, 789)
	s.Set(1, uint64(128))

	v, ok := s.Get(123)
	require.True(t, ok)
	_, ok = v.(struct{})
	require.True(t, ok)

	v, ok = s.Get(456)
	require.True(t, ok)
	_v, ok := v.(int)
	require.True(t, ok)
	require.EqualValues(t, 789, _v)

	v, ok = s.Get(1)
	require.True(t, ok)
	__v, ok := v.(uint64)
	require.True(t, ok)
	require.EqualValues(t, 128, __v)
}

func TestFIMapOps(t *testing.T) {
	for i := 0; i < 5; i++ {
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
		_, ok := s.Get(x)
		require.False(t, ok)

		oldSize := s.size

		s.Remove(x)
		require.EqualValues(t, oldSize, s.size)
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
	require.EqualValues(t, len(m), s.Size())

	for _, k := range m {
		x, ok := s.Get(k)
		require.True(t, ok)
		require.EqualValues(t, k+1, x)

		_, ok = s.Get(k + 2000000000)
		require.False(t, ok)
	}

	// try to clone
	cl := s.Clone()
	require.EqualValues(t, s, cl)

	// try to remove on clone
	testRemoveOneByOne(t, cl, m)

	// iterate
	s.Set(0, uint64(0)) // set free key
	_ = s.Iterate(func(k uint64, v interface{}) error {
		if _v, ok := v.(uint64); !ok || (k != 0 && _v != k+1) || (k == 0 && _v != 0) {
			t.Log(_v, k)
			t.Fatal()
		}
		return nil
	})

	// iterate but stop with fake error
	fakeErr := fmt.Errorf("fake error")
	require.EqualValues(t, fakeErr, s.Iterate(func(k uint64, v interface{}) error { return fakeErr }))

	// reset
	oldLen := len(s.keys)
	s.Reset()
	require.EqualValues(t, oldLen, len(s.keys))
	require.EqualValues(t, oldLen, len(s.values))
	require.EqualValues(t, 0, s.size)
	require.Equal(t, nil, s.freeVal)
	require.False(t, s.hasFreeKey)
	for i := range s.keys {
		require.Equal(t, freeKey, s.keys[i])
		require.Equal(t, nil, s.values[i])
	}
}

func testRemoveOneByOne(t *testing.T, s *Map, m []uint64) {
	// remove all
	for _, k := range m {
		x, ok := s.Get(k)
		require.True(t, ok)
		require.EqualValues(t, k+1, x)

		oldSize := s.size
		s.Remove(k)
		require.EqualValues(t, oldSize-1, s.size)

		_, ok = s.Get(k)
		require.False(t, ok)
	}

	// check again
	for i := range s.keys {
		require.Equal(t, freeKey, s.keys[i])
		require.Equal(t, nil, s.values[i])
	}

	// check shrink happened?
	require.EqualValues(t, 2, len(s.keys))
}
