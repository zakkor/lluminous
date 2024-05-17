//go:build release

package main

import "embed"

//go:embed dist-client/*
//go:embed dist-client/**/*
var staticFiles embed.FS
var embedStaticFiles = true
