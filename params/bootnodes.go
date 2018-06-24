package params

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	//"84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@18.219.254.124:30304",
}

//todo add ip white list
var CommitteeNodes = []string{
	"13.232.79.76",
}

var WitnessNodes = map[string]bool{
	"13.56.255.143:30301": true,
}

var SpNodes = map[string]bool{
	"54.183.115.79:30301": true,
	"13.56.223.196:30301": true,
	//"explorer": true,
}
