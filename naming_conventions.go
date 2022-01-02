package config

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

// UpperSnakeCase takes the path to a struct-field and returns it in upper-case snake-case.
// For example:
//  UpperSnakeCase([]string{"Document","HTMLParser"}) // returns "DOCUMENT_HTML_PARSER"
func UpperSnakeCase(path []string) string {
	return strings.ToUpper(strings.Join(SplitPathIntoWords(path), "_"))
}

// LowerSnakeCase takes the path to a struct-field and returns it in lower-case snake-case.
// For example:
//  LowerSnakeCase([]string{"Document","HTMLParser"}) // returns "document_html_parser"
func LowerSnakeCase(path []string) string {
	return strings.ToLower(strings.Join(SplitPathIntoWords(path), "_"))
}

// KebabCase will convert the provided struct-field path to kebab-case, e.g.:
//  KebabCase([]string{"Document","HTMLParser"}) // returns "document-html-parser"
func KebabCase(path []string) string {
	return strings.ToLower(strings.Join(SplitPathIntoWords(path), "-"))
}

// CamelCase will convert the provided struct-field path to camel-case. For example:
//  CamelCase([]string{"Document","HTMLParser"}) // returns "documentHTMLParser"
func CamelCase(path []string) string {
	runes := []rune(PascalCase(path))
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// PascalCase will convert the provided struct-field path to pascal-case. For example:
//  PascalCase([]string{"Document","HTMLParser"}) // returns "DocumentHTMLParser"
// As you can see above the input is already expected to be field names in pascal-case
// and the names will be joined.
func PascalCase(path []string) string {
	return strings.Join(path, "") // Should already be the case, so let's just return

}

// SplitPathIntoWords is a utility function used by some of the *Case-functions.
// It takes a field path and returns the list of individual words. For example:
//  SplitPathIntoWords([]string{"Document", "HTMLParser"}) // returns []string{"Document", "HTML", "Parser"}
func SplitPathIntoWords(path []string) []string {
	// Let's be optimisitic and set the capacity to path assuming that often
	// we don't have to split words.
	words := make([]string, 0, len(path))
	for _, elem := range path {
		splits := camelcase.Split(elem)
		words = append(words, splits...)
	}
	return words
}
