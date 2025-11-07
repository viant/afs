package url

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func Test_IsRelative(t *testing.T) {
	type tc struct {
		in   string
		want bool
		name string
	}
	cases := []tc{
		{in: "", want: true, name: "empty"},
		{in: "foo/bar", want: true, name: "relative posix"},
		{in: "/var/tmp", want: false, name: "posix abs"},
		{in: "file:///var/tmp", want: false, name: "scheme"},
		{in: "gs://bucket/key", want: false, name: "scheme gs"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases,
			tc{in: `C:\\tmp\\x.txt`, want: false, name: "win drive"},
			tc{in: `\\\\server\\share\\dir\\x.txt`, want: false, name: "win UNC"},
		)
	} else {
		cases = append(cases,
			tc{in: `C:\\tmp\\x.txt`, want: false, name: "drive string treated abs"},
			tc{in: `\\\\server\\share\\dir\\x.txt`, want: false, name: "UNC treated abs"},
		)
	}
	for _, c := range cases {
		got := IsRelative(c.in)
		assert.EqualValues(t, c.want, got, c.name)
	}
}
