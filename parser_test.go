package sqlcmp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_parseSQLCond(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
		expectedValue any
	}{
		{
			input: "(t1.key = t3.key and t3.date > now())", expectedQuery: "((t1.key = t3.key) AND (t3.date > now()))", expectedValue: &SQLCondition{
				Expression: &SQLCondition{
					Expression: &SQLCondition{
						Expression: &InfixExpression{
							Token: Token{Type: SQLAnd, Literal: "and"},
							Left: &InfixExpression{
								Token:    Token{Type: ASSIGN, Literal: "="},
								Left:     &Identifier{Token: Token{Type: IDENT, Literal: "t1"}, Value: "t1.key"},
								Operator: ASSIGN,
								Right:    &Identifier{Token: Token{Type: IDENT, Literal: "t3"}, Value: "t3.key"},
							},
							Operator: SQLAnd,
							Right: &InfixExpression{
								Token:    Token{Type: GT, Literal: ">"},
								Left:     &Identifier{Token: Token{Type: IDENT, Literal: "t3"}, Value: "t3.date"},
								Operator: GT,
								Right: &CallExpression{
									Token: Token{Type: LPAREN, Literal: "("},
									Function: &Identifier{
										Token: Token{Type: IDENT, Literal: "now"},
										Value: "now",
									},
								},
							},
						},
					},
				},
			},
		},
		{input: "id=1", expectedQuery: "(id = 1)", expectedValue: &SQLCondition{
			Expression: &InfixExpression{
				Token:    Token{Type: ASSIGN, Literal: ASSIGN.String()},
				Left:     &Identifier{Token: Token{Type: IDENT, Literal: "id"}, Value: "id"},
				Operator: ASSIGN,
				Right:    &IntegerLiteral{Token: Token{Type: INT, Literal: "1"}, Value: 1},
			},
		}},
		{input: "t.id=1", expectedQuery: "(t.id = 1)", expectedValue: &SQLCondition{
			Expression: &InfixExpression{
				Token:    Token{Type: ASSIGN, Literal: ASSIGN.String()},
				Left:     &Identifier{Token: Token{Type: IDENT, Literal: "t"}, Value: "t.id"},
				Operator: ASSIGN,
				Right:    &IntegerLiteral{Token: Token{Type: INT, Literal: "1"}, Value: 1},
			},
		}},
		{
			input: "id>100", expectedQuery: "(id > 100)", expectedValue: &SQLCondition{
				Expression: &InfixExpression{
					Token:    Token{Type: GT, Literal: ">"},
					Left:     &Identifier{Token: Token{Type: IDENT, Literal: "id"}, Value: "id"},
					Operator: GT,
					Right:    &IntegerLiteral{Token: Token{Type: INT, Literal: "100"}, Value: 100},
				},
			},
		},
		{
			input: "name = 'test*'", expectedQuery: "(name = test*)", expectedValue: &SQLCondition{
				Expression: &InfixExpression{
					Token:    Token{Type: ASSIGN, Literal: "="},
					Left:     &Identifier{Token: Token{Type: IDENT, Literal: "name"}, Value: "name"},
					Operator: ASSIGN,
					Right:    &StringLiteral{Token: Token{Type: STRING, Literal: "test*"}, Value: "test*"},
				},
			},
		},
		{
			input: "(name = 'test*')", expectedQuery: "(name = test*)", expectedValue: &SQLCondition{
				Expression: &SQLCondition{
					Expression: &InfixExpression{
						Token:    Token{Type: ASSIGN, Literal: "="},
						Left:     &Identifier{Token: Token{Type: IDENT, Literal: "name"}, Value: "name"},
						Operator: ASSIGN,
						Right:    &StringLiteral{Token: Token{Type: STRING, Literal: "test*"}, Value: "test*"},
					},
				},
			},
		},
		{
			input: "x=1 and y=2", expectedQuery: "((x = 1) AND (y = 2))", expectedValue: &SQLCondition{
				Expression: &SQLCondition{
					Expression: &InfixExpression{
						Token: Token{Type: SQLAnd, Literal: "and"},
						Left: &InfixExpression{
							Token:    Token{Type: ASSIGN, Literal: ASSIGN.String()},
							Left:     &Identifier{Token: Token{Type: IDENT, Literal: "x"}, Value: "x"},
							Operator: ASSIGN,
							Right:    &IntegerLiteral{Token: Token{Type: INT, Literal: "1"}, Value: 1},
						},
						Operator: SQLAnd,
						Right: &InfixExpression{
							Token:    Token{Type: ASSIGN, Literal: ASSIGN.String()},
							Left:     &Identifier{Token: Token{Type: IDENT, Literal: "y"}, Value: "y"},
							Operator: ASSIGN,
							Right:    &IntegerLiteral{Token: Token{Type: INT, Literal: "2"}, Value: 2},
						},
					},
				},
			},
		},
		{
			input:         "(email like '%abc%')",
			expectedQuery: "(email LIKE %abc%)",
			expectedValue: &SQLCondition{
				Expression: &SQLCondition{
					Expression: &InfixExpression{
						Token:    Token{Type: SQLLike, Literal: "like"},
						Left:     &Identifier{Token: Token{Type: IDENT, Literal: "email"}, Value: "email"},
						Operator: SQLLike,
						Right: &StringLiteral{
							Token: Token{Type: STRING, Literal: "%abc%"},
							Value: "%abc%",
						},
					},
				},
			},
		},
		{
			input:         "date between '2023-10-01' and '2023-10-15'",
			expectedQuery: "date BETWEEN 2023-10-01 AND 2023-10-15",
			expectedValue: &SQLCondition{
				Expression: &BetweenExpression{
					Token:  Token{Type: SQLBetween, Literal: "between"},
					Column: &Identifier{Token: Token{Type: IDENT, Literal: "date"}, Value: "date"},
					From: &Identifier{
						Token: Token{Type: STRING, Literal: "2023-10-01"},
						Value: "2023-10-01",
					},
					To: &Identifier{
						Token: Token{Type: STRING, Literal: "2023-10-15"},
						Value: "2023-10-15",
					},
				},
			},
		},
		{
			input:         "range between 10 and 999",
			expectedQuery: "range BETWEEN 10 AND 999",
			expectedValue: &SQLCondition{
				Expression: &BetweenExpression{
					Token:  Token{Type: SQLBetween, Literal: "between"},
					Column: &Identifier{Token: Token{Type: IDENT, Literal: "range"}, Value: "range"},
					From: &Identifier{
						Token: Token{Type: INT, Literal: "10"},
						Value: "10",
					},
					To: &Identifier{
						Token: Token{Type: INT, Literal: "999"},
						Value: "999",
					},
				},
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		p := NewParser(NewLexer(tc.input))
		exp := p.parseSQLCondition()

		require.EqualValuesf(t, 0, len(p.Errors()), "%v", p.Errors())
		require.Equal(t, tc.expectedQuery, exp.String())
		require.Equal(t, tc.expectedValue, exp)
	}
}

func TestParser_parseSQLColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
		expectedExp   Expression
	}{
		{
			input:         "date",
			expectedQuery: "date",
			expectedExp: &Identifier{
				Token: Token{Type: IDENT, Literal: "date"},
				Value: "date",
			},
		},
		{
			input:         "now() as date",
			expectedQuery: "now() AS date",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &CallExpression{
					Token: Token{Type: LPAREN, Literal: "("},
					Function: &Identifier{
						Token: Token{Type: IDENT, Literal: "now"},
						Value: "now",
					},
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "date"},
					Value: "date",
				},
			},
		},
		{
			input:         "db.table as t1",
			expectedQuery: "db.table AS t1",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &Identifier{
					Token: Token{Type: IDENT, Literal: "db"},
					Value: "db.table",
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "t1"},
					Value: "t1",
				},
			},
		},
		{
			input:         "name as nm",
			expectedQuery: "name AS nm",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &Identifier{
					Token: Token{Type: IDENT, Literal: "name"},
					Value: "name",
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "nm"},
					Value: "nm",
				},
			},
		},
		{
			input:         "1001 as ID",
			expectedQuery: "1001 AS ID",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &IntegerLiteral{
					Token: Token{Type: INT, Literal: "1001"},
					Value: 1001,
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "ID"},
					Value: "ID",
				},
			},
		},
		// all
		{
			input:         "*",
			expectedQuery: "*",
			expectedExp: &Identifier{
				Token: Token{Type: IDENT, Literal: "*"},
				Value: "*",
			},
		},
		{
			input:         "t1.*",
			expectedQuery: "t1.*",
			expectedExp: &Identifier{
				Token: Token{Type: IDENT, Literal: "t1"},
				Value: "t1.*",
			},
		},
		{
			input:         "`date` as `dt`",
			expectedQuery: "date AS dt",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &StringLiteral{
					Token: Token{Type: STRING, Literal: "date"},
					Value: "date",
				},
				Right: &StringLiteral{
					Token: Token{Type: STRING, Literal: "dt"},
					Value: "dt",
				},
			},
		},
		{
			input:         " sum(price) AS amount",
			expectedQuery: "sum(price) AS amount",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "AS"},
				Operator: SQLAs,
				Left: &CallExpression{
					Token: Token{Type: LPAREN, Literal: "("},
					Function: &Identifier{
						Token: Token{Type: IDENT, Literal: "sum"},
						Value: "sum",
					},
					Arguments: []Expression{
						&Identifier{
							Token: Token{Type: IDENT, Literal: "price"},
							Value: "price",
						},
					},
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "amount"},
					Value: "amount",
				},
			},
		},
		{
			input:         "(select 1) as id",
			expectedQuery: "(SELECT 1) AS id",
			expectedExp: &InfixExpression{
				Token:    Token{Type: SQLAs, Literal: "as"},
				Operator: SQLAs,
				Left: &SQLSubSelectExpression{
					Token: Token{Type: LPAREN, Literal: "("},
					Select: &SQLSelectStatement{
						Token: Token{Type: SQLSelect, Literal: "select"},
						SQLSelectColumns: []Expression{
							&IntegerLiteral{
								Token: Token{Type: INT, Literal: "1"},
								Value: 1,
							},
						},
					},
				},
				Right: &Identifier{
					Token: Token{Type: IDENT, Literal: "id"},
					Value: "id",
				},
			},
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		exp := p.parseSQLColumn()
		checkParserErrors(t, p)

		require.Equal(t, tt.expectedQuery, exp.String())
		require.EqualValuesf(t, tt.expectedExp, exp, "input: %s", tt.input)
	}
}

func TestParser_parseSQLSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
		expectedExp   Expression
	}{
		{
			input:         "date",
			expectedQuery: "date",
			expectedExp: &SQLSource{
				Token: Token{Type: IDENT, Literal: "date"},
				Value: "date",
			},
		},
		{
			input:         "now() as date",
			expectedQuery: "now() AS date",
			expectedExp: &SQLSource{
				Token: Token{Type: IDENT, Literal: "now"},
				Value: "now()",
				Alias: "date",
			},
		},
		{
			input:         "db.table as t1",
			expectedQuery: "db.table AS t1",
			expectedExp: &SQLSource{
				Token: Token{Type: IDENT, Literal: "db"},
				Value: "db.table",
				Alias: "t1",
			},
		},
		{
			input:         "name as nm",
			expectedQuery: "name AS nm",
			expectedExp: &SQLSource{
				Token: Token{Type: IDENT, Literal: "name"},
				Value: "name",
				Alias: "nm",
			},
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		exp := p.parseSQLSource()
		checkParserErrors(t, p)

		require.Equal(t, tt.expectedQuery, exp.String())
		require.EqualValuesf(t, tt.expectedExp, exp, "input: %s", tt.input)
	}
}

func TestParser_parseSQLSelect(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input  string
		output string
		isNil  bool
	}{
		{
			input: "SELECT 1",
			isNil: true,
		},
		{
			input:  "(SELECT 1)",
			output: "(SELECT 1)",
		},
		{
			input:  "(SELECT name from t)",
			output: "(SELECT name FROM t)",
		},
		{
			input: "(SELECT name from t",
			isNil: true,
		},
		{
			input:  "(SELECT name from t join t2 ON (a=b))",
			output: "(SELECT name FROM t JOIN t2 ON (a = b))",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			p := NewParser(NewLexer(tc.input))
			exp := p.parseSQLSubSelect()
			if tc.isNil {
				require.Nil(t, exp)
			} else {
				require.NotNilf(t, exp, "errors: %v", p.Errors())
				require.Equal(t, tc.output, exp.String())
			}
		})
	}
}

func TestParser_parseSQLSelectStatementMulti(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
	}{
		{
			input:         "Select (SELECT 1) AS nm",
			expectedQuery: "SELECT (SELECT 1) AS nm;",
		},
		{
			input:         "select (select max(price) from orders where orders.user_id = users.id) as max_price, id from users where id > 1",
			expectedQuery: "SELECT (SELECT max(price) FROM orders WHERE (orders.user_id = users.id)) AS max_price, id FROM users WHERE (id > 1);",
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		stmt := p.parseSQLSelectStatement()
		checkParserErrors(t, p)

		if !testSelectStatement(t, stmt, tt.expectedQuery) {
			return
		}
	}
}

func TestParser_parseSQLSelectStatementError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
	}{
		{
			input: "select * from t1 as 1",
		},
		{
			input: "select x as 3 fr om 1",
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		stmt := p.parseSQLSelectStatement()
		require.Truef(t, len(p.Errors()) != 0, "input: %s, %s", tt.input, stmt.String())
		_ = stmt
		t.Log(stmt.String())
		t.Log(p.errors)
	}
}

func TestParser_parseSQLSelectStatementWithJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
	}{
		{
			input:         "select * from t1 inner join t2 ON (t1.id= t2.id) WHERE t1.id > 100 order by date",
			expectedQuery: "SELECT * FROM t1 INNER JOIN t2 ON (t1.id = t2.id) WHERE (t1.id > 100) ORDER BY date;",
		},
		{
			input: "select * from t1 inner join t2 ON (t1.id= t2.id) left join t3 ON (t1.key = t3.key and t3.date > now()) " +
				"WHERE t1.id > 100 group by date",
			expectedQuery: "SELECT * FROM t1 INNER JOIN t2 ON (t1.id = t2.id) LEFT JOIN t3 ON ((t1.key = t3.key) AND (t3.date > now())) " +
				"WHERE (t1.id > 100) GROUP BY date;",
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		stmt := p.parseSQLSelectStatement()
		checkParserErrors(t, p)

		if !testSelectStatement(t, stmt, tt.expectedQuery) {
			return
		}
	}
}

func TestParser_parseSQLSelectStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
	}{
		{input: "Select;", expectedQuery: "SELECT;"},
		{input: "Select now();", expectedQuery: "SELECT now();"},
		{input: "Select now() as dt;", expectedQuery: "SELECT now() AS dt;"},
		{input: "Select name", expectedQuery: "SELECT name;"},
		{input: "Select id, name", expectedQuery: "SELECT id, name;"},
		{input: "Select id, name, `date` as `dt`;", expectedQuery: "SELECT id, name, date AS dt;"},
		{input: "Select id from table", expectedQuery: "SELECT id FROM table;"},
		{input: "select * from `users`", expectedQuery: "SELECT * FROM users;"},
		{input: "select t.* from `users` AS t", expectedQuery: "SELECT t.* FROM users AS t;"},
		{input: "select *,id AS \"ID\" from `users`", expectedQuery: "SELECT *, id AS ID FROM users;"},
		{input: "select name as nm from users", expectedQuery: "SELECT name AS nm FROM users;"},
		{input: "Select a.id, b.date as dt from table as a, users as b", expectedQuery: "SELECT a.id, b.date AS dt FROM table AS a, users AS b;"},
		{input: "select * from t WHERE id = 1", expectedQuery: "SELECT * FROM t WHERE (id = 1);"},
		{
			input:         "select * from t WHERE id = 1 and date > '2023-01-01' GROUP BY name, id ORDER BY name, id DESC",
			expectedQuery: "SELECT * FROM t WHERE ((id = 1) AND (date > 2023-01-01)) GROUP BY name, id ORDER BY name, id DESC;",
		},
		{input: "select * from t WHERE id = 1 LIMIT 10", expectedQuery: "SELECT * FROM t WHERE (id = 1) LIMIT 10;"},
		{input: "select * from t WHERE id = 1 LIMIT 5, 10", expectedQuery: "SELECT * FROM t WHERE (id = 1) LIMIT 5, 10;"},
		{
			input:         "select * from t where date between '2020-01-01' AND '2023-10-10' AND id > 3",
			expectedQuery: "SELECT * FROM t WHERE (date BETWEEN 2020-01-01 AND 2023-10-10 AND (id > 3));",
		},
		{
			input:         "select (select max(price) from orders) as max_price, name from users",
			expectedQuery: "SELECT (SELECT max(price) FROM orders) AS max_price, name FROM users;",
		},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		stmt := p.parseSQLSelectStatement()
		checkParserErrors(t, p)

		if !testSelectStatement(t, stmt, tt.expectedQuery) {
			return
		}
	}
}

func TestLetStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := NewLexer(tt.input)
		p := NewParser(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
		val := stmt.(*LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func testLiteralExpression(
	t *testing.T,
	exp Expression,
	expected any,
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	}

	t.Errorf("type of exp not handled. got=%T", exp)

	return false
}

func testIdentifier(t *testing.T, exp Expression, value string) bool {
	ident, ok := exp.(*Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il Expression, value int64) bool {
	integ, ok := il.(*IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp Expression, value bool) bool {
	bo, ok := exp.(*Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}
	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}
	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()

	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}

	t.Errorf("input: %s", p.l.input)

	t.FailNow()
}

func testSelectStatement(t *testing.T, s Statement, exp string) bool {
	t.Helper()

	if !strings.EqualFold(s.TokenLiteral(), "select") {
		t.Errorf("s.TokenLiteral not 'select'. got=%q", s.TokenLiteral())
		return false
	}

	require.Equal(t, exp, s.String())

	return true
}

func testLetStatement(t *testing.T, s Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}
	return true
}

func TestIntegerLiteralExpression(t *testing.T) {
	t.Parallel()

	input := "5;"
	l := NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	literal, ok := stmt.Expression.(*IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5",
			literal.TokenLiteral())
	}
}
