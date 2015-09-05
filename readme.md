simplevendor
============

This is quite possibly the simplest vendoring tool possible. It:

- Assumes `$PWD/vendor` is the vendor directory.
- Goes over all the packages in the current directory.
- Finds all the transitive dependencies for the packages and their tests.
- Vendors those packages, excluding any non source files as well as not vendoring their tests.
- The exception to the above rule is readme and license files.

That's it. No manifest. No source modifications. No import rewriting.
