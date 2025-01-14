package main

import (
	_ "embed"
)

// Version is a version
//
//go:embed version.txt
var Version string
