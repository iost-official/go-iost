package account

// Account type of a permission tree
type Account struct {
	ID          string                 `json:"id"`
	Referrer    string                 `json:"referrer"`
	Groups      map[string]*Group      `json:"groups"`
	Permissions map[string]*Permission `json:"permissions"`
}

// Item identity of a permission owner
type Item struct {
	ID         string `json:"id"` // key pair id
	Permission string `json:"permission"`
	IsKeyPair  bool   `json:"is_key_pair"`
	Weight     int    `json:"weight"`
}

// Group group of permissions
type Group struct {
	Name  string  `json:"name"`
	Items []*Item `json:"items"`
}

// Permission permission struct
type Permission struct {
	Name      string   `json:"name"`
	Groups    []string `json:"groups"`
	Items     []*Item  `json:"items"`
	Threshold int      `json:"threshold"`
}

// NewAccount a new empty account
func NewAccount(id string) *Account {
	return &Account{
		ID:          id,
		Groups:      make(map[string]*Group),
		Permissions: make(map[string]*Permission),
	}
}

// NewAccountFromKeys new account with owner and active key
func NewAccountFromKeys(id, ownerKey, activeKey string) *Account {
	a := &Account{
		ID:          id,
		Groups:      make(map[string]*Group),
		Permissions: make(map[string]*Permission),
	}
	a.Permissions["owner"] = &Permission{
		Name:      "owner",
		Threshold: 1,
		Items: []*Item{
			{
				ID:        ownerKey,
				IsKeyPair: true,
				Weight:    1,
			},
		},
	}
	a.Permissions["active"] = &Permission{
		Name:      "active",
		Threshold: 1,
		Items: []*Item{
			{
				ID:        activeKey,
				IsKeyPair: true,
				Weight:    1,
			},
		},
	}
	return a
}
