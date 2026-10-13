package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseValueCoversSupportedTypesAndInvalidInputs(t *testing.T) {
	timeValue := time.Date(2026, time.October, 12, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		valueType ValueType
		raw       string
		want      any
		wantErr   bool
	}{
		{name: "string", valueType: ValueTypeString, raw: "valor", want: "valor"},
		{name: "uint64", valueType: ValueTypeUint64, raw: "42", want: uint64(42)},
		{name: "uint64 inválido", valueType: ValueTypeUint64, raw: "x", wantErr: true},
		{name: "float64", valueType: ValueTypeFloat64, raw: "12.5", want: float64(12.5)},
		{name: "float64 inválido", valueType: ValueTypeFloat64, raw: "x", wantErr: true},
		{name: "bool", valueType: ValueTypeBool, raw: "true", want: true},
		{name: "bool inválido", valueType: ValueTypeBool, raw: "sim", wantErr: true},
		{name: "timestamp RFC3339", valueType: ValueTypeTime, raw: timeValue.Format(time.RFC3339), want: timeValue},
		{name: "data ISO", valueType: ValueTypeTime, raw: "2026-10-12", want: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)},
		{name: "data inválida", valueType: ValueTypeTime, raw: "12/10/2026", wantErr: true},
		{name: "tipo desconhecido", valueType: ValueType("unknown"), raw: "valor", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseValue(test.valueType, test.raw)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
