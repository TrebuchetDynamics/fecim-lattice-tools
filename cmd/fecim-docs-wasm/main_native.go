//go:build !js || !wasm

package main

import "fmt"

func main() { fmt.Println("Build fecim-docs-wasm with GOOS=js GOARCH=wasm") }
