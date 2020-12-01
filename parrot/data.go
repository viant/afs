package parrot

import (
	"fmt"
	"strings"
)

//Data represents parrot data
type Data []byte

//AsBytesLiteral returns literal bytes representation
func (d Data) AsBytesLiteral(ASCII bool) string {
	if ASCII {
		data := string(d)
		var count = strings.Count(data, "`")
		if count > 0 {
			data = strings.Replace(data, "`", "`+\"`\"+`", count)
			return fmt.Sprintf("[]byte(`%s`)", data)
		}
		return fmt.Sprintf("[]byte(`%s`)", data)

	}
	var parts = make([]string, 0)
	for i := 0; i < len(d); i += 16 {
		part := make([]string, 16)
		j := 0
		for j = 0; (j+i) < len(d) && j < 16; j++ {
			part[j] = fmt.Sprintf("0x%x", d[i+j])
		}
		parts = append(parts, strings.Join(part[:j], ","))
	}
	return fmt.Sprintf("[]byte{%s}", strings.Join(parts, ",\n"))
}
