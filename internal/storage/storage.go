package storage

import "errors"

var (
	ErrURLNotFound = errors.New("err not found")
	ErrURLExists   = errors.New("url exists")
)
