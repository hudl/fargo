# fargo

Netflix Eureka client in golang. Named for the show Eureka.

```go
c = fargo.NewConn("http", "127.0.0.1", "8080")
c.GetApps() // returns a map[String]fargo.Application
```

# Hacking

## Just Let Me Import Already

`go get github.com/hudl/fargo`

```go

package main

import (
    "github.com/hudl/fargo"
)

func main() {
    e := fargo.NewConn()
}

```

## Adding Stuff

`go test` is your friend. I use the excellent [gocheck](launchpad.net/gocheck)
library in addition to the standard lib's `testing` package. If you add
something, write a test. If something is broken, write a test that reproduces
your issue or post repro steps to the issues on this repository. The tests
require that you have a eureka install and are designed to run against the
included [vagrant](http://vagrantup.com) specs.

## Using Vagrant

The Vagrantfile in this repo will set up a two-server eureka cluster using the
OpsCode bento boxes. By default, the VMs are named node1.localdomain and
node2.localdomain. They'll be automatically provisioned so long as you've
already configured/built eureka.

To build eureka `git clone https://github.com/Netflix/eureka.git && cd eureka
&& git checkout 1.1.118 && ./gradlew clean build`

# MIT License

```
The MIT License (MIT)

Copyright (c) 2013 Ryan S. Brown <sb@ryansb.com>
Copyright (c) 2013 Hudl <@Hudl>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
```
