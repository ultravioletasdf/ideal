package parser

import (
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/ultravioletasdf/ideal/lexer"
)

var Known = []string{"string", "int", "int8", "int16", "int32", "int64", "float64", "float32", "bool"}

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
	if name.Type != lexer.Identifier {
		unexpected(name, "identifier", name.Value)
	}
	if t := p.peek(); t.Type == lexer.Identifier {
		if !isDigitsOnly(t.Value) {
			unexpected(t, "a postive integer", t.Value)
		}
		return OptionNode{Name: name.Value, Value: p.next().Value}
	} else {
		q1 := p.next()
		value := p.next()
		q2 := p.next()
		if q1.Type != lexer.Quote {
			unexpected(q1, "\"", q1.Value)
		}
		if value.Type != lexer.Identifier {
			unexpected(value, "identifier", value.Value)
		}
		if q2.Type != lexer.Quote {
			unexpected(q2, "\"", q2.Value)
		}
		return OptionNode{Name: name.Value, Value: value.Value}
	}

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
	} else if result.Type == lexer.LeftSquareBracket {
		size := p.next()         // Could also be ]
		rightBracket := p.next() // Could also be a type name
		if size.Type == lexer.RightSquareBracket {
			outputs = append(outputs, "[]"+rightBracket.Value)
			goto skip
		}
		typeName := p.next()

		if size.Type != lexer.Identifier {
			unexpected(size, "an identifier", size.Value)
		}
		if typeName.Type != lexer.Identifier {
			unexpected(typeName, "an identifier", typeName.Value)
		}
		sizeInt, err := strconv.Atoi(size.Value)
		if err != nil {
			fmt.Printf("%d:%d Couldn't parse identifier as an integer: %v\n", size.Pos.Line, size.Pos.Column, err)
			os.Exit(1)
		}
		if sizeInt < 1 {
			fmt.Printf("%d:%d Arrays must be at least 1 in size, got %d\n", size.Pos.Line, size.Pos.Column, sizeInt)
			os.Exit(1)
		}
		if rightBracket.Type != lexer.RightSquareBracket {
			unexpected(rightBracket, "]", rightBracket.Value)
		}
		outputs = append(outputs, fmt.Sprintf("[%d]%s", sizeInt, typeName.Value))
	} else if result.Type != lexer.Identifier {
		unexpected(result, "identifier", result.Value)
	} else {
		outputs = []string{result.Value}
	}
skip:
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
				fmt.Println("Expected Comma")
				unexpected(token, ",", token.Value)
				os.Exit(1)
			}
		}
		if token.Type == lexer.LeftSquareBracket {
			size := p.next()
			rightBracket := p.next()
			if size.Type == lexer.RightSquareBracket {
				list = append(list, "[]"+rightBracket.Value)
				continue
			}
			typeName := p.next()

			if size.Type != lexer.Identifier {
				unexpected(size, "an identifier", size.Value)
			}
			if typeName.Type != lexer.Identifier {
				unexpected(typeName, "an identifier", typeName.Value)
			}
			sizeInt, err := strconv.Atoi(size.Value)
			if err != nil {
				fmt.Printf("%d:%d Couldn't parse identifier as an integer: %v\n", size.Pos.Line, size.Pos.Column, err)
				os.Exit(1)
			}
			if sizeInt < 1 {
				fmt.Printf("%d:%d Arrays must be at least 1 in size, got %d\n", size.Pos.Line, size.Pos.Column, sizeInt)
				os.Exit(1)
			}
			if rightBracket.Type != lexer.RightSquareBracket {
				unexpected(rightBracket, "]", rightBracket.Value)
			}
			list = append(list, fmt.Sprintf("[%d]%s", sizeInt, typeName.Value))
			expectsComma = true
		} else if token.Type == lexer.Identifier {
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
		if one.Type != lexer.Identifier {
			unexpected(one, "identifier", one.Value)
		} else if two.Type == lexer.LeftSquareBracket {
			size := p.next()         // Could also be ]
			rightBracket := p.next() // Could also be a type name
			if size.Type == lexer.RightSquareBracket {
				fields = append(fields, FieldNode{Name: one.Value, Type: rightBracket.Value})
				continue
			}
			typeName := p.next()
			if size.Type != lexer.Identifier {
				unexpected(size, "an identifier", size.Value)
			}
			if typeName.Type != lexer.Identifier {
				unexpected(typeName, "an identifier", typeName.Value)
			}
			sizeInt, err := strconv.Atoi(size.Value)
			if err != nil {
				fmt.Printf("%d:%d Couldn't parse identifier as an integer: %v\n", size.Pos.Line, size.Pos.Column, err)
				os.Exit(1)
			}
			if sizeInt < 1 {
				fmt.Printf("%d:%d Arrays must be at least 1 in size, got %d\n", size.Pos.Line, size.Pos.Column, sizeInt)
				os.Exit(1)
			}
			if rightBracket.Type != lexer.RightSquareBracket {
				unexpected(rightBracket, "]", rightBracket.Value)
			}
			fields = append(fields, FieldNode{Name: one.Value, Type: fmt.Sprintf("[%d]%s", sizeInt, typeName.Value)})

		} else if one.Type != lexer.Identifier || two.Type != lexer.Identifier {
			fmt.Printf("%d:%d Expected two identifiers, got %v and %v\n", one.Pos.Line, one.Pos.Column, one.Value, two.Value)
			os.Exit(1)
		} else {
			fields = append(fields, FieldNode{Name: one.Value, Type: two.Value})
		}
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
func isDigitsOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}
