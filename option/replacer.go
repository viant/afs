package option

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//Replace return modification handler with the specified replacements map
func Replace(replacements map[string]string) func(info os.FileInfo, reader io.ReadCloser) (io.ReadCloser, error) {
	return func(info os.FileInfo, reader io.ReadCloser) (io.ReadCloser, error) {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		_ = reader.Close()
		text := string(data)
		for k, v := range replacements {
			if count := strings.Count(text, k); count > 0 {
				text = strings.Replace(text, k, v, count)
			}
		}
		return ioutil.NopCloser(strings.NewReader(text)), nil
	}
}
