package sqlcmp

import "strings"

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// List of Identifiers.

	IDENT TokenType = "IDENT" // default TokenType like: add, foobar, x, y, ...
	INT   TokenType = "INT"

	// List of delimiters.

	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	DOT       TokenType = "."

	// List of keywords.

	FUNCTION TokenType = "FUNCTION"
	LET      TokenType = "LET"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	SETS     TokenType = "SET"
	HASH     TokenType = "HASH"

	// List of SQL allow tokens.

	SQLSelect  TokenType = "SELECT"
	SQLFrom    TokenType = "FROM"
	SQLWhere   TokenType = "WHERE"
	SQLAnd     TokenType = "AND"
	SQLOr      TokenType = "OR"
	SQLLike    TokenType = "LIKE"
	SQLOrder   TokenType = "ORDER"
	SQLGroup   TokenType = "GROUP"
	SQLBy      TokenType = "BY"
	SQLAs      TokenType = "AS"
	SQLDesc    TokenType = "DESC"
	SQLAsc     TokenType = "ASC"
	SQLLimit   TokenType = "LIMIT"
	SQLJoin    TokenType = "JOIN"
	SQLInner   TokenType = "INNER"
	SQLLeft    TokenType = "LEFT"
	SQLRight   TokenType = "RIGHT"
	SQLOuter   TokenType = "OUTER"
	SQLCross   TokenType = "CROSS"
	SQLOn      TokenType = "ON"
	SQLNot     TokenType = "NOT"
	SQLIn      TokenType = "IN"
	SQLBetween TokenType = "BETWEEN"

	// List of allow operators.

	ASSIGN      TokenType = "="
	PLUS        TokenType = "+"
	MINUS       TokenType = "-"
	BANG        TokenType = "!"
	ASTERISK    TokenType = "*"
	SLASH       TokenType = "/"
	BinarySlash TokenType = "\\"
	LT          TokenType = "<"
	GT          TokenType = ">"
	EQ          TokenType = "=="
	NotEq       TokenType = "!="
	STRING      TokenType = "STRING"
	LBRACKET    TokenType = "["
	RBRACKET    TokenType = "]"
	COLON       TokenType = ":"
	BinaryOr    TokenType = "|"
	BinaryAnd   TokenType = "&"
)

type TokenType string

func (t TokenType) String() string {
	return string(t)
}

// Token todo.
type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"set":    SETS,
	"S":      SETS,
	"HASH":   HASH,
	"hash":   HASH,
	"H":      HASH,

	"select":  SQLSelect,
	"from":    SQLFrom,
	"where":   SQLWhere,
	"and":     SQLAnd,
	"or":      SQLOr,
	"like":    SQLLike,
	"order":   SQLOrder,
	"group":   SQLGroup,
	"by":      SQLBy,
	"as":      SQLAs,
	"desc":    SQLDesc,
	"asc":     SQLAsc,
	"limit":   SQLLimit,
	"join":    SQLJoin,
	"left":    SQLLeft,
	"right":   SQLRight,
	"inner":   SQLInner,
	"cross":   SQLCross,
	"outer":   SQLOuter,
	"on":      SQLOn,
	"not":     SQLNot,
	"in":      SQLIn,
	"between": SQLBetween,
}

// LookupIdent converts string to TokenType.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}

	return IDENT
}
