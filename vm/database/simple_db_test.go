package database

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/bitly/go-simplejson"
)

func TestDB(t *testing.T) {
	db := NewDatabase()
	db.AddSystem("system.json")
	db.Put("state", "i-1", "sabc")
	err := db.Save("state.json")
	if err != nil {
		t.Fatal(err)
	}
	db2 := NewDatabase()
	db2.AddSystem("system.json")
	db2.Load("state.json")
	v, err := db2.Get("state", "i-1")
	if err != nil {
		t.Fatal(err)
	}
	if v != "sabc@" {
		t.Fatal(v)
	}
	os.Remove("state.json")
}

func TestSjson(t *testing.T) {
	s := `
{
  "a":100,
  "b":"str",
  "c":true,
  "d": null
}`
	j, err := simplejson.NewJson([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(reflect.TypeOf(j.Get("a").Interface()).Name())
	fmt.Println(reflect.TypeOf(j.Get("b").Interface()).Name())
	fmt.Println(reflect.TypeOf(j.Get("c").Interface()).Name())
	fmt.Println(j.Get("d").Interface())

}
