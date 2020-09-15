// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bytes"
	"strings"
)

// Expander represents a property set that interprets special character
// sequences in property values as references to other property values for
// replacement.
//
// For example, the following properties:
//     color.alert = red
//     color.info = blue
//     color.text = black
//
//     css.alert = border: 1px solid ${color.alert}; color: ${color.text};
//     css.info = border: 1px solid ${color.info}; color: ${color.text};
// Would result in the following values:
//     "css.alert": "border: 1px solid red; color: black;"
//     "css.info":  "border: 1px solid blue; color: black;"
//
// Nested and recursive property expansions are permitted. If a property value
// does not exist, the property reference will be left unchanged.
type Expander struct {
	// Prefix indicates the start of a property expansion.
	Prefix string
	// Suffix indicates the end of a property expansion.
	Suffix string
	*Properties
}

// NewExpander creates an empty property set with the default expansion
// Prefix "${" and Suffix "}".
func NewExpander() *Expander {
	e := &Expander{
		Prefix:     "${",
		Suffix:     "}",
		Properties: NewProperties(),
	}
	return e
}

// Get retrieves the value of a property with all property references expanded.
// If the property does not exist, an empty string will be returned.
func (e *Expander) Get(key string) string {
	v := e.Properties.Get(key)
	return e.expand(v)
}

// GetDefault retrieves the value of a property with all property references
// expanded. If the property does not exist, the default value will be returned
// with all its property references expanded.
func (e *Expander) GetDefault(key, defVal string) string {
	v := e.Properties.GetDefault(key, defVal)
	return e.expand(v)
}

// expand any embedded property references in a string
func (e *Expander) expand(v string) string {
	if len(v) <= 0 || strings.Index(v, e.Prefix) < 0 ||
		strings.Index(v, e.Suffix) < 0 {

		return v
	}

	var out bytes.Buffer
	start := 0
	nest := 0
	for i := 0; i < len(v); i++ {
		if !strings.HasPrefix(v[i:], e.Prefix) {
			continue
		}

		out.WriteString(v[start:i])
		start = i + len(e.Prefix)

		for j := start; j < len(v); j++ {
			if strings.HasPrefix(v[j:], e.Suffix) {
				if nest == 0 {
					exp := e.expand(v[start:j])
					val := e.Properties.Get(exp)
					if len(val) == 0 {
						out.WriteString(e.Prefix)
						out.WriteString(exp)
						out.WriteString(e.Suffix)
					} else {
						out.WriteString(val)
					}
					start = j + len(e.Suffix)
					i = start - 1
					break
				} else {
					nest--
				}
			} else if strings.HasPrefix(v[j:], e.Prefix) {
				nest++
			}
		}
	}

	if start < len(v) {
		out.WriteString(v[start:])
	}

	result := out.String()

	// expand properties recursively
	if v == result {
		return out.String()
	} else {
		return e.expand(out.String())
	}
}
