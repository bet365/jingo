# Jingo

This package provides the ability to encode golang structs to a buffer as JSON. 

The main take-aways are

* It's very fast. *(We can't find a faster one)*
* Very low allocs, 0 in a lot of cases.
* Clear API - similar to the stdlib. It just uses struct tags.
* No other library dependencies.
* It doesn't require a build step, like `go generate`. 

## Another JSON Library...why?

Performance. Check out these numbers - they were generated with the gojay (which is fast) perf data, SmallPayload and LargePayload respectively.

> Numbers were generated on a Centos 7 Machine, Quad-core Intel(R) Core(TM) i5-4590 CPU @ 3.30GHz. 

|         Lib          |   Iter   | ns/op | B/op | allocs/op | +/-  |
| -------------------- | -------- | ----- | ---- | --------- | ---- |
| jingo                | 10000000 |   208 |    0 |         0 | 4.8x |
| stdlib encoding/json |  1000000 |  1008 |  160 |         1 | 1x   |
| gojay                |  2000000 |   605 |  512 |         1 | 1.6x |
| json-iterator        |  2000000 |   825 |  168 |         2 | 1.2x |

|         Lib          |  Iter  | ns/op |  B/op | allocs/op | +/-  |
| -------------------- | ------ | ----- | ----- | --------- | ---- |
| jingo                | 200000 |  9748 |     0 |         0 | 3x   |
| stdlib encoding/json |  50000 | 29854 |  4866 |         1 | 1x   |
| gojay                | 100000 | 16884 | 18308 |         5 | 1.7x |
| json-iterator        | 100000 | 21033 |  4873 |         2 | 1.4x |

These results can be even more pronounced depending on the shape of the struct - these results are based on a struct with a lot of string data in:

|         Lib          |   Iter   | ns/op | B/op | allocs/op |  +/-  |
| -------------------- | -------- | ----- | ---- | --------- | ----- |
| jingo                | 10000000 |   212 |    0 |         0 | 11.5x |
| stdlib encoding/json |   500000 |  2443 |  720 |         4 | 1x    |
| gojay                |  1000000 |  1147 |  512 |         1 | 2.1x  |
| json-iterator        |   500000 |  2606 |  744 |         5 | 0.9x  |

## Example Usage

The usage is similar to that of the stdlib `json.Marshal`, but we do most of our work upon instantiation of the encoders.

The encoders you have available to you are 

* `jingo.StructEncoder`
* `jingo.SliceEncoder`

They both reference each other and they work in exactly the same way. You'll see, like the stdlib `encode/json`, there is very little wire-up involved. 

```go
package main

import (
    "fmt"
    "github.com/bet365/jingo"
)

// sample struct we'll encode
type MyPayload struct {
    Name string `json:"name"`
    Age int     `json:"age"`
    ID int      // anything we don't annotate doesn't get emitted. 
}

// Create an encoder, letting it know which type of struct we're going to be encoding. 
// You only do this once per type!
var enc = jingo.NewStructEncoder(MyPayload{})

func main() {
    // now lets encode something 
    p := MyPayload{
        Name: "Mr Payload",
        Age: 33,
    }

    // pull a buffer from the pool and pass it along with the struct to Marshal
    buf := jingo.NewBufferFromPool()
    enc.Marshal(&p, buf)

    fmt.Println(buf.String()) // {"name":"Mr Payload","age":33}

    // return the buffer to the pool now we're done
    buf.ReturnToPool()
}

```

## Buffer

Buffer is a simple custom buffer type which complies with `io.Writer`. Its main benefit being it has pooling built-in. This goes a long way to helping make jingo fast by reducing its allocations and ensuring good write speeds.

## Options

There are a couple of subtle ways you can configure the encoders. 

* You can specify a default capacity for buffer using `NewBufferFromPoolWithCap(int)*Buffer`
* It supports the same `json:"tag,options"` syntax as the stdlib, but not the same options. Currently the options you have are
    - `,stringer`, which instead of the standard serialization method for a given type, nominates that its `.String()` function is invoked instead to provide the serialization value.
    - `,raw`, which allows byteslice-like items (like `[]byte` and `string`) to be written to the buffer directly with no conversion, quoting or otherwise. `nil` or empty fields annotated as `raw` will output `null`. 
    - `,encoder` which instead of the standard serialization method for a given type, nominates that its `.JSONEncode(*jingo.Buffer)` function or `EncodeJSON(io.Writer)` function are invoked instead. From there you can manually write to the buffer or writer for that particular field. There are a choice of 2 interfaces you need to comply with depending on your use case, either `jingo.JSONEncoder` (which introduces a dependency on `Buffer`), or `jingo.JSONMarshaler` which allows writing directly to an `io.Writer`.
    - `,escape`, which safely escapes `"`,`\`, line feed (`\n`), carriage return (`\r`) and tab (`\t`) characters to valid JSON whilst writing. To get the same functionality when using `SliceEncoder` on its own, use `jingo.EscapeString` to initialize the encoder - e.g `NewSliceEncoder([]jingo.EscapeString)` - instead of `string` directly. There is obviously a performance impact on the write speed using this option, the benchmarks show it takes twice the time of a standard string write, so whilst it is still faster than using the stdlib, to get the best performance it is recommended to only be used when needed and only then when the escaping work can't be done up-front.
    - `,omitempty`, this will omit the field from being output excluding structs. This has a performance impact and should be used when required. This is still more performant than the stdlib.


## How does it work

When you create an instance of an encoder it recursively generates an instruction set which defines how to iteratively encode your structs. This gives it the ability to provide a clear API but with the same benefits as a build-time optimized encoder. It's almost exclusively able to do all type assertions and reflection activity during the compile, then makes ample use of the `unsafe` package during the instruction-set execution (the `Marshal` call) to make reading and writing very fast. 

As part of the instruction set compilation it also generates static meta-data, i.e field names, brackets, braces etc. These are then chunked into instructions on demand.

## Drawbacks?

The package is designed to be performant and as such it is not 100% functionally compatible with stdlib. Specifically. 

* The `,string` tag option isn't supported, only strings are quoted by default - use `,stringer` instead to achieve the same results.  This may be added in future releases. 
* Maps are currently not supported. Initial thoughts were given that this is a performance focused library it doesn't make much sense to iterate maps and would advise against doing so for performance sensitive applications - **however - maps are being added**!

## Contribution Guidelines

Contributions are welcome! Fork the repo and submit a pull request to get your change added. 

Please take into consideration whether or not the change aligns with the agenda of the project to avoid having them rejected. For example, when adding a new feature, try to make sure you're creating a new instruction/set for the feature being added - don't add logic to existing instructions at the cost of performance for all other code paths currently using them.  It's best to have more instructions with no logic than fewer instructions with a few conditionals that execute at runtime.  

Feel free to raise an issue here beforehand to discuss anything with others before your implementation. 