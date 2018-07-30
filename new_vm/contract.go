package new_vm

type VersionCode string

type ContractInfo struct {
	Name    string
	Lang    string
	Version VersionCode
}

type Contract struct {
	ContractInfo
	Code string
}
