package sqlcmp

import (
	"testing"
)

func TestNextTokenSQLSelect(t *testing.T) {
	t.Parallel()

	input := `SELECT id, name, date as dt, now() as dt2`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SQLSelect, "SELECT"},
		{IDENT, "id"},
		{COMMA, ","},
		{IDENT, "name"},
		{COMMA, ","},
		{IDENT, "date"},
		{SQLAs, "as"},
		{IDENT, "dt"},
		{COMMA, ","},
		{IDENT, "now"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{SQLAs, "as"},
		{IDENT, "dt2"},
		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenSQL(t *testing.T) {
	t.Parallel()

	input := `SELECT * 
from table1 
where id=1 
  AND (name LIKE '%ab"c_' OR email LIKE "xx'xx")  
GROUP BY id,name
ORDER BY name`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SQLSelect, "SELECT"},
		{ASTERISK, "*"},
		{SQLFrom, "from"},
		{IDENT, "table1"},
		// where
		{SQLWhere, "where"},
		{IDENT, "id"},
		{ASSIGN, "="},
		{INT, "1"},
		{SQLAnd, "AND"},

		{LPAREN, "("},
		{IDENT, "name"},
		{SQLLike, "LIKE"},
		{STRING, "%ab\"c_"},
		{SQLOr, "OR"},
		{IDENT, "email"},
		{SQLLike, "LIKE"},
		{STRING, "xx'xx"},
		{RPAREN, ")"},
		// group
		{SQLGroup, "GROUP"},
		{SQLBy, "BY"},
		{IDENT, "id"},
		{COMMA, ","},
		{IDENT, "name"},

		// order
		{SQLOrder, "ORDER"},
		{SQLBy, "BY"},
		{IDENT, "name"},

		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken(t *testing.T) {
	t.Parallel()

	input := `let five = 5;
let ten = 10;
let add = fn(x, y) {
x + y;
};
let result = add(five, ten);
!-/*5;
5 < 10 > 5;
if (5 < 10) {
return true;
} else {
return false;
}
10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
set{1, 2, 4}
set{1, 2, 4} | set{1, 2, 4}
set{1, 2, 4} & set{1, 2, 4}
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LET, "let"},
		{IDENT, "five"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "ten"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "add"},
		{ASSIGN, "="},
		{FUNCTION, "fn"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COMMA, ","},
		{IDENT, "y"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "five"},
		{COMMA, ","},
		{IDENT, "ten"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},

		{BANG, "!"},
		{MINUS, "-"},
		{SLASH, "/"},
		{ASTERISK, "*"},
		{INT, "5"},
		{SEMICOLON, ";"},

		{INT, "5"},
		{LT, "<"},
		{INT, "10"},
		{GT, ">"},
		{INT, "5"},
		{SEMICOLON, ";"},

		{IF, "if"},
		{LPAREN, "("},
		{INT, "5"},
		{LT, "<"},
		{INT, "10"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{TRUE, "true"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},

		{ELSE, "else"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{FALSE, "false"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},

		{INT, "10"},
		{EQ, "=="},
		{INT, "10"},
		{SEMICOLON, ";"},

		{INT, "10"},
		{NotEq, "!="},
		{INT, "9"},
		{SEMICOLON, ";"},

		{STRING, "foobar"},
		{STRING, "foo bar"},
		{LBRACKET, "["},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{RBRACKET, "]"},
		{SEMICOLON, ";"},

		{LBRACE, "{"},
		{STRING, "foo"},
		{COLON, ":"},
		{STRING, "bar"},
		{RBRACE, "}"},

		{SETS, "set"},
		{LBRACE, "{"},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "4"},
		{RBRACE, "}"},

		//
		{SETS, "set"},
		{LBRACE, "{"},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "4"},
		{RBRACE, "}"},
		{BinaryOr, "|"},
		{SETS, "set"},
		{LBRACE, "{"},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "4"},
		{RBRACE, "}"},
		//
		{SETS, "set"},
		{LBRACE, "{"},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "4"},
		{RBRACE, "}"},
		{BinaryAnd, "&"},
		{SETS, "set"},
		{LBRACE, "{"},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "4"},
		{RBRACE, "}"},

		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
