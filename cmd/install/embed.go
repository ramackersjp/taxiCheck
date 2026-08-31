//go:build windows

package main

import _ "embed"

// Payload files are copied by `make build-windows-installer` immediately
// before this package is compiled.

//go:embed payload/taxiprijs.exe
var payloadEXE []byte

//go:embed payload/env.example
var payloadEnv []byte

//go:embed payload/LICENSE
var payloadLicense []byte
