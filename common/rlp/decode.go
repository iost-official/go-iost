package rlp

import "errors"

// 表示存储在RLP流中值的类型
type Kind int

// Byte是0，String是1，List是2
const (
	Byte Kind = iota
	String
	List
)

var (
	// EOL is returned when the end of the current list
	// has been reached during streaming.
	EOL = errors.New("rlp: end of list")

	// Actual Errors
	ErrExpectedString = errors.New("rlp: expected String or Byte")
	ErrExpectedList   = errors.New("rlp: expected List")
	ErrCanonInt       = errors.New("rlp: non-canonical integer format")
	ErrCanonSize      = errors.New("rlp: non-canonical size information")
	ErrElemTooLarge   = errors.New("rlp: element is larger than containing list")
	ErrValueTooLarge  = errors.New("rlp: value size exceeds available input length")

	// This error is reported by DecodeBytes if the slice contains
	// additional data after the first RLP value.
	ErrMoreThanOneValue = errors.New("rlp: input contains more than one value")

	// internal errors
	errNotInList    = errors.New("rlp: call of ListEnd outside of any list")
	errNotAtEOL     = errors.New("rlp: call of ListEnd not positioned at EOL")
	errUintOverflow = errors.New("rlp: uint overflow")
)
