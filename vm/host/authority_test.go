package host

import (
	"encoding/json"
	"testing"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/vm/database"
)

func TestRequireAuth_ByKey(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"keya": 1})
	ctx.Set("signer_list", map[string]bool{"a@pa": true})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["pa"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Items: []*account.Item{
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

	ans, cost := host.RequireSignerAuth("a", "pa")
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
	ctx.Set("signer_list", map[string]bool{"b@active": true})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["pa"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Items: []*account.Item{
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
			Items: []*account.Item{
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
	ctx.Set("signer_list", map[string]bool{"a@active": true})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").AnyTimes().DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["owner"] = &account.Permission{
			Name:   "owner",
			Groups: []string{},
			Items: []*account.Item{
				{
					ID:         "keyowner",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		ac.Permissions["active"] = &account.Permission{
			Name:   "active",
			Groups: []string{},
			Items: []*account.Item{
				{
					ID:         "keya",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		ac.Permissions["pa"] = &account.Permission{
			Name:   "pa",
			Groups: []string{},
			Items: []*account.Item{
				{
					ID:         "keypa",
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

	ans, _ = host.RequireAuth("a", "owner")
	if ans {
		t.Fatal(ans)
	}

}

func TestAuthority_Owner(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"keya": 1})
	ctx.Set("signer_list", map[string]bool{"a@owner": true})

	db, host := myinit(t, ctx)

	db.EXPECT().Commit().Return()
	db.EXPECT().Get("state", "m-auth.iost-auth-a").DoAndReturn(func(a, b string) (string, error) {
		ac := account.NewAccount("a")
		ac.Permissions["owner"] = &account.Permission{
			Name:   "owner",
			Groups: []string{},
			Items: []*account.Item{
				{
					ID:         "keya",
					Permission: "",
					IsKeyPair:  true,
					Weight:     1,
				},
			},
			Threshold: 1,
		}
		ac.Permissions["active"] = &account.Permission{
			Name:   "active",
			Groups: []string{},
			Items: []*account.Item{
				{
					ID:         "keyactive",
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
	ans, cost := host.RequireAuth("a", "active")
	if !ans {
		t.Fatal(ans)
	}
	if cost.ToGas() == 0 {
		t.Fatal(cost)
	}
}
