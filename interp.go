package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Value struct {
	valueType ValueType

	intVal    int
	stringVal string
	boolVal   bool
	bitsetVal BitsetMap
	oidVal    string
	addrVal   string
}

func (v *Value) String() string {
	str := ""
	switch v.valueType {
	case ValueBoolean:
		str += fmt.Sprintf("<Boolean: %t>", v.boolVal)
	case ValueInteger:
		str += fmt.Sprintf("<Integer: %d>", v.intVal)
	case ValueString:
		str += fmt.Sprintf("<String: %s>", v.stringVal)
	case ValueBitset:
		str += fmt.Sprintf("<Bitset: %v>", v.bitsetVal)
	case ValueOid:
		str += fmt.Sprintf("<OID: %s>", v.oidVal)
	case ValueNone:
		str += "<none>"
	}
	return str
}

type Interpreter struct {
	variables   *Variables
	values      map[string]*Value // variable id --> Value
	oid2Values  map[string]*Value // oid --> Value
	oid2ValLock sync.RWMutex
}

// GetValueForOid is a thread safe version of getting value from oid map
func (interp *Interpreter) GetValueForOid(oidStr string) (val *Value, found bool) {
	interp.oid2ValLock.RLock()
	defer interp.oid2ValLock.RUnlock()
	val, found = interp.oid2Values[oidStr]
	if !found {
		return nil, false
	}
	return val, true
}

// SetValueForOid is a thread safe version of setting value in the oid map
func (interp *Interpreter) SetValueForOid(oidStr string, val *Value) {
	interp.oid2ValLock.Lock()
	defer interp.oid2ValLock.Unlock()
	interp.oid2Values[oidStr] = val
}

// Prompt for input of variables
// If there is an error then ask again
func promptForInput(id string, val *Value, variables *Variables) {
	// prompt for input
	for {
		fmt.Printf("Input %s: ", id)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, "\n ")
		var err error
		fail := false
		switch val.valueType {
		case ValueString:
			val.stringVal = text
		case ValueInteger, ValueCounter, ValueTimeticks:
			val.intVal, err = strconv.Atoi(text)
			if err != nil {
				fail = true
				fmt.Printf("Invalid integer/counter/ticks: %v\n", err)
			}
		case ValueBoolean:
			val.boolVal, err = strconv.ParseBool(text)
			if err != nil {
				fail = true
				fmt.Printf("Invalid boolean: %v\n", err)
			}
		case ValueIpv4address:
			val.addrVal = text
			err = isValidIpv4Address(text)
			if err != nil {
				fail = true
				fmt.Printf("Invalid ipv4address: %v\n", err)
			}
		case ValueOid:
			val.oidVal = text
			err = isValidOID(text)
			if err != nil {
				fail = true
				fmt.Printf("Invalid OID: %v\n", err)
			}
		case ValueBitset:
			// Bitset more complex and can refer to aliases
			// So run our parser over it and use existing aliases from program
			l := lex("", text)
			parser := NewParser(l)
			parser.variables = new(Variables)
			parser.variables.intAliases = variables.intAliases
			err := parser.match(itemLeftSquareBracket, "bitset input")
			if err != nil {
				fail = true
				fmt.Printf("Invalid bitset map: %v\n", err)
			} else {
				bitsetMap, err := parser.parseBitsetLiteral()
				if err != nil {
					fail = true
					fmt.Printf("Invalid bitset map: %v\n", err)
				}
				val.bitsetVal = bitsetMap
			}
		}
		if !fail {
			return
		}
	}
}

// Init initializes the interpreter
// Must call before interpreting program
func (interp *Interpreter) Init(prog *Program) {
	interp.variables = prog.variables

	/* initialise variables based on the types */
	interp.values = make(map[string]*Value)
	interp.oid2Values = make(map[string]*Value)

	// sort types according to line#s for deterministic order input
	var ids []string
	for i := range interp.variables.types {
		ids = append(ids, i)
	}
	sort.Slice(ids, func(i, j int) bool {
		return interp.variables.types[ids[i]].lineNum <
			interp.variables.types[ids[j]].lineNum
	})

	for _, id := range ids {
		typ := interp.variables.types[id]
		//fmt.Printf("id = %s, typ = %v\n", id, typ)
		val := new(Value)
		val.valueType = typ.valueType

		if typ.externalInput {
			promptForInput(id, val, interp.variables)
		}

		interp.values[id] = val
		if len(typ.oid) > 0 {
			//fmt.Printf("%s: %v\n", typ.oid, val)
			interp.oid2Values[typ.oid] = val
		}
	}
}

func isValidOID(str string) (err error) {
	fields := strings.Split(str, ".")
	for _, x := range fields {
		_, err := strconv.ParseUint(x, 10, 32)
		if err != nil {
			return err
		}
	}
	return nil
}

func isValidIpv4Address(str string) (err error) {
	fields := strings.Split(str, ".")

	if len(fields) != 4 {
		return errors.New("Address not 4 fields")
	}

	for _, x := range fields {
		_, err := strconv.ParseUint(x, 10, 8)
		if err != nil {
			return err
		}
	}
	return nil
}

// InterpProgram Interprets the program aka runs the program
// prog - the program parse tree to run
func (interp *Interpreter) InterpProgram(prog *Program) (err error) {
	_, err = interp.interpStatementList(prog.stmtList)
	if err != nil {
		return err
	}
	return nil
}

func (interp *Interpreter) interpStatementList(stmtList []*Statement) (isExit bool, err error) {
	for _, stmt := range stmtList {
		exit, err := interp.interpStatement(stmt)
		if err != nil {
			return false, err
		}
		if exit {
			return true, nil
		}
	}
	return false, nil
}

func (interp *Interpreter) interpStatement(stmt *Statement) (isExit bool, err error) {
	err = nil
	isExit = false
	switch stmt.stmtType {
	case StmtAssignment:
		err = interp.interpAssignmentStmt(stmt.assignmentStmt)
	case StmtIf:
		isExit, err = interp.interpIfStmt(stmt.ifStmt)
	case StmtLoop:
		err = interp.interpLoopStmt(stmt.loopStmt)
	case StmtPrint:
		err = interp.interpPrintStmt(stmt.printStmt)
	case StmtSleep:
		err = interp.interpSleepStmt(stmt.sleepStmt)
	case StmtExit:
		return true, nil
	}
	return isExit, err
}

func (interp *Interpreter) interpIfStmt(ifStmt *IfStatement) (isExit bool, err error) {
	val, err := interp.interpBoolExpression(ifStmt.boolExpression)
	if err != nil {
		return false, err
	}
	if val {
		return interp.interpStatementList(ifStmt.stmtList)
	}
	for _, elif := range ifStmt.elsifList {
		val, err = interp.interpBoolExpression(elif.boolExpression)
		if err != nil {
			return false, err
		}
		if val {
			return interp.interpStatementList(elif.stmtList)
		}
	}
	// no matches - check out the else if there is one
	return interp.interpStatementList(ifStmt.elseStmtList)
}

func (interp *Interpreter) interpPrintStmt(printStmt *PrintStatement) (err error) {
	val, err := interp.interpStringExpression(printStmt.exprn)
	if err != nil {
		return err
	}
	fmt.Println(val) // TODO: handle backslash characters
	return nil
}

func (interp *Interpreter) interpSleepStmt(sleepStmt *SleepStatement) (err error) {
	duration, err := interp.interpIntExpression(sleepStmt.exprn)
	if err != nil {
		return err
	}
	switch sleepStmt.units {
	case TimeSecs:
		time.Sleep(time.Duration(duration) * time.Second)
	case TimeMillis:
		time.Sleep(time.Duration(duration) * time.Millisecond)
	}
	return nil
}

func (interp *Interpreter) interpLoopStmt(loopStmt *LoopStatement) (err error) {
	switch loopStmt.loopType {
	case LoopForever:
		for {
			exit, err := interp.interpStatementList(loopStmt.stmtList)
			if err != nil {
				return err
			}
			if exit {
				break
			}
		}
	case LoopTimes:
		n, err := interp.interpIntExpression(loopStmt.intExpression)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			exit, err := interp.interpStatementList(loopStmt.stmtList)
			if err != nil {
				return err
			}
			if exit {
				break
			}
		}
	case LoopWhile:
		for {
			val, err := interp.interpBoolExpression(loopStmt.boolExpression)
			if err != nil {
				return err
			}
			if !val {
				break
			}
			exit, err := interp.interpStatementList(loopStmt.stmtList)
			if err != nil {
				return err
			}
			if exit {
				break
			}
		}
	}
	return nil
}

func (interp *Interpreter) interpAssignmentStmt(assign *AssignmentStatement) (err error) {
	value, err := interp.interpExpression(assign.exprn)
	if err != nil {
		return err
	}
	varType := interp.variables.types[assign.identifier]
	value.valueType = varType.valueType // ensure counter/timeticks overrides integer type expression

	interp.values[assign.identifier] = value
	interp.SetValueForOid(varType.oid, value)
	//fmt.Printf("setvalue: %s %v\n", typ.oid, value)
	return nil
}

func (interp *Interpreter) interpExpression(exprn *Expression) (val *Value, err error) {
	val = new(Value)
	switch exprn.exprnType {
	case ExprnBoolean:
		val.valueType = ValueBoolean
		val.boolVal, err = interp.interpBoolExpression(exprn.boolExpression)
	case ExprnInteger:
		val.valueType = ValueInteger
		val.intVal, err = interp.interpIntExpression(exprn.intExpression)
	case ExprnString:
		val.valueType = ValueString
		val.stringVal, err = interp.interpStringExpression(exprn.stringExpression)
	case ExprnBitset:
		val.valueType = ValueBitset
		val.bitsetVal, err = interp.interpBitsetExpression(exprn.bitsetExpression)
	case ExprnOid:
		val.valueType = ValueOid
		val.oidVal, err = interp.interpOidExpression(exprn.oidExpression)
	case ExprnAddr:
		val.valueType = ValueIpv4address
		val.addrVal, err = interp.interpAddrExpression(exprn.addrExpression)
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (interp *Interpreter) interpBitsetExpression(exprn *BitsetExpression) (val BitsetMap, err error) {
	newBitset := BitsetMap(make(map[uint]bool))
	for _, plusTerm := range exprn.plusTerms {
		bitset, err := interp.interpBitsetTerm(plusTerm)
		if err != nil {
			return nil, err
		}
		for k, v := range bitset {
			newBitset[k] = v
		}
	}
	for _, minusTerm := range exprn.minusTerms {
		bitset, err := interp.interpBitsetTerm(minusTerm)
		if err != nil {
			return nil, err
		}
		for k := range bitset {
			delete(newBitset, k)
		}

	}
	return newBitset, nil
}

func (interp *Interpreter) interpBitsetTerm(term *BitsetTerm) (val BitsetMap, err error) {
	switch term.bitsetTermType {
	case BitsetTermValue:
		return term.bitsetVal, nil
	case BitsetTermId:
		val := interp.values[term.identifier]
		return val.bitsetVal, nil
	case BitsetTermBracket:
		return interp.interpBitsetExpression(term.bracketedExprn)
	}
	return nil, fmt.Errorf("Invalid bitset type: %d", term.bitsetTermType)
}

func (interp *Interpreter) interpStringExpression(strExprn *StringExpression) (string, error) {
	str := ""
	for _, term := range strExprn.addTerms {
		s, err := interp.interpStringTerm(term)
		if err != nil {
			return "", err
		}
		str += s
	}
	return str, nil
}

func (interp *Interpreter) interpOidExpression(oidExprn *OidExpression) (string, error) {
	str := ""
	for _, term := range oidExprn.addTerms {
		s, err := interp.interpOidTerm(term)
		if err != nil {
			return "", err
		}
		str += s
	}
	return str, nil
}

func (interp *Interpreter) interpOidTerm(oidTerm *OidTerm) (string, error) {
	switch oidTerm.oidTermType {
	case OidTermValue:
		return oidTerm.oidVal, nil
	case OidTermBracket:
		return interp.interpOidExpression(oidTerm.bracketedExprn)
	case OidTermId:
		val := interp.values[oidTerm.identifier]
		return val.oidVal, nil
	}
	return "", nil
}

func (interp *Interpreter) interpAddrExpression(addrExprn *AddrExpression) (string, error) {
	switch addrExprn.addrExprnType {
	case AddrExprnValue:
		return addrExprn.addrVal, nil
	case AddrExprnId:
		val := interp.values[addrExprn.identifier]
		return val.addrVal, nil
	}
	return "", nil
}

func (interp *Interpreter) interpStringTerm(strTerm *StringTerm) (string, error) {
	switch strTerm.strTermType {
	case StringTermValue:
		return strTerm.strVal, nil
	case StringTermBracket:
		return interp.interpStringExpression(strTerm.bracketedExprn)
	case StringTermId:
		val := interp.values[strTerm.identifier]
		return val.stringVal, nil
	case StringTermStringedBoolExprn:
		b, err := interp.interpBoolExpression(strTerm.stringedBoolExprn)
		if err != nil {
			return "", err
		}
		return strconv.FormatBool(b), nil
	case StringTermStringedIntExprn:
		i, err := interp.interpIntExpression(strTerm.stringedIntExprn)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(i), nil
	case StringTermStringedOidExprn:
		o, err := interp.interpOidExpression(strTerm.stringedOidExprn)
		if err != nil {
			return "", err
		}
		return o, nil
	case StringTermStringedAddrExprn:
		a, err := interp.interpAddrExpression(strTerm.stringedAddrExprn)
		if err != nil {
			return "", err
		}
		return a, nil
	case StringTermStringedBitsetExprn:
		b, err := interp.interpBitsetExpression(strTerm.stringedBitsetExprn)
		if err != nil {
			return "", err
		}
		return b.String(), nil
	}
	return "", nil
}

func (interp *Interpreter) interpBoolExpression(boolExprn *BoolExpression) (val bool, err error) {
	for _, term := range boolExprn.boolOrTerms {
		val, err = interp.interpBoolTerm(term)
		if err != nil {
			return false, err
		}
		if val {
			return true, nil
		}
	}
	return false, nil
}

func (interp *Interpreter) interpBoolTerm(boolTerm *BoolTerm) (val bool, err error) {
	for _, factor := range boolTerm.boolAndFactors {
		val, err = interp.interpBoolFactor(factor)
		if err != nil {
			return false, err
		}
		if !val {
			return false, nil
		}
	}
	return true, nil
}

func (interp *Interpreter) interpBoolFactor(boolFactor *BoolFactor) (val bool, err error) {
	switch boolFactor.boolFactorType {
	case BoolFactorConst:
		return boolFactor.boolConst, nil
	case BoolFactorNot:
		val, err = interp.interpBoolFactor(boolFactor.notBoolFactor)
		return !val, err
	case BoolFactorBracket:
		return interp.interpBoolExpression(boolFactor.bracketedExprn)
	case BoolFactorId:
		value := interp.values[boolFactor.boolIdentifier]
		return value.boolVal, nil
	case BoolFactorIntComparison:
		return interp.interpIntComparison(boolFactor.intComparison)
	}
	return false, nil
}

func (interp *Interpreter) interpIntComparison(intComparison *IntComparison) (bool, error) {
	lhs, err := interp.interpIntExpression(intComparison.lhsIntExpression)
	if err != nil {
		return false, err
	}
	rhs, err := interp.interpIntExpression(intComparison.rhsIntExpression)
	if err != nil {
		return false, err
	}
	switch intComparison.intComparator {
	case IntCompEquals:
		return lhs == rhs, nil
	case IntCompGreaterEquals:
		return lhs >= rhs, nil
	case IntCompGreaterThan:
		return lhs > rhs, nil
	case IntCompLessEquals:
		return lhs <= rhs, nil
	case IntCompLessThan:
		return lhs < rhs, nil
	}
	return false, nil
}

func (interp *Interpreter) interpIntExpression(intExpression *IntExpression) (int, error) {
	val := 0
	for _, term := range intExpression.plusTerms {
		plusVal, err := interp.interpIntTerm(term)
		if err != nil {
			return 0, err
		}
		val += plusVal
	}
	for _, term := range intExpression.minusTerms {
		minusVal, err := interp.interpIntTerm(term)
		if err != nil {
			return 0, err
		}
		val -= minusVal
	}
	return val, nil
}

func (interp *Interpreter) interpIntTerm(intTerm *IntTerm) (int, error) {
	val := 1
	for _, factor := range intTerm.timesFactors {
		timesVal, err := interp.interpIntFactor(factor)
		if err != nil {
			return 1, err
		}
		val *= timesVal
	}
	for _, factor := range intTerm.divideFactors {
		divideVal, err := interp.interpIntFactor(factor)
		if err != nil {
			return 1, err
		}
		val /= divideVal
	}
	return val, nil
}

func (interp *Interpreter) interpIntFactor(intFactor *IntFactor) (int, error) {
	switch intFactor.intFactorType {
	case IntFactorConst:
		return intFactor.intConst, nil
	case IntFactorBracket:
		return interp.interpIntExpression(intFactor.bracketedExprn)
	case IntFactorId:
		value := interp.values[intFactor.intIdentifier]
		return value.intVal, nil
	case IntFactorMinus:
		value, err := interp.interpIntFactor(intFactor.minusIntFactor)
		if err != nil {
			return 0, err
		}
		return -value, nil
	}
	return 0, nil
}
