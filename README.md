##LARS
<img align="right" src="https://raw.githubusercontent.com/go-playground/lars/master/examples/README/test.gif">
![Project status](https://img.shields.io/badge/version-2.2-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/projects/4351aa2d-2f94-40be-a6ef-85c248490378/679708/badge.svg)](https://semaphoreci.com/joeybloggs/lars)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/lars/badge.svg?branch=master)](https://coveralls.io/github/go-playground/lars?branch=master)
[![Go Report Card](https://goreportcard.com/badge/go-playground/lars)](https://goreportcard.com/report/go-playground/lars)
[![GoDoc](https://godoc.org/github.com/go-playground/lars?status.svg)](https://godoc.org/github.com/go-playground/lars)
![License](https://img.shields.io/dub/l/vibe-d.svg)

LARS is a fast radix-tree based, zero allocation, HTTP router for Go.  [ view examples](https://github.com/go-playground/lars/tree/master/examples)

Why Another HTTP Router?
------------------------
Have you ever been painted into a corner by a framework, **ya me too!** and I've noticed that allot of routers out there, IMHO, are adding so much functionality that they are turning into Web Frameworks, (which is fine, frameworks are important) however, not at the expense of flexibility and configurability. So with no further ado, introducing LARS an HTTP router that can be your launching pad in creating a framework for your needs. How? Context is an interface [see example here](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go), where you can add as little or much as you want or need and most importantly...**under your control**.

Key & Unique Features 
--------------
- [x] Context is an interface allowing passing of framework/globals/application specific variables. [example](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go)
- [x] Contains helpful logic to help prevent adding bad routes, keeping your url's consistent. i.e. /user/:id and /user/:user_id - the second one will fail to add letting you know that :user_id should be :id
- [x] Has an uber simple middleware + handler definitions!!! middleware and handlers actually have the exact same definition!
- [x] Can register custom handlers for making other middleware + handler patterns usable with this router; the best part about this is can register one for your custom context and not have to do type casting everywhere [see here](https://github.com/go-playground/lars/blob/master/examples/custom-handler/main.go)
- [x] Full support for standard/native http Handler + HandlerFunc + some others [see here](https://github.com/go-playground/lars/blob/master/examples/native/main.go)
  * When Parsing a form call Context's ParseForm amd ParseMulipartForm functions and the URL params will be added into the Form object, just like query parameters are, so no extra work
- [x] lars uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter) so incredibly fast and efficient.


**Note:** Since this router has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns /user/new and /user/:user for the same request method at the same time. The routing of different request methods is independent from each other. I was initially against this, and this router allowed it in a previous version, however it nearly cost me in a big app where the dynamic param value say :type actually could have matched another static route and that's just too dangerous, so it is no longer allowed.

Installation
-----------

Use go get 

```go
go get github.com/go-playground/lars
```

Usage
------
Below is a simple example, for a full example [see here](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go)
```go
package main

import (
	"net/http"

	"github.com/go-playground/lars"
	mw "github.com/go-playground/lars/examples/middleware/logging-recovery"
)

func main() {

	l := lars.New()
	l.Use(mw.LoggingAndRecovery) // LoggingAndRecovery is just an example copy paste and modify to your needs

	l.Get("/", HelloWorld)

	http.ListenAndServe(":3007", l.Serve())
}

// HelloWorld ...
func HelloWorld(c lars.Context) {
	c.Response().Write([]byte("Hello World"))
}
```

Middleware
-----------
There are some pre-defined middlewares within the middleware folder; NOTE: that the middleware inside will
comply with the following rule(s):

* Are completely reusable by the community without modification

Other middleware will be listed under the examples/middleware/... folder for a quick copy/paste modify. as an example a logging or
recovery middleware are very application dependent and therefore will be listed under the examples/middleware/...

Benchmarks
-----------
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.6 darwin/amd64

NOTICE: lars uses a custom version of [httprouter](https://github.com/julienschmidt/httprouter), benchmarks can be found [here](https://github.com/joeybloggs/go-http-routing-benchmark/tree/lars-only)

```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   LARS: 49016 Bytes

#GPlusAPI Routes: 13
   LARS: 3624 Bytes

#ParseAPI Routes: 26
   LARS: 6616 Bytes

#Static Routes: 157
   LARS: 30104 Bytes

PASS
BenchmarkLARS_Param       	20000000	        75.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param5      	10000000	       126 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param20     	 5000000	       311 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParamWrite  	10000000	       144 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubStatic	20000000	       101 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubParam 	10000000	       154 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubAll   	   50000	     33295 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusStatic 	20000000	        72.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusParam  	20000000	        99.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlus2Params	10000000	       124 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusAll    	 1000000	      1640 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseStatic 	20000000	        73.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseParam  	20000000	        79.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Parse2Params	20000000	        97.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseAll    	  500000	      2974 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_StaticAll   	   50000	     23641 ns/op	       0 B/op	       0 allocs/op
```

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

This package is inspired by the following 
-----------
- [httptreemux](https://github.com/dimfeld/httptreemux)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [echo](https://github.com/labstack/echo)
- [gin](https://github.com/gin-gonic/gin)

Licenses
--------
[MIT License](https://raw.githubusercontent.com/go-playground/lars/master/LICENSE) (MIT), Copyright (c) 2015 Dean Karn
[BSD License](https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE), Copyright (c) 2013 Julien Schmidt. All rights reserved.
