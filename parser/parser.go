package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	// 字句解析器インスタンスへのポインタ
	l *lexer.Lexer

	// 現トークン
	curToken token.Token
	// 次トークン
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.nextToken() // curToken()
	p.nextToken() // peekToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program = newProgramASTNode()

	advanceTokens()

	for currentToken() != EOF_TOKEN {
		statement = parseLetStatement()
	}
	return nil
}
