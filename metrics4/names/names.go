package names

import (
	"net/url"
	"strings"
)

type Service struct {
	Service   string
	Host      string
	Path      string
	TargetURL *url.URL
}

func (s Service) String() string {
	return s.Service
}

const DotSeparator = "."
const PipeSeparator = "|"

func Flatten(name string, labels []string, separator string) string {
	if len(labels) == 0 {
		return name
	}
	return name + separator + strings.Join(labels, separator)
}

// todo(fs): this function probably allocates like crazy. If on the stack then it might be ok.
// todo(fs): otherwise, give some love.
func Labels(labels []string, values[], stringsprefix, fieldsep, recsep string) string {
	if len(labels) == 0 {
		return ""
	}
	if len(labels) > len(values) {
		v2 := make([]string, len(labels))
		copy(v2, values)
		for i := len(values); i < len(v2); i++ {
			v2[i] = ""
		}
		values = v2
	}

	return prefix + strings.Join(fields, recsep)
}
