package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ValueType int
type StatementType int
type LoopType int
type ExpressionType int
type IntExpressionType int
type IntOperatorType int
type IntComparatorType int
type BoolExpressionType int
type BoolOperatorType int
type StringExpressionType int
type StringOperatorType int
type TimeUnit int

const (
	TimeSecs TimeUnit = iota
	TimeMillis
)

const (
	ValueInteger ValueType = iota
	ValueString
	ValueBoolean
	ValueBitset
	ValueOid
	ValueCounter
	ValueTimeticks
	ValueIpv4address
	ValueNone
)

const (
	StmtLoop StatementType = iota
	StmtIf
	StmtAssignment
	StmtPrint
	StmtSleep
	StmtExit
	StmtRead
)

const (
	LoopForever LoopType = iota
	LoopWhile
	LoopTimes
)

const (
	ExprnInteger ExpressionType = iota
	ExprnBoolean
	ExprnString
	ExprnBitset
	ExprnOid
	ExprnAddr
)

const (
	IntExprnValue IntExpressionType = iota
	IntExprnVariable
	IntExprnBinary
	IntExprnUnary
)

const (
	IntUnaryOpNegative IntOperatorType = iota
	IntBinaryOp
	IntBinaryOpAdd
	IntBinaryOpMinus
	IntBinaryOpTimes
	IntBinaryOpDivide
)

const (
	IntCompLessThan IntComparatorType = iota
	IntCompGreaterThan
	IntCompLessEquals
	IntCompGreaterEquals
	IntCompEquals
)

const (
	StringExprnValue StringExpressionType = iota
	StringExprnVariable
	StringExprnBinary
)

const (
	StringBinaryOpAdd StringOperatorType = iota
	StringBinaryOpMinus
)

type Program struct {
	variables *Variables
	stmtList  []*Statement
}

type Variables struct {
	types        map[string]*Type
	typesFromOid map[string]*Type
	intAliases   map[string]int
}

type Parser struct {
	prefixOid string // OID prefix used if oid not prefixed by dot
	variables *Variables

	lex   *lexer
	token item
	hold  bool // don't get next but hold where we are
}

//-------------------------------------------------------------------------------

// nextItem returns the nextItem token from lexer or saved from peeking.
func (parser *Parser) nextItem() item {
	if parser.hold {
		parser.hold = false
	} else {
		parser.token = parser.lex.nextItem()
	}
	//fmt.Println("-> token: ", parser.token)
	return parser.token
}

// peek returns but does not consume the nextItem token.
func (parser *Parser) peek() item {
	if parser.hold {
		return parser.token
	}
	parser.hold = true
	parser.token = parser.lex.nextItem()
	return parser.token
}

func (parser *Parser) matchItem(itemTyp itemType, context string) (item item, err error) {
	item = parser.nextItem()
	//fmt.Printf("-> matching on item: %v, got token: %v\n", itemTyp, item)
	if item.typ != itemTyp {
		return item, parser.errorf("Expecting %v in %s but got \"%v\"", itemTyp, context, item.typ)
	}
	return item, nil
}

func (parser *Parser) match(itemTyp itemType, context string) (err error) {
	_, err = parser.matchItem(itemTyp, context)
	return err
}

//-------------------------------------------------------------------------------

func printIndent(indent int) {
	for indent > 0 {
		fmt.Print("  ")
		indent--
	}
}

func printfIndent(indent int, format string, a ...interface{}) {
	printIndent(indent)
	fmt.Printf(format, a...)
}

func PrintProgram(prog *Program, indent int) {
	printfIndent(indent, "Program\n")
	PrintVariables(prog.variables, indent+1)
	PrintStatementList(prog.stmtList, indent+1)
}

func PrintVariables(vars *Variables, indent int) {
	printfIndent(indent, "Variables\n")

	// types
	// sort for testing predictability
	printfIndent(indent+1, "Types\n")
	ids := make([]string, 0)
	for id := range vars.types {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		printfIndent(indent+2, "%s: %v\n", id, vars.types[id])
	}

	// types
	// sort for testing predictability
	if len(vars.intAliases) > 0 {
		printfIndent(indent+1, "Aliases\n")
		ids = make([]string, 0)
		for id := range vars.intAliases {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			printfIndent(indent+2, "%s: %v\n", id, vars.intAliases[id])
		}
	}
}

func PrintStatementList(stmtList []*Statement, indent int) {
	printfIndent(indent, "StatementList\n")
	for _, stmt := range stmtList {
		PrintOneStatement(stmt, indent+1)
	}
}

func PrintOneStatement(stmt *Statement, indent int) {
	printfIndent(indent, "Statement (type code: %v)\n", stmt.stmtType)

	switch stmt.stmtType {
	case StmtAssignment:
		PrintAssignmentStmt(stmt.assignmentStmt, indent+1)
	case StmtIf:
		PrintIfStmt(stmt.ifStmt, indent+1)
	case StmtLoop:
		PrintLoopStmt(stmt.loopStmt, indent+1)
	case StmtPrint:
		PrintPrintStmt(stmt.printStmt, indent+1)
	case StmtRead:
		PrintReadStmt(stmt.readStmt, indent+1)
	case StmtExit:
		printfIndent(indent, "Exit\n")
	}
}

func PrintAssignmentStmt(assign *AssignmentStatement, indent int) {
	printfIndent(indent, "Assignment\n")

	printfIndent(indent, "lhs var = %s\n", assign.identifier)
	PrintExpression(assign.exprn, indent+1)
}

func PrintIfStmt(ifStmt *IfStatement, indent int) {
	printfIndent(indent, "If Statement\n")

	printfIndent(indent, "predicate\n")
	PrintBooleanExpression(ifStmt.boolExpression, indent+1)

	printfIndent(indent, "if stmts\n")
	PrintStatementList(ifStmt.stmtList, indent+1)

	// print the elseif parts
	for i, elseif := range ifStmt.elsifList {
		printfIndent(indent+1, "[%d] elsif\n", i)
		printElseIfStmt(elseif, indent+1)
	}

	if len(ifStmt.elseStmtList) > 0 {
		printfIndent(indent, "else stmts\n")
		PrintStatementList(ifStmt.elseStmtList, indent+1)
	}
}

func PrintPrintStmt(printStmt *PrintStatement, indent int) {
	printfIndent(indent, "Print Statement\n")
	PrintStringExpression(printStmt.exprn, indent+1)
}

func PrintSleepStmt(sleepStmt *SleepStatement, indent int) {
	printfIndent(indent, "Sleep Statement\n")
	PrintIntExpression(sleepStmt.exprn, indent+1)
	switch sleepStmt.units {
	case TimeSecs:
		printfIndent(indent+1, "secs\n")
	case TimeMillis:
		printfIndent(indent+1, "msecs\n")
	}
}

func PrintReadStmt(readStmt *ReadStatement, indent int) {
	printfIndent(indent, "Read Statement\n")
	printfIndent(indent, "id: %s", readStmt.identifier)
}

func PrintLoopStmt(loopStmt *LoopStatement, indent int) {
	printfIndent(indent, "Loop Statement (%v)\n", loopStmt.loopType)
	switch loopStmt.loopType {
	case LoopWhile:
		PrintBooleanExpression(loopStmt.boolExpression, indent+1)
	case LoopTimes:
		PrintIntExpression(loopStmt.intExpression, indent+1)
	}
	PrintStatementList(loopStmt.stmtList, indent+1)
}

func printElseIfStmt(elseif *ElseIf, indent int) {
	printfIndent(indent, "elsif expression\n")
	PrintBooleanExpression(elseif.boolExpression, indent+1)

	printfIndent(indent, "elsif stmts\n")
	PrintStatementList(elseif.stmtList, indent+1)
}

func PrintExpression(exprn *Expression, indent int) {
	printfIndent(indent, "Expression\n")
	switch exprn.exprnType {
	case ExprnBoolean:
		PrintBooleanExpression(exprn.boolExpression, indent+1)
	case ExprnInteger:
		PrintIntExpression(exprn.intExpression, indent+1)
	case ExprnString:
		PrintStringExpression(exprn.stringExpression, indent+1)
	case ExprnBitset:
		PrintBitsetExpression(exprn.bitsetExpression, indent+1)
	}
}

func PrintBitsetExpression(exprn *BitsetExpression, indent int) {
	printfIndent(indent, "Bitset Expression\n")
	printfIndent(indent, "Add terms\n")
	for i, bitsetTerm := range exprn.plusTerms {
		PrintBitsetTerm(i, bitsetTerm, indent+1)
	}
	if len(exprn.minusTerms) > 0 {
		printfIndent(indent, "Minus terms\n")
		for i, bitsetTerm := range exprn.minusTerms {
			PrintBitsetTerm(i, bitsetTerm, indent+1)
		}
	}
}

func PrintBitsetTerm(index int, bitsetTerm *BitsetTerm, indent int) {
	printfIndent(indent, "[%d]: bitset term\n", index)
	switch bitsetTerm.bitsetTermType {
	case BitsetTermValue:
		printfIndent(indent, "Value: %v\n", bitsetTerm.bitsetVal)
	case BitsetTermId:
		printfIndent(indent, "Identifier: %s\n", bitsetTerm.identifier)
	case BitsetTermBracket:
		printfIndent(indent, "Bracketed Bitset Expression\n")
		PrintBitsetExpression(bitsetTerm.bracketedExprn, indent+1)
	}
}

func PrintStringExpression(exprn *StringExpression, indent int) {
	printfIndent(indent, "String Expression\n")
	PrintStrAddTerms(exprn.addTerms, indent)
}

func PrintOidExpression(exprn *OidExpression, indent int) {
	printfIndent(indent, "OID Expression\n")
	PrintOidAddTerms(exprn.addTerms, indent)
}

func PrintStrAddTerms(addTerms []*StringTerm, indent int) {
	printfIndent(indent, "Add String Terms\n")
	for i, term := range addTerms {
		PrintStrAddTerm(i, term, indent+1)
	}
}

func PrintOidAddTerms(addTerms []*OidTerm, indent int) {
	printfIndent(indent, "Add Oid Terms\n")
	for i, term := range addTerms {
		PrintOidAddTerm(i, term, indent+1)
	}
}

func PrintOidAddTerm(i int, term *OidTerm, indent int) {
	printfIndent(indent, "[%d]: oid term\n", i)
	switch term.oidTermType {
	case OidTermValue:
		printfIndent(indent, "Literal: \"%s\"\n", term.oidVal)
	case OidTermId:
		printfIndent(indent, "Identifier: %s\n", term.identifier)
	case OidTermBracket:
		printfIndent(indent, "Bracketed Oid Expression\n")
		PrintOidExpression(term.bracketedExprn, indent+1)
	}
}

func PrintStrAddTerm(i int, term *StringTerm, indent int) {
	printfIndent(indent, "[%d]: string term\n", i)
	switch term.strTermType {
	case StringTermValue:
		printfIndent(indent, "Literal: \"%s\"\n", term.strVal)
	case StringTermId:
		printfIndent(indent, "Identifier: %s\n", term.identifier)
	case StringTermBracket:
		printfIndent(indent, "Bracketed String Expression\n")
		PrintStringExpression(term.bracketedExprn, indent+1)
	case StringTermStringedBoolExprn:
		printfIndent(indent, "Stringify Bool Expression\n")
		PrintBooleanExpression(term.stringedBoolExprn, indent+1)
	case StringTermStringedIntExprn:
		printfIndent(indent, "Stringify Int Expression\n")
		PrintIntExpression(term.stringedIntExprn, indent+1)
	}
}

func PrintBooleanExpression(exprn *BoolExpression, indent int) {
	printfIndent(indent, "Boolean Expression\n")
	PrintOrTerms(exprn.boolOrTerms, indent)
}

func PrintOrTerms(orTerms []*BoolTerm, indent int) {
	printfIndent(indent, "Or Terms\n")
	for i, term := range orTerms {
		PrintOrTerm(i, term, indent+1)
	}
}

func PrintOrTerm(i int, term *BoolTerm, indent int) {
	printfIndent(indent, "[%d]: term\n", i)
	PrintAndFactors(term.boolAndFactors, indent+1)
}

func PrintAndFactors(andFactors []*BoolFactor, indent int) {
	printfIndent(indent, "And Factors\n")
	for i, factor := range andFactors {
		PrintBoolFactor(i, factor, indent+1)
	}
}

func PrintBoolFactor(i int, factor *BoolFactor, indent int) {
	printfIndent(indent, "[%d]: factor\n", i)
	switch factor.boolFactorType {
	case BoolFactorNot:
		printfIndent(indent, "Not factor\n")
		PrintBoolFactor(i, factor.notBoolFactor, indent+1)
	case BoolFactorConst:
		printfIndent(indent, "Const factor: %t\n", factor.boolConst)
	case BoolFactorId:
		printfIndent(indent, "Id factor: %s\n", factor.boolIdentifier)
	case BoolFactorBracket:
		printfIndent(indent, "Bracket expression\n")
		PrintBooleanExpression(factor.bracketedExprn, indent+1)
	case BoolFactorIntComparison:
		printfIndent(indent, "Integer comparison\n")
		printfIndent(indent, "%v\n", factor.intComparison.intComparator)
		PrintIntExpression(factor.intComparison.lhsIntExpression, indent+1)
		PrintIntExpression(factor.intComparison.rhsIntExpression, indent+1)
	}
}

func PrintIntExpression(exprn *IntExpression, indent int) {
	printfIndent(indent, "Integer Expression\n")
	if len(exprn.plusTerms) > 0 {
		PrintPlusTerms(exprn.plusTerms, indent)
	}
	if len(exprn.minusTerms) > 0 {
		PrintMinusTerms(exprn.minusTerms, indent)
	}
}

func PrintPlusTerms(plusTerms []*IntTerm, indent int) {
	printfIndent(indent, "Plus Terms\n")
	for i, term := range plusTerms {
		PrintPlusTerm(i, term, indent+1)
	}
}

func PrintMinusTerms(minusTerms []*IntTerm, indent int) {
	printfIndent(indent, "Minus Terms\n")
	for i, term := range minusTerms {
		PrintMinusTerm(i, term, indent+1)
	}
}

func PrintPlusTerm(i int, term *IntTerm, indent int) {
	printfIndent(indent, "[%d]: plus term\n", i)
	if len(term.timesFactors) > 0 {
		PrintTimesFactors(term.timesFactors, indent+1)
	}
	if len(term.divideFactors) > 0 {
		PrintDivideFactors(term.divideFactors, indent+1)
	}
}

func PrintMinusTerm(i int, term *IntTerm, indent int) {
	printfIndent(indent, "[%d]: minus term\n", i)
	PrintTimesFactors(term.timesFactors, indent+1)
	PrintDivideFactors(term.divideFactors, indent+1)
}

func PrintTimesFactors(timesFactors []*IntFactor, indent int) {
	printfIndent(indent, "Times Factors\n")
	for i, factor := range timesFactors {
		PrintIntFactor(i, factor, indent+1)
	}
}

func PrintDivideFactors(divideFactors []*IntFactor, indent int) {
	printfIndent(indent, "Divide Factors\n")
	for i, factor := range divideFactors {
		PrintIntFactor(i, factor, indent+1)
	}
}

func PrintIntFactor(i int, factor *IntFactor, indent int) {
	printfIndent(indent, "[%d]: factor\n", i)
	switch factor.intFactorType {
	case IntFactorMinus:
		printfIndent(indent, "Minus factor\n")
		PrintIntFactor(i, factor.minusIntFactor, indent+1)
	case IntFactorConst:
		printfIndent(indent, "Const factor: %d\n", factor.intConst)
	case IntFactorId:
		printfIndent(indent, "Id factor: %s\n", factor.intIdentifier)
	case IntFactorBracket:
		printfIndent(indent, "Bracket expression\n")
		PrintIntExpression(factor.bracketedExprn, indent+1)
	}
}

func NewParser(l *lexer) *Parser {
	return &Parser{
		lex:       l,
		prefixOid: ".1.3.6.1",
	}
}

func (parser *Parser) ParseProgram() (prog *Program, err error) {
	prog = new(Program)
	prog.variables, err = parser.parseVariables()
	if err != nil {
		return nil, err
	}
	parser.variables = prog.variables

	err = parser.match(itemRun, "program")
	if err != nil {
		return nil, err
	}

	err = parser.match(itemNewLine, "program")
	if err != nil {
		return nil, err
	}

	prog.stmtList, err = parser.parseStatementList()
	if err != nil {
		return nil, err
	}

	err = parser.match(itemEndRun, "program")
	if err != nil {
		return nil, err
	}
	return prog, nil
}

func (parser *Parser) parseVariables() (vars *Variables, err error) {
	vars = new(Variables)
	vars.types = make(map[string]*Type)
	vars.typesFromOid = make(map[string]*Type)
	vars.intAliases = make(map[string]int)

	item := parser.peek()
	if item.typ != itemVar {
		// no variables to process
		return vars, nil
	}
	item = parser.nextItem()

	err = parser.match(itemNewLine, "Var start")
	if err != nil {
		return nil, err
	}

	// we have potentially some variables (could be empty)
	for {
		item = parser.nextItem()
		switch item.typ {
		case itemEndVar:
			// end of variable declaration
			err = parser.match(itemNewLine, "Var End")
			if err != nil {
				return nil, err
			}
			return vars, nil
		case itemEOF:
			// end of any input which is an error
			err := parser.errorf("Cannot find EndVar")
			return nil, err
		case itemIdentifier:
			idStr := item.val
			var initMode InitMode

			switch parser.nextItem().typ {
			case itemColon:
				initMode = InitModeZero
			case itemGreaterThan:
				initMode = InitModeExternal
			default:
				return nil, parser.errorf("Non valid char after variable identifier")
			}

			typ, err := parser.parseType(vars, initMode)
			if err != nil {
				return nil, err
			}

			vars.types[idStr] = typ
			vars.typesFromOid[typ.oid] = typ

			err = parser.match(itemNewLine, "Variable declaration")
			if err != nil {
				return nil, err
			}
		default:
			return nil, parser.errorf("Unexpected token: %s in variables section", item)
		}
	}
}

func (parser *Parser) parseType(vars *Variables, initMode InitMode) (typ *Type, err error) {
	typ = new(Type)
	typ.initMode = initMode

	item := parser.nextItem()
	typ.lineNum = item.line

	// optional oid
	if item.typ == itemOidLiteral || item.typ == itemIntegerLiteral {
		if strings.HasPrefix(item.val, ".") {
			typ.oid = item.val
		} else {
			typ.oid = parser.prefixOid + "." + item.val
		}
		item = parser.nextItem()

		// optional rw or rwb snmp mode
		if item.typ == itemRW {
			typ.snmpMode = SnmpModeReadWrite
			typ.externalValue = make(chan *Value)
			item = parser.nextItem()
		}
	}
	//fmt.Printf("item is %v\n", item)

	switch item.typ {
	case itemString:
		typ.valueType = ValueString
	case itemInteger:
		typ.valueType = ValueInteger
	case itemCounter:
		if typ.snmpMode == SnmpModeReadWrite {
			return nil, parser.errorf("Counter type can not be in rw mode as cannot be set")
		}
		typ.valueType = ValueCounter
	case itemTimeticks:
		typ.valueType = ValueTimeticks
	case itemBoolean:
		typ.valueType = ValueBoolean
		if typ.oid != "" {
			return nil, parser.errorf("Bool type can not have an OID")
		}
	case itemIpv4address:
		typ.valueType = ValueIpv4address
	case itemBitset:
		typ.valueType = ValueBitset
	case itemOid:
		typ.valueType = ValueOid
	default:
		return nil, parser.errorf("Expecting a variable type")
	}

	//fmt.Printf("var type: %v\n", typ)

	// optional aliases: [ 1 = 'blah', 2 = 'bloh', 3 = 'bleh', ]
	if (typ.valueType == ValueInteger || typ.valueType == ValueBitset) &&
		parser.peek().typ == itemLeftSquareBracket {

		parser.nextItem()

		// loop through each alias - can be empty
		for {
			// num = 'value' , ... ]

			if parser.peek().typ == itemRightSquareBracket {
				parser.nextItem()
				break
			}

			numItem, err := parser.matchItem(itemIntegerLiteral, "alias")
			if err != nil {
				return nil, err
			}

			err = parser.match(itemEquals, "alias")
			if err != nil {
				return nil, err
			}

			aliasItem, err := parser.matchItem(itemAlias, "alias")
			if err != nil {
				return nil, err
			}

			x, _ := strconv.Atoi(numItem.val)
			if _, ok := vars.intAliases[aliasItem.val]; ok {
				return nil, parser.errorf("Cannot redfine existing alias \"%s\"", aliasItem.val)
			}
			vars.intAliases[aliasItem.val] = x

			// optional comma
			if parser.peek().typ == itemComma {
				parser.nextItem()
			}
		}
	}

	return typ, nil
}

func (parser *Parser) lookupType(id string) ValueType {
	typ, ok := parser.variables.types[id]
	if ok {
		return typ.valueType
	}
	return ValueNone
}

func isStmtListEndKeyword(i item) bool {
	return i.typ == itemEndRun || i.typ == itemEndLoop || i.typ == itemEndIf ||
		i.typ == itemElse || i.typ == itemElseIf

}

func (parser *Parser) parseStatementList() ([]*Statement, error) {
	var stmtList []*Statement
	for {
		if isStmtListEndKeyword(parser.peek()) {
			return stmtList, nil
		}
		stmt, err := parser.parseStatement()
		if err != nil {
			return nil, err
		}
		stmtList = append(stmtList, stmt)
	}
}

func (parser *Parser) errorf(format string, a ...interface{}) error {
	preamble := fmt.Sprintf("Error at line %d: ", parser.token.line)
	return fmt.Errorf(preamble+format, a...)
}

func (parser *Parser) parseStatement() (stmt *Statement, err error) {
	stmt = new(Statement)

	item := parser.peek()
	switch item.typ {
	case itemIdentifier:
		stmt.stmtType = StmtAssignment
		stmt.assignmentStmt, err = parser.parseAssignment()
		if err != nil {
			return nil, err
		}
	case itemIf:
		parser.nextItem()
		stmt.stmtType = StmtIf
		stmt.ifStmt, err = parser.parseIfStatement()
		if err != nil {
			return nil, err
		}
	case itemLoop:
		parser.nextItem()
		stmt.stmtType = StmtLoop
		stmt.loopStmt, err = parser.parseLoopStatement()
		if err != nil {
			return nil, err
		}
	case itemPrint:
		parser.nextItem()
		stmt.stmtType = StmtPrint
		stmt.printStmt, err = parser.parsePrintStatement()
		if err != nil {
			return nil, err
		}
	case itemSleep:
		parser.nextItem()
		stmt.stmtType = StmtSleep
		stmt.sleepStmt, err = parser.parseSleepStatement()
		if err != nil {
			return nil, err
		}
	case itemExit:
		parser.nextItem()
		stmt.stmtType = StmtExit
		parser.match(itemNewLine, "exit")
		// Note: there is nothing else with it to store
	case itemRead:
		parser.nextItem()
		stmt.stmtType = StmtRead
		stmt.readStmt, err = parser.parseReadStatement()
		if err != nil {
			return nil, err
		}

	default:
		return nil, parser.errorf("Missing leading statement token. Got %v", item)
	}

	return stmt, err
}

func (parser *Parser) parsePrintStatement() (printStmt *PrintStatement, err error) {
	printStmt = new(PrintStatement)

	printStmt.exprn, err = parser.parseStrExpression()
	if err != nil {
		return nil, err
	}
	err = parser.match(itemNewLine, "print statement")
	if err != nil {
		return nil, err
	}
	return printStmt, nil
}

//
// read identifier
//
func (parser *Parser) parseReadStatement() (readStmt *ReadStatement, err error) {
	readStmt = new(ReadStatement)

	item := parser.nextItem()
	id := item.val
	typ, ok := parser.variables.types[id]
	if !ok {
		return nil, parser.errorf("Unable to read on undefined variable")
	}
	if typ.oid == "" {
		return nil, parser.errorf("Unable to read on non OID variable")
	}
	if typ.snmpMode != SnmpModeReadWrite {
		return nil, parser.errorf("Unable to read on non rw OID variable")
	}

	readStmt.identifier = id

	err = parser.match(itemNewLine, "read")
	if err != nil {
		return nil, err
	}

	return readStmt, nil
}

func (parser *Parser) parseSleepStatement() (sleepStmt *SleepStatement, err error) {
	sleepStmt = new(SleepStatement)

	sleepStmt.exprn, err = parser.parseIntExpression()
	if err != nil {
		return nil, err
	}

	item := parser.nextItem()
	switch item.typ {
	case itemSecs:
		sleepStmt.units = TimeSecs
	case itemMillis:
		sleepStmt.units = TimeMillis
	default:
		return nil, parser.errorf("Expecting time units in sleep statement but got \"%v\"", item.typ)
	}

	err = parser.match(itemNewLine, "sleep statement")
	if err != nil {
		return nil, err
	}
	return sleepStmt, nil
}

// Note: other parsers use panic/recover instead of returning an error

// Grammar
//	<loop> ::= loop \n {<statement>} endloop \n |
//             loop times <int-expression> \n {<statement>} endloop \n |
//             loop <bool-expression> \n {<statement>} endloop \n
//
func (parser *Parser) parseLoopStatement() (loopStmt *LoopStatement, err error) {
	loopStmt = new(LoopStatement)

	switch parser.peek().typ {
	case itemNewLine:
		// forever loop
		// just statements and no conditional part of loop construct
		loopStmt.loopType = LoopForever
	case itemLoopTimes:
		parser.nextItem() // move over the "times" keyword
		loopStmt.loopType = LoopTimes
		loopStmt.intExpression, err = parser.parseIntExpression()
		if err != nil {
			return nil, err
		}

	default:
		// while loop
		loopStmt.loopType = LoopWhile
		loopStmt.boolExpression, err = parser.parseBoolExpression()
		if err != nil {
			return nil, err
		}
	}

	// now parse the newline and statement list...

	err = parser.match(itemNewLine, "loop")
	if err != nil {
		return nil, err
	}
	loopStmt.stmtList, err = parser.parseStatementList()
	if err != nil {
		return nil, err
	}

	err = parser.match(itemEndLoop, "loop")
	if err != nil {
		return nil, err
	}
	err = parser.match(itemNewLine, "loop")
	if err != nil {
		return nil, err
	}

	return loopStmt, nil
}

// Grammar
// <if> ::= if <bool-expression> \n {<statement>}
//    {elseif <bool-expression> \n {<statement>}} [else \n {<statement>}] endif \n
//
func (parser *Parser) parseIfStatement() (ifStmt *IfStatement, err error) {
	ifStmt = new(IfStatement)

	ifStmt.boolExpression, err = parser.parseBoolExpression()
	if err != nil {
		return nil, err
	}
	err = parser.match(itemNewLine, "if statement")
	if err != nil {
		return nil, err
	}
	ifStmt.stmtList, err = parser.parseStatementList()
	if err != nil {
		return nil, err
	}
	for {
		item := parser.nextItem()
		switch item.typ {
		case itemElseIf:
			elseIf, err := parser.parseElseIf()
			if err != nil {
				return nil, err
			}
			ifStmt.elsifList = append(ifStmt.elsifList, elseIf)
		case itemElse:
			err = parser.match(itemNewLine, "else")
			if err != nil {
				return nil, err
			}
			ifStmt.elseStmtList, err = parser.parseStatementList()
			if err != nil {
				return nil, err
			}
		case itemEndIf:
			err = parser.match(itemNewLine, "if")
			if err != nil {
				return nil, err
			}
			return ifStmt, nil
		default:
			return nil, parser.errorf("Bad token in if statement")
		}
	}

}

// grammar:
//    elseif <bool-expression> \n {<statement>}
//
func (parser *Parser) parseElseIf() (elseIf *ElseIf, err error) {
	elseIf = new(ElseIf)
	elseIf.boolExpression, err = parser.parseBoolExpression()
	if err != nil {
		return nil, err
	}
	err = parser.match(itemNewLine, "elseif")
	if err != nil {
		return nil, err
	}
	elseIf.stmtList, err = parser.parseStatementList()
	if err != nil {
		return nil, err
	}
	return elseIf, nil
}

func (parser *Parser) parseAssignment() (assign *AssignmentStatement, err error) {
	assign = new(AssignmentStatement)
	idItem := parser.nextItem()
	assign.identifier = idItem.val

	err = parser.match(itemEquals, "Assignment")
	if err != nil {
		return nil, err
	}

	idType := parser.lookupType(assign.identifier)
	switch idType {
	case ValueBoolean:
		boolExprn, err := parser.parseBoolExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnBoolean
		assign.exprn.boolExpression = boolExprn
	case ValueInteger, ValueCounter, ValueTimeticks:
		intExprn, err := parser.parseIntExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnInteger
		assign.exprn.intExpression = intExprn
	case ValueString:
		strExprn, err := parser.parseStrExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnString
		assign.exprn.stringExpression = strExprn
	case ValueBitset:
		bitsetExprn, err := parser.parseBitsetExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnBitset
		assign.exprn.bitsetExpression = bitsetExprn
	case ValueOid:
		oidExprn, err := parser.parseOidExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnOid
		assign.exprn.oidExpression = oidExprn
	case ValueIpv4address:
		addrExprn, err := parser.parseAddrExpression()
		if err != nil {
			return nil, err
		}
		assign.exprn = new(Expression)
		assign.exprn.exprnType = ExprnAddr
		assign.exprn.addrExpression = addrExprn
	default:
		return nil, parser.errorf("Assignment to undeclared variable: %s", idItem.val)
	}

	err = parser.match(itemNewLine, "assignment")
	if err != nil {
		return nil, err
	}

	return assign, nil
}

//
// e.g: (a + 3 * (c - 4)) < 10 & (d & e | f)
//
// This function is broken
// Change grammar to:
//
//<bool-expression>::=<bool-term>{<or><bool-term>}
//<bool-term>::=<bool-factor>{<and><bool-factor>}
//<bool-factor>::=<bool-constant>|<not><bool-factor>|(<bool-expression>)|<int-comparison>
//
// Leave out int-comparison for 1st testing
//
//<int-comparison>::=<int-expression><int-comp><int-expression>
//
//<int-expression>::=<int-term>{<plus-or-minus><int-term>}
//<int-term>::=<int-factor>{<times-or-divice><int-factor>}
//<int-factor>::=<int-constant>|<minus><int-factor>|(<int-expression>)
//...
//<bool-constant>::= false|true
//<or>::='|'
//<and>::='&'
//<not>::='!'
//<plus-or-minus>::='+' | '-'
//<times-or-divide>::= '*' | '/'
//<minus>::='-'
//

//<bool-expression>::=<bool-term>{<or><bool-term>}
func (parser *Parser) parseBoolExpression() (boolExprn *BoolExpression, err error) {
	boolExprn = new(BoolExpression)

	// process 1st term
	boolTerm, err := parser.parseBoolTerm()
	if err != nil {
		return nil, err
	}
	boolExprn.boolOrTerms = append(boolExprn.boolOrTerms, boolTerm)

	// optionally process others
	for parser.peek().typ == itemOr {
		parser.nextItem()
		boolTerm, err = parser.parseBoolTerm()
		if err != nil {
			return nil, err
		}
		boolExprn.boolOrTerms = append(boolExprn.boolOrTerms, boolTerm)
	}
	return boolExprn, nil
}

//
// <bitset-expression> ::= <bitset-term> {<bitset-operator> <bitset-term>}
// <bitset-term> ::= <bitset-literal> | <identifier> |
//                       <lparen> <bitset-expression> <rparen>
func (parser *Parser) parseBitsetExpression() (bitsetExprn *BitsetExpression, err error) {
	bitsetExprn = new(BitsetExpression)

	// process 1st term
	bitsetTerm, err := parser.parseBitsetTerm()
	if err != nil {
		return nil, err
	}
	bitsetExprn.plusTerms = append(bitsetExprn.plusTerms, bitsetTerm)

	// optionally process others
	var usingPlus bool
loop:
	for {
		switch parser.peek().typ {
		case itemPlus:
			usingPlus = true
		case itemMinus:
			usingPlus = false
		default:
			break loop
		}
		parser.nextItem()
		bitsetTerm, err := parser.parseBitsetTerm()
		if err != nil {
			return nil, err
		}
		if usingPlus {
			bitsetExprn.plusTerms = append(bitsetExprn.plusTerms, bitsetTerm)
		} else {
			bitsetExprn.minusTerms = append(bitsetExprn.minusTerms, bitsetTerm)
		}
	}

	return bitsetExprn, nil
}

func (parser *Parser) parseBitsetTerm() (bitsetTerm *BitsetTerm, err error) {
	bitsetTerm = new(BitsetTerm)

	item := parser.nextItem()
	switch item.typ {
	case itemIdentifier:
		if parser.lookupType(item.val) != ValueBitset {
			return nil, parser.errorf("Not bitset variable in bitset expression")
		}
		bitsetTerm.bitsetTermType = BitsetTermId
		bitsetTerm.identifier = item.val
	case itemLeftSquareBracket:
		// [3, 5] or ['alias1', 'alias2', 'alias3']
		bitsetTerm.bitsetTermType = BitsetTermValue
		bitsetTerm.bitsetVal, err = parser.parseBitsetValue()
		if err != nil {
			return nil, err
		}
	case itemLeftParen:
		bitsetTerm.bitsetTermType = BitsetTermBracket
		bitsetTerm.bracketedExprn, err = parser.parseBitsetExpression()
		if err != nil {
			return nil, parser.errorf("Can not process bracketed expression")
		}
		err = parser.match(itemRightParen, "Bracketed expression")
		if err != nil {
			return nil, err
		}
	default:
		return nil, parser.errorf("Invalid bitset term")
	}
	return bitsetTerm, nil
}

func (parser *Parser) parseBitsetLiteral() (bitsetMap BitsetMap, err error) {
	bitsetMap = make(BitsetMap)

	// [3, 5] or ['alias1', 'alias2', 'alias3']
loop:
	for {
		item := parser.nextItem()
		switch item.typ {
		case itemRightSquareBracket:
			break loop
		case itemIntegerLiteral:
			x, err := strconv.ParseUint(item.val, 10, 32)
			if err != nil {
				return nil, err
			}
			bitsetMap[uint(x)] = true
		case itemAlias:
			x, ok := parser.variables.intAliases[item.val]
			if !ok {
				return nil, parser.errorf("Invalid bitsring alias")
			}
			if x < 0 {
				return nil, parser.errorf("Non positive integer alias in bitset: %s", item.val)
			}
			parser.nextItem()
			bitsetMap[uint(x)] = true
		case itemComma:
			parser.nextItem()
			// ignore commas
		default:
			return nil, fmt.Errorf("Invalid bitset literal")
		}
	}
	return bitsetMap, nil
}

// parse the value of a bitset
// if literalOnly then the values must be known at parse time
// otherwise it can have a bit position int expression to be calculated
// at interpretation time
func (parser *Parser) parseBitsetValue() (bitsetValue *BitsetValue, err error) {
	bitsetValue = new(BitsetValue)
	bitsetValue.bitPosExprns = make([]*IntExpression, 0)

	// [3, 5] or ['alias1', 'alias2', 'alias3']
loop:
	for {
		item := parser.peek()
		switch item.typ {
		case itemRightSquareBracket:
			parser.nextItem()
			break loop
		case itemComma:
			parser.nextItem()
			// ignore commas
		default:
			intExprn, err := parser.parseIntExpression()
			if err != nil {
				return nil, err
			}
			bitsetValue.bitPosExprns = append(bitsetValue.bitPosExprns, intExprn)
		}
	}
	return bitsetValue, nil
}

//
//	<string-expression> ::= <str-term> {<binary-str-operator> <str-term>}
//	<string-term> ::= <string-literal> | <identifier> | str(<expression>)
//	                     | <lparen><string-expression><rparen>
func (parser *Parser) parseStrExpression() (strExprn *StringExpression, err error) {
	strExprn = new(StringExpression)

	// process 1st erm
	strTerm, err := parser.parseStrTerm()
	if err != nil {
		return nil, err
	}
	strExprn.addTerms = append(strExprn.addTerms, strTerm)

	// optionally process others
	for parser.peek().typ == itemPlus {
		parser.nextItem()
		strTerm, err = parser.parseStrTerm()
		if err != nil {
			return nil, err
		}
		strExprn.addTerms = append(strExprn.addTerms, strTerm)
	}
	return strExprn, nil

}

func (parser *Parser) parseOidExpression() (oidExprn *OidExpression, err error) {
	oidExprn = new(OidExpression)

	// process 1st erm
	oidTerm, err := parser.parseOidTerm()
	if err != nil {
		return nil, err
	}
	oidExprn.addTerms = append(oidExprn.addTerms, oidTerm)

	// optionally process others
	for parser.peek().typ == itemPlus {
		parser.nextItem()
		oidTerm, err = parser.parseOidTerm()
		if err != nil {
			return nil, err
		}
		oidExprn.addTerms = append(oidExprn.addTerms, oidTerm)
	}
	return oidExprn, nil
}

func (parser *Parser) parseIntExpression() (intExprn *IntExpression, err error) {
	intExprn = new(IntExpression)

	// process 1st term
	intTerm, err := parser.parseIntTerm()
	if err != nil {
		return nil, err
	}
	intExprn.plusTerms = append(intExprn.plusTerms, intTerm)

	// optionally process others
	var usingPlus bool
loop:
	for {
		switch parser.peek().typ {
		case itemPlus:
			usingPlus = true
		case itemMinus:
			usingPlus = false
		default:
			break loop
		}
		parser.nextItem()
		intTerm, err := parser.parseIntTerm()
		if err != nil {
			return nil, err
		}
		if usingPlus {
			intExprn.plusTerms = append(intExprn.plusTerms, intTerm)
		} else {
			intExprn.minusTerms = append(intExprn.minusTerms, intTerm)
		}
	}
	return intExprn, nil
}

func (parser *Parser) parseIntTerm() (intTerm *IntTerm, err error) {
	intTerm = new(IntTerm)

	// process 1st factor
	intFactor, err := parser.parseIntFactor()
	if err != nil {
		return nil, err
	}
	intTerm.timesFactors = append(intTerm.timesFactors, intFactor)

	// optionally process others
	var usingTimes bool
loop:
	for {
		switch parser.peek().typ {
		case itemTimes:
			usingTimes = true
		case itemDivide:
			usingTimes = false
		default:
			break loop
		}
		parser.nextItem()
		intFactor, err := parser.parseIntFactor()
		if err != nil {
			return nil, err
		}
		if usingTimes {
			intTerm.timesFactors = append(intTerm.timesFactors, intFactor)
		} else {
			intTerm.divideFactors = append(intTerm.divideFactors, intFactor)
		}
	}
	return intTerm, nil
}

//<bool-term>::=<bool-factor>{<and><bool-factor>}
func (parser *Parser) parseBoolTerm() (boolTerm *BoolTerm, err error) {
	boolTerm = new(BoolTerm)

	// process 1st factor
	boolFactor, err := parser.parseBoolFactor()
	if err != nil {
		return nil, err
	}
	boolTerm.boolAndFactors = append(boolTerm.boolAndFactors, boolFactor)

	// optionally process others
	for parser.peek().typ == itemAnd {
		parser.nextItem()
		boolFactor, err = parser.parseBoolFactor()
		if err != nil {
			return nil, err
		}
		boolTerm.boolAndFactors = append(boolTerm.boolAndFactors, boolFactor)
	}
	return boolTerm, err
}

//	<string-term> ::= <string-literal> | <identifier>
//				| strInt(<int-expression>) | strBool(<bool-expression>)
//	            | <lparen><string-expression><rparen>
func (parser *Parser) parseStrTerm() (strTerm *StringTerm, err error) {
	strTerm = new(StringTerm)

	item := parser.nextItem()
	switch item.typ {
	case itemIdentifier:
		if parser.lookupType(item.val) != ValueString {
			return nil, parser.errorf("Not string variable in string expression")
		}
		strTerm.strTermType = StringTermId
		strTerm.identifier = item.val
	case itemStringLiteral:
		strTerm.strTermType = StringTermValue
		strTerm.strVal = item.val
	case itemLeftParen:
		strTerm.strTermType = StringTermBracket
		strTerm.bracketedExprn, err = parser.parseStrExpression()
		if err != nil {
			return nil, parser.errorf("Can not process bracketed expression")
		}
		err = parser.match(itemRightParen, "Bracketed expression")
		if err != nil {
			return nil, err
		}
	case itemStrBool:
		err = parser.match(itemLeftParen, "strBool")
		if err != nil {
			return nil, err
		}
		strTerm.strTermType = StringTermStringedBoolExprn
		strTerm.stringedBoolExprn, err = parser.parseBoolExpression()
		if err != nil {
			return nil, parser.errorf("Can not process stringed expression")
		}
		err = parser.match(itemRightParen, "Stringify expression")
		if err != nil {
			return nil, err
		}
	case itemStrInt, itemStrCounter, itemStrTimeticks:
		err = parser.match(itemLeftParen, "strInt/Counter/Timeticks")
		if err != nil {
			return nil, err
		}
		strTerm.strTermType = StringTermStringedIntExprn
		strTerm.stringedIntExprn, err = parser.parseIntExpression()
		if err != nil {
			return nil, parser.errorf("Can not process stringed expression")
		}
		err = parser.match(itemRightParen, "Stringify expression")
		if err != nil {
			return nil, err
		}
	case itemStrOid:
		err = parser.match(itemLeftParen, "strOId")
		if err != nil {
			return nil, err
		}
		strTerm.strTermType = StringTermStringedOidExprn
		strTerm.stringedOidExprn, err = parser.parseOidExpression()
		if err != nil {
			return nil, parser.errorf("Can not process stringed expression")
		}
		err = parser.match(itemRightParen, "Stringify expression")
		if err != nil {
			return nil, err
		}
	case itemStrIpaddress:
		err = parser.match(itemLeftParen, "strIpaddress")
		if err != nil {
			return nil, err
		}
		strTerm.strTermType = StringTermStringedAddrExprn
		strTerm.stringedAddrExprn, err = parser.parseAddrExpression()
		if err != nil {
			return nil, parser.errorf("Can not process stringed expression")
		}
		err = parser.match(itemRightParen, "Stringify expression")
		if err != nil {
			return nil, err
		}
	case itemStrBitset:
		err = parser.match(itemLeftParen, "strBitset")
		if err != nil {
			return nil, err
		}
		strTerm.strTermType = StringTermStringedBitsetExprn
		strTerm.stringedBitsetExprn, err = parser.parseBitsetExpression()
		if err != nil {
			return nil, parser.errorf("Can not process stringed expression")
		}
		err = parser.match(itemRightParen, "Stringify expression")
		if err != nil {
			return nil, err
		}
	default:
		return nil, parser.errorf("Invalid string term")
	}
	return strTerm, nil
}

func (parser *Parser) validateAddrStr(addrStr string) (err error) {
	components := strings.Split(addrStr, ".")
	if len(components) != 4 {
		return parser.errorf("Wrong number of parts of address: %d", len(components))
	}
	for _, comp := range components {
		_, err := strconv.ParseUint(comp, 10, 8)
		if err != nil {
			return parser.errorf("Invalid address component: %v", err)
		}
	}
	return nil
}

func (parser *Parser) parseAddrExpression() (addrExprn *AddrExpression, err error) {
	addrExprn = new(AddrExpression)

	item := parser.nextItem()
	switch item.typ {
	case itemIdentifier:
		if parser.lookupType(item.val) != ValueIpv4address {
			return nil, parser.errorf("Not address variable in address expression")
		}
		addrExprn.addrExprnType = AddrExprnId
		addrExprn.identifier = item.val
	case itemOidLiteral:
		err := parser.validateAddrStr(item.val)
		if err != nil {
			return nil, err
		}
		addrExprn.addrExprnType = AddrExprnValue
		addrExprn.addrVal = item.val
	default:
		return nil, parser.errorf("Invalid address expression")
	}
	return addrExprn, nil
}

func (parser *Parser) parseOidTerm() (oidTerm *OidTerm, err error) {
	oidTerm = new(OidTerm)

	item := parser.nextItem()
	switch item.typ {
	case itemIdentifier:
		if parser.lookupType(item.val) != ValueOid {
			return nil, parser.errorf("Not oid variable in oid expression")
		}
		oidTerm.oidTermType = OidTermId
		oidTerm.identifier = item.val
	case itemOidLiteral:
		oidTerm.oidTermType = OidTermValue
		oidTerm.oidVal = item.val
	case itemLeftParen:
		oidTerm.oidTermType = OidTermBracket
		oidTerm.bracketedExprn, err = parser.parseOidExpression()
		if err != nil {
			return nil, parser.errorf("Can not process bracketed expression")
		}
		err = parser.match(itemRightParen, "Bracketed expression")
		if err != nil {
			return nil, err
		}
	default:
		return nil, parser.errorf("Invalid oid term")
	}
	return oidTerm, nil
}

//<bool-factor>::=<bool-constant>|<not><bool-factor>|(<bool-expression>)
//                |<int-comparison>
func (parser *Parser) parseBoolFactor() (boolFactor *BoolFactor, err error) {
	boolFactor = new(BoolFactor)

	item := parser.peek()
	match := false
	switch item.typ {
	case itemIdentifier:
		// only match on boolean variables
		if parser.lookupType(item.val) == ValueBoolean {
			parser.nextItem()
			boolFactor.boolFactorType = BoolFactorId
			boolFactor.boolIdentifier = item.val
			match = true
		}
	case itemTrue:
		parser.nextItem()
		boolFactor.boolFactorType = BoolFactorConst
		boolFactor.boolConst = true
		match = true
	case itemFalse:
		parser.nextItem()
		boolFactor.boolFactorType = BoolFactorConst
		boolFactor.boolConst = false
		match = true
	case itemNot:
		parser.nextItem()
		boolFactor.boolFactorType = BoolFactorNot
		boolFactor.notBoolFactor, err = parser.parseBoolFactor()
		if err != nil {
			return nil, parser.errorf("Not missing factor")
		}
		match = true
	case itemLeftParen:
		parser.nextItem()
		boolFactor.boolFactorType = BoolFactorBracket
		boolFactor.bracketedExprn, err = parser.parseBoolExpression()
		if err != nil {
			return nil, parser.errorf("Can not process bracketed expression")
		}

		err = parser.match(itemRightParen, "Bracketed expression")
		if err != nil {
			return nil, err
		}
		match = true
	}
	if !match {
		boolFactor.boolFactorType = BoolFactorIntComparison
		boolFactor.intComparison, err = parser.parseIntComparison()
		if err != nil {
			return nil, err
		}
	}
	return boolFactor, nil
}

func (parser *Parser) parseIntComparison() (intComp *IntComparison, err error) {
	intComp = new(IntComparison)

	intComp.lhsIntExpression, err = parser.parseIntExpression()
	if err != nil {
		return nil, err
	}

	item := parser.nextItem()
	switch item.typ {
	case itemLessThan:
		intComp.intComparator = IntCompLessThan
	case itemLessEquals:
		intComp.intComparator = IntCompLessEquals
	case itemGreaterThan:
		intComp.intComparator = IntCompGreaterThan
	case itemGreaterEquals:
		intComp.intComparator = IntCompGreaterEquals
	case itemEquals:
		intComp.intComparator = IntCompEquals
	default:
		return nil, parser.errorf("Bad operator for integer")
	}

	intComp.rhsIntExpression, err = parser.parseIntExpression()
	if err != nil {
		return nil, err
	}

	return intComp, nil
}

func (parser *Parser) parseIntFactor() (intFactor *IntFactor, err error) {
	intFactor = new(IntFactor)

	item := parser.nextItem()
	switch item.typ {
	case itemIdentifier:
		valType := parser.lookupType(item.val)
		if valType != ValueInteger && valType != ValueCounter && valType != ValueTimeticks {
			return nil, parser.errorf("Not numeric variable in integer expression")
		}
		intFactor.intFactorType = IntFactorId
		intFactor.intIdentifier = item.val
	case itemIntegerLiteral:
		intFactor.intFactorType = IntFactorConst
		intFactor.intConst, err = strconv.Atoi(item.val)
		if err != nil {
			return nil, parser.errorf("Invalid integer literal")
		}
	case itemAlias:
		intFactor.intFactorType = IntFactorConst
		x, ok := parser.variables.intAliases[item.val]
		if !ok {
			return nil, parser.errorf("Invalid integer alias")
		}
		intFactor.intConst = x
	case itemMinus:
		intFactor.intFactorType = IntFactorMinus
		intFactor.minusIntFactor, err = parser.parseIntFactor()
		if err != nil {
			return nil, parser.errorf("Minus missing int factor")
		}
	case itemLeftParen:
		intFactor.intFactorType = IntFactorBracket
		intFactor.bracketedExprn, err = parser.parseIntExpression()
		if err != nil {
			return nil, parser.errorf("Can not process bracketed expression")
		}

		err = parser.match(itemRightParen, "Bracketed expression")
		if err != nil {
			return nil, err
		}
	default:
		return nil, parser.errorf("Invalid item/operator in integer factor")
	}
	return intFactor, nil
}

type InitMode int

const (
	InitModeZero InitMode = iota
	InitModeExternal
)

type SnmpMode int

const (
	SnmpModeRead SnmpMode = iota
	SnmpModeReadWrite
)

type Type struct {
	valueType     ValueType
	oid           string
	initMode      InitMode
	snmpMode      SnmpMode
	externalValue chan *Value
	lineNum       int
}

func (typ Type) String() string {
	var str string
	switch typ.valueType {
	case ValueInteger:
		str = "Integer"
	case ValueCounter:
		str = "Counter"
	case ValueTimeticks:
		str = "Timeticks"
	case ValueString:
		str = "String"
	case ValueBoolean:
		str = "Boolean"
	case ValueBitset:
		str = "Bitset"
	case ValueOid:
		str = "Oid"
	case ValueNone:
		str = "None"
	}
	if len(typ.oid) > 0 {
		str += fmt.Sprintf(" oid: %s", typ.oid)
	}
	return str
}

func (loopTyp LoopType) String() string {
	switch loopTyp {
	case LoopForever:
		return "forever"
	case LoopTimes:
		return "times"
	case LoopWhile:
		return "while"
	}
	return "unknown loop"
}

func (intComp IntComparatorType) String() string {
	switch intComp {
	case IntCompEquals:
		return "Equals ="
	case IntCompGreaterEquals:
		return "Greater or Equals >="
	case IntCompGreaterThan:
		return "Greater than >"
	case IntCompLessEquals:
		return "Less or Equals <="
	case IntCompLessThan:
		return "Less than <"
	}
	return "unknown operator"
}

type Statement struct {
	stmtType StatementType

	assignmentStmt *AssignmentStatement
	ifStmt         *IfStatement
	loopStmt       *LoopStatement
	printStmt      *PrintStatement
	sleepStmt      *SleepStatement
	readStmt       *ReadStatement
}

type LoopStatement struct {
	loopType LoopType

	intExpression  *IntExpression
	boolExpression *BoolExpression
	stmtList       []*Statement
}

type IfStatement struct {
	boolExpression *BoolExpression
	stmtList       []*Statement
	elsifList      []*ElseIf
	elseStmtList   []*Statement
}

type ElseIf struct {
	boolExpression *BoolExpression
	stmtList       []*Statement
}

type AssignmentStatement struct {
	identifier string
	exprn      *Expression
}

type PrintStatement struct {
	exprn *StringExpression
}

type ReadStatement struct {
	identifier string
}

type SleepStatement struct {
	exprn *IntExpression
	units TimeUnit
}

type Expression struct {
	exprnType ExpressionType

	intExpression    *IntExpression
	boolExpression   *BoolExpression
	stringExpression *StringExpression
	bitsetExpression *BitsetExpression
	oidExpression    *OidExpression
	addrExpression   *AddrExpression
}

//<bool-expression>::=<bool-term>{<or><bool-term>}
//<bool-term>::=<bool-factor>{<and><bool-factor>}
//<bool-factor>::=<bool-constant>|<bool-identifier>|<not><bool-factor>|(<bool-expression>)|<int-comparison>
//<int-comparison>::=<int-expression><int-comp><int-expression>

type BoolExpression struct {
	boolOrTerms []*BoolTerm
}

type BoolTerm struct {
	boolAndFactors []*BoolFactor
}

type BoolFactorType int

const (
	BoolFactorConst BoolFactorType = iota
	BoolFactorId
	BoolFactorNot
	BoolFactorBracket
	BoolFactorIntComparison
)

type BoolFactor struct {
	boolFactorType BoolFactorType

	boolConst      bool
	boolIdentifier string
	notBoolFactor  *BoolFactor
	bracketedExprn *BoolExpression
	intComparison  *IntComparison
}

type IntComparison struct {
	// integer comparisons: <, >, <=, >=, =
	intComparator IntComparatorType

	lhsIntExpression *IntExpression
	rhsIntExpression *IntExpression
}

//<int-expression>::=<int-term>{<plus-or-minus><int-term>}
//<int-term>::=<int-factor>{<times-or-divide><int-factor>}
//<int-factor>::=<int-constant>|<int-identifier>|<minus><int-factor>|(<int-expression>)

type IntExpression struct {
	plusTerms  []*IntTerm
	minusTerms []*IntTerm
}

type IntTerm struct {
	timesFactors  []*IntFactor
	divideFactors []*IntFactor
}

type IntFactorType int

const (
	IntFactorConst IntFactorType = iota
	IntFactorId
	IntFactorMinus
	IntFactorBracket
)

type IntFactor struct {
	intFactorType IntFactorType

	intConst       int
	intIdentifier  string
	minusIntFactor *IntFactor
	bracketedExprn *IntExpression
}

// <string-expression> ::= <str-term> {<binary-str-operator> <str-term>}
// <string-term> ::= <string-literal> | <identifier> | str(<expression>) | <lparen><string-expression><rparen>

type StringExpression struct {
	addTerms []*StringTerm
}

type BitsetExpression struct {
	plusTerms  []*BitsetTerm
	minusTerms []*BitsetTerm
}

type OidExpression struct {
	addTerms []*OidTerm
}

type AddrExprnType int

const (
	AddrExprnValue AddrExprnType = iota
	AddrExprnId
)

type AddrExpression struct {
	addrExprnType AddrExprnType
	addrVal       string
	identifier    string
}

type OidTermType int

const (
	OidTermValue OidTermType = iota
	OidTermId
	OidTermBracket
)

type OidTerm struct {
	oidTermType OidTermType

	oidVal         string
	identifier     string
	bracketedExprn *OidExpression
}

type StringTermType int

const (
	StringTermValue StringTermType = iota
	StringTermId
	StringTermBracket
	StringTermStringedIntExprn
	StringTermStringedBoolExprn
	StringTermStringedOidExprn
	StringTermStringedAddrExprn
	StringTermStringedBitsetExprn
)

type StringTerm struct {
	strTermType StringTermType

	strVal              string
	identifier          string
	bracketedExprn      *StringExpression
	stringedIntExprn    *IntExpression
	stringedBoolExprn   *BoolExpression
	stringedOidExprn    *OidExpression
	stringedAddrExprn   *AddrExpression
	stringedBitsetExprn *BitsetExpression
}

type BitsetTermType int

const (
	BitsetTermValue BitsetTermType = iota
	BitsetTermId
	BitsetTermBracket
)

type BitsetMap map[uint]bool

func (bitsetValue BitsetMap) String() (str string) {
	var keys []int // use int instead of uint so we can use sort.Ints()
	for k := range bitsetValue {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	// To perform the opertion you want
	str = "{"
	for i, k := range keys {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%d", k)
	}
	str += "}"
	return str
}

type BitsetTerm struct {
	bitsetTermType BitsetTermType

	bitsetVal      *BitsetValue
	identifier     string
	bracketedExprn *BitsetExpression
}

type BitsetValue struct {
	bitPosExprns []*IntExpression // regular case
}
