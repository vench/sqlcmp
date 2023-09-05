package sqlcmp

import "strings"

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"
	// 1343456

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	SETS     = "SET"
	HASH     = "HASH"
	// SQL
	SQLSelect = "SELECT"
	SQLFrom   = "FROM"
	SQLWhere  = "WHERE"
	SQLAnd    = "AND"
	SQLOr     = "OR"
	SQLLike   = "LIKE"
	SQLOrder  = "ORDER"
	SQLGroup  = "GROUP"
	SQLBy     = "BY"
	SQLAs     = "AS"

	// Operators
	ASSIGN      = "="
	PLUS        = "+"
	MINUS       = "-"
	BANG        = "!"
	ASTERISK    = "*"
	SLASH       = "/"
	BinarySlash = "\\"
	LT          = "<"
	GT          = ">"
	EQ          = "=="
	NotEq       = "!="
	STRING      = "STRING"
	LBRACKET    = "["
	RBRACKET    = "]"
	COLON       = ":"
	BinaryOr    = "|"
	BinaryAnd   = "&"
)

type TokenType string

// Type - is allow type
// Literal - parse string in source code belong current Type
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
	"select": SQLSelect,
	"from":   SQLFrom,
	"where":  SQLWhere,
	"and":    SQLAnd,
	"or":     SQLOr,
	"like":   SQLLike,
	"order":  SQLOrder,
	"group":  SQLGroup,
	"by":     SQLBy,
	"as":     SQLAs,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}

	return IDENT
}
