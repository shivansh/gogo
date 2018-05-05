// Package ast implements utility functions for generating abstract syntax
// trees from Go BNF.

package ast

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shivansh/gogo/src/utils"
	"github.com/shivansh/gogo/tmp/token"
)

type symTabType map[string][]string

type SymInfo struct {
	varSymTab symTabType
	parent    *SymInfo
}

// DeferStackItem is an individual item stored when a call to defer is made. It
// contains the code for the corresponding function call which is placed at the
// end of function body.
type DeferStackItem []string

var (
	tmpIndex   int
	labelIndex int
	varIndex   int
	// funcSymtabCreated keeps track whether a symbol table corresponding to a
	// function declaration has to be instantiated. This is because usually
	// a new symbol table is created when the corresponding block begins.
	// However, in case of functions the arguments also need to be added to
	// the symbol table. Thus the symbol table is instantiated when the
	// production rule corresponding to the arguments is reached and not
	// when the block begins.
	funcSymtabCreated bool
	forSymtabCreated  bool
	symTab            symTabType // symbol table for temporaries ; TODO: Update this.
	// currSymTab keeps track of the currently active symbol table
	// depending on scope.
	currSymTab *SymInfo
	// globalSymTab keeps track of the global struct and function declarations.
	// NOTE: structs and functions can only be declared globally.
	globalSymTab symTabType
	// deferStack stores the deferred function calls which are then called
	// when the surrounding function block ends.
	deferStack *utils.Stack
	re         *regexp.Regexp
)

func init() {
	symTab = make(symTabType)
	globalSymTab = make(symTabType)
	// currSymTab now allocates space for a global symbol table for variable
	// declarations. Ideally, instead of creating a new symbol table it
	// should've been global symbol table, but since the type of
	// currSymTab.varSymTab is not pointer, updates made elsewhere will not
	// be reflected globally.
	// TODO: Update type of currSymTab.varSymTab to a pointer.
	currSymTab = &SymInfo{make(symTabType), nil}
	deferStack = utils.CreateStack()
	re = regexp.MustCompile("(^-?[0-9]+$)") // integers
}

// SearchInScope returns the symbol table entry for a given variable in the
// current scope. If not found, the parent symbol table is looked up until the
// topmost symbol table is reached. If not found in all these symbol tables,
// then the global symbol table is looked up which contains the entries
// corresponding to structs and functions.
func SearchInScope(v string) ([]string, bool) {
	currScope := currSymTab
	for currScope != nil {
		symTabEntry, ok := currScope.varSymTab[v]
		if ok {
			return symTabEntry, true
		} else {
			currScope = currScope.parent
		}
	}
	// Lookup in global scope in case the variable corresponds to a struct
	// or a function name.
	for k, symTabEntry := range globalSymTab {
		if k == v {
			return symTabEntry, true
		}
	}
	return []string{}, false
}

// GetRealName extracts the original name of variable from its renamed version.
func GetRealName(s string) string {
	realName := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			break
		} else {
			realName = realName + string(s[i])
		}
	}
	return realName
}

// NewTmp generates a unique temporary variable.
func NewTmp() string {
	t := fmt.Sprintf("t%d", tmpIndex)
	tmpIndex++
	return t
}

// NewLabel generates a unique label name.
func NewLabel() string {
	l := fmt.Sprintf("l%d", labelIndex)
	labelIndex++
	return l
}

// Node represents a node in the AST of a given program.
type Node struct {
	// If the node represents an expression, then place stores the name of
	// the variable storing the value of the expression.
	Place string
	Code  []string // IR instructions
}

type Attrib interface{}

// NewVar generates a unique variable name used for renaming. A variable named
// var will be renamed to 'var.int_lit' where int_lit is an integer. Since
// variable names cannot contain a '.', this will not result in a naming
// conflict with an existing variable. The renamed variable will only occur in
// the IR (there is no constraint on variable names in IR as of now).
func RenameVariable(v string) string {
	ret := fmt.Sprintf("%s.%d", v, varIndex)
	varIndex++
	return ret
}

func PrintIR(src Attrib) (Attrib, error) {
	re := regexp.MustCompile("\n(\n)*")
	c := src.(Node).Code
	for _, v := range c {
		v := strings.TrimSpace(v)
		// Compress multiple newlines within IR statements into
		// a single newline.
		v = re.ReplaceAllString(v, "\n")
		if v != "" {
			fmt.Println(v)
		}
	}
	return nil, nil
}

// InitNode initializes a AST node with the given place and code values.
func InitNode(place string, code []string) (Node, error) {
	return Node{place, code}, nil
}

// NewNode creates a new AST node from the given attribute.
func NewNode(attr Attrib) (Node, error) {
	return Node{attr.(Node).Place, attr.(Node).Code}, nil
}

// --- [ Top level declarations ] ----------------------------------------------

// NewTopLevelDecl returns a top level declaration.
func NewTopLevelDecl(topDecl, repeatTopDecl Attrib) (Node, error) {
	n := Node{"", topDecl.(Node).Code}
	n.Code = append(n.Code, repeatTopDecl.(Node).Code...)
	return n, nil
}

// NewTypeDef returns a type definition.
func NewTypeDef(ident, typ Attrib) (Node, error) {
	return Node{fmt.Sprintf("%s, %s", typ.(Node).Place, string(ident.(*token.Token).Lit)), typ.(Node).Code}, nil
}

// NewElementList returns a keyed element list.
func NewElementList(key, keyList Attrib) (Node, error) {
	n := Node{fmt.Sprintf("%s, %s", key.(Node).Place, keyList.(Node).Place), key.(Node).Code}
	n.Code = append(n.Code, keyList.(Node).Code...)
	return n, nil
}

// AppendKeyedElement appends a keyed element to a list of keyed elements.
func AppendKeyedElement(key, keyList Attrib) (Node, error) {
	n := Node{fmt.Sprintf("%s, %s", key.(Node).Place, keyList.(Node).Place), key.(Node).Code}
	n.Code = append(n.Code, keyList.(Node).Code...)
	return n, nil
}

// --- [ Variable declarations ] -----------------------------------------------

// NewVarSpec creates a new variable specification.
// The accepted variadic arguments (args) in their order are -
// 	- IdentifierList
// 	- Type
// 	- ExpressionList
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewVarSpec(typ int, args ...Attrib) (Node, error) {
	n := Node{"", []string{}}
	expr := []string{}
	// Add the IR instructions for ExpressionList.
	switch typ {
	case 1:
		n.Code = args[2].(Node).Code
		expr = utils.SplitAndSanitize(args[2].(Node).Place, ",")
	case 2:
		n.Code = args[1].(Node).Code
		expr = utils.SplitAndSanitize(args[1].(Node).Place, ",")
	case 3:
		return Node{}, nil
	}
	for k, v := range args[0].(Node).Code {
		renamedVar := RenameVariable(v)
		// TODO: Handle other types
		currSymTab.varSymTab[v] = []string{renamedVar, "int"}
		if typ == 0 {
			n.Code = append(n.Code, fmt.Sprintf("=, %s, 0", renamedVar))
		} else if typ == 1 || typ == 2 {
			n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
		}
	}
	return n, nil
}

// --- [ Type declarations ] ---------------------------------------------------

func NewTypeDecl(args ...Attrib) (Node, error) {
	typeInfo := utils.SplitAndSanitize(args[1].(Node).Place, ",")
	structName := strings.TrimSpace(typeInfo[1])
	typ := strings.TrimSpace(typeInfo[0])
	switch typ {
	case "struct":
		// Create a global symbol table entry.
		// NOTE: The symbol table entry of a struct is of the form -
		//      structName : []{"struct", memberName1, memberType1, ...}
		globalSymTab[structName] = []string{"struct"}
		globalSymTab[structName] = append(globalSymTab[structName], args[1].(Node).Code...)
	default: // TODO: Add remaining types.
		return Node{}, fmt.Errorf("Unknown type %s", typ)
	}
	// TODO: Member initialization will be done when a new object is
	// instantiated.
	return Node{"", []string{}}, nil
}

// --- [ Constant declarations ] -----------------------------------------------

func NewConstSpec(typ int, args ...Attrib) (Node, error) {
	n := Node{"", []string{}}
	expr := []string{}
	if typ == 1 {
		n.Code = append(n.Code, args[1].(Node).Code...)
		expr = utils.SplitAndSanitize(args[1].(Node).Place, ",")
	}
	for k, v := range args[0].(Node).Code {
		renamedVar := RenameVariable(v)
		currSymTab.varSymTab[v] = []string{renamedVar, "int"}
		switch typ {
		case 0:
			n.Code = append(n.Code, fmt.Sprintf("=, %s, 0", renamedVar))
		case 1:
			n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
		}
	}
	return n, nil
}

// --- [ Expressions ] ---------------------------------------------------------

// NewExpr returns a new expression.
func NewExpr(expr Attrib) (Node, error) {
	return Node{expr.(Node).Place, expr.(Node).Code}, nil
}

// AppendExpr appends a list of expressions to a given expression.
func AppendExpr(expr, exprlist Attrib) (Node, error) {
	n := Node{"", expr.(Node).Code}
	n.Code = append(n.Code, exprlist.(Node).Code...)
	n.Place = fmt.Sprintf("%s,%s", expr.(Node).Place, exprlist.(Node).Place)
	return n, nil
}

// NewBoolExpr returns a new logical expression.
func NewBoolExpr(op, leftexpr, rightexpr Attrib) (Node, error) {
	n := Node{"", leftexpr.(Node).Code}
	n.Code = append(n.Code, rightexpr.(Node).Code...)
	n.Place = NewTmp()
	afterLabel := NewLabel()
	switch string(op.(*token.Token).Lit) {
	case "||":
		trueLabel := NewLabel()
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, leftexpr.(Node).Place),
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, rightexpr.(Node).Place),
			fmt.Sprintf("=, %s, 0", n.Place),
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", trueLabel),
			fmt.Sprintf("=, %s, 1", n.Place),
			fmt.Sprintf("label, %s", afterLabel),
		)
	case "&&":
		falseLabel := NewLabel()
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, leftexpr.(Node).Place),
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, rightexpr.(Node).Place),
			fmt.Sprintf("=, %s, 1", n.Place),
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", falseLabel),
			fmt.Sprintf("=, %s, 0", n.Place),
			fmt.Sprintf("label, %s", afterLabel),
		)
	}
	return n, nil
}

// NewRelExpr returns a new relational expression.
func NewRelExpr(op, leftexpr, rightexpr Attrib) (Node, error) {
	n := Node{"", leftexpr.(Node).Code}
	n.Place = NewTmp()
	n.Code = append(n.Code, rightexpr.(Node).Code...)
	branchOp := ""
	falseLabel := NewLabel()
	afterLabel := NewLabel()
	switch op.(Node).Place {
	case "==":
		branchOp = "bne"
	case "!=":
		branchOp = "beq"
	case "<=":
		branchOp = "bgt"
	case "<":
		branchOp = "bge"
	case ">=":
		branchOp = "blt"
	case ">":
		branchOp = "ble"
	}
	n.Code = utils.AppendToSlice(
		n.Code,
		fmt.Sprintf("%s, %s, %s, %s", branchOp, falseLabel, leftexpr.(Node).Place, rightexpr.(Node).Place),
		fmt.Sprintf("=, %s, 1", n.Place),
		fmt.Sprintf("j, %s", afterLabel),
		fmt.Sprintf("label, %s", falseLabel),
		fmt.Sprintf("=, %s, 0", n.Place),
		fmt.Sprintf("label, %s", afterLabel),
	)
	return n, nil
}

// NewArithExpr returns a new arithmetic expression.
func NewArithExpr(op, leftexpr, rightexpr Attrib) (Node, error) {
	n := Node{"", leftexpr.(Node).Code}
	op = string(op.(*token.Token).Lit)
	n.Code = append(n.Code, rightexpr.(Node).Code...)
	if re.MatchString(leftexpr.(Node).Place) && re.MatchString(rightexpr.(Node).Place) {
		// Expression is of the form "1 op 2".
		leftval, err := strconv.Atoi(leftexpr.(Node).Place)
		if err != nil {
			return Node{}, err
		}
		rightval, err := strconv.Atoi(rightexpr.(Node).Place)
		if err != nil {
			return Node{}, err
		}
		switch op {
		case "+":
			n.Place = strconv.Itoa(leftval + rightval)
		case "-":
			n.Place = strconv.Itoa(leftval - rightval)
		case "*":
			n.Place = strconv.Itoa(leftval * rightval)
		case "/":
			n.Place = strconv.Itoa(leftval / rightval)
		case "%":
			n.Place = strconv.Itoa(leftval % rightval)
		default:
			return Node{}, fmt.Errorf("Invalid operation %s", op)
		}
	} else if re.MatchString(leftexpr.(Node).Place) {
		// Expression is of the form "1 + b", which needs to be
		// converted to the equivalent form "b + 1" to form valid IR.
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, rightexpr.(Node).Place, leftexpr.(Node).Place))
	} else {
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, leftexpr.(Node).Place, rightexpr.(Node).Place))
	}
	return n, nil
}

func NewUnaryExpr(op, expr Attrib) (Node, error) {
	n := Node{"", expr.(Node).Code}
	switch op.(Node).Place {
	case "-":
		if re.MatchString(expr.(Node).Place) {
			// expression is of the form 1+2
			term3val, err := strconv.Atoi(expr.(Node).Place)
			if err != nil {
				return Node{}, err
			}
			n.Place = strconv.Itoa(term3val * -1)
		} else {
			n.Place = NewTmp()
			n.Code = append(n.Code, fmt.Sprintf("*, %s, %s, -1", n.Place, expr.(Node).Place))
		}
	case "!":
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("not, %s, %s", n.Place, expr.(Node).Place))
	case "+":
		n.Place = expr.(Node).Place
	case "&":
		// Place of any pointer variable starts with "pointer:" and then followed by string returned by NewTmp()
		n.Place = "pointer:" + expr.(Node).Place
	case "*":
		n.Place = "deref:" + expr.(Node).Place
	default:
		return n, fmt.Errorf("%s operator not supported", op.(Node).Place)
	}
	return n, nil
}

// NewPrimaryExprSel returns an AST node for PrimaryExpr Selector.
func NewPrimaryExprSel(expr, selector Attrib) (Node, error) {
	// The symbol table entry of a selector is of the form -
	//	(exprPlace).(selectorPlace)
	varName := fmt.Sprintf("%s.%s", expr.(Node).Place, selector.(Node).Place)
	symTabEntry, found := SearchInScope(varName)
	if found {
		if _, ok := globalSymTab[varName]; ok {
			return Node{}, fmt.Errorf("%s not in scope", varName)
		} else {
			return Node{symTabEntry[0], []string{}}, nil
		}
	} else {
		return Node{}, fmt.Errorf("%s not in scope", varName)
	}
}

// NewPrimaryExprIndex returns an AST node for PrimaryExpr Index.
func NewPrimaryExprIndex(expr, index Attrib) (Node, error) {
	n := Node{"", []string{}}
	n.Place = NewTmp()
	// NOTE: Indexing is only supported by array types and not pointer types.
	symTab[n.Place] = []string{fmt.Sprintf("%s, %s", expr.(Node).Place, index.(Node).Place), "array"}
	n.Code = append(n.Code, index.(Node).Code...)
	n.Code = append(n.Code, fmt.Sprintf("from, %s, %s, %s", n.Place, expr.(Node).Place, index.(Node).Place))
	return n, nil
}

// NewPrimaryExprArgs returns an AST node for PrimaryExpr Arguments.
// NOTE: This is the production rule for a function call.
func NewPrimaryExprArgs(expr, args Attrib) (Node, error) {
	n := Node{"", args.(Node).Code}
	typeName := globalSymTab[expr.(Node).Place][0]
	returnLen := 0
	if strings.HasPrefix(typeName, "func") {
		returnLen, _ = strconv.Atoi(typeName[5:])
	} else {
		return Node{}, fmt.Errorf("%s is not a function", expr.(Node).Place)
	}
	argExpr := utils.SplitAndSanitize(args.(Node).Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.(Node).Place, k, v))
	}
	n.Code = append(n.Code, fmt.Sprintf("call, %s", expr.(Node).Place))
	for k := 0; k < returnLen; k++ {
		n.Place = fmt.Sprintf("%s, return.%d", n.Place, k)
	}
	return n, nil
}

func NewCompositeLit(typ, val Attrib) (Node, error) {
	// NOTE: Structs don't support initialization of members
	// currently. All data members are initialized to default
	// values when a struct instance is defined.
	n := Node{typ.(Node).Place, []string{}}
	// Check if the LiteralType corresponds to ArrayType.
	// This is done because unlike structs it is not required
	// to add a symbol table entry for place value of arrays
	// (which is of the form "array:<length_of_array>"), thus
	// returning early.
	typeName := typ.(Node).Place
	if strings.HasPrefix(typeName, "array") {
		return n, nil
	}
	// In case the corresponds to a struct, add the
	// code for its data member initialization.
	symTabEntry, found := SearchInScope(typ.(Node).Place)
	if found {
		switch symTabEntry[0] {
		case "struct":
			// The place value for struct is of the form -
			//      struct:<number of members of struct>:<name of struct>
			n.Place = fmt.Sprintf("struct:%d:%s", (len(symTabEntry)-1)/2, n.Place)
			litVals := utils.SplitAndSanitize(val.(Node).Place, ",")
			litValCodes := val.(Node).Code
			structInit := []string{}
			// In case of integral (or any type) initializations, the corresponding
			// lexeme is placed at the place value, justifying the length check which
			// is made on 'litVals' instead of 'litValCodes'. If there are no place
			// values for the data members, then initialize all to their default
			// values. Otherwise initialize them to the corresponding place value.
			if len(litVals) == 0 {
				for k, v := range symTabEntry[1:] {
					if k%2 == 0 {
						structInit = append(structInit, v)
						// TODO: Update default values depending on type
						structInit = append(structInit, "0")
					}
				}
			} else {
				for k, v := range symTabEntry[1:] {
					if k%2 == 0 {
						structInit = append(structInit, v)
						structInit = append(structInit, litVals[k/2])
					}
				}
			}
			// When these code values will be utilized above, the litValCodes will be
			// placed above the code corresponding to structInit (litVals can be expressions).
			n.Code = append(n.Code, structInit...)
			n.Code = append(n.Code, litValCodes...)
		}
	} else {
		// TODO: Update error message.
		return Node{}, fmt.Errorf("%s not in scope", typ.(Node).Place)
	}
	return n, nil
}

// NewIdentifier returns a new identifier.
func NewIdentifier(ident Attrib) (Node, error) {
	varName := string(ident.(*token.Token).Lit)
	symTabEntry, found := SearchInScope(varName)
	if found {
		if _, ok := globalSymTab[varName]; ok {
			return Node{varName, []string{}}, nil
		} else {
			return Node{symTabEntry[0], []string{}}, nil
		}
	} else {
		return Node{}, fmt.Errorf("%s not declared", varName)
	}
}

// --- [ Functions ] -----------------------------------------------------------

// NewFuncDecl returns a new function declaration.
func NewFuncDecl(marker, body Attrib) (Node, error) {
	n := Node{"", marker.(Node).Code}
	n.Code = append(n.Code, body.(Node).Code...)
	funcSymtabCreated = false // end of function block
	// Return statement insertion will be handled when defer
	// stack is emptied and the deferred calls are inserted.
	addRetStmt := false
	if deferStack.Len > 0 {
		addRetStmt = true
	}
	for deferStack.Len > 0 {
		deferFuncCode := deferStack.Pop().(DeferStackItem)
		n.Code = append(n.Code, deferFuncCode...)
	}
	if addRetStmt {
		n.Code = append(n.Code, fmt.Sprintf("ret,"))
	}
	return n, nil
}

// NewFuncMarker returns a marker non-terminal used in the production rule for
// function declaration.
func NewFuncMarker(name, signature Attrib) (Node, error) {
	n := Node{name.(Node).Place, []string{fmt.Sprintf("func, %s", name.(Node).Place)}}
	// Assign values to arguments.
	for k, v := range signature.(Node).Code {
		n.Code = append(n.Code, fmt.Sprintf("=, %s, %s.%d", v, name.(Node).Place, k))
	}
	if _, found := globalSymTab[name.(Node).Place]; !found {
		globalSymTab[name.(Node).Place] = []string{fmt.Sprintf("func:%s", signature.(Node).Place)}
	} else {
		return Node{}, fmt.Errorf("Function %s is already declared\n", name)
	}
	return n, nil
}

// NewSignature returns a function signature.
// The accepted variadic arguments (args) in their order are -
// 	- parameters
//	- result
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewSignature(typ int, args ...Attrib) (Node, error) {
	// The parent symbol table in case of function declaration will be nil
	// as functions can only be declared globally.
	childSymTab := SymInfo{make(symTabType), currSymTab}
	// Update the current symbol table to point to the newly created symbol
	// table.
	currSymTab = &childSymTab
	for _, v := range args[0].(Node).Code {
		if v == "" {
			break
		}
		currSymTab.varSymTab[v] = []string{v, "int"}
	}
	if typ == 0 {
		return Node{"0", args[0].(Node).Code}, nil
	} else {
		return Node{fmt.Sprintf("%s", args[1].(Node).Place), args[0].(Node).Code}, nil
	}
}

// NewResult defines the return type of a function.
func NewResult(params Attrib) (Node, error) {
	returnLength := 0
	// finding number of return variable
	for _, v := range params.(Node).Code {
		if v == "int" {
			returnLength++
		}
	}
	return Node{fmt.Sprintf("%d", returnLength), []string{}}, nil
}

// NewParamList returns a list of parameters.
func NewParamList(decl, declList Attrib) (Node, error) {
	n := Node{"", decl.(Node).Code}
	n.Code = append(n.Code, declList.(Node).Code...)
	return n, nil
}

// AppendParam appends a parameter to a list of parameters.
func AppendParam(decl, declList Attrib) (Node, error) {
	n := Node{"", decl.(Node).Code}
	n.Code = append(n.Code, declList.(Node).Code...)
	return n, nil
}

// NewFieldDecl returns a field declaration.
func NewFieldDecl(identList, typ Attrib) (Node, error) {
	n := Node{"", []string{}}
	for _, v := range identList.(Node).Code {
		n.Code = append(n.Code, v)
		n.Code = append(n.Code, typ.(Node).Place)
	}
	return n, nil
}

// AppendFieldDecl appends a field declaration to a list of field declarations.
func AppendFieldDecl(decl, declList Attrib) (Node, error) {
	n := Node{"", decl.(Node).Code}
	n.Code = append(n.Code, declList.(Node).Code...)
	return n, nil
}

// AppendIdent appends an identifier to a list of identifiers.
func AppendIdent(ident, identList Attrib) (Node, error) {
	// The lexemes corresponding to the individual identifiers
	// are appended to the slice for code to avoid adding
	// comma-separated string in place since the identifiers
	// don't have any IR code to be added.
	n := Node{"", []string{string(ident.(*token.Token).Lit)}}
	n.Code = append(n.Code, identList.(Node).Code...)
	return n, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewStmtList returns a statement list.
func NewStmtList(stmt, stmtList Attrib) (Node, error) {
	n := Node{"", stmt.(Node).Code}
	n.Code = append(n.Code, stmtList.(Node).Code...)
	return n, nil
}

// NewLabelStmt returns a labeled statement.
func NewLabelStmt(label, stmt Attrib) (Node, error) {
	n := Node{"", []string{fmt.Sprintf("label, %s", label.(Node).Place)}}
	n.Code = append(n.Code, stmt.(Node).Code...)
	return n, nil
}

// NewReturnStmt returns a return statement.
// A return statement can be of the following types -
//	- an empty return: In this case the argument expr is empty.
//	- non-empty return: In this case the argument expr contains the return
//	  expression.
func NewReturnStmt(expr ...Attrib) (Node, error) {
	if len(expr) == 0 {
		// The return statement is empty.
		// The defer statements need to be inserted before the return stmt (and
		// not at the end of function block as was the previous misconception).
		// When defer stmt is used, the return stmt for main() is also inserted
		// when all the defer calls from stack are popped and inserted in IR.
		if deferStack.Len > 0 {
			// Return statement insertion will be handled when defer
			// stack is emptied and the deferred calls are inserted.
			return Node{"", []string{}}, nil
		} else {
			return Node{"", []string{"ret,"}}, nil
		}
	} else {
		n := Node{"", []string{}}
		if deferStack.Len == 0 {
			retExpr := utils.SplitAndSanitize(expr[0].(Node).Place, ",")
			n.Code = append(n.Code, expr[0].(Node).Code...)
			for k, v := range retExpr {
				n.Code = append(n.Code, fmt.Sprintf("=, return.%d, %s", k, v))
			}
			n.Code = append(n.Code, fmt.Sprintf("ret,"))
		}
		return n, nil
	}
}

// --- [ Blocks ] --------------------------------------------------------------

// NewBlock returns a block.
func NewBlock(stmt Attrib) (Node, error) {
	// start of block
	currSymTab = currSymTab.parent // end of block
	return Node{"", stmt.(Node).Code}, nil
}

// NewBlockMarker returns a marker non-terminal used in the production rule for
// block declaration. This marker demarcates the beginning of a new block and
// the corresponding symbol table is instantiated here.
func NewBlockMarker() (Attrib, error) {
	if funcSymtabCreated {
		// The symbol table for functions is created when the
		// rule for Signature is reached so that the arguments
		// can also be added. At this point the function block
		// (if there was any) has completed.
		childSymTab := SymInfo{make(symTabType), currSymTab}
		// Update the current symbol table to point to the newly
		// created symbol table.
		currSymTab = &childSymTab
	} else {
		// Allow creation of symbol table for another function.
		funcSymtabCreated = true
	}
	return nil, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewIfStmt returns an if statement.
func NewIfStmt(typ int, args ...Attrib) (Node, error) {
	n := Node{"", args[0].(Node).Code}
	afterLabel := NewLabel()
	elseLabel := NewLabel()
	switch typ {
	case 0:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(Node).Place),
			args[1].(Node).Code,
		)
	case 1:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].(Node).Place),
			args[1].(Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].(Node).Code,
		)
	case 2:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].(Node).Place),
			args[1].(Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].(Node).Code,
		)
	case 3:
		n.Code = utils.AppendToSlice(
			args[1].(Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[1].(Node).Place),
			args[2].(Node).Code,
		)
	case 4:
		fallthrough
	case 5:
		n.Code = utils.AppendToSlice(
			args[1].(Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[1].(Node).Place),
			args[2].(Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[3].(Node).Code,
		)
	}
	n.Code = append(n.Code, fmt.Sprintf("label, %s", afterLabel))
	return n, nil
}

// NewSwitchStmt returns a switch statement.
func NewSwitchStmt(expr, caseClause Attrib) (Node, error) {
	n := Node{"", expr.(Node).Code}
	caseLabels := []string{}
	caseStmts := caseClause.(Node).Code
	// SplitAndSanitize cannot be used here as removal of empty
	// entries seems to be causing erroneous index calculations.
	// Regression caused in "test/codegen/switch.go".
	caseTemporaries := strings.Split(caseClause.(Node).Place, ",")
	afterLabel := NewLabel()
	defaultLabel := afterLabel
	// The last value in caseTemporaries will be the place value
	// returned by Empty (arising from the production rule
	// RepeatExprCaseClause -> Empty).
	// This has to be ignored.
	for k, v := range caseTemporaries[:len(caseTemporaries)-1] {
		caseLabel := NewLabel()
		caseLabels = append(caseLabels, caseLabel)
		n.Code = append(n.Code, caseStmts[2*k])
		if strings.TrimSpace(v) == "default" {
			defaultLabel = caseLabel
		} else {
			n.Code = append(n.Code, fmt.Sprintf("beq, %s, %s, %s", caseLabel, expr.(Node).Place, v))
		}
	}
	n.Code = append(n.Code, fmt.Sprintf("j, %s", defaultLabel))
	for k, v := range caseLabels {
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("label, %s", v),
			caseStmts[2*k+1],
			fmt.Sprintf("j, %s", afterLabel),
		)
	}
	n.Code = append(n.Code, fmt.Sprintf("label, %s", afterLabel))
	return n, nil
}

// NewExprCaseClause returns an expression case clause.
func NewExprCaseClause(expr, stmtList Attrib) (Node, error) {
	n := Node{expr.(Node).Place, []string{}}
	exprCode := ""
	for _, v := range expr.(Node).Code {
		exprCode += v
		exprCode += "\n"
	}
	n.Code = append(n.Code, exprCode)
	stmtCode := ""
	for _, v := range stmtList.(Node).Code {
		stmtCode += v
		stmtCode += "\n"
	}
	n.Code = append(n.Code, stmtCode)
	return n, nil
}

// AppendExprCaseClause appends an expression case clause to a list of same.
func AppendExprCaseClause(expr, exprList Attrib) (Node, error) {
	n := Node{"", expr.(Node).Code}
	n.Code = append(n.Code, exprList.(Node).Code...)
	n.Place = fmt.Sprintf("%s, %s", expr.(Node).Place, exprList.(Node).Place)
	return n, nil
}

// NewForStmt returns a for statement.
// The accepted variants of the variadic arguments (args) are -
//	- Block
//	- Condition, Block
//	- ForClause, Block
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewForStmt(typ int, args ...Attrib) (Node, error) {
	n := Node{"", []string{}}
	startLabel := NewLabel()
	afterLabel := NewLabel()
	switch typ {
	case 0:
		n.Code = append(n.Code, fmt.Sprintf("label, %s", startLabel))
		for _, v := range args[0].(Node).Code {
			v := strings.TrimSpace(v)
			switch v {
			case "break":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", afterLabel))
			case "continue":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", startLabel))
			default:
				n.Code = append(n.Code, v)
			}
		}
	case 1:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("label, %s", startLabel),
			args[0].(Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(Node).Place),
		)
		for _, v := range args[1].(Node).Code {
			v := strings.TrimSpace(v)
			switch v {
			case "break":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", afterLabel))
			case "continue":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", startLabel))
			default:
				n.Code = append(n.Code, v)
			}
		}
	case 2:
		n.Code = utils.AppendToSlice(
			n.Code,
			args[0].(Node).Code[0], // init stmt
			fmt.Sprintf("label, %s", startLabel),
			args[0].(Node).Code[1], // condition
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(Node).Place),
		)
		for _, v := range args[1].(Node).Code {
			v := strings.TrimSpace(v)
			switch v {
			case "break":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", afterLabel))
			case "continue":
				n.Code = append(n.Code, fmt.Sprintf("j, %s", startLabel))
			default:
				n.Code = append(n.Code, v)
			}
		}
		n.Code = append(n.Code, args[0].(Node).Code[2]) // post stmt
	}
	n.Code = utils.AppendToSlice(
		n.Code,
		fmt.Sprintf("j, %s", startLabel),
		fmt.Sprintf("label, %s", afterLabel),
	)
	return n, nil
}

// NewForClause returns a for clause.
func NewForClause(typ int, args ...Attrib) (Node, error) {
	var initStmtCode, condCode, postStmtCode string
	switch typ {
	case 0:
		for _, v := range args[0].(Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		return Node{"1", []string{initStmtCode, "", ""}}, nil
	case 1:
		for _, v := range args[0].(Node).Code {
			condCode += fmt.Sprintf("%s\n", v)
		}
		return Node{args[0].(Node).Place, []string{"", condCode, ""}}, nil
	case 2:
		for _, v := range args[0].(Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return Node{"1", []string{"", "", postStmtCode}}, nil
	case 3:
		for _, v := range args[0].(Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(Node).Code {
			condCode += fmt.Sprintf("%s\n", v)
		}
		return Node{args[1].(Node).Place, []string{initStmtCode, condCode, ""}}, nil
	case 4:
		for _, v := range args[0].(Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return Node{"1", []string{initStmtCode, "", postStmtCode}}, nil
	case 5:
		for _, v := range args[0].(Node).Code {
			condCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return Node{args[0].(Node).Place, []string{"", condCode, postStmtCode}}, nil
	case 6:
		for _, v := range args[0].(Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(Node).Code {
			condCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[2].(Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return Node{args[1].(Node).Place, []string{initStmtCode, condCode, postStmtCode}}, nil
	}
	return Node{}, fmt.Errorf("NewForClause: Invalid type %d", typ)
}

// NewDeferStmt returns a defer statement.
func NewDeferStmt(expr, args Attrib) (Node, error) {
	// Add code corresponding to the arguments.
	n := Node{"", args.(Node).Code}
	argExpr := utils.SplitAndSanitize(args.(Node).Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.(Node).Place, k, v))
	}
	n.Place = NewTmp()
	// Push the code for the actual function call to the defer stack.
	deferCode := make(DeferStackItem, 0)
	deferCode = append(deferCode, fmt.Sprintf("call, %s", expr.(Node).Place))
	deferCode = append(deferCode, fmt.Sprintf("store, %s", n.Place))
	deferStack.Push(deferCode)
	return n, nil
}

// NewIOStmt returns an I/O statement.
func NewIOStmt(typ, expr Attrib) (Node, error) {
	n := Node{"", expr.(Node).Code}
	typ = string(typ.(*token.Token).Lit)
	switch typ {
	case "printInt":
		// The IR for printInt is supposed to look as following -
		//	printInt, a, a
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s", typ, expr.(Node).Place, expr.(Node).Place))
	case "printStr":
		fallthrough
	case "scanInt":
		n.Code = append(n.Code, fmt.Sprintf("%s, %s", typ, expr.(Node).Place))
	}
	return n, nil
}

// NewIncDecStmt returns an increment or a decrement statement.
func NewIncDecStmt(op, expr Attrib) (Node, error) {
	op = string(op.(*token.Token).Lit)
	n := Node{"", expr.(Node).Code}
	switch op {
	case "++":
		n.Code = append(n.Code, fmt.Sprintf("+, %s, %s, 1", expr.(Node).Place, expr.(Node).Place))
	case "--":
		n.Code = append(n.Code, fmt.Sprintf("-, %s, %s, 1", expr.(Node).Place, expr.(Node).Place))
	default:
		return Node{}, fmt.Errorf("Invalid operator %s", op)
	}
	return n, nil
}

// NewAssignStmt returns an assignment statement.
func NewAssignStmt(typ int, op, leftExpr, rightExpr Attrib) (Node, error) {
	switch typ {
	case 0:
		op := string(op.(*token.Token).Lit)[0]
		n := Node{"", leftExpr.(Node).Code}
		n.Code = append(n.Code, rightExpr.(Node).Code...)
		leftExpr := utils.SplitAndSanitize(leftExpr.(Node).Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.(Node).Place, ",")
		for k, v := range leftExpr {
			n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, v, v, rightExpr[k]))
			_, ok := symTab[v]
			if ok {
				if symTab[v][1] == "array" {
					arrayInfo := utils.SplitAndSanitize(symTab[v][0], ",")
					// arrayInfo contains the following info -
					//      0th index: array name
					//      1st index: array index
					n.Code = append(n.Code, fmt.Sprintf("into, %s, %s, %s, %s", arrayInfo[0], arrayInfo[0], arrayInfo[1], v))
				}
			}
		}
		return n, nil
	case 1:
		n := Node{"", leftExpr.(Node).Code}
		n.Code = append(n.Code, rightExpr.(Node).Code...)
		leftExpr := utils.SplitAndSanitize(leftExpr.(Node).Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.(Node).Place, ",")
		if len(leftExpr) != len(rightExpr) {
			return Node{}, errors.New("No. of entities in LHS ≠ RHS")
		}
		for k, v := range leftExpr {
			if len(currSymTab.varSymTab[GetRealName(v)]) >= 2 && currSymTab.varSymTab[GetRealName(v)][1] == "pointer" {
				if strings.HasPrefix(rightExpr[k], "pointer") {
					currSymTab.varSymTab[GetRealName(v)][3] = currSymTab.varSymTab[GetRealName(rightExpr[k][8:])][0]
				} else {
					currSymTab.varSymTab[GetRealName(v)][3] = currSymTab.varSymTab[GetRealName(rightExpr[k])][3]
				}
			} else if strings.HasPrefix(rightExpr[k], "deref") && strings.HasPrefix(v, "deref") {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", currSymTab.varSymTab[GetRealName(v[6:])][3], currSymTab.varSymTab[GetRealName(rightExpr[k][6:])][3]))
			} else if strings.HasPrefix(rightExpr[k], "deref") {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", currSymTab.varSymTab[GetRealName(v)][0], currSymTab.varSymTab[GetRealName(rightExpr[k][6:])][3]))
			} else if strings.HasPrefix(v, "deref") {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", currSymTab.varSymTab[GetRealName(v[6:])][3], rightExpr[k]))
			} else {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", v, rightExpr[k]))
				_, ok := symTab[v]
				if ok {
					if symTab[v][1] == "array" {
						arrayInfo := utils.SplitAndSanitize(symTab[v][0], ",")
						// arrayInfo[0]: array name
						// arrayInfo[1]: array index
						n.Code = append(n.Code, fmt.Sprintf("into, %s, %s, %s, %s", arrayInfo[0], arrayInfo[0], arrayInfo[1], v))
					}
				}
			}
		}
		return n, nil
	case 2:
		n := Node{"", []string{}}
		// TODO: Structs do not support multiple short declarations in a
		// single statement for now.
		exprName := rightExpr.(Node).Place
		if strings.HasPrefix(exprName, "struct") {
			return Node{}, errors.New("Use short declaration for declaring structs")
		} else {
			n.Code = rightExpr.(Node).Code
			expr := utils.SplitAndSanitize(rightExpr.(Node).Place, ",")
			if len(expr) != len(leftExpr.(Node).Code) {
				return Node{}, errors.New("No. of entities in LHS ≠ RHS")
			}
			for k, v := range leftExpr.(Node).Code {
				symTabEntry, found := SearchInScope(v)
				if found {
					renamedVar := symTabEntry[0]
					if strings.HasPrefix(expr[k], "array") {
						return Node{}, errors.New("Use short declaration for declaring arrays")
					} else if len(currSymTab.varSymTab[v]) >= 2 && currSymTab.varSymTab[v][1] != "pointer" {
						n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
					}
				} else {
					return Node{}, fmt.Errorf("%s not declared", v)
				}
			}
		}
		return n, nil
	}
	return Node{}, nil
}

// NewShortDecl returns a short variable declaration.
func NewShortDecl(identList, exprList Attrib) (Node, error) {
	n := Node{"", []string{}}
	// TODO: Structs do not support multiple short declarations in a
	// single statement for now.
	exprName := exprList.(Node).Place
	if strings.HasPrefix(exprName, "struct") {
		// NOTE: The following index calculations assume that
		// struct names cannot include a ':' character.
		colonIndex := strings.LastIndexAny(exprName, ":")
		structLen, err := strconv.Atoi(exprName[7:colonIndex])
		if err != nil {
			return Node{}, err
		}
		// TODO: Multiple struct initializations are not handled currently.
		structName := identList.(Node).Code[0]
		// keeping structName in the symbol table with type as Struct
		currSymTab.varSymTab[structName] = []string{structName, "struct"}
		// The individual struct member initializers can contain
		// expressions whose code need to be added before the
		// members are initialized.
		n.Code = append(n.Code, exprList.(Node).Code[2*structLen:]...)
		// Add code for struct member initializations.
		var varName, varVal string
		for k, v := range exprList.(Node).Code[:2*structLen] {
			if k%2 == 0 {
				// Member names are located at even locations.
				varName = v
			} else {
				// (Initialized) member values are located at odd locations.
				varVal = v
				renamedVar := RenameVariable(fmt.Sprintf("%s.%s", structName, varName))
				currSymTab.varSymTab[fmt.Sprintf("%s.%s", structName, varName)] = []string{renamedVar, "int"}
				// TODO: Add the struct initializations to symbol table. Also,
				// handle member accesses as -
				//      node := Node{1}
				//      b := node.val  // member access
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, varVal))
			}
		}
	} else {
		n.Code = exprList.(Node).Code
		// TODO: Check this -- placeVals and expr are same??
		placeVals := strings.Split(exprList.(Node).Place, ",")
		expr := utils.SplitAndSanitize(exprList.(Node).Place, ",")
		if len(expr) != len(identList.(Node).Code) {
			return Node{}, errors.New("No. of entities in LHS ≠ RHS")
		}
		for k, v := range identList.(Node).Code {
			renamedVar := RenameVariable(v)
			_, ok := currSymTab.varSymTab[v]
			if !ok {
				// TODO: All types are int currently.
				if strings.HasPrefix(expr[k], "pointer") {
					currSymTab.varSymTab[v] = []string{renamedVar, "pointer", "int", expr[k][8:]}
				} else if len(currSymTab.varSymTab[GetRealName(expr[k])]) >= 2 && currSymTab.varSymTab[GetRealName(expr[k])][1] == "pointer" {
					currSymTab.varSymTab[v] = []string{renamedVar, "pointer", currSymTab.varSymTab[GetRealName(expr[k])][2], currSymTab.varSymTab[GetRealName(expr[k])][3]}
				} else if strings.HasPrefix(expr[k], "deref") {
					currSymTab.varSymTab[v] = []string{renamedVar, "int"}
				} else {
					currSymTab.varSymTab[v] = []string{renamedVar, "int"}
				}
			} else {
				return Node{}, fmt.Errorf("%s is already declared", v)
			}
			if strings.HasPrefix(expr[k], "array") {
				// TODO: rename arrays
				n.Code = append(n.Code, fmt.Sprintf("decl, %s, %s", renamedVar, expr[k][6:]))
			} else if strings.HasPrefix(expr[k], "deref") {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, currSymTab.varSymTab[GetRealName(expr[k][6:])][3]))
			} else if strings.HasPrefix(placeVals[k], "string") {
				// Check if the RHS is string.
				n.Code = append(n.Code, fmt.Sprintf("declStr, %s, %s", renamedVar, expr[k][7:]))
			} else if len(currSymTab.varSymTab[v]) >= 2 && currSymTab.varSymTab[v][1] != "pointer" {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
			}
		}
	}
	return n, nil
}
