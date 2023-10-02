package sqlcmp

func (p *Parser) parseInfixBetweenExpression(left Expression) Expression {
	if !p.curTokenIs(SQLBetween) {
		p.addError("check Between")
		return nil
	}

	exp := &BetweenExpression{
		Token: p.curToken, Column: left,
	}

	p.nextToken() // skip between

	exp.From = p.parseIdentifier()

	if !p.peekTokenIs(SQLAnd) {
		p.addError("check And")
		return nil
	}
	p.nextToken()
	p.nextToken()
	exp.To = p.parseIdentifier()

	return exp
}

func (p *Parser) parseInfixInExpression(left Expression) Expression {
	if !p.curTokenIs(SQLIn) {
		p.addError("check IN")
		return nil
	}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	exp := &InExpression{Token: p.curToken, Column: left}

	exp.Arguments = p.parseExpressionList(RPAREN)

	return exp
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

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(RPAREN)
	return exp
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
