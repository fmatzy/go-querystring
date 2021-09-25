# go-querystring

This package is intended to replace `url.ParseQuery` and support` transform.Transformer`.

## Install

```
go get github.com/fmatzy/go-querystring
```

## Usage

```go
package main

import (
	"fmt"

	qs "github.com/fmatzy/go-querystring"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	// Form POST from non UTF-8 system
	s := "sjis=%3C%93%FA%96%7B%8C%EA+SJIS%3E"
	q, err := qs.Parse(s, japanese.ShiftJIS.NewDecoder())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", q)
}
```