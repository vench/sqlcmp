package sqlcmp

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func TestSQLCondition_Structcher(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		in   SQLCondition
		out  string
	}{
		{
			name: "empty",
		},
		{
			name: "infix eq",
			in: SQLCondition{
				Expression: &InfixExpression{
					Operator: ASSIGN,
					Left:     &Identifier{Token: Token{Type: IDENT, Literal: "t1"}, Value: "t1.key"},
					Right:    &IntegerLiteral{Value: 100},
				},
			},
			out: "(t1.key = ?)",
		},
		{
			name: "infix eq 2 wrap",
			in: SQLCondition{
				Expression: &SQLCondition{
					Expression: &InfixExpression{
						Operator: ASSIGN,
						Left:     &Identifier{Token: Token{Type: IDENT, Literal: "t1"}, Value: "t1.key"},
						Right:    &IntegerLiteral{Value: 100},
					},
				},
			},
			out: "(t1.key = ?)",
		},
		{
			name: "infix in int",
			in: SQLCondition{
				Expression: &InExpression{
					Column: &Identifier{Token: Token{Type: IDENT, Literal: "t1"}, Value: "t1.key"},
					Arguments: []Expression{
						&IntegerLiteral{Token: Token{Type: INT, Literal: "100"}, Value: 100},
						&IntegerLiteral{Token: Token{Type: INT, Literal: "200"}, Value: 200},
					},
				},
			},
			out: "t1.key IN (?)",
		},
		{
			name: "infix in string",
			in: SQLCondition{
				Expression: &InExpression{
					Column: &Identifier{Token: Token{Type: IDENT, Literal: "t2"}, Value: "t2.name"},
					Arguments: []Expression{
						&StringLiteral{Token: Token{Type: STRING, Literal: "abc"}, Value: "abc"},
						&StringLiteral{Token: Token{Type: STRING, Literal: "xyz"}, Value: "xyz"},
					},
				},
			},
			out: "t2.name IN (?)",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out := tc.in.Structcher()
			require.Equal(t, tc.out, out)
		})
	}
}
