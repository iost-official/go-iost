package native

var tokenABIsV4 *abiSet

func init() {
	tokenABIsV4 = newAbiSet()
	tokenABIsV4.Register(initTokenABI, true)
	tokenABIsV4.Register(createTokenABI)
	tokenABIsV4.Register(balanceOfTokenABI)
	tokenABIsV4.Register(supplyTokenABI)
	tokenABIsV4.Register(totalSupplyTokenABI)

	// modified methods for V2
	tokenABIsV4.Register(issueTokenABIV2)
	tokenABIsV4.Register(transferFreezeTokenABIV2)
	tokenABIsV4.Register(destroyTokenABIV2)

	// modified methods for V3
	tokenABIsV4.Register(transferTokenABIV3)

	// modified methods for V4
	tokenABIsV4.Register(updateTokenTotalSupplyABI)
}
