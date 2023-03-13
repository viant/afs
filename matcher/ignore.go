package matcher

import (
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

/*

Ignore matcher represents matcher that matches file that are not in the ignore rules.
The syntax of ignore borrows heavily from that of .gitignore; see https://git-scm.com/docs/gitignore or man gitignore for a full reference.

Each line is one of the following:

    pattern: a pattern specifies file names to ignore (or explicitly include) in the upload. If multiple patterns match the file name, the last matching pattern takes precedence.
    comment: comments begin with # and are ignored (see "ADVANCED TOPICS" for an exception). If you want to include a # at the beginning of a pattern, you must escape it: \#.
    blank line: A blank line is ignored and useful for readability.

*/
type Ignore struct {
	Rules []string
	Ext   map[string]bool
}

//Load loads matcher rules from location
func (i *Ignore) Load(location string) error {
	content, err := ioutil.ReadFile(location)
	if err != nil {
		return err
	}
	i.Rules = make([]string, 0)
	for _, item := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(item, "#") {
			continue
		}
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		i.Rules = append(i.Rules, strings.TrimSpace(item))
	}
	return nil
}

//Match matches returns true for any resource that does not match ignore rules
func (i *Ignore) Match(parent string, info os.FileInfo) bool {
	return !i.shouldSkip(parent, info)

}

func (i *Ignore) init() {
	var rules = i.Rules
	i.Ext = make(map[string]bool)
	if len(rules) == 0 {
		return
	}
	var updated = make([]string, 0)
	for k, rule := range rules {
		if strings.HasPrefix(rule, "*.") {
			i.Ext[rule[1:]] = true
			continue
		} else if strings.HasPrefix(rule, ".") {
			i.Ext[rules[k]] = true
			continue
		}
		updated = append(updated, rules[k])
	}
	i.Rules = updated
}

func (i *Ignore) shouldSkipFolderExpression(expr, location string) bool {
	if strings.HasPrefix(expr, "/") {
		prefix := expr[1:]
		if strings.HasPrefix(location, prefix) && prefix != location {
			return true
		}
	} else if strings.HasSuffix(expr, "/**") {
		index := strings.LastIndex(expr, "/**")
		prefix := string(expr[0:index])
		if strings.HasPrefix(location, prefix) {
			return true
		}
	} else if strings.HasSuffix(expr, "/") {
		index := strings.LastIndex(expr, "/")
		prefix := string(expr[0:index])
		if strings.HasPrefix(location, prefix) {
			return true
		}
	} else if strings.HasPrefix(expr, "**/") {
		index := strings.Index(expr, "**/")
		suffix := string(expr[index+3:])
		if strings.HasSuffix(location, suffix) {
			return true
		}
	}
	return false
}

func (i *Ignore) shouldSkipWildcardExpression(expr, location string, info os.FileInfo) bool {
	if strings.HasSuffix(expr, "*") {
		index := strings.Index(expr, "*")
		prefix := expr[:index]
		if strings.HasPrefix(location, prefix) || strings.HasPrefix(info.Name(), prefix) {
			return true
		}

	} else if strings.HasPrefix(expr, "*") {
		index := strings.Index(expr, "*")
		suffix := expr[index+1:]
		if strings.HasSuffix(location, suffix) {
			return true
		}

	} else if strings.Contains(expr, "*") {
		index := strings.Index(expr, "*")
		prefix := expr[:index]
		suffix := expr[index+1:]
		if strings.HasPrefix(location, prefix) && strings.HasSuffix(location, suffix) {
			return true
		}
		if strings.HasPrefix(info.Name(), prefix) && strings.HasSuffix(info.Name(), suffix) {
			return true
		}
	}
	return false
}

func (i *Ignore) shouldSkip(parent string, info os.FileInfo) bool {
	location := path.Join(parent, info.Name())
	if strings.HasPrefix(location, "/") {
		location = string(location[1:])
	}
	if len(i.Ext) > 0 {
		ext := path.Ext(info.Name())
		if i.Ext[ext] {
			return true
		}
	}

	if len(i.Rules) == 0 {
		return false
	}

	for _, expr := range i.Rules {
		if info.Name() == expr {
			return true
		} else if strings.Contains(expr, "/") {
			if i.shouldSkipFolderExpression(expr, location) {
				return true
			}
		} else {
			if i.shouldSkipWildcardExpression(expr, location, info) {
				return true
			}
		}

	}
	return false
}

//NewIgnore creates a new exclusion rule
func NewIgnore(options ...storage.Option) (*Ignore, error) {
	location := &option.Location{}
	ignore := &Ignore{
		Rules: make([]string, 0),
	}
	option.Assign(options, &location, &ignore.Rules)
	if location.Path != "" {
		return ignore, ignore.Load(location.Path)
	}
	ignore.init()
	return ignore, nil
}
