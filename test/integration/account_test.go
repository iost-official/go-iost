package integration

import (
	"testing"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthority(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	s := verifier.NewSimulator()
	defer s.Clear()
	Convey("test of Auth", t, func() {
		ca, err := s.Compile("auth.iost", "../../contract/account", "../../contract/account.js")
		So(err, ShouldBeNil)
		s.Visitor.SetContract(ca)
		s.Visitor.SetContract(native.GasABI())
		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 1e8)
		s.SetRAM(testID[0], 1000)
		s.SetRAM("myidid", 1000)

		r, err := s.Call("auth.iost", "SignUp", array2json([]interface{}{"myidid", kp.ID, kp.ID}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldStartWith, `{"id":"myidid",`)

		r, err = s.Call("auth.iost", "SignUp", array2json([]interface{}{"invalid#id", kp.ID, kp.ID}), kp.ID, kp)
		So(err, ShouldBeNil)
		ilog.Info(r.Status.Message)
		So(r.Status.Message, ShouldContainSubstring, "id contains invalid character")

		acc, _ := host.ReadAuth(s.Visitor, "myidid")
		So(acc.Referrer, ShouldEqual, kp.ID)
		So(acc.ReferrerUpdateTime, ShouldEqual, s.Head.Time)
		s.SetGas("myidid", 10000000)
		r, err = s.Call("auth.iost", "UpdateReferrer", array2json([]interface{}{"myidid", "hahaha"}), "myidid", kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "referrer can only be updated one time per 30 days")
		s.Head.Time += 30 * 24 * 3600 * 1e9
		r, err = s.Call("auth.iost", "UpdateReferrer", array2json([]interface{}{"myidid", "hahaha"}), "myidid", kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		acc, _ = host.ReadAuth(s.Visitor, "myidid")
		So(acc.Referrer, ShouldEqual, "hahaha")

		r, err = s.Call("auth.iost", "AddPermission", array2json([]interface{}{"myidid", "perm1", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)

		r, err = s.Call("auth.iost", "AddPermission", array2json([]interface{}{"myidid", "perm1", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "permission already exist")

		r, err = s.Call("auth.iost", "DropPermission", array2json([]interface{}{"myidid", "perm1"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldNotContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)

		r, err = s.Call("auth.iost", "AssignPermission", array2json([]interface{}{"myidid", "active", "acc1", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "unexpected item")

		r, err = s.Call("auth.iost", "AssignPermission", array2json([]interface{}{"myidid", "active", "IOST1234", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldContainSubstring, `{"id":"IOST1234","is_key_pair":true,"weight":1}`)

		r, err = s.Call("auth.iost", "AssignPermission", array2json([]interface{}{"myidid", "active", "acc1@active", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldContainSubstring, `{"id":"acc1","permission":"@active","is_key_pair":false,"weight":1}`)

		r, err = s.Call("auth.iost", "RevokePermission", array2json([]interface{}{"myidid", "active", "acc1"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldNotContainSubstring, `{"id":"acc1","permission":"@active","is_key_pair":false,"weight":1}`)

		r, err = s.Call("auth.iost", "RevokePermission", array2json([]interface{}{"myidid", "active", "acc2"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "item not found")
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldNotContainSubstring, `{"id":"acc1","permission":"@active","is_key_pair":false,"weight":1}`)

		r, err = s.Call("auth.iost", "AddGroup", array2json([]interface{}{"myidid", "grp0"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldContainSubstring, `"groups":{"grp0":{"name":"grp0","items":[]}}`)

		r, err = s.Call("auth.iost", "AddGroup", array2json([]interface{}{"myidid", "grp0"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "group already exist")
		ilog.Info(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")))
		r, err = s.Call("auth.iost", "DropGroup", array2json([]interface{}{"myidid", "grp0"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(database.Unmarshal(s.Visitor.MGet("auth.iost-auth", "myidid")), ShouldNotContainSubstring, `"groups":{"grp0":{"name":"grp0","items":[]}}`)
	})
}
