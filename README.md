LCARS
===========
![Project status](http://img.shields.io/status/experimental.png?color=red)

LCARS (Library Access Retrieval System), a fast radix-tree based, HTTP router for Go.

Unique Features 
--------------
LCARS has the following features 

- 

Installation
-----------

Use go get 

```go
go get github.com/go-experimental/LCARS
``` 

or to update

```go
go get -u github.com/go-experimental/LCARS
``` 

Then import LCARS package into your code.

```go
import "github.com/go-experimental/LCARS"
``` 

Getting Started
----------------

Usage
------

Benchmarks
-----------
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.5.3 darwin/amd64


```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   lcars: 81016 Bytes

#GPlusAPI Routes: 13
   lcars: 6904 Bytes

#ParseAPI Routes: 26
   lcars: 7808 Bytes

#Static Routes: 157
   lcars: 79240 Bytes

PASS
BenchmarkLCARS_Param       	20000000	        87.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_Param5      	10000000	       144 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_Param20     	 5000000	       382 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_ParamWrite  	10000000	       168 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GithubStatic	20000000	       109 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GithubParam 	10000000	       151 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GithubAll   	   50000	     38100 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GPlusStatic 	20000000	        73.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GPlusParam  	20000000	       100 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GPlus2Params	10000000	       138 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_GPlusAll    	 1000000	      1838 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_ParseStatic 	20000000	        90.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_ParseParam  	20000000	       123 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_Parse2Params	10000000	       133 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_ParseAll    	  300000	      3902 ns/op	       0 B/op	       0 allocs/op
BenchmarkLCARS_StaticAll   	   50000	     24861 ns/op	       0 B/op	       0 allocs/op

```

This package is inspired by the following 
- Dimfeld/httpTreeMux
- julienschmidt/httprouter
- labstack/echo

License 
--------
This project is licensed unter MIT, for more information look into the LICENSE file.
Copyright (c) 2016 Go Playground


