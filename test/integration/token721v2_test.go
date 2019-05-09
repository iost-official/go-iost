package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/vm/native"
	"github.com/stretchr/testify/assert"

	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
)

func Test_Token721v2Approval(t *testing.T) {
	ilog.Start()
	s := NewSimulator()
	defer s.Clear()

	createAccountsWithResource(s)
	s.SetContract(native.SystemContractABI("token721.iost", "1.1.0"))
	r, err := s.Call("token721.iost", "create", fmt.Sprintf(`["iost","%s",100000]`, acc0.ID), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	for i := 0; i < 6; i++ {
		r, err = s.Call("token721.iost", "issue", fmt.Sprintf(`["iost","%s","{\"id\":%d}"]`, acc0.ID, i), acc0.ID, acc0.KeyPair)
		assert.Nil(t, err)
		assert.Equal(t, "", r.Status.Message)
		assert.Equal(t, fmt.Sprintf(`["%d"]`, i), r.Returns[0])
	}

	ca, err := s.Compile("", "./test_data/token721v2", "./test_data/token721v2.js")
	assert.Nil(t, err)
	assert.NotNil(t, ca)

	cname, r, err := s.DeployContract(ca, acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	r, err = s.Call("token721.iost", "approve", fmt.Sprintf(`["iost","%s","%s","%d"]`, acc0.ID, cname, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	r, err = s.Call(cname, "approve", fmt.Sprintf(`["%s","%s","%d"]`, acc0.ID, cname, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Contains(t, r.Status.Message, "transaction has no permission")

	r, err = s.Call(cname, "approveWithAuth", fmt.Sprintf(`["%s","%s","%d"]`, acc0.ID, cname, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Contains(t, r.Status.Message, "transaction has no permission")

	r, err = s.Call(cname, "transfer", fmt.Sprintf(`["%s","%s","%d"]`, acc0.ID, cname, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Contains(t, r.Status.Message, "transaction has no permission")

	r, err = s.Call(cname, "transferWithAuth", fmt.Sprintf(`["%s","%s","%d"]`, acc0.ID, cname, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	s.Head.Time += 1000
	cname2, r, err := s.DeployContract(ca, acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	r, err = s.Call(cname, "approve", fmt.Sprintf(`["%s","%s","%d"]`, cname, cname2, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Contains(t, r.Status.Message, "transaction has no permission")

	r, err = s.Call(cname, "approveWithAuth", fmt.Sprintf(`["%s","%s","%d"]`, cname, cname2, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)

	r, err = s.Call(cname2, "transfer", fmt.Sprintf(`["%s","%s","%d"]`, cname, cname2, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Contains(t, r.Status.Message, "transaction has no permission")

	r, err = s.Call(cname2, "transferWithAuth", fmt.Sprintf(`["%s","%s","%d"]`, cname, cname2, 0), acc0.ID, acc0.KeyPair)
	assert.Nil(t, err)
	assert.Equal(t, "", r.Status.Message)
}
