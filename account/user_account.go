package account

type UserAccount struct {
  addr UserAddress
  nonce uint32
  balance uint64
  //current_utxo UTXOPool
}
