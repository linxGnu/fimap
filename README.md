# fimap

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
BenchmarkMapSmall-8      	   20000	     81113 ns/op	   47818 B/op	      66 allocs/op
BenchmarkMapMedium-8     	       5	 327177149 ns/op	99879297 B/op	   76594 allocs/op
BenchmarkMapLarge-8      	       1	1853012017 ns/op	403992160 B/op	  306862 allocs/op
BenchmarkFIMapSmall-8    	   30000	     47431 ns/op	   98032 B/op	      17 allocs/op
BenchmarkFIMapMedium-8   	       5	 257082510 ns/op	201326320 B/op	      39 allocs/op
BenchmarkFIMapLarge-8    	       1	1369296578 ns/op	805306096 B/op	      43 allocs/op
```
