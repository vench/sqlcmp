package sqlcmp

import (
	"bytes"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// SQLStructcher todo,
type SQLStructcher interface {
	Structcher() string
}

// Program this structure represents a program as a list of statements.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}

// String
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// LetStatement todo.
type LetStatement struct {
	Token Token // the LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// Identifier todo.
type Identifier struct {
	Token Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// ReturnStatement todo.
type ReturnStatement struct {
	Token       Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// SQLSelectStatement todo.
type SQLSelectStatement struct {
	Token            Token // the 'select' token
	SQLSelectColumns []Expression
	From             []Expression
	Join             []Expression
	Cond             []Expression
	Order            []Expression
	Group            []Expression

	Offset Expression
	Limit  Expression
}

func (rs *SQLSelectStatement) statementNode()       {}
func (rs *SQLSelectStatement) TokenLiteral() string { return rs.Token.Literal }

// func (rs *SQLSelectStatement) Structcher() string { }

func (rs *SQLSelectStatement) toString(skipSemicolon bool) string {
	var out bytes.Buffer
	out.WriteString(SQLSelect.String())

	if rs.SQLSelectColumns != nil {
		for i := range rs.SQLSelectColumns {
			if i != 0 {
				out.WriteString(",")
			}
			out.WriteString(" ")

			out.WriteString(rs.SQLSelectColumns[i].String())
		}
	}

	if rs.From != nil {
		out.WriteString(" " + SQLFrom.String())

		for i := range rs.From {
			if i != 0 {
				out.WriteString(",")
			}

			out.WriteString(" ")
			out.WriteString(rs.From[i].String())
		}
	}

	if rs.Join != nil {
		for i := range rs.Join {
			out.WriteString(" ")
			out.WriteString(rs.Join[i].String())
		}
	}

	if rs.Cond != nil {
		out.WriteString(" " + SQLWhere.String())

		for i := range rs.Cond {
			if i != 0 {
				out.WriteString(",")
			}

			out.WriteString(" ")
			out.WriteString(rs.Cond[i].String())
		}
	}

	if rs.Group != nil {
		out.WriteString(" " + SQLGroup.String() + " " + SQLBy.String())

		for i := range rs.Group {
			if i != 0 {
				out.WriteString(",")
			}

			out.WriteString(" ")
			out.WriteString(rs.Group[i].String())
		}
	}

	if rs.Order != nil {
		out.WriteString(" " + SQLOrder.String() + " " + SQLBy.String())

		for i := range rs.Order {
			if i != 0 {
				out.WriteString(",")
			}

			out.WriteString(" ")
			out.WriteString(rs.Order[i].String())
		}
	}

	if rs.Limit != nil {
		if rs.Offset != nil {
			out.WriteString(" " + SQLLimit.String() + " " + rs.Offset.String() + ", " + rs.Limit.String())
		} else {
			out.WriteString(" " + SQLLimit.String() + " " + rs.Limit.String())
		}
	}

	if skipSemicolon {
		out.WriteString(";")
	}

	return out.String()
}

func (rs *SQLSelectStatement) String() string {
	return rs.toString(true)
}

// ExpressionStatement todo.
type ExpressionStatement struct {
	Token      Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

// IntegerLiteral todo.
type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// PrefixExpression todo.
type PrefixExpression struct {
	Token    Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression todo.
type InfixExpression struct {
	Token    Token // The operator token, e.g. +
	Left     Expression
	Operator TokenType
	Right    Expression
}

func (oe *InfixExpression) Structcher() string {
	var out bytes.Buffer

	if oe.Operator != SQLAs {
		out.WriteString("(")
	}

	out.WriteString(structcher(oe.Left))
	out.WriteString(" " + oe.Operator.String() + " ")
	out.WriteString(structcher(oe.Right))

	if oe.Operator != SQLAs {
		out.WriteString(")")
	}

	return out.String()
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	if oe.Operator != SQLAs {
		out.WriteString("(")
	}

	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator.String() + " ")
	out.WriteString(oe.Right.String())

	if oe.Operator != SQLAs {
		out.WriteString(")")
	}

	return out.String()
}

// Boolean todo.
type Boolean struct {
	Token Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// IfExpression todo.
type IfExpression struct {
	Token       Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

type BetweenExpression struct {
	Token  Token // The 'in' token
	Column Expression
	From   Expression
	To     Expression
}

func (ce *BetweenExpression) Structcher() string {
	var out bytes.Buffer
	out.WriteString(ce.Column.String() + " ")
	out.WriteString(SQLBetween.String() + " ? AND ?")

	return out.String()
}

func (ce *BetweenExpression) expressionNode()      {}
func (ce *BetweenExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *BetweenExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Column.String() + " ")
	out.WriteString(SQLBetween.String() + " ")
	out.WriteString(ce.From.String())
	out.WriteString(" AND ")
	out.WriteString(ce.To.String())

	return out.String()
}

// InExpression todo.
type InExpression struct {
	Token     Token // The 'in' token
	Column    Expression
	Arguments []Expression
}

func (ce *InExpression) Structcher() string {
	var out bytes.Buffer
	out.WriteString(ce.Column.String() + " ")
	out.WriteString(SQLIn.String() + " ")
	out.WriteString("(?)")

	return out.String()
}

func (ce *InExpression) expressionNode()      {}
func (ce *InExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *InExpression) String() string {
	var out bytes.Buffer
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Column.String() + " ")
	out.WriteString(SQLIn.String() + " ")
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// BlockStatement todo.
type BlockStatement struct {
	Token      Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// FunctionLiteral todo.
type FunctionLiteral struct {
	Token      Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	var params []string
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// SelectLiteral todo.
type SelectLiteral struct {
	Token   Token // The 'select' token
	Columns []Expression
	From    *BlockStatement
	Where   []*Identifier
	GroupBy *Identifier
	OrderBy *Identifier
}

func (fl *SelectLiteral) expressionNode()      {}
func (fl *SelectLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *SelectLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("SELECT *...")
	return out.String()
}

// CallExpression todo.
type CallExpression struct {
	Token    Token // The '(' token
	Function Expression
	// Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// StringLiteral todo.
type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// SQLSource that structure represents representations of a different source in the DB.
type SQLSource struct {
	Token Token
	Value string
	Alias string
}

func (sl *SQLSource) expressionNode()      {}
func (sl *SQLSource) TokenLiteral() string { return sl.Token.Literal }
func (sl *SQLSource) String() string {
	if sl.Alias != "" {
		return sl.Value + " AS " + sl.Alias
	}
	return sl.Value
}

// SQLOrderExp todo.
type SQLOrderExp struct {
	Token     Token
	Value     string
	Direction Token
}

func (sl *SQLOrderExp) expressionNode()      {}
func (sl *SQLOrderExp) TokenLiteral() string { return sl.Token.Literal }
func (sl *SQLOrderExp) String() string {
	if sl.Direction.Literal != "" {
		return sl.Value + " " + sl.Direction.Type.String()
	}

	return sl.Value
}

// SQLJoinExp todo.
type SQLJoinExp struct {
	Token Token
	Type  TokenType
	Table Expression
	Cond  []Expression
}

func (sl *SQLJoinExp) expressionNode()      {}
func (sl *SQLJoinExp) TokenLiteral() string { return sl.Token.Literal }
func (sl *SQLJoinExp) String() string {
	str := ""
	if sl.Type.String() != "" {
		str = sl.Type.String() + " "
	}

	str += SQLJoin.String() + " " + sl.Table.String()
	if sl.Cond != nil {
		str += " ON "
		for i := range sl.Cond {
			str += sl.Cond[i].String()
		}
	}

	return str
}

// SQLCondition wrapper for Expression.
type SQLCondition struct {
	Expression Expression
}

func (sl *SQLCondition) Structcher() string {
	return structcher(sl.Expression)
}

func (sl *SQLCondition) expressionNode()      {}
func (sl *SQLCondition) TokenLiteral() string { return sl.Expression.TokenLiteral() }
func (sl *SQLCondition) String() string       { return sl.Expression.String() }

func structcher(exp Expression) string {
	if exp == nil {
		return ""
	}

	switch tp := exp.(type) {
	case *IntegerLiteral:
		return "?"
	case *StringLiteral:
		return "?"
	case SQLStructcher:
		return tp.Structcher()
	}

	return exp.String()
}

// ArrayLiteral todo.
type ArrayLiteral struct {
	Token    Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	var elements []string
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// IndexExpression todo.
type IndexExpression struct {
	Token Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// DotExpression todo.
type DotExpression struct {
	Token Token // The . token
	Left  Expression
	Right Expression
}

func (ie *DotExpression) expressionNode()      {}
func (ie *DotExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *DotExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Left.String())
	out.WriteString(".")
	out.WriteString(ie.Right.String())
	return out.String()
}

// SQLSubSelectExpression todo.
type SQLSubSelectExpression struct {
	Token  Token // The ( token
	Select *SQLSelectStatement
}

func (ie *SQLSubSelectExpression) expressionNode()      {}
func (ie *SQLSubSelectExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *SQLSubSelectExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Select.toString(false))
	out.WriteString(")")
	return out.String()
}
