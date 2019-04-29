package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	// 字句解析器インスタンスへのポインタ
	l *lexer.Lexer

	// エラー
	errors []string

	// 現トークン
	curToken token.Token
	// 次トークン
	peekToken token.Token
}

// Parserのコンストラクタ
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// 2つトークンを読み込む。curTokenとpeekTokenがセットされる
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// curTokenとpeekTokenを進めるヘルパーメソッド
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseしてASTを生成するプログラム
func (p *Parser) ParseProgram() *ast.Program {
	// ASTルートノード生成
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// 終端まで入力トークンを繰り返し読み出し
	for !p.curTokenIs(token.EOF) {
		// 文の構文解析
		stmt := p.parseStatement()
		// 対象が含まれていたらstatementに追加
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// ポインタを進める
		p.nextToken()
	}
	return program
}

// 文の構文解析
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		// LETトークンの構文解析
		return p.parseLetStatement()
	case token.RETURN:
		// RETURNトークンの構文解析
		return p.parseReturnStatement()
	default:
		return nil
	}
}

// LETトークンの構文解析
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	// 次トークンが識別子でなければスキップ
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 次トークンが=でなければスキップ
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: セミコロンに遭遇するまで式を読み飛ばしてしまう
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// curTokenが指定トークンかどうか判別
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenが指定トークンかどうか判別
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// アサーション関数
// peekTokenの型チェック(指定トークンかどうか)がOKならポインタを進める
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// peekTokenが期待値と合わない場合にエラー出力を行うヘルパー関数
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// TODO: セミコロンに遭遇するまで式を読み飛ばしてしまう
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
