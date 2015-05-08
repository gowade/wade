# Wade.Go
Wade.Go is an upcoming brand new way to develop web sites and applications.
It's a *client-centric* web development library, but NOT for Javascript!

Isomorphic Javascript is cool but what could be better than that? **Isomorphic Go**.  
Advantages:
* Isomorphism: Write ui/client once, in Go and HTML, render seemlessly on both client and server (no SEO problems). Go code is transpiled to Javascript on browser.
* Pleasure: Modern **React**-like development model, in Go (strict types ftw!).
* Maintainability: No more maintainability headache like with Javascript, and we could *go easy* on tests.
It helps tremendously to have strict typing and a nice compiler, especially for large projects.
* Convenience: Easy collaboration between client and server since they use the same great programming language.

# Development Status
* Mar 12, 2015: Iteration 5 starts.
* May 03, 2015: Core rendering and template/component functionalities working. Still early stage, not yet have end-to-end tests for the DOM diff engine.

# Run the test app
Make sure you have a working Go installation and [Gopherjs](https://github.com/gopherjs/gopherjs), then

1. `go get -u github.com/gowade/wade`
2. Install `fuel` the code generator: `go install github.com/gowade/wade/fuel`
3. Go to "browser_tests/worklog/main", run `fuel build`, then run `./run_gopherjs`
4. Use browser to open the file `browser_tests/worklog/main/public/index.html`

# LICENSE
Wade.Go is [BSD licensed](https://github.com/gowade/wade/blob/master/LICENSE)
