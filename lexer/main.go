package lexer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
)

type Type int

const (
	Service      Type = iota // service
	Package                  // package
	Option                   // option
	Structure                // struct
	LeftBrace                // {
	RightBrace               // }
	LeftBracket              // (
	RightBracket             // )
	EndOfFile                // EOF
	Colon                    // :
	Comma                    // ,
	Quote                    // "
	Identifier
	Illegal // Unrecognized character
)

type Token struct {
	Type  Type
	Value string
	Pos   Position
}

type Position struct {
	Line   int
	Column int
}
type Lexer struct {
	pos               Position
	reader            *bufio.Reader
	unterminatedQuote bool
}

func New(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    Position{Line: 1, Column: 0},
		reader: bufio.NewReader(reader),
	}
}
func (l *Lexer) Lex() Token {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{Type: EndOfFile, Pos: l.pos}
			}
			panic(err)
		}
		l.pos.Column++
		switch r {
		case '\n':
			l.resetPosition()
		case '#':
			l.skipLine()
		case '{':
			return Token{Type: LeftBrace, Pos: l.pos, Value: "{"}
		case '}':
			return Token{Type: RightBrace, Pos: l.pos, Value: "}"}
		case '(':
			return Token{Type: LeftBracket, Pos: l.pos, Value: "("}
		case ')':
			return Token{Type: RightBracket, Pos: l.pos, Value: ")"}
		case ':':
			return Token{Type: Colon, Pos: l.pos, Value: ":"}
		case ',':
			return Token{Type: Comma, Pos: l.pos, Value: ","}
		case '"':
			l.unterminatedQuote = !l.unterminatedQuote
			return Token{Type: Quote, Pos: l.pos, Value: "\""}
		default:
			if unicode.IsSpace(r) {
				continue
			} else if l.unterminatedQuote || unicode.IsLetter(r) || r == '_' {
				// backup and let lexIdent rescan the beginning of the ident
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				switch lit {
				case "package":
					return Token{Type: Package, Pos: startPos, Value: "package"}
				case "service":
					return Token{Type: Service, Pos: startPos, Value: "service"}
				case "struct":
					return Token{Type: Structure, Pos: startPos, Value: "struct"}
				case "option":
					return Token{Type: Option, Pos: startPos, Value: "option"}
				}
				return Token{Type: Identifier, Value: lit, Pos: startPos}
			} else {
				fmt.Printf("%d:%d Illegal token '%s'\n", l.pos.Line, l.pos.Column, string(r))
				os.Exit(1)
				return Token{Type: Illegal, Value: string(r), Pos: l.pos}
			}
		}
	}
}
func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.Column--
}
func (l *Lexer) lexIdent() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the identifier
				return lit
			}
		}

		l.pos.Column++
		if r == '"' {
			l.backup()
			return lit
		} else if l.unterminatedQuote || unicode.IsLetter(r) || r == '_' {
			lit = lit + string(r)
		} else {
			// scanned something not in the identifier
			l.backup()
			return lit
		}
	}
}
func (l *Lexer) resetPosition() {
	l.pos.Line++
	l.pos.Column = 0
}
func (l *Lexer) skipLine() {
	_, err := l.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}
	l.resetPosition()
}
