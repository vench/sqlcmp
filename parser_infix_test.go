package sqlcmp

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func Test_structcher(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		in   Expression
		out  string
	}{
		{
			name: "empty",
		},
		{
			name: "integer",
			in: &IntegerLiteral{
				Value: 100,
			},
			out: "?",
		},
		{
			name: "string",
			in: &StringLiteral{
				Value: "xyz",
			},
			out: "?",
		},
		{
			name: "string",
			in: &InfixExpression{
				Left: &Identifier{
					Value: "name",
				},
				Operator: "=",
				Right: &StringLiteral{
					Value: "xyz",
				},
			},
			out: "(name = ?)",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out := structcher(tc.in)
			require.Equal(t, tc.out, out)
		})
	}
}
