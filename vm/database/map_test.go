package database

import (
	"testing"

	"strings"

	"github.com/gin-gonic/gin/json"
)

func TestJson(t *testing.T) {
	buf, err := json.Marshal([]string{"abc", "def"})
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != `["abc","def"]` {
		t.Fatal(string(buf))
	}
}

func TestString(t *testing.T) {
	ss := strings.Replace("@a@b@c", "@b", "", 1)
	if ss != `@a@c` {
		t.Fatal(ss)
	}

	sl := strings.Split("@a@b@c", "@")
	if sl[1] != "a" {
		t.Fatal(sl)
	}
}
