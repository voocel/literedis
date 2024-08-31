package sortedset

import (
	"errors"
	"regexp"
	"strings"
)

type Pattern struct {
	exp *regexp.Regexp
}

var replaceMap = map[byte]string{
	'+': `\+`,
	')': `\)`,
	'$': `\$`,
	'.': `\.`,
	'{': `\{`,
	'}': `\}`,
	'|': `\|`,
	'*': ".*",
	'?': ".",
}

var errEndWithEscape = "end with escape \\"

func CompilePattern(src string) (*Pattern, error) {
	regexSrc := strings.Builder{}
	regexSrc.WriteByte('^')
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if ch == '\\' {
			if i == len(src)-1 {
				return nil, errors.New(errEndWithEscape)
			}
			regexSrc.WriteByte(ch)
			regexSrc.WriteByte(src[i+1])
			i++
		} else if ch == '^' {
			if i == 0 {
				regexSrc.WriteString(`\^`)
			} else if i == 1 {
				if src[i-1] == '[' {
					regexSrc.WriteString(`^`)
				} else {
					regexSrc.WriteString(`\^`)
				}
			} else {
				if src[i-1] == '[' && src[i-2] != '\\' {
					regexSrc.WriteString(`^`)
				} else {
					regexSrc.WriteString(`\^`)
				}
			}
		} else if escaped, toEscape := replaceMap[ch]; toEscape {
			regexSrc.WriteString(escaped)
		} else {
			regexSrc.WriteByte(ch)
		}
	}
	regexSrc.WriteByte('$')
	re, err := regexp.Compile(regexSrc.String())
	if err != nil {
		return nil, err
	}
	return &Pattern{
		exp: re,
	}, nil
}

func (p *Pattern) IsMatch(s string) bool {
	return p.exp.Match([]byte(s))
}
