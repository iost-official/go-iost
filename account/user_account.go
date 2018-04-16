package account

const (
  UserAccount = iota
  SuperAccount
)

type Account struct {
  priority int
  addr Address
  // todo: contract_code, storage_db, current_state
}
