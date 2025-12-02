//go:build !go1.18

package io

// Provide `any` alias for Go < 1.18
type any = interface{}
