# Jsonco (commented json)

`jsonco` is an [io.Reader](http://golang.org/pkg/io/#Reader) for JSON which strips C-style comments and trailing commas,
allowing use of JSON as a *reasonable* config format. It also has a utility method which translates byte offset into the stream
into line and character positions for easier error interpretation. The package is aware of multibyte characters.

Single line comments start with `//` and continue to the end of the line. Multiline comments are enclosed in `/*` and `*/`.
If a trailing comma is in front of `]` or `}` it is stripped as well.

This implementation is used by https://github.com/tinode/chat and as such it's up to date and supported.


## Examples

Given `settings.json`

```js
{
	"key": "value", // k:v

	// a list of numbers
	"list": [1, 2, 3],

	/* 
	a list of numbers
	which are important
	*/
	"numbers": [1, 2, 3],
}
```

You can read it in as a *normal* json file:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tinode/jsonco"
)

func main() {
	var v interface{}
	f, _ := os.Open("settings.json")
	// Wrap the reader before passing it to the json decoder.
	jr := jsonco.New(f)
	json.NewDecoder(jr).Decode(&v)
	fmt.Println(v)
}
```

If a parsing error is encountered, its location can be found like this:

```go
if err := json.NewDecoder(jr).Decode(&v); err != nil {
	switch jerr := err.(type) {
	case *json.UnmarshalTypeError:
		lnum, cnum, _ := jr.LineAndChar(file, jerr.Offset)
		fmt.Fatalf("Unmarshall error in %s at %d:%d (offset %d bytes): %s",
			jerr.Field, lnum, cnum, jerr.Offset, jerr.Error())
	case *json.SyntaxError:
		lnum, cnum, _ := jr.LineAndChar(file, jerr.Offset)
		fmt.Fatalf("Syntax error at %d:%d (offset %d bytes): %s",
			lnum, cnum, jerr.Offset, jerr.Error())
	default:
		fmt.Fatalln("Failed to parse: ", err)
	}
}

```

## Godoc

https://pkg.go.dev/github.com/tinode/jsonco?tab=doc

## License

MIT

## References

This code is forked from https://github.com/DisposaBoy/JsonConfigReader with added go.mod, offset translation
and multibyte character support.
