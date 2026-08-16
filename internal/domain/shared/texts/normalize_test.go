package texts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeUcFirst(t *testing.T) {
	result := NormalizeUcFirst("  joÃO   da SILVA  ")
	require.Equal(t, "João Da Silva", result)
}

func TestNormalizeUcFirstEmpty(t *testing.T) {
	result := NormalizeUcFirst("   ")
	require.Empty(t, result)
}

func TestNormalizeSpaces(t *testing.T) {
	result := NormalizeSpaces("  Motor   com   problema  ")
	require.Equal(t, "Motor com problema", result)
}

func TestNormalizeLower(t *testing.T) {
	result := NormalizeLower("  EMAIL@TESTE.COM  ")
	require.Equal(t, "email@teste.com", result)
}

func TestNormalizeUpper(t *testing.T) {
	result := NormalizeUpper("  abc1d23  ")
	require.Equal(t, "ABC1D23", result)
}

func TestOnlyNumbers(t *testing.T) {
	result := OnlyNumbers("529.982.247-25")
	require.Equal(t, "52998224725", result)
}

func TestOnlyNumbersEmpty(t *testing.T) {
	result := OnlyNumbers("ABC.-/")
	require.Empty(t, result)
}
