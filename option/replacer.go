package option

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//Replace return modification handler with the specified replacements map
func Replace(replacements map[string]string) Modifier {
	return func(info os.FileInfo, reader io.Reader) (io.Reader, error) {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		text := string(data)
		for k, v := range replacements {
			if count := strings.Count(text, k); count > 0 {
				text = strings.Replace(text, k, v, count)
			}
		}
		return ioutil.NopCloser(strings.NewReader(text)), nil
	}
}
