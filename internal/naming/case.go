package naming

import (
	"strings"
	"unicode"
)

func ToPascal(value string) string {
	words := splitWords(value)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
	}
	return strings.Join(words, "")
}

func ToCamel(value string) string {
	pascal := ToPascal(value)
	if pascal == "" {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func ToSnake(value string) string {
	return strings.Join(lowerWords(value), "_")
}

func ToKebab(value string) string {
	return strings.Join(lowerWords(value), "-")
}

func lowerWords(value string) []string {
	words := splitWords(value)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return words
}

func splitWords(value string) []string {
	var words []string
	var current []rune
	for _, r := range value {
		switch {
		case r == '-' || r == '_' || r == '/' || unicode.IsSpace(r):
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		case unicode.IsUpper(r) && len(current) > 0:
			words = append(words, string(current))
			current = []rune{r}
		default:
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}
