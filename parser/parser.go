package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota // 優先順位付けのためインクリメントする値を追加
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
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

	// 前置構文解析関数のmap
	prefixParseFns map[token.TokenType]prefixParseFn
	// 中置構文解析関数のmap
	infixParseFns map[token.TokenType]infixParseFn
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

	// 前置構文解析関数のmapを初期化
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

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

// 文の構文解析(トークンタイプに応じてASTノードを生成)
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		// LETトークンの構文解析
		return p.parseLetStatement()
	case token.RETURN:
		// RETURNトークンの構文解析
		return p.parseReturnStatement()
	default:
		// 式文の構文解析
		return p.parseExpressionStatement()
	}
}

// 'let'文の構文解析(LET用ASTノードを生成)
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	// 次トークンが識別子でなければスキップ
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// ASTの識別子ノードを返す
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

// 'return'文の構文解析(RETURN用ASTノードを生成)
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// TODO: セミコロンに遭遇するまで式を読み飛ばしてしまう
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

type (
	// 前置構文解析関数
	prefixParseFn func() ast.Expression
	// 中置構文解析関数(演算子の左側を引数として取る)
	infixParseFn func(ast.Expression) ast.Expression
)

// 前置構文解析関数のmapへの追加
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// 中置構文解析関数のmapへの追加
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// 式文の構文解析(Expression用ASTノードを生成)
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt

}

// 現トークンの前置に関連付けられた構文解析関数があれば返却
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	return leftExp
}

// ASTの識別子ノードを返す
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// 整数リテラルの構文解析(整数リテラル用ASTノードを生成)
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// 文字列をint64に変換
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}
