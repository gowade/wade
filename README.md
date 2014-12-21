
>>> December 21, 2014: Wade.Go i3 on react-esque virtual DOM technique is underway (on the "vdom" branch). The internal API rewrite and code migration is done, it now passed the *wadereddi* functional test. The remaining task is Javascript adaptation, mainly integrating with a Javascript virtual DOM rendering library.

# Wade.Go
**Wade.Go** is a **client-centric** web framework for Go, it's like nothing you ever heard. It brings these awesome things together
* Compiled, statically typed programming with Go (compiled to Javascript on client side by [gopherjs](https://github.com/gopherjs/gopherjs))
* Client-centric web development model with HTML data binding (think AngularJs)
* Hybrid rendering: Write code once, render on both client and server (think server-side ReactJS)

With the creation of Wade.Go, you can now write both client side and server side in a single programming language that is not Javascript. Go brings the best concurrency pattern (goroutines) and the static type system that makes maintenance a breeze.

Although being client-centric in approach, Wade.Go is independent and designed to work perfectly when Javascript is not available, the ability to render the site on the server is built in. It even has a **functional test** API that runs the app in native Go with `go test` (no browser).

Wade.Go is made for web *sites*, not just *apps*, it can help build content-heavy sites like blog, forum as well as dynamic web applications.

# Getting started
* [Tutorial](https://github.com/phaikawl/wade/wiki/Wade.Go-Quick-Start-Guide) (Outdated)
* [godoc API reference](http://godoc.org/github.com/phaikawl/wade) (Outdated)
* [Wadereddi](https://github.com/phaikawl/wadereddi) the demo app (Test only)

# License

Copyright (c) 2014, Hai Thanh Nguyen
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.





