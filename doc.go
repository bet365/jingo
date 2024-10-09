/*
Package jingo provides the ability to encode go structs to a buffer as JSON.

The main features are
It's very fast.
Very low allocs, 0 in a lot of cases.
Clear API - similar to the stdlib. It just uses struct tags.
No other library dependencies.
It doesn't require a build step, like `go generate`.

You only need to create an instance of an encoder once per struct/slice type
you wish to marshal.
*/
package jingo
