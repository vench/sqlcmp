package sqlcmp

import (
	"fmt"
	"strconv"
)

const (
	// operation priority
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
		EQ:          EQUALS,
		NotEq:       EQUALS,
		ASSIGN:      EQUALS,
		LT:          LESSGREATER,
		GT:          LESSGREATER,
		PLUS:        SUM,
		MINUS:       SUM,
		SLASH:       PRODUCT,
		ASTERISK:    PRODUCT,
		LPAREN:      CALL,
		LBRACKET:    INDEX,
		BinaryOr:    SUM,
		BinaryAnd:   SUM,
		BinarySlash: SUM,

		SQLOr:  LOWEST,
		SQLAnd: LOWEST,
		SQLAs:  EQUALS,
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
	p.registerPrefix(TRUE, p.parseBoolean)
	p.registerPrefix(FALSE, p.parseBoolean)
	p.registerPrefix(LPAREN, p.parseSQLGroupedCondition)
	p.registerPrefix(ASTERISK, p.parseAsterisk)
	p.registerPrefix(IF, p.parseIfExpression)
	p.registerPrefix(FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(STRING, p.parseStringLiteral)
	p.registerPrefix(LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(SETS, p.parseSetsLiteral)
	p.registerPrefix(LBRACE, p.parseSetsLiteralShort)
	p.registerPrefix(HASH, p.parseHashLiteral)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(BinaryOr, p.parseInfixExpression)
	p.registerInfix(BinaryAnd, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(BinarySlash, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
	p.registerInfix(EQ, p.parseInfixExpression)
	p.registerInfix(NotEq, p.parseInfixExpression)
	p.registerInfix(ASSIGN, p.parseInfixExpression)
	p.registerInfix(LT, p.parseInfixExpression)
	p.registerInfix(GT, p.parseInfixExpression)
	p.registerInfix(LPAREN, p.parseCallExpression)
	p.registerInfix(LBRACKET, p.parseIndexExpression)
	p.registerInfix(SQLAnd, p.parseInfixCondExpression)
	p.registerInfix(SQLOr, p.parseInfixCondExpression)
	p.registerInfix(SQLAs, p.parseInfixAsExpression)
	// DOT

	//nolint:gocritic
	// p.registerPrefix(STRING, p.parseSQLSource)

	//nolint:gocritic
	// p.registerPrefix(SQLSelect, p.parseSQLSelect)

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
		if stmt := p.parseStatement(); stmt != nil {
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

func (p *Parser) curTokenIs(t ...TokenType) bool {
	for i := range t {
		if p.curToken.Type == t[i] {
			return true
		}
	}

	return false
}

func (p *Parser) peekTokenIs(t ...TokenType) bool {
	for i := range t {
		if p.peekToken.Type == t[i] {
			return true
		}
	}

	return false
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

//nolint:funlen,gocyclo,gocritic
func (p *Parser) parseSQLSelectStatement() *SQLSelectStatement {
	stmt := &SQLSelectStatement{Token: p.curToken}
	p.nextToken()

	// parse columns
	for !p.curTokenIs(SEMICOLON, EOF, SQLFrom) {
		if p.curTokenIs(COMMA) {
			p.nextToken() // next arg
		}

		if v := p.parseSQLColumn(); v != nil {
			stmt.SQLSelectColumns = append(stmt.SQLSelectColumns, v)
		}

		p.nextToken()
	}

	// parse from
	if !p.curTokenIs(SQLFrom) {
		if !p.curTokenIs(SEMICOLON, EOF) {
			p.peekError(SEMICOLON)
		}

		return stmt
	}

	// skip from token
	p.nextToken()

	for !p.curTokenIs(SEMICOLON, EOF, SQLWhere, SQLGroup, SQLOrder, SQLLimit, SQLInner, SQLLeft, SQLRight, SQLCross, SQLJoin) {
		if p.curTokenIs(COMMA) {
			p.nextToken() // next table
		}
		if v := p.parseSQLFrom(); v != nil {
			stmt.From = append(stmt.From, v)
		}
		p.nextToken()
	}
	/*
		stmt.From = append(stmt.From, p.parseExpression(LOWEST))
		for p.peekTokenIs(COMMA) {
			p.nextToken()
			p.nextToken()
			stmt.From = append(stmt.From, p.parseExpression(LOWEST))
		}*/

	// parse join
	for p.curTokenIs(SQLInner, SQLLeft, SQLRight, SQLCross, SQLJoin) {
		exp := &SQLJoinExp{Token: Token{Type: SQLJoin}}
		for !p.curTokenIs(SQLJoin) { // get type
			exp.Type = p.curToken.Type
			p.nextToken()
		}

		if p.curTokenIs(SQLOuter) { // skip outer
			p.nextToken()
		}

		if !p.curTokenIs(SQLJoin) {
			p.peekError(SQLJoin)

			return nil
		}
		// skip join token
		p.nextToken()

		// get source
		if v := p.parseSQLSource(); v != nil {
			exp.Table = v
		} else {
			p.peekError(SQLFrom) // TODO: special case error
			return nil
		}

		if p.curTokenIs(SQLOn) { // parse cond
			p.nextToken()

			for !p.curTokenIs(SEMICOLON, EOF, SQLOrder, SQLGroup, SQLLimit, SQLWhere, RPAREN, SQLInner, SQLLeft, SQLRight, SQLCross, SQLJoin) {
				if cond := p.parseSQLCondition(); cond != nil {
					exp.Cond = append(exp.Cond, cond)
				}
				p.nextToken()
			}
		}

		stmt.Join = append(stmt.Join, exp)
	}

	// parse where
	if p.curTokenIs(SQLWhere) {
		p.nextToken()

		for !p.curTokenIs(SEMICOLON, EOF, SQLOrder, SQLGroup, SQLLimit, RPAREN) {
			if cond := p.parseSQLCondition(); cond != nil {
				stmt.Cond = append(stmt.Cond, cond)
			}
			p.nextToken()
		}
	}

	if p.curTokenIs(SQLGroup) {
		if !p.peekTokenIs(SQLBy) {
			p.peekError(SQLBy)
			return nil
		}

		p.nextToken()
		p.nextToken()

		for !p.curTokenIs(SEMICOLON, EOF, SQLOrder, SQLLimit) {
			if p.curTokenIs(COMMA) {
				p.nextToken() // next arg
			}
			if v := p.parseSQLSource(); v != nil {
				stmt.Group = append(stmt.Group, v)
			}
		}
	}

	if p.curTokenIs(SQLOrder) {
		if !p.peekTokenIs(SQLBy) {
			p.peekError(SQLBy)
			return nil
		}
		p.nextToken()
		p.nextToken()

		for !p.curTokenIs(SEMICOLON, EOF, SQLLimit) {
			if p.curTokenIs(COMMA) {
				p.nextToken() // next arg
			}
			if v := p.parseSQLOrder(); v != nil {
				stmt.Order = append(stmt.Order, v)
			}
		}
	}

	if p.curTokenIs(SQLLimit) {
		p.nextToken()

		limit := p.parseIntegerLiteral()
		p.nextToken()

		if p.curTokenIs(COMMA) {
			p.nextToken()
			stmt.Offset = limit
			stmt.Limit = p.parseIntegerLiteral()
			p.nextToken()
		} else {
			stmt.Limit = limit
		}
	}

	if !p.peekTokenIs(SEMICOLON, EOF) {
		p.peekError(SEMICOLON)

		return nil
	}

	return stmt
}

func (p *Parser) parseSQLCondition() Expression {
	prefixes := p.prefixParseFns[p.curToken.Type]
	if len(prefixes) == 0 {
		p.noPrefixParseFnError(p.curToken)

		return nil
	}

	for _, prefix := range prefixes {
		leftExp := prefix()

		if leftExp == nil {
			continue
		}

		// try parse infix
		for !p.peekTokenIs(SEMICOLON, EOF) && LOWEST <= p.peekPrecedence() {
			infix := p.infixParseFns[p.peekToken.Type]
			if infix == nil {
				return &SQLCondition{Expression: leftExp}
			}

			p.nextToken()
			leftExp = infix(leftExp)
		}

		return &SQLCondition{Expression: leftExp}
	}

	return nil
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
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
		p.noPrefixParseFnError(p.curToken)

		return nil
	}

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
	exp := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(DOT) {
		p.nextToken()

		// @todo: maybe DOT like infix
		exp.Value += DOT.String()

		if !p.peekTokenIs(IDENT, ASTERISK) {
			p.peekError(IDENT)
			return nil
		}
		p.nextToken()

		exp.Value += p.curToken.Literal
	}

	return exp
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

func (p *Parser) noPrefixParseFnError(t Token) {
	msg := fmt.Sprintf(
		"no prefix parse function for %s found, literal: %s, cur token: %s",
		t.Type.String(), t.Literal, p.curToken.Literal)

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
	if v, ok := precedences[p.peekToken.Type]; ok {
		return v
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if operator, ok := precedences[p.curToken.Type]; ok {
		return operator
	}

	return LOWEST
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	defer untrace(trace("parseInfixExpression"))

	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Type,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()

	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() Expression {
	return &Boolean{Token: p.curToken, Value: p.curTokenIs(TRUE)}
}

func (p *Parser) parseAsterisk() Expression {
	return &Identifier{
		Token: Token{Type: IDENT, Literal: p.curToken.Literal},
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseSQLGroupedCondition() Expression {
	p.nextToken()
	exp := p.parseSQLCondition()
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
	for !p.curTokenIs(RBRACE, EOF) {
		if stmt := p.parseStatement(); stmt != nil {
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

//nolint:unused
func (p *Parser) parseCallArguments() []Expression {
	var args []Expression
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

func (p *Parser) parseSQLFrom() Expression {
	exp := p.parseExpression(LOWEST)
	// validate
	switch tp := exp.(type) {
	case *InfixExpression:
		if _, ok := tp.Left.(*IntegerLiteral); ok {
			p.addError("left node is integer")

			return exp
		}

		if _, ok := tp.Right.(*IntegerLiteral); ok {
			p.addError("right node is integer")

			return exp
		}
	}

	return exp
}

func (p *Parser) parseSQLColumn() Expression {
	exp := p.parseExpression(LOWEST)

	// validate
	switch tp := exp.(type) {
	case *InfixExpression:
		if _, ok := tp.Right.(*IntegerLiteral); ok {
			p.addError("right node is integer")

			return exp
		}
	}

	return exp
}

func (p *Parser) parseSQLSource() Expression {
	stopTokens := []TokenType{
		COMMA, EOF, SQLFrom, SEMICOLON, SQLWhere, SQLGroup, SQLOrder, SQLLimit, SQLInner, SQLLeft, SQLRight, SQLCross, SQLJoin, SQLOn,
	}

	col := &SQLSource{Token: p.curToken, Value: ""}
	col.Value += p.curToken.Literal
	var alias bool

	for !p.peekTokenIs(stopTokens...) {
		p.nextToken()

		if p.curTokenIs(SQLAs) {
			alias = true
			continue
		}

		if alias {
			if col.Alias == "" && p.curTokenIs(INT) {
				p.peekError(INT)

				return nil
			}

			col.Alias += p.curToken.Literal
		} else {
			col.Value += p.curToken.Literal
		}
	}

	p.nextToken()

	return col
}

func (p *Parser) parseSQLOrder() Expression {
	col := &SQLOrderExp{Token: p.curToken, Value: ""}
	col.Value += p.curToken.Literal

	for !p.peekTokenIs(COMMA, EOF, SEMICOLON, SQLLimit) {
		p.nextToken()

		if p.curTokenIs(SQLAsc, SQLDesc) {
			col.Direction = p.curToken
			break
		}

		col.Value += p.curToken.Literal
	}

	p.nextToken()

	return col
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
	var list []Expression
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

//nolint:gocritic
func (p *Parser) parseInfixAsExpression(left Expression) Expression {
	// panic("parseInfixAsExpression")
	if exp := p.parseInfixExpression(left); exp != nil {
		// e := exp.(*InfixExpression)
		// panic(reflect.TypeOf(exp))
		return exp
	}

	return nil
}

func (p *Parser) parseInfixCondExpression(left Expression) Expression {
	if exp := p.parseInfixExpression(left); exp != nil {
		return &SQLCondition{
			Expression: exp,
		}
	}

	return nil
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
