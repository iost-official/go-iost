package host

import (
	"encoding/json"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/vm/database"
)

func TestRequireAuth_ByKey(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"keya": 1})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["pa"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Users: []*account.Item{
				{
					ID:         "keya",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
				{
					ID:         "b",
					Permission: "active",
					IsKeyPair:  false,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		j, err := json.Marshal(ac)
		if err != nil {
			t.Fatal(err)
		}
		return database.MustMarshal(string(j)), nil
	})

	ans, cost := host.RequireAuth("a", "pa")
	if !ans {
		t.Fatal(ans)
	}
	if cost.ToGas() == 0 {
		t.Fatal(cost)
	}
}

func TestAuthority_ByUser(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"keyb": 1})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["pa"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Users: []*account.Item{
				{
					ID:         "keya",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
				{
					ID:         "b",
					Permission: "pb",
					IsKeyPair:  false,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		j, err := json.Marshal(ac)
		if err != nil {
			t.Fatal(err)
		}
		return database.MustMarshal(string(j)), nil
	})
	db.EXPECT().Get("state", "m-auth.iost-auth-b").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("b")
		ac.Permissions["active"] = &account.Permission{
			Name:   "active",
			Groups: []string{},
			Users: []*account.Item{
				{
					ID:         "keyb",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		j, err := json.Marshal(ac)
		if err != nil {
			t.Fatal(err)
		}
		return database.MustMarshal(string(j)), nil
	})

	ans, cost := host.RequireAuth("a", "pa")
	if !ans {
		t.Fatal(ans)
	}
	if cost.ToGas() == 0 {
		t.Fatal(cost)
	}
}
func TestAuthority_Active(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"keya": 1})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["active"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Users: []*account.Item{
				{
					ID:         "keya",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		j, err := json.Marshal(ac)
		if err != nil {
			t.Fatal(err)
		}
		return database.MustMarshal(string(j)), nil
	})
	ans, cost := host.RequireAuth("a", "pa")
	if !ans {
		t.Fatal(ans)
	}
	if cost.ToGas() == 0 {
		t.Fatal(cost)
	}
}
