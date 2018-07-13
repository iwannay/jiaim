package main

import (
	"errors"
)

var (
	// ring
	ErrorRingEmpty = errors.New("ring buffer empty")
	ErrorRingFull  = errors.New("ring buffer full")
	ErrorGroupDrop = errors.New("group has drop")
)
