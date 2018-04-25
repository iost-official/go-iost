package vm

type Contract struct {
	name  string

	main  Method
	apis  []Method

	owner Signature
	signs []Signature
}

