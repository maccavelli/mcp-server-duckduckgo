package main

import "strings"

// RawVersion is the build-time release identity.
//
// The Makefile has always passed -X main.RawVersion, but no such variable
// existed, so the linker silently did nothing and this server had no version
// identity at all. Defining it here is what makes the stamp effective; "dev"
// is the default so an unstamped binary is never mistaken for a release.
var RawVersion = "dev"

// RawBuildKind is "release" only for a tag build. A bool cannot be set with
// the Go linker's -X flag, so this is a string and only that exact value
// counts; anything else is a local build that update refuses to replace
// without --force.
var RawBuildKind = "local"

// Version is the trimmed display value only. Ordering uses RawVersion.
var Version = strings.TrimPrefix(RawVersion, "v")
