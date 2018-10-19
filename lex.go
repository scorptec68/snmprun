package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 60:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q (type %s)", i.val, i.typ)
}

func (t itemType) String() string {
	// lookup others
	str := others[t]
	if str != "" {
		return str
	}
	// lookup keywords
	for key, value := range keywords {
		if value == t {
			return key
		}
	}
	// lookup symbols
	for key, value := range symbols {
		if value == t {
			return key
		}
	}
	// else just number
	return fmt.Sprintf("%d", t)
}

// itemType identifies the type of lex items.
type (
	itemType      int
	Pos           int
	processResult int
)

const (
	resultMatch processResult = iota
	resultNoMatch
	resultMatchError
)

const (
	itemError              itemType = iota // error occurred; value is text of error
	itemIntegerLiteral                     // integer value
	itemOidLiteral                         // 1.3.6.1.*
	itemAlias                              // 'alias'
	itemStringLiteral                      // "string"
	itemEquals                             // '=' or '=='
	itemNotEquals                          // '#'
	itemLessEquals                         // '<='
	itemGreaterEquals                      // '>='
	itemLessThan                           // '<'
	itemGreaterThan                        // '>'
	itemAnd                                // '&'
	itemOr                                 // '|'
	itemNot                                // '~'
	itemPlus                               // '+'
	itemMinus                              // '-'
	itemTimes                              // '*'
	itemDivide                             // '/'
	itemLeftParen                          // '('
	itemRightParen                         // ')'
	itemLeftSquareBracket                  // '['
	itemRightSquareBracket                 // ']'
	itemColon                              // ':'
	itemComma                              // ','
	itemNewLine                            // '\n'
	itemEOF
	itemIdentifier // alphanumeric identifier
	// Keywords appear after all the rest.
	itemKeyword   // used only to delimit the keywords
	itemIf        // if keyword
	itemElse      // else keyword
	itemElseIf    // elseif keyword
	itemEndIf     // endif keyword
	itemLoop      // loop keyword
	itemLoopTimes // times keyword in for loop
	itemEndLoop   // endloop keyword
	itemPrint     // print keyword
	itemStrInt    // str keywords...
	itemStrBool
	itemStrCounter
	itemStrOid
	itemStrTimeticks
	itemStrIpaddress
	itemStrBitset
	itemStrGuage
	itemBoolean     // boolean keyword
	itemString      // string keyword
	itemInteger     // integer keyword
	itemBitset      // bitset keyword
	itemOid         // oid keyword
	itemCounter     // counter keyword
	itemTimeticks   // timeticks keyword
	itemIpv4address // ipaddress keyword
	itemGauge       // guage keyword (guage type = uint32)
	itemTrue        // true
	itemFalse       // false
	itemVar         // var
	itemEndVar      // endvar
	itemRun         // run
	itemEndRun      // endrun
	itemExit        // exit
	itemSleep       // sleep
	itemSecs        // secs
	itemMillis      // msecs
	itemRW          // rw
	itemRead        // read
	itemContains    // contains
	itemNone
)

var others = map[itemType]string{
	itemError:          "error",
	itemIntegerLiteral: "int literal",
	itemOidLiteral:     "OID",
	itemAlias:          "alias",
	itemStringLiteral:  "string literal",
	itemNewLine:        "new line",
	itemEOF:            "EOF",
	itemIdentifier:     "identifier",
	itemNone:           "none",
}

var keywords = map[string]itemType{
	"var":          itemVar,
	"endvar":       itemEndVar,
	"run":          itemRun,
	"endrun":       itemEndRun,
	"if":           itemIf,
	"else":         itemElse,
	"elseif":       itemElseIf,
	"endif":        itemEndIf,
	"loop":         itemLoop,
	"endloop":      itemEndLoop,
	"print":        itemPrint,
	"strInt":       itemStrInt,
	"strBool":      itemStrBool,
	"strCounter":   itemStrCounter,
	"strTimeticks": itemStrTimeticks,
	"strIpaddress": itemStrIpaddress,
	"strOid":       itemStrOid,
	"strBitset":    itemStrBitset,
	"strGuage":     itemStrGuage,
	"boolean":      itemBoolean,
	"string":       itemString,
	"integer":      itemInteger,
	"int":          itemInteger, // mimic C, java, go
	"counter":      itemCounter,
	"timeticks":    itemTimeticks,
	"ipaddress":    itemIpv4address,
	"bitset":       itemBitset,
	"oid":          itemOid,
	"guage":        itemGauge,
	"true":         itemTrue,
	"false":        itemFalse,
	"times":        itemLoopTimes,
	"exit":         itemExit,
	"sleep":        itemSleep,
	"secs":         itemSecs,
	"msecs":        itemMillis,
	"rw":           itemRW,
	"read":         itemRead,
	"contains":     itemContains,
}

var symbols = map[string]itemType{
	"<":  itemLessThan,
	">":  itemGreaterThan,
	"<=": itemLessEquals,
	">=": itemGreaterEquals,
	"=":  itemEquals,
	"==": itemEquals, // added to mimic C, java, go, etc...
	"+":  itemPlus,
	"-":  itemMinus,
	"~":  itemNot,
	"#":  itemNotEquals,
	"*":  itemTimes,
	"/":  itemDivide,
	"&":  itemAnd,
	"|":  itemOr,
	":":  itemColon,
	"(":  itemLeftParen,
	")":  itemRightParen,
	"[":  itemLeftSquareBracket,
	"]":  itemRightSquareBracket,
	",":  itemComma,
}

type processFn func(*lexer) processResult

const (
	eof        = -1
	spaceChars = " \t\r\n" // These are the space characters defined by Go itself.
)

// lexer holds the state of the scanner.
type lexer struct {
	name         string    // the name of the input; used only for error reports
	input        string    // the string being scanned
	pos          Pos       // current position in the input
	start        Pos       // start position of this item
	width        Pos       // width of last rune read from input
	prevItemType itemType  // previous item
	items        chan item // channel of scanned items
	line         int       // 1+number of newlines seen
}

//--------------------------------------------------------------------------------------------
// core rune processing

// nextItem returns the nextItem rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the nextItem rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) itemString() string {
	return l.input[l.start:l.pos]
}

// backup steps back one rune. Can only be called once per call of nextItem.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// resets to start of token
func (l *lexer) reset() {
	l.line -= strings.Count(l.input[l.start:l.pos], "\n")
	l.pos = l.start
}

//--------------------------------------------------------------------------------------------

// accept consumes the nextItem rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptNot(valid string) bool {
	if !strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
// returns number of accepted chars
func (l *lexer) acceptRun(valid string) int {
	i := 0
	for strings.ContainsRune(valid, l.next()) {
		i++
	}
	l.backup()
	return i
}

func (l *lexer) acceptNotRun(valid string) bool {
	for !strings.ContainsRune(valid, l.next()) {
		if l.width == 0 {
			// reached eof
			return false
		}
	}
	l.backup()
	return true
}

//--------------------------------------------------------------------------------------------
// items channel functions

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.line}
	l.prevItemType = t
	l.start = l.pos
}

// errorf returns an error token
func (l *lexer) errorf(format string, args ...interface{}) {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.line}
}

// nextItem returns the nextItem item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

//--------------------------------------------------------------------------------------------

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:         name,
		input:        input,
		items:        make(chan item),
		line:         1,
		prevItemType: itemNone,
	}
	go l.run()
	return l
}

var processFunctions = []processFn{
	processComment,
	processSymbol,
	processStringLiteral,
	processAlias,
	processOidLiteral,
	processNumericLiteral,
	processKeyword,
	processIdentifier}

// run runs lexer over the input
func (l *lexer) run() {
mainLoop:
	for {
		if !processWhitespace(l) {
			break
		}
		//fmt.Println("testing", string(l.peek()))
		found := false
	processLoop:
		for _, processFunc := range processFunctions {
			//fmt.Println("func =", processFunc)
			result := processFunc(l)
			//fmt.Println("peek = ", string(l.peek()))
			switch result {
			case resultMatch:
				found = true
				break processLoop
			case resultMatchError:
				break mainLoop
			}
		}
		if !found {
			l.errorf("Invalid token: '%s'", string(l.peek()))
			break
		}
	}
	l.emit(itemEOF)
	close(l.items)
}

/*
 * Eat up any whitespace to the nextItem non-whitespace
 * Return true if found a non-whitespace rune to process
 * Return false if didn't find anything and reached EOF
 */
func processWhitespace(l *lexer) bool {
	for {
		rune := l.next()

		switch rune {
		case eof:
			return false
		case '\n':
			if !isLineContinuedItem(l.prevItemType) && l.prevItemType != itemNewLine && l.prevItemType != itemNone {
				l.emit(itemNewLine)
			}
			// otherwise binary operator prior to end of line continue to nextItem line
			// OR comma at end of line
			// OR multiple new lines
			// OR new line at start of file
		}

		if !strings.ContainsRune(spaceChars, rune) {
			l.backup()
			l.ignore()
			return true
		}
	}
}

/*
 * Process a comment
 * Return true if successfully processed a comment
 * Return false if this was not a comment and other processing should be done.
 */
func processComment(l *lexer) processResult {
	rune := l.next()
	if rune == '/' {
		rune := l.next()
		if rune == '/' {
			// eat up until \n
			for l.next() != '\n' {
			}
			l.ignore()
			return resultMatch
		} else {
			l.backup()
			l.backup()
			return resultNoMatch
		}
	} else {
		l.backup()
		return resultNoMatch
	}
}

/*
 * Process a symbol/operator of 1 or 2 runes
 * Return true if got a match and emitted symbol
 * Return false otherwise
 */
func processSymbol(l *lexer) processResult {
	rune1 := l.next()
	rune2 := l.next()
	key1 := string(rune1)
	key12 := string(rune1) + string(rune2)
	if item, ok := symbols[key12]; ok {
		l.emit(item)
		return resultMatch
	} else if item, ok := symbols[key1]; ok {
		l.backup()
		l.emit(item)
		return resultMatch
	}
	// no 1 or 2 char symbol matches
	l.backup()
	l.backup()
	return resultNoMatch
}

func processStringLiteral(l *lexer) processResult {
	if l.peek() != '"' {
		return resultNoMatch
	}

	l.next()
	l.ignore()

	// now look for matching "
	if l.acceptNotRun("\"") {
		l.emit(itemStringLiteral)
		l.next()
		l.ignore()
		return resultMatch
	} else {
		l.errorf("Could not find string terminator")
		return resultMatchError
	}
}

func processAlias(l *lexer) processResult {
	if l.peek() != '\'' {
		return resultNoMatch
	}

	l.next()
	l.ignore()

	// now look for matching "
	if l.acceptNotRun("'") {
		l.emit(itemAlias)
		l.next()
		l.ignore()
		return resultMatch
	} else {
		l.errorf("Could not find alias terminator")
		return resultMatchError
	}
}

// processOID matches with:
// .1
// .1.2
// 1.2.3
// not with 1
// Need at least one dot
func processOidLiteral(l *lexer) processResult {
	leadingDot := false

	// optional leading dot
	if l.peek() == '.' {
		leadingDot = true
		l.accept(".")
	}

	digits := "0123456789"
	n := l.acceptRun(digits)
	if n == 0 {
		l.reset()
		return resultNoMatch
	}

	for i := 0; ; i++ {
		if !l.accept(".") {
			// finished
			if !leadingDot && i == 0 {
				// its a number not an oid
				l.reset()
				return resultNoMatch
			}
			break
		}
		n = l.acceptRun(digits)
		if n == 0 {
			l.reset()
			return resultNoMatch
		}
	}

	l.emit(itemOidLiteral)
	return resultMatch
}

func processNumericLiteral(l *lexer) processResult {
	r := l.next()
	if r != '+' && r != '-' && !('0' <= r && r <= '9') {
		l.backup()
		return resultNoMatch
	}

	// Optional leading sign.
	l.accept("+-")

	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)

	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		return resultMatchError
	}
	l.emit(itemIntegerLiteral)
	return resultMatch
}

func processKeyword(l *lexer) processResult {

	if !isAlpha(l.peek()) {
		return resultNoMatch
	}

	// extract word up to space or end of line
	// test word in the keywords list
	for {
		rune := l.next()
		if isEndOfWord(rune) {
			l.backup()

			// now look up word
			word := l.itemString()
			// fmt.Printf("-> Look up %s\n", word)
			if item, ok := keywords[word]; ok {
				l.emit(item)
				return resultMatch
			} else {
				l.reset()
				return resultNoMatch
			}
		}
		if !isAlpha(rune) {
			l.reset()
			return resultNoMatch
		}
	}
}

func processIdentifier(l *lexer) processResult {
	// extract word up to space or end of line
	// ensure word is all alpha numeric

	if !isAlpha(l.peek()) {
		return resultNoMatch
	}

	for {
		rune := l.next()
		if !isAlphaNumeric(rune) {
			l.backup()
			l.emit(itemIdentifier)
			return resultMatch
		}
	}
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}

func isEndOfWord(r rune) bool {
	return isSpace(r) || isEndOfLine(r) || r == eof || r == '('
}

// Is item allow arguments to span on next line or not?
func isLineContinuedItem(t itemType) bool {
	switch t {
	case itemLessThan, itemLessEquals, itemGreaterThan, itemGreaterEquals, itemEquals, itemPlus,
		itemMinus, itemNotEquals, itemTimes, itemDivide, itemAnd, itemOr, itemComma:
		return true
	default:
		return false
	}
}
