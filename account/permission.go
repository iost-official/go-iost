package account

import "encoding/json"

type Account struct {
	ID          string                 `json:"id"`
	Groups      map[string]*Group      `json:"groups"`
	Permissions map[string]*Permission `json:"permissions"`
}
type User struct {
	ID         string `json:"id"` // key pair id
	Permission string `json:"permission"`
	IsKeyPair  bool   `json:"is_key_pair"`
	Weight     int    `json:"weight"`
}
type Group struct {
	Name  string  `json:"name"`
	Users []*User `json:"users"`
}
type Permission struct {
	Name      string   `json:"name"`
	Groups    []string `json:"groups"`
	Users     []*User  `json:"users"`
	Threshold int      `json:"threshold"`
}

func NewAccount(id string) *Account {
	return &Account{
		ID:          id,
		Groups:      make(map[string]*Group),
		Permissions: make(map[string]*Permission),
	}
}

func NewInitAccount(id, ownerKey, activeKey string) *Account {
	a := &Account{
		ID:          id,
		Groups:      make(map[string]*Group),
		Permissions: make(map[string]*Permission),
	}
	a.Permissions["owner"] = &Permission{
		Name:      "owner",
		Threshold: 1,
		Users: []*User{
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
		Users: []*User{
			{
				ID:        activeKey,
				IsKeyPair: true,
				Weight:    1,
			},
		},
	}
	return a
}

func (a *Account) Encode() []byte {
	buf, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return buf
}
