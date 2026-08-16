package texts

import (
	"strings"
	"unicode"
)

// NormalizeUcFirst  remove espaços desnecessários e padroniza
// cada palavra com a primeira letra maiúscula.
//
// Exemplo:
// "  joÃO   da SILVA  " -> "João Da Silva"
func NormalizeUcFirst(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)

	if value == "" {
		return ""
	}

	words := strings.Fields(value)

	for index, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[index] = string(runes)
	}

	return strings.Join(words, " ")
}

// NormalizeSpaces remove espaços no início e no final
// e reduz múltiplos espaços internos para apenas um.
//
// Exemplo:
// "  Motor   com problema  " -> "Motor com problema"
func NormalizeSpaces(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// NormalizeLower remove espaços externos e converte
// todos os caracteres para minúsculo.
//
// Exemplo:
// "  EMAIL@TESTE.COM  " -> "email@teste.com"
func NormalizeLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// NormalizeUpper remove espaços externos e converte
// todos os caracteres para maiúsculo.
//
// Exemplo:
// "  abc1d23  " -> "ABC1D23"
func NormalizeUpper(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// OnlyNumbers remove todos os caracteres que não sejam números.
//
// Exemplo:
// "529.982.247-25" -> "52998224725"
func OnlyNumbers(value string) string {
	var builder strings.Builder

	for _, character := range value {
		if unicode.IsDigit(character) {
			builder.WriteRune(character)
		}
	}

	return builder.String()
}
