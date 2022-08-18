package gendoc

import (
    "errors"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

var (
	paraPattern         = regexp.MustCompile(`(\n|\r|\r\n)\s*`)
	spacePattern        = regexp.MustCompile("( )+")
	multiNewlinePattern = regexp.MustCompile(`(\r\n|\r|\n){2,}`)
	specialCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
    crossReferencePattern = regexp.MustCompile(`<ref>((?:.|\r\n)*?)<\/ref>`)
)

// PFilter splits the content by new lines and wraps each one in a <p> tag.
func PFilter(content string) template.HTML {
	paragraphs := paraPattern.Split(content, -1)
	return template.HTML(fmt.Sprintf("<p>%s</p>", strings.Join(paragraphs, "</p><p>")))
}

// ParaFilter splits the content by new lines and wraps each one in a <para> tag.
func ParaFilter(content string) string {
	paragraphs := paraPattern.Split(content, -1)
	return fmt.Sprintf("<para>%s</para>", strings.Join(paragraphs, "</para><para>"))
}

// NoBrFilter removes single CR and LF from content.
func NoBrFilter(content string) string {
	normalized := strings.Replace(content, "\r\n", "\n", -1)
	paragraphs := multiNewlinePattern.Split(normalized, -1)
	for i, p := range paragraphs {
		withoutCR := strings.Replace(p, "\r", " ", -1)
		withoutLF := strings.Replace(withoutCR, "\n", " ", -1)
		paragraphs[i] = spacePattern.ReplaceAllString(withoutLF, " ")
	}
	return strings.Join(paragraphs, "\n\n")
}

// AnchorFilter replaces all special characters with URL friendly dashes
func AnchorFilter(str string) string {
	return specialCharsPattern.ReplaceAllString(strings.ReplaceAll(str, "/", "_"), "-")
}

func FieldMessageFilter(lookupMap map[string]*Message) func() (map[string]*Message) {
    return func() map[string]*Message {
      return lookupMap
    }
}

func HtmlFilter(lookupMap map[string]lookup) func(string) (string, error) {

    return func(str string) (string, error) {
      matches := crossReferencePattern.FindAllStringSubmatchIndex(str, -1)
      replacements := make([]string, 0);
      cursor := 0
      for i := 0; i < len(matches); i++ {
        match := str[matches[i][2]:matches[i][3]];
        // Try to find the match
        if lookup, ok := lookupMap[match]; ok {
          // copy from cursor to matches[i][0]
          // fmt.Fprintf(os.Stderr, "it's a method: %v: %v -> %v\n", matches[i], match, lookup);
          replacements = append(replacements, str[cursor:matches[i][0]]);
          replacements = append(replacements, fmt.Sprintf("<a href='#%s'>%s</a>", lookup.id, lookup.name));
          cursor = matches[i][1];
        } else {
          return "", errors.New(fmt.Sprintf("invalid cross reference: %s\n", match));
        }
      }
      replacements = append(replacements, str[cursor:])
      return strings.Join(replacements, ""), nil
    }
}

func DictBuilder(values ...interface{}) (map[string]interface{}, error) {
    if len(values)%2 != 0 {
        return nil, errors.New("invalid dict call")
    }
    dict := make(map[string]interface{}, len(values)/2)
    for i := 0; i < len(values); i+=2 {
        key, ok := values[i].(string)
        if !ok {
            return nil, errors.New("dict keys must be strings")
        }
        dict[key] = values[i+1]
    }
    return dict, nil
}
