//go:build !release

package main

import "io/fs"

var staticFiles fs.FS = nil
var embedStaticFiles = false
