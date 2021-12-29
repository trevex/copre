package config

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

func UpperSnakeCase(path []string) string {
	return strings.ToUpper(strings.Join(SplitPathIntoWords(path), "_"))
}

func LowerSnakeCase(path []string) string {
	return strings.ToLower(strings.Join(SplitPathIntoWords(path), "_"))
}

func KebabCase(path []string) string {
	return strings.ToLower(strings.Join(SplitPathIntoWords(path), "-"))
}

func CamelCase(path []string) string {
	runes := []rune(PascalCase(path))
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func PascalCase(path []string) string {
	return strings.Join(path, "") // Should already be the case, so let's just return

}

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
