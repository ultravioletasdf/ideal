package parser

import (
	"fmt"
	"os"

	"idl/lexer"
)

type Nodes struct {
	Package    string
	Options    []OptionNode
	Services   []ServiceNode
	Structures []StructureNode
}
type OptionNode struct {
	Name  string
	Value string
}
type ServiceNode struct {
	Name      string
	Functions []FunctionNode
}
type FunctionNode struct {
	Name    string
	Inputs  []string
	Outputs []string
}
type StructureNode struct {
	Name   string
	Fields []FieldNode
}
type FieldNode struct {
	Name string
	Type string
}
type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) next() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EndOfFile}
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}
func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EndOfFile}
	}
	return p.tokens[p.pos]
}
func (p *Parser) parsePackage() string {
	tok := p.next()
	if tok.Type != lexer.Package {
		unexpected(tok, "package", tok.Value)
	}
	return p.next().Value
}
func (p *Parser) parseOption() OptionNode {
	tok := p.next()
	if tok.Type != lexer.Option {
		unexpected(tok, "option", tok.Value)
	}
	name := p.next()
	q1 := p.next()
	value := p.next()
	q2 := p.next()
	if q1.Type != lexer.Quote {
		unexpected(q1, "\"", q1.Value)
	}
	if name.Type != lexer.Identifier {
		unexpected(name, "identifier", name.Value)
	}
	if value.Type != lexer.Identifier {
		unexpected(value, "identifier", value.Value)
	}
	if q2.Type != lexer.Quote {
		unexpected(q2, "\"", q2.Value)
	}
	return OptionNode{Name: name.Value, Value: value.Value}
}
func (p *Parser) parseService() ServiceNode {
	tok := p.next()
	if tok.Type != lexer.Service {
		unexpected(tok, "service", tok.Value)
	}
	name := p.next().Value
	p.next()

	var functions []FunctionNode
	for p.peek().Type != lexer.RightBrace {
		functions = append(functions, p.parseFunction())
	}

	p.next()
	return ServiceNode{Name: name, Functions: functions}
}
func (p *Parser) parseFunction() FunctionNode {
	name := p.next().Value
	nextToken := p.next()
	if nextToken.Type != lexer.LeftBracket && nextToken.Type != lexer.Colon {
		unexpected(nextToken, "( or :", nextToken.Value)
	}
	var inputs []string
	if nextToken.Type == lexer.LeftBracket {
		inputs = p.parseList()
		p.next() // consume :
	}
	result := p.next()
	var outputs []string
	if result.Type == lexer.LeftBracket {
		outputs = p.parseList()
	} else if result.Type != lexer.Identifier {
		unexpected(result, "identifier", result.Value)
	} else {
		outputs = []string{result.Value}
	}
	return FunctionNode{Name: name, Inputs: inputs, Outputs: outputs}
}
func (p *Parser) parseList() (list []string) {
	expectsComma := false
	for p.peek().Type != lexer.RightBracket {
		token := p.next()
		if expectsComma {
			if token.Type == lexer.Comma {
				if p.peek().Type == lexer.RightBracket {
					fmt.Printf("%d:%d Trailing commas are not allowed\n", token.Pos.Line, token.Pos.Column)
					os.Exit(1)
				}
				expectsComma = false
				continue
			} else {
				unexpected(token, ",", token.Value)
				fmt.Println("Expected Comma")
				os.Exit(1)
			}
		}
		if token.Type == lexer.Identifier {
			list = append(list, token.Value)
			expectsComma = true
		} else {
			unexpected(token, "identifier", token.Value)
		}
	}
	p.next()
	return
}
func (p *Parser) parseStruct() StructureNode {
	tok := p.next()
	if tok.Type != lexer.Structure {
		unexpected(tok, "structure", tok.Value)
	}
	name := p.next().Value
	p.next()
	var fields []FieldNode
	for p.peek().Type != lexer.RightBrace {
		one := p.next()
		two := p.next()
		if one.Pos.Line != two.Pos.Line {
			fmt.Printf("%d:%d Must follow format: FieldName type\n", one.Pos.Line, one.Pos.Column)
			os.Exit(1)
		}
		if one.Type != lexer.Identifier || two.Type != lexer.Identifier {
			fmt.Printf("%d:%d Expected two identifiers, got %v and %v\n", one.Pos.Line, one.Pos.Column, one.Value, two.Value)
			os.Exit(1)
		}
		fields = append(fields, FieldNode{Name: one.Value, Type: two.Value})
	}
	p.next()
	return StructureNode{Name: name, Fields: fields}
}
func (p *Parser) Parse() Nodes {
	var pkg = p.parsePackage()

	var options []OptionNode
	var services []ServiceNode
	var structs []StructureNode

	for t := p.peek(); t.Type != lexer.EndOfFile; t = p.peek() {
		switch t.Type {
		case lexer.Option:
			options = append(options, p.parseOption())
		case lexer.Service:
			services = append(services, p.parseService())
		case lexer.Structure:
			structs = append(structs, p.parseStruct())
		default:
			unexpected(t, "service or struct", t.Value)
		}
	}

	return Nodes{Package: pkg, Options: options, Services: services, Structures: structs}
}
func unexpected(token lexer.Token, expected, unexpected string) {
	fmt.Printf("%d:%d Unexpected token '%v', expected '%v'\n", token.Pos.Line, token.Pos.Column, unexpected, expected)
	os.Exit(1)
}
