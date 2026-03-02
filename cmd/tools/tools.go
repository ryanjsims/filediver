//go:build tools

// see https://github.com/golang/go/issues/25922
package main

import (
	_ "golang.org/x/tools/cmd/stringer"
)
