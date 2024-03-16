package modifier

import (
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"io"
	"os"
	"strings"
)

// Replace return modification handler with the specified replacements map
func Replace(replacements map[string]string) option.Modifier {
	return func(_ string, info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error) {
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, nil, err
		}
		_ = reader.Close()
		text := string(data)
		for k, v := range replacements {
			if count := strings.Count(text, k); count > 0 {
				text = strings.Replace(text, k, v, count)
			}
		}
		info = file.AdjustInfoSize(info, len(text))
		return info, io.NopCloser(strings.NewReader(text)), nil
	}
}
