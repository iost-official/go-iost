package pob

type ClientState int

const (
	Alive ClientState = iota
	Missing
	Error

)