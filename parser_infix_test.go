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
			name: "infix str",
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
		{
			name: "infix tree",
			in: &InfixExpression{
				Left: &InfixExpression{
					Left: &Identifier{
						Value: "xyz",
					},
					Operator: "=",
					Right: &StringLiteral{
						Value: "456",
					},
				},
				Operator: "AND",
				Right: &InfixExpression{
					Left: &Identifier{
						Value: "abc",
					},
					Operator: "=",
					Right: &StringLiteral{
						Value: "123",
					},
				},
			},
			out: "(abc = ?) AND(xyz = ?) AND",
		},
		{
			name: "infix tree 3",
			in: &InfixExpression{
				Left: &InfixExpression{
					Left: &Identifier{
						Value: "z",
					},
					Operator: "=",
					Right: &StringLiteral{
						Value: "1",
					},
				},
				Operator: "AND",
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left: &Identifier{
							Value: "x",
						},
						Operator: "=",
						Right: &StringLiteral{
							Value: "2",
						},
					},
					Operator: "AND",
					Right: &InfixExpression{
						Left: &Identifier{
							Value: "y",
						},
						Operator: "=",
						Right: &StringLiteral{
							Value: "3",
						},
					},
				},
			},
			out: "(x = ?) AND AND(y = ?) AND AND(z = ?) AND",
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
