package sqlcmp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSQLSelectStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         string
		expectedQuery string
		expectedValue any
	}{
		{input: "Select;", expectedQuery: "SELECT;"},
		{input: "Select now();", expectedQuery: "SELECT now();"},
		{input: "Select now() as dt;", expectedQuery: "SELECT now() AS dt;"},
		{input: "Select name", expectedQuery: "SELECT name;"},
		{input: "Select id, name", expectedQuery: "SELECT id, name;"},
		{input: "Select id, name, `date` as `dt`;", expectedQuery: "SELECT id, name, `date` AS `dt`;"},
		{input: "Select id from table", expectedQuery: "SELECT id FROM table;"},
		{input: "select * from `users`", expectedQuery: "SELECT * FROM `users`;"},
		{input: "select t.* from `users` AS t", expectedQuery: "SELECT t.* FROM `users` AS t;"},
		{input: "select *,id AS \"ID\" from `users`", expectedQuery: "SELECT *, id AS ID FROM `users`;"},
		{input: "select name as nm from users", expectedQuery: "SELECT name AS nm FROM users;"},
		{input: "Select a.id, b.date as dt from table as a, users as b", expectedQuery: "SELECT a.id, b.date AS dt FROM table AS a, users AS b;"},
	}

	for _, tt := range tests {
		p := NewParser(NewLexer(tt.input))

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]

		if !testSelectStatement(t, stmt, tt.expectedQuery) {
			return
		}

		val, ok := stmt.(*SQLSelectStatement)
		require.True(t, ok)

		t.Log(val.String())
		//nolint:gocritic
		// if !testLiteralExpression(t, val, tt.expectedValue) {
		//	return
		// }
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

//nolint:unused
func testInfixExpression(t *testing.T, exp Expression, left any,
	operator string, right any) bool {
	opExp, ok := exp.(*InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
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