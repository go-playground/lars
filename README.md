# LARS
![Project status](http://img.shields.io/status/experimental.png?color=red)

LARS - Library Access Retrieval System

#### Unique Features 



#### Usage

Use go get 

```go
go get github.com/go-experimental/lars
``` 

or to update

```go
go get -u github.com/go-experimental/lars
``` 

Then import LARS package into your code.

```go
import "github.com/go-experimental/lars"
``` 


#### Benchmarks
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.5.3 darwin/amd64


```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   LARS: 81016 Bytes

#GPlusAPI Routes: 13
   LARS: 6904 Bytes

#ParseAPI Routes: 26
   LARS: 7808 Bytes

#Static Routes: 157
   LARS: 79240 Bytes

PASS
BenchmarkLARS_Param       	20000000	        87.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param5      	10000000	       144 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param20     	 5000000	       382 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParamWrite  	10000000	       168 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubStatic	20000000	       109 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubParam 	10000000	       151 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubAll   	   50000	     38100 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusStatic 	20000000	        73.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusParam  	20000000	       100 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlus2Params	10000000	       138 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusAll    	 1000000	      1838 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseStatic 	20000000	        90.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseParam  	20000000	       123 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Parse2Params	10000000	       133 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseAll    	  300000	      3902 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_StaticAll   	   50000	     24861 ns/op	       0 B/op	       0 allocs/op


### License 
This project is licensed unter MIT, for more information look into the LICENSE file.
Copyright (c) 2016 Go Playground


