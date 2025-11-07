package url

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func Test_ToFileURL(t *testing.T) {
	type tc struct {
		in   string
		want string
		name string
	}

	cases := []tc{
		{in: "file:///var/tmp/a.txt", want: "file:///var/tmp/a.txt", name: "already uri posix"},
		{in: "/var/tmp/a.txt", want: "file:///var/tmp/a.txt", name: "posix abs"},
		{in: "gs://bucket/key", want: "gs://bucket/key", name: "gs unchanged"},
		{in: "agents/foo.txt", want: "file://agents/foo.txt", name: "relative posix"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases,
			tc{in: `C:\\tmp\\x.txt`, want: `file:///C:/tmp/x.txt`, name: "win drive"},
			tc{in: `\\\\server\\share\\dir\\x.txt`, want: `file://server/share/dir/x.txt`, name: "win UNC"},
		)
	} else {
		cases = append(cases, tc{in: `C:\\tmp\\x.txt`, want: `file:///C:/tmp/x.txt`, name: "win drive on posix"})
	}

	for _, c := range cases {
		got := ToFileURL(c.in)
		assert.EqualValues(t, c.want, got, c.name)
	}
}
