package token

import "strconv"

var keywords map[string]Token

// Token represents a token.
type Token int

// List of tokens
const (
	// Special tokens
	Illegal Token = iota
	EOF
	Comment
	StringFragment // foo

	_literalBeg
	// Identifiers and basic type literals
	Ident // foo
	Int   // 12345
	Float // 123.45
	Char  // 'x'
	_literalEnd

	_operatorBeg
	// Operators and delimiters
	Add               // +
	Sub               // -
	Mul               // *
	Quo               // /
	Rem               // %
	And               // &
	Or                // |
	Xor               // ^
	AndNot            // &^
	Shl               // <<
	Shr               // >>
	Nullish           // ??
	AddAssign         // +=
	SubAssign         // -=
	MulAssign         // *=
	QuoAssign         // /=
	RemAssign         // %=
	AndAssign         // &=
	OrAssign          // |=
	XorAssign         // ^=
	AndNotAssign      // &^=
	ShlAssign         // <<=
	ShrAssign         // >>=
	NullishAssign     // ??=
	LAnd              // &&
	LOr               // ||
	Inc               // ++
	Dec               // --
	Equal             // ==
	Less              // <
	Greater           // >
	Assign            // =
	Not               // !
	NotEqual          // !=
	LessEq            // <=
	GreaterEq         // >=
	Define            // :=
	Ellipsis          // ...
	LParen            // (
	LBrack            // [
	LBrace            // {
	Comma             // ,
	Period            // .
	RParen            // )
	RBrack            // ]
	RBrace            // }
	Semicolon         // ;
	Colon             // :
	Question          // ?
	Arrow             // =>
	DoubleQuote       // "
	Backtick          // `
	DoubleSingleQuote // ''
	_operatorEnd

	_keywordBeg
	// Keywords
	Break
	Continue
	Else
	For
	Func
	If
	Return
	Defer
	Try
	Throw
	True
	False
	In
	Nil
	Import
	_keywordEnd
)

var tokens = [...]string{
	// Special tokens
	Illegal:        "ILLEGAL",
	EOF:            "EOF",
	Comment:        "COMMENT",
	StringFragment: "STRFRAG",

	// Identifiers and basic type literals
	Ident: "IDENT",
	Int:   "INT",
	Float: "FLOAT",
	Char:  "CHAR",

	// Operators and delimiters
	Add:               "+",
	Sub:               "-",
	Mul:               "*",
	Quo:               "/",
	Rem:               "%",
	And:               "&",
	Or:                "|",
	Xor:               "^",
	AndNot:            "&^",
	Shl:               "<<",
	Shr:               ">>",
	Nullish:           "??",
	AddAssign:         "+=",
	SubAssign:         "-=",
	MulAssign:         "*=",
	QuoAssign:         "/=",
	RemAssign:         "%=",
	AndAssign:         "&=",
	OrAssign:          "|=",
	XorAssign:         "^=",
	AndNotAssign:      "&^=",
	ShlAssign:         "<<=",
	ShrAssign:         ">>=",
	NullishAssign:     "??=",
	LAnd:              "&&",
	LOr:               "||",
	Inc:               "++",
	Dec:               "--",
	Equal:             "==",
	Less:              "<",
	Greater:           ">",
	Assign:            "=",
	Not:               "!",
	NotEqual:          "!=",
	LessEq:            "<=",
	GreaterEq:         ">=",
	Define:            ":=",
	Ellipsis:          "...",
	LParen:            "(",
	LBrack:            "[",
	LBrace:            "{",
	Comma:             ",",
	Period:            ".",
	RParen:            ")",
	RBrack:            "]",
	RBrace:            "}",
	Semicolon:         ";",
	Colon:             ":",
	Question:          "?",
	Arrow:             "=>",
	DoubleQuote:       "\"",
	Backtick:          "`",
	DoubleSingleQuote: "''",

	// Keywords
	Break:    "break",
	Continue: "continue",
	Else:     "else",
	For:      "for",
	Func:     "fn",
	If:       "if",
	Return:   "return",
	Defer:    "defer",
	Try:      "try",
	Throw:    "throw",
	True:     "true",
	False:    "false",
	In:       "in",
	Nil:      "nil",
	Import:   "import",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// LowestPrec represents lowest operator precedence.
const LowestPrec = 0

// Precedence returns the precedence for the operator token.
func (tok Token) Precedence() int {
	switch tok {
	case Nullish:
		return 1
	case LOr:
		return 2
	case LAnd:
		return 3
	case Equal, NotEqual, Less, LessEq, Greater, GreaterEq:
		return 4
	case Add, Sub, Or, Xor:
		return 5
	case Mul, Quo, Rem, Shl, Shr, And, AndNot:
		return 6
	}
	return LowestPrec
}

// IsLiteral returns true if the token is a literal.
func (tok Token) IsLiteral() bool {
	return _literalBeg < tok && tok < _literalEnd
}

// IsOperator returns true if the token is an operator.
func (tok Token) IsOperator() bool {
	return _operatorBeg < tok && tok < _operatorEnd
}

// IsKeyword returns true if the token is a keyword.
func (tok Token) IsKeyword() bool {
	return _keywordBeg < tok && tok < _keywordEnd
}

// Lookup returns corresponding keyword if ident is a keyword.
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return Ident
}

func init() {
	keywords = make(map[string]Token)
	for i := _keywordBeg + 1; i < _keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
}
