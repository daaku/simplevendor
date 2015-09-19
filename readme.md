simplevendor
============

This is quite possibly the simplest vendoring tool possible. It:

- Assumes `$PWD/vendor` is the vendor directory.
- Goes over all the packages in the current directory.
- Finds all the transitive dependencies for the packages and their tests.
- Vendors those packages, excluding any non source files and test files.
- The exception to the above rule is readme and license files.

That's it. No manifest. No source modifications. No import rewriting.


Install
--

A go developer go gets:

```sh
go get -u github.com/daaku/simplevendor
```


Usage
--

Run `simplevendor` in your project directory:

```sh
cd $MY_PROJECT
simplevendor
```


Keeping-Up
--

When you want to upgrade your dependencies or vendor a new dependency:

```sh
cd $MY_PROJECT
rm -rf vendor
simplevendor
```


CI Keeping-Up
--

You may want to leverage your CI to help stay updated with upstream. Here's an
example if you use Travis CI:

https://travis-ci.org/daaku/rell
https://github.com/daaku/rell/blob/master/.travis.yml


Noisy Tests
--

Running tests may be noisy, look here for discussion:
https://github.com/golang/go/issues/11659

```sh
go test $(go list ./... | grep -v /vendor/)
```
