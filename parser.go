package sqlcmp

import (
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	// ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var (
	precedences = map[TokenType]int{
		EQ:           EQUALS,
		NOT_EQ:       EQUALS,
		LT:           LESSGREATER,
		GT:           LESSGREATER,
		PLUS:         SUM,
		MINUS:        SUM,
		SLASH:        PRODUCT,
		ASTERISK:     PRODUCT,
		LPAREN:       CALL,
		LBRACKET:     INDEX,
		BINARY_OR:    SUM,
		BINARY_AND:   SUM,
		BINARY_SLASH: SUM,
	}

	showEnteringLeaving = false
)

type (
	prefixParseFn func() Expression
	infixParseFn  func(a Expression) Expression
)

type Parser struct {
	l              *Lexer
	curToken       Token
	peekToken      Token
	errors         []string
	prefixParseFns map[TokenType][]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[TokenType][]prefixParseFn)
	p.registerPrefix(IDENT, p.parseIdentifier)
	p.registerPrefix(INT, p.parseIntegerLiteral)
	p.registerPrefix(BANG, p.parsePrefixExpression)
	p.registerPrefix(MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(BINARY_OR, p.parseInfixExpression)
	p.registerInfix(BINARY_AND, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(BINARY_SLASH, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
	p.registerInfix(EQ, p.parseInfixExpression)
	p.registerInfix(NOT_EQ, p.parseInfixExpression)
	p.registerInfix(LT, p.parseInfixExpression)
	p.registerInfix(GT, p.parseInfixExpression)

	p.registerPrefix(TRUE, p.parseBoolean)
	p.registerPrefix(FALSE, p.parseBoolean)

	p.registerPrefix(LPAREN, p.parseGroupedExpression)
	p.registerPrefix(IF, p.parseIfExpression)
	p.registerPrefix(FUNCTION, p.parseFunctionLiteral)
	p.registerInfix(LPAREN, p.parseCallExpression)
	p.registerPrefix(STRING, p.parseStringLiteral)
	p.registerPrefix(STRING, p.parseSQLColumn)
	p.registerPrefix(LBRACKET, p.parseArrayLiteral)
	p.registerInfix(LBRACKET, p.parseIndexExpression)

	p.registerPrefix(SETS, p.parseSetsLiteral)
	p.registerPrefix(LBRACE, p.parseSetsLiteralShort)
	p.registerPrefix(HASH, p.parseHashLiteral)

	p.registerPrefix(SQLSelect, p.parseSQLSelect)

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}
	for p.curToken.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case LET:
		return p.parseLetStatement()
	case RETURN:
		return p.parseReturnStatement()
	case SQLSelect:
		return p.parseSQLSelectStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *LetStatement {
	stmt := &LetStatement{Token: p.curToken}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()

		return true
	}

	p.peekError(t)

	return false
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSQLSelectStatement() *SQLSelectStatement {
	stmt := &SQLSelectStatement{Token: p.curToken}
	p.nextToken()

	v := p.parseExpression(LOWEST)
	stmt.SQLSelectValue = append(stmt.SQLSelectValue, v)
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
		// end expression
		return stmt
	}

	if p.peekTokenIs(COMMA) {
		p.nextToken() // next arg
		p.nextToken()

		v2 := p.parseExpression(LOWEST)
		stmt.SQLSelectValue = append(stmt.SQLSelectValue, v2)

	}

	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) Errors() []string {
	return p.errors
}
func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = append(p.prefixParseFns[tokenType], fn)
}

func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))

	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	defer untrace(trace("parseExpression"))

	prefixs := p.prefixParseFns[p.curToken.Type]
	if len(prefixs) == 0 {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	//fmt.Println(p.curToken.Literal)
	for _, prefix := range prefixs {
		leftExp := prefix()

		if leftExp == nil {
			continue
		}
		for !p.peekTokenIs(SEMICOLON) && precedence < p.peekPrecedence() {
			infix := p.infixParseFns[p.peekToken.Type]
			if infix == nil {
				return leftExp
			}
			p.nextToken()
			leftExp = infix(leftExp)
		}
		return leftExp
	}
	return nil
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	defer untrace(trace("parseIntegerLiteral"))

	lit := &IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) noPrefixParseFnError(t TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixExpression() Expression {
	defer untrace(trace("parsePrefixExpression"))

	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	defer untrace(trace("parseInfixExpression"))

	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	//if expression.Operator == "+" {
	//	expression.Right = p.parseExpression(precedence - 1)
	//} else {
	expression.Right = p.parseExpression(precedence)
	//}
	return expression
}

func (p *Parser) parseBoolean() Expression {
	return &Boolean{Token: p.curToken, Value: p.curTokenIs(TRUE)}
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() Expression {
	expression := &IfExpression{Token: p.curToken}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(ELSE) {
		p.nextToken()
		if !p.expectPeek(LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}
	p.nextToken()
	for !p.curTokenIs(RBRACE) && !p.curTokenIs(EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseSQLSelect() Expression {
	lit := &SelectLiteral{Token: p.curToken}
	p.nextToken()

	exp := p.parseSQLColumn()
	lit.Columns = append(lit.Columns, exp)

	return lit
}

func (p *Parser) parseFunctionParameters() []*Identifier {
	var identifiers []*Identifier
	if p.peekTokenIs(RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()

	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)
	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !p.expectPeek(RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(RPAREN)
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}
	if p.peekTokenIs(RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return args
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseSQLColumn() Expression {
	return &SQLColumn{StringLiteral: StringLiteral{Token: p.curToken, Value: p.curToken.Literal}}
}

func (p *Parser) parseArrayLiteral() Expression {
	array := &ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(RBRACKET)
	return array
}

func (p *Parser) parseSetsLiteralShort() Expression {
	sets := &SetsLiteral{Token: p.curToken}
	sets.Elements = p.parseExpressionList(RBRACE)
	return sets
}

func (p *Parser) parseSetsLiteral() Expression {
	sets := &SetsLiteral{Token: p.curToken}
	p.nextToken() // skeep {
	sets.Elements = p.parseExpressionList(RBRACE)
	return sets
}

func (p *Parser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseHashLiteral() Expression {
	hash := &HashLiteral{Token: p.curToken}
	p.nextToken() // skeep {
	hash.Pairs = make(map[Expression]Expression)
	for !p.peekTokenIs(RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(RBRACE) && !p.expectPeek(COMMA) {
			return nil
		}
	}
	if !p.expectPeek(RBRACE) {
		return nil
	}
	return hash
}

// util
func trace(s string) string {
	if showEnteringLeaving {
		fmt.Println("entering:", s)
	}
	return s
}

func untrace(s string) {
	if showEnteringLeaving {
		fmt.Println("leaving:", s)
	}
}
