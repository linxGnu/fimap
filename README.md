# fimap

[![Build Status](https://travis-ci.org/linxGnu/fimap.svg?branch=master)](https://travis-ci.org/linxGnu/fimap)
[![Go Report Card](https://goreportcard.com/badge/github.com/linxGnu/fimap)](https://goreportcard.com/report/github.com/linxGnu/fimap)
[![Coverage Status](https://coveralls.io/repos/github/linxGnu/fimap/badge.svg?branch=master)](https://coveralls.io/github/linxGnu/fimap?branch=master)
[![godoc](https://img.shields.io/badge/docs-GoDoc-green.svg)](https://godoc.org/github.com/linxGnu/fimap)

A fast map that uses `uint64` for the index key and `interface{}` for the data, based on http://java-performance.info/implementing-world-fastest-java-int-to-int-hash-map/

It is 1.5-2X faster than the builtin map.

### Installing

To start using `fimap`, install Go and run `go get`:

```
go get -u github.com/linxGnu/fimap
```

### Example

```go
import "github.com/linxGnu/fimap"

// 1000: expect capacity
// 0.5: threshold. When cardinality >= threshold * capacity
// the map would grow 2x
s, _ := fimap.New(1000, 0.5) 

// set
s.Set(123, struct{}{})
s.Set(345, 128)

// get
v, exist := s.Get(123)

// get number of elements in map
s.Size()

// create new map and clone from orginal
c := s.Clone()

// iterate
s.Iterate(func(key uint64, value interface{}) error {
  // do some thing with key value
  return nil // return non nil error if you want to stop iteration
})
```

### Benchmark
```scala
system_profiler SPHardwareDataType

Hardware:

    Hardware Overview:

      Model Name: MacBook Pro
      Model Identifier: MacBookPro14,3
      Processor Name: Intel Core i7
      Processor Speed: 2.8 GHz
      Number of Processors: 1
      Total Number of Cores: 4
      L2 Cache (per Core): 256 KB
      L3 Cache: 6 MB
      Memory: 16 GB
      Boot ROM Version: 185.0.0.0.0
      SMC Version (system): 2.45f0
```

```scala
goversion: 1.12

goos: darwin
goarch: amd64
pkg: github.com/linxGnu/fimap
BenchmarkFIMapSmall-8    	   30000	     47060 ns/op	   98368 B/op	      23 allocs/op
BenchmarkFIMapMedium-8   	       5	 267164702 ns/op	201326656 B/op	      45 allocs/op
BenchmarkFIMapLarge-8    	       1	1665465441 ns/op	805306432 B/op	      49 allocs/op
BenchmarkMapSmall-8      	   20000	     79285 ns/op	   47814 B/op	      65 allocs/op
BenchmarkMapMedium-8     	       3	 383101179 ns/op	99880053 B/op	   76588 allocs/op
BenchmarkMapLarge-8      	       1	2242347725 ns/op	404004624 B/op	  307234 allocs/op
```
