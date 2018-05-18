// Package ast implements utility functions for generating abstract syntax
// trees from Go BNF.

package ast

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shivansh/gogo/src/utils"
	"github.com/shivansh/gogo/tmp/token"
)

type (
	// Attrib represents attributes for symbols in the grammar.
	Attrib interface{}
	// DeferStackItem is an individual item stored when a call to defer is
	// made. It contains the code for the corresponding function call which
	// is placed at the end of the function body.
	DeferStackItem []string
)

var (
	// deferStack stores the deferred function calls which are then called
	// when the surrounding function block ends.
	deferStack *utils.Stack
	re         *regexp.Regexp
)

func init() {
	globalSymTab = make(symTabType)
	// currScope now allocates space for a global symbol table for variable
	// declarations. Ideally, instead of creating a new symbol table it
	// should've been global symbol table, but since the type of
	// currScope.symTab is not pointer, updates made elsewhere will not
	// be reflected globally.
	// TODO: Update type of currScope.symTab to a pointer.
	currScope = &SymInfo{make(symTabType), nil}
	deferStack = utils.CreateStack()
	re = regexp.MustCompile("(^-?[0-9]+$)") // integers
}

// Node represents a node in the AST of a given program.
type Node struct {
	// If the node represents an expression, then place stores the name of
	// the variable storing the value of the expression.
	Place string
	Code  []string // IR instructions
}

// PrintIR generates the IR instructions accumulated in the "Code" attribute of
// the "SourceFile" non-terminal.
func PrintIR(src Attrib) (Attrib, error) {
	c := src.(*Node).Code
	for _, v := range c {
		v := strings.TrimSpace(v)
		if v != "" {
			fmt.Println(v)
		}
	}
	return nil, nil
}

// InitNode initializes an AST node with the given "Place" and "Code" attributes.
func InitNode(place string, code []string) (*Node, error) {
	return &Node{place, code}, nil
}

// NewNode creates a new AST node from the attributes of the given non-terminal.
func NewNode(attr Attrib) (*Node, error) {
	return &Node{attr.(*Node).Place, attr.(*Node).Code}, nil
}

// --- [ Top level declarations ] ----------------------------------------------

// NewTopLevelDecl returns a top level declaration.
func NewTopLevelDecl(topDecl, repeatTopDecl Attrib) (*Node, error) {
	n := &Node{"", topDecl.(*Node).Code}
	n.Code = append(n.Code, repeatTopDecl.(*Node).Code...)
	return n, nil
}

// NewTypeDef returns a type definition.
func NewTypeDef(ident, typ Attrib) (*Node, error) {
	return &Node{fmt.Sprintf("%s, %s", typ.(*Node).Place, string(ident.(*token.Token).Lit)), typ.(*Node).Code}, nil
}

// NewElementList returns a keyed element list.
func NewElementList(key, keyList Attrib) (*Node, error) {
	n := &Node{fmt.Sprintf("%s, %s", key.(*Node).Place, keyList.(*Node).Place), key.(*Node).Code}
	n.Code = append(n.Code, keyList.(*Node).Code...)
	return n, nil
}

// AppendKeyedElement appends a keyed element to a list of keyed elements.
func AppendKeyedElement(key, keyList Attrib) (*Node, error) {
	n := &Node{fmt.Sprintf("%s, %s", key.(*Node).Place, keyList.(*Node).Place), key.(*Node).Code}
	n.Code = append(n.Code, keyList.(*Node).Code...)
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
func NewVarSpec(typ int, args ...Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	expr := []string{}
	// Add the IR instructions for ExpressionList.
	switch typ {
	case 1:
		n.Code = args[2].(*Node).Code
		expr = utils.SplitAndSanitize(args[2].(*Node).Place, ",")
	case 2:
		n.Code = args[1].(*Node).Code
		expr = utils.SplitAndSanitize(args[1].(*Node).Place, ",")
	case 3:
		return &Node{}, nil
	}
	for k, v := range args[0].(*Node).Code {
		renamedVar := RenameVariable(v)
		// TODO: Handle other types
		InsertSymbol(v, INTEGER, renamedVar)
		if typ == 0 {
			n.Code = append(n.Code, fmt.Sprintf("=, %s, 0", renamedVar))
		} else if typ == 1 || typ == 2 {
			n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
		}
	}
	return n, nil
}

// --- [ Type declarations ] ---------------------------------------------------

// NewTypeDecl returns a type declaration.
func NewTypeDecl(args ...Attrib) (*Node, error) {
	typeInfo := utils.SplitAndSanitize(args[1].(*Node).Place, ",")
	structName := strings.TrimSpace(typeInfo[1])
	typ := strings.TrimSpace(typeInfo[0])
	switch typ {
	case "struct":
		// Create a global symbol table entry.
		// NOTE: The symbol table entry of a struct is of the form -
		//      structName : []{"struct", memberName1, memberType1, ...}
		globalSymTab[structName] = symTabEntry{
			kind:    STRUCT,
			symbols: args[1].(*Node).Code,
		}
	default: // TODO: Add remaining types.
		return &Node{}, fmt.Errorf("Unknown type %s", typ)
	}
	// TODO: Member initialization will be done when a new object is
	// instantiated.
	return &Node{"", []string{}}, nil
}

// --- [ Constant declarations ] -----------------------------------------------

// NewConstSpec returns a constant declaration.
func NewConstSpec(typ int, args ...Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	expr := []string{}
	if typ == 1 {
		n.Code = append(n.Code, args[1].(*Node).Code...)
		expr = utils.SplitAndSanitize(args[1].(*Node).Place, ",")
	}
	for k, v := range args[0].(*Node).Code {
		renamedVar := RenameVariable(v)
		InsertSymbol(v, INTEGER, renamedVar)
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
func NewExpr(expr Attrib) (*Node, error) {
	return NewNode(expr)
}

// AppendExpr appends a list of expressions to a given expression.
func AppendExpr(expr, exprlist Attrib) (*Node, error) {
	n := &Node{"", expr.(*Node).Code}
	n.Code = append(n.Code, exprlist.(*Node).Code...)
	n.Place = fmt.Sprintf("%s,%s", expr.(*Node).Place, exprlist.(*Node).Place)
	return n, nil
}

// NewBoolExpr returns a new logical expression.
func NewBoolExpr(op, leftexpr, rightexpr Attrib) (*Node, error) {
	n := &Node{"", leftexpr.(*Node).Code}
	n.Code = append(n.Code, rightexpr.(*Node).Code...)
	n.Place = NewTmp()
	afterLabel := NewLabel()
	switch string(op.(*token.Token).Lit) {
	case "||":
		trueLabel := NewLabel()
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, leftexpr.(*Node).Place),
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, rightexpr.(*Node).Place),
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
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, leftexpr.(*Node).Place),
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, rightexpr.(*Node).Place),
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
func NewRelExpr(op, leftexpr, rightexpr Attrib) (*Node, error) {
	n := &Node{"", leftexpr.(*Node).Code}
	n.Place = NewTmp()
	n.Code = append(n.Code, rightexpr.(*Node).Code...)
	branchOp := ""
	falseLabel := NewLabel()
	afterLabel := NewLabel()
	switch op.(*Node).Place {
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
		fmt.Sprintf("%s, %s, %s, %s", branchOp, falseLabel, leftexpr.(*Node).Place, rightexpr.(*Node).Place),
		fmt.Sprintf("=, %s, 1", n.Place),
		fmt.Sprintf("j, %s", afterLabel),
		fmt.Sprintf("label, %s", falseLabel),
		fmt.Sprintf("=, %s, 0", n.Place),
		fmt.Sprintf("label, %s", afterLabel),
	)
	return n, nil
}

// NewArithExpr returns an arithmetic expression.
func NewArithExpr(op, leftexpr, rightexpr Attrib) (*Node, error) {
	n := &Node{"", leftexpr.(*Node).Code}
	op = string(op.(*token.Token).Lit)
	n.Code = append(n.Code, rightexpr.(*Node).Code...)
	if re.MatchString(leftexpr.(*Node).Place) && re.MatchString(rightexpr.(*Node).Place) {
		// Expression is of the form "1 op 2". Such expression can
		// be reduced by evaluating its value during compilation itself.
		leftval, err := strconv.Atoi(leftexpr.(*Node).Place)
		if err != nil {
			return &Node{}, err
		}
		rightval, err := strconv.Atoi(rightexpr.(*Node).Place)
		if err != nil {
			return &Node{}, err
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
			return &Node{}, fmt.Errorf("Invalid operation %s", op)
		}
	} else if re.MatchString(leftexpr.(*Node).Place) {
		// Expression is of the form "1 + b", which needs to be
		// converted to the equivalent form "b + 1" to be counted as a
		// valid IR statement.
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, rightexpr.(*Node).Place, leftexpr.(*Node).Place))
	} else {
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, leftexpr.(*Node).Place, rightexpr.(*Node).Place))
	}
	return n, nil
}

// NewUnaryExpr returns a unary expression.
func NewUnaryExpr(op, expr Attrib) (*Node, error) {
	n := &Node{"", expr.(*Node).Code}
	switch op.(*Node).Place {
	case "-":
		if re.MatchString(expr.(*Node).Place) {
			// expression is of the form 1+2
			term3val, err := strconv.Atoi(expr.(*Node).Place)
			if err != nil {
				return &Node{}, err
			}
			n.Place = strconv.Itoa(term3val * -1)
		} else {
			n.Place = NewTmp()
			n.Code = append(n.Code, fmt.Sprintf("*, %s, %s, -1", n.Place, expr.(*Node).Place))
		}
	case "!":
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("not, %s, %s", n.Place, expr.(*Node).Place))
	case "+":
		n.Place = expr.(*Node).Place
	case "&":
		// Place attribute of a pointer variable starts with "pointer:"
		// followed by its place attribute.
		n.Place = PTR + ":" + expr.(*Node).Place
	case "*":
		n.Place = DRF + ":" + expr.(*Node).Place
	default:
		return n, fmt.Errorf("%s operator not supported", op.(*Node).Place)
	}
	return n, nil
}

// NewPrimaryExprSel returns an AST node for PrimaryExpr Selector.
func NewPrimaryExprSel(expr, selector Attrib) (*Node, error) {
	// The symbol table entry of a selector is of the form -
	//	(exprPlace).(selectorPlace)
	varName := fmt.Sprintf("%s.%s", expr.(*Node).Place, selector.(*Node).Place)
	if symEntry, found := Lookup(varName); found {
		if _, ok := globalSymTab[varName]; ok {
			return &Node{}, fmt.Errorf("%s not in scope", varName)
		} else {
			return &Node{symEntry.symbols[0], []string{}}, nil
		}
	} else {
		return &Node{}, fmt.Errorf("%s not in scope", varName)
	}
}

// NewPrimaryExprIndex returns an AST node for PrimaryExpr Index.
func NewPrimaryExprIndex(expr, index Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	n.Place = NewTmp()
	// NOTE: Indexing is only supported by array types and not pointer types.
	InsertSymbol(n.Place, ARRAY, expr.(*Node).Place, index.(*Node).Place)
	n.Code = append(n.Code, index.(*Node).Code...)
	n.Code = append(n.Code, fmt.Sprintf("from, %s, %s, %s", n.Place, expr.(*Node).Place, index.(*Node).Place))
	return n, nil
}

// NewPrimaryExprArgs returns an AST node for PrimaryExpr Arguments.
// NOTE: This is the production rule for a function call.
func NewPrimaryExprArgs(expr, args Attrib) (*Node, error) {
	n := &Node{"", args.(*Node).Code}
	symEntry := globalSymTab[expr.(*Node).Place]
	returnLen := 0
	if symEntry.kind == FUNCTION {
		ret, err := strconv.Atoi(symEntry.symbols[0])
		if err != nil {
			return &Node{}, err
		}
		returnLen = ret
	} else {
		return &Node{}, fmt.Errorf("%s is not a function", expr.(*Node).Place)
	}
	argExpr := utils.SplitAndSanitize(args.(*Node).Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.(*Node).Place, k, v))
	}
	n.Code = append(n.Code, fmt.Sprintf("call, %s", expr.(*Node).Place))
	for k := 0; k < returnLen; k++ {
		n.Place = fmt.Sprintf("%s, return.%d", n.Place, k)
	}
	return n, nil
}

func NewCompositeLit(typ, val Attrib) (*Node, error) {
	n := &Node{typ.(*Node).Place, []string{}}
	// Check if the LiteralType corresponds to ArrayType. This is done
	// because unlike structs it is not required to add a symbol table entry
	// for place values of arrays (which is of the form "array:<length>"),
	// thus returning early.
	typeName := typ.(*Node).Place
	if strings.HasPrefix(typeName, ARR) {
		return n, nil
	}
	// In case the corresponds to a struct, add the code for its data member
	// initialization.
	if symEntry, found := Lookup(typ.(*Node).Place); found {
		switch symEntry.kind {
		case STRUCT:
			// The place value for struct is of the form -
			//      struct:<number of struct members>:<struct name>
			n.Place = fmt.Sprintf("struct:%d:%s", len(symEntry.symbols)/2, n.Place)
			litVals := utils.SplitAndSanitize(val.(*Node).Place, ",")
			litValCodes := val.(*Node).Code
			structInit := []string{}
			// In case of integral (or any type) initializations, the corresponding
			// lexeme is placed at the place value, justifying the length check which
			// is made on 'litVals' instead of 'litValCodes'. If there are no place
			// values for the data members, then initialize all to their default
			// values. Otherwise initialize them to the corresponding place value.
			if len(litVals) == 0 {
				for k, v := range symEntry.symbols {
					if k%2 == 0 {
						structInit = append(structInit, v)
						// TODO: Update default values depending on type
						structInit = append(structInit, "0")
					}
				}
			} else {
				for k, v := range symEntry.symbols {
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
		return &Node{}, fmt.Errorf("%s not in scope", typ.(*Node).Place)
	}
	return n, nil
}

// NewIdentifier returns a new identifier.
func NewIdentifier(ident Attrib) (*Node, error) {
	varName := string(ident.(*token.Token).Lit)
	if symEntry, found := Lookup(varName); found {
		if _, ok := globalSymTab[varName]; ok {
			return &Node{varName, []string{}}, nil
		} else {
			return &Node{symEntry.symbols[0], []string{}}, nil
		}
	} else {
		return &Node{}, fmt.Errorf("undefined: %s", varName)
	}
}

// --- [ Functions ] -----------------------------------------------------------

// NewFuncDecl returns a function declaration.
func NewFuncDecl(marker, body Attrib) (*Node, error) {
	n := &Node{"", marker.(*Node).Code}
	n.Code = append(n.Code, body.(*Node).Code...)
	funcSymtabCreated = false // end of function block
	// Return statement insertion will be handled when defer stack is
	// emptied and the code for deferred calls has been inserted.
	if deferStack.Len > 0 {
		defer func() { n.Code = append(n.Code, fmt.Sprintf("ret,")) }()
	}
	for deferStack.Len > 0 {
		deferFuncCode := deferStack.Pop().(DeferStackItem)
		n.Code = append(n.Code, deferFuncCode...)
	}
	return n, nil
}

// NewFuncMarker returns a marker non-terminal used in the production rule for
// function declaration.
func NewFuncMarker(name, signature Attrib) (*Node, error) {
	n := &Node{name.(*Node).Place, []string{fmt.Sprintf("func, %s", name.(*Node).Place)}}
	// Assign values to arguments.
	for k, v := range signature.(*Node).Code {
		n.Code = append(n.Code, fmt.Sprintf("=, %s, %s.%d", v, name.(*Node).Place, k))
	}
	if _, found := globalSymTab[name.(*Node).Place]; !found {
		globalSymTab[name.(*Node).Place] = symTabEntry{
			kind:    FUNCTION,
			symbols: []string{signature.(*Node).Place},
		}
		// globalSymTab[name.(*Node).Place] = []string{fmt.Sprintf("%s:%s", FNC, signature.(*Node).Place)}
	} else {
		return &Node{}, fmt.Errorf("Function %s is already declared\n", name)
	}
	return n, nil
}

// NewSignature returns a function signature.
// The accepted variadic arguments (args) in their order are -
// 	- parameters
//	- result
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewSignature(typ int, args ...Attrib) (*Node, error) {
	NewScope()
	for _, v := range args[0].(*Node).Code {
		if v == "" {
			break
		}
		InsertSymbol(v, INTEGER, v)
	}
	if typ == 0 {
		return &Node{"0", args[0].(*Node).Code}, nil
	} else {
		return &Node{fmt.Sprintf("%s", args[1].(*Node).Place), args[0].(*Node).Code}, nil
	}
}

// NewResult defines the return type of a function.
func NewResult(params Attrib) (*Node, error) {
	returnLength := 0
	// finding number of return variable
	for _, v := range params.(*Node).Code {
		if v == INT {
			returnLength++
		}
	}
	return &Node{fmt.Sprintf("%d", returnLength), []string{}}, nil
}

// NewParamList returns a list of parameters.
func NewParamList(decl, declList Attrib) (*Node, error) {
	n := &Node{"", decl.(*Node).Code}
	n.Code = append(n.Code, declList.(*Node).Code...)
	return n, nil
}

// AppendParam appends a parameter to a list of parameters.
func AppendParam(decl, declList Attrib) (*Node, error) {
	n := &Node{"", decl.(*Node).Code}
	n.Code = append(n.Code, declList.(*Node).Code...)
	return n, nil
}

// NewFieldDecl returns a field declaration.
func NewFieldDecl(identList, typ Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	for _, v := range identList.(*Node).Code {
		n.Code = append(n.Code, v)
		n.Code = append(n.Code, typ.(*Node).Place)
	}
	return n, nil
}

// AppendFieldDecl appends a field declaration to a list of field declarations.
func AppendFieldDecl(decl, declList Attrib) (*Node, error) {
	n := &Node{"", decl.(*Node).Code}
	n.Code = append(n.Code, declList.(*Node).Code...)
	return n, nil
}

// AppendIdent appends an identifier to a list of identifiers.
func AppendIdent(ident, identList Attrib) (*Node, error) {
	// The lexemes corresponding to the individual identifiers are appended
	// to the slice for code to avoid adding comma-separated string in place
	// since the identifiers don't have any IR code to be added.
	n := &Node{"", []string{string(ident.(*token.Token).Lit)}}
	n.Code = append(n.Code, identList.(*Node).Code...)
	return n, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewStmtList returns a statement list.
func NewStmtList(stmt, stmtList Attrib) (*Node, error) {
	n := &Node{"", stmt.(*Node).Code}
	n.Code = append(n.Code, stmtList.(*Node).Code...)
	return n, nil
}

// NewLabelStmt returns a labeled statement.
func NewLabelStmt(label, stmt Attrib) (*Node, error) {
	n := &Node{"", []string{fmt.Sprintf("label, %s", label.(*Node).Place)}}
	n.Code = append(n.Code, stmt.(*Node).Code...)
	return n, nil
}

// NewReturnStmt returns a return statement.
// A return statement can be of the following types -
//	- an empty return: In this case the argument expr is empty.
//	- non-empty return: In this case the argument expr contains the return
//	  expression.
func NewReturnStmt(expr ...Attrib) (*Node, error) {
	if len(expr) == 0 {
		// The return statement is empty.
		// The defer statements need to be inserted before the return stmt (and
		// not at the end of function block as was the previous misconception).
		// When defer stmt is used, the return stmt for main() is also inserted
		// when all the defer calls from stack are popped and inserted in IR.
		if deferStack.Len > 0 {
			// Return statement insertion will be handled when defer
			// stack is emptied and the deferred calls are inserted.
			return &Node{"", []string{}}, nil
		} else {
			return &Node{"", []string{"ret,"}}, nil
		}
	} else {
		n := &Node{"", []string{}}
		if deferStack.Len == 0 {
			retExpr := utils.SplitAndSanitize(expr[0].(*Node).Place, ",")
			n.Code = append(n.Code, expr[0].(*Node).Code...)
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
func NewBlock(stmt Attrib) (*Node, error) {
	currScope = currScope.parent // end of the previous block
	return &Node{"", stmt.(*Node).Code}, nil
}

// NewBlockMarker returns a marker non-terminal used in the production rule for
// block declaration. This marker demarcates the beginning of a new block and
// the corresponding symbol table is instantiated here.
func NewBlockMarker() (Attrib, error) {
	if funcSymtabCreated {
		// The symbol table for functions is created when the rule for
		// Signature is reached so that the arguments can also be added
		// At this point the function block/scope (if there was any) has
		// ended.
		NewScope()
	} else {
		// Allow creation of symbol table for another function.
		funcSymtabCreated = true
	}
	return nil, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewIfStmt returns an if statement.
func NewIfStmt(typ int, args ...Attrib) (*Node, error) {
	n := &Node{"", args[0].(*Node).Code}
	afterLabel := NewLabel()
	elseLabel := NewLabel()
	switch typ {
	case 0:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(*Node).Place),
			args[1].(*Node).Code,
		)
	case 1:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].(*Node).Place),
			args[1].(*Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].(*Node).Code,
		)
	case 2:
		n.Code = utils.AppendToSlice(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].(*Node).Place),
			args[1].(*Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].(*Node).Code,
		)
	case 3:
		n.Code = utils.AppendToSlice(
			args[1].(*Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[1].(*Node).Place),
			args[2].(*Node).Code,
		)
	case 4:
		fallthrough
	case 5:
		n.Code = utils.AppendToSlice(
			args[1].(*Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[1].(*Node).Place),
			args[2].(*Node).Code,
			fmt.Sprintf("j, %s", afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[3].(*Node).Code,
		)
	}
	n.Code = append(n.Code, fmt.Sprintf("label, %s", afterLabel))
	return n, nil
}

// NewSwitchStmt returns a switch statement.
func NewSwitchStmt(expr, caseClause Attrib) (*Node, error) {
	n := &Node{"", expr.(*Node).Code}
	caseLabels := []string{}
	caseStmts := caseClause.(*Node).Code
	// SplitAndSanitize cannot be used here as removal of empty
	// entries seems to be causing erroneous index calculations.
	// Regression caused in "test/codegen/switch.go".
	caseTemporaries := strings.Split(caseClause.(*Node).Place, ",")
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
			n.Code = append(n.Code, fmt.Sprintf("beq, %s, %s, %s", caseLabel, expr.(*Node).Place, v))
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
func NewExprCaseClause(expr, stmtList Attrib) (*Node, error) {
	n := &Node{expr.(*Node).Place, []string{}}
	var exprCode, stmtCode string
	for _, v := range expr.(*Node).Code {
		exprCode += fmt.Sprintf("%s\n", v)
	}
	n.Code = append(n.Code, exprCode)
	for _, v := range stmtList.(*Node).Code {
		stmtCode += fmt.Sprintf("%s\n", v)
	}
	n.Code = append(n.Code, stmtCode)
	return n, nil
}

// AppendExprCaseClause appends an expression case clause to a list of same.
func AppendExprCaseClause(expr, exprList Attrib) (*Node, error) {
	n := &Node{"", expr.(*Node).Code}
	n.Code = append(n.Code, exprList.(*Node).Code...)
	n.Place = fmt.Sprintf("%s, %s", expr.(*Node).Place, exprList.(*Node).Place)
	return n, nil
}

// NewForStmt returns a for statement.
// The accepted variants of the variadic arguments (args) are -
//	- Block
//	- Condition, Block
//	- ForClause, Block
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewForStmt(typ int, args ...Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	startLabel := NewLabel()
	afterLabel := NewLabel()
	switch typ {
	case 0:
		n.Code = append(n.Code, fmt.Sprintf("label, %s", startLabel))
		for _, v := range args[0].(*Node).Code {
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
			args[0].(*Node).Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(*Node).Place),
		)
		for _, v := range args[1].(*Node).Code {
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
			args[0].(*Node).Code[0], // init stmt
			fmt.Sprintf("label, %s", startLabel),
			args[0].(*Node).Code[1], // condition
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].(*Node).Place),
		)
		for _, v := range args[1].(*Node).Code {
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
		n.Code = append(n.Code, args[0].(*Node).Code[2]) // post stmt
	}
	n.Code = utils.AppendToSlice(
		n.Code,
		fmt.Sprintf("j, %s", startLabel),
		fmt.Sprintf("label, %s", afterLabel),
	)
	return n, nil
}

// NewForClause returns a for clause.
func NewForClause(typ int, args ...Attrib) (*Node, error) {
	var initStmtCode, condStmtCode, postStmtCode string
	switch typ {
	case 0:
		for _, v := range args[0].(*Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{initStmtCode, "", ""}}, nil
	case 1:
		for _, v := range args[0].(*Node).Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[0].(*Node).Place, []string{"", condStmtCode, ""}}, nil
	case 2:
		for _, v := range args[0].(*Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{"", "", postStmtCode}}, nil
	case 3:
		for _, v := range args[0].(*Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(*Node).Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[1].(*Node).Place, []string{initStmtCode, condStmtCode, ""}}, nil
	case 4:
		for _, v := range args[0].(*Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(*Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{initStmtCode, "", postStmtCode}}, nil
	case 5:
		for _, v := range args[0].(*Node).Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(*Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[0].(*Node).Place, []string{"", condStmtCode, postStmtCode}}, nil
	case 6:
		for _, v := range args[0].(*Node).Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].(*Node).Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[2].(*Node).Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[1].(*Node).Place, []string{initStmtCode, condStmtCode, postStmtCode}}, nil
	}
	return &Node{}, fmt.Errorf("NewForClause: Invalid type %d", typ)
}

// NewDeferStmt returns a defer statement.
func NewDeferStmt(expr, args Attrib) (*Node, error) {
	// Add code corresponding to the arguments.
	n := &Node{"", args.(*Node).Code}
	argExpr := utils.SplitAndSanitize(args.(*Node).Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.(*Node).Place, k, v))
	}
	n.Place = NewTmp()
	// Push the code for the actual function call to the defer stack.
	deferCode := make(DeferStackItem, 0)
	deferCode = append(deferCode, fmt.Sprintf("call, %s", expr.(*Node).Place))
	deferCode = append(deferCode, fmt.Sprintf("store, %s", n.Place))
	deferStack.Push(deferCode)
	return n, nil
}

// NewIOStmt returns an I/O statement.
func NewIOStmt(typ, expr Attrib) (*Node, error) {
	n := &Node{"", expr.(*Node).Code}
	typ = string(typ.(*token.Token).Lit)
	switch typ {
	case "printInt":
		// The IR for printInt is supposed to look as following -
		//	printInt, a, a
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s", typ, expr.(*Node).Place, expr.(*Node).Place))
	case "printStr":
		fallthrough
	case "scanInt":
		n.Code = append(n.Code, fmt.Sprintf("%s, %s", typ, expr.(*Node).Place))
	}
	return n, nil
}

// NewIncDecStmt returns an increment or a decrement statement.
func NewIncDecStmt(op, expr Attrib) (*Node, error) {
	op = string(op.(*token.Token).Lit)
	n := &Node{"", expr.(*Node).Code}
	switch op {
	case "++":
		n.Code = append(n.Code, fmt.Sprintf("+, %s, %s, 1", expr.(*Node).Place, expr.(*Node).Place))
	case "--":
		n.Code = append(n.Code, fmt.Sprintf("-, %s, %s, 1", expr.(*Node).Place, expr.(*Node).Place))
	default:
		return &Node{}, fmt.Errorf("Invalid operator %s", op)
	}
	return n, nil
}

// NewAssignStmt returns an assignment statement.
func NewAssignStmt(typ int, op, leftExpr, rightExpr Attrib) (*Node, error) {
	switch typ {
	case 0:
		op := string(op.(*token.Token).Lit)[0]
		n := &Node{"", leftExpr.(*Node).Code}
		n.Code = append(n.Code, rightExpr.(*Node).Code...)
		leftExpr := utils.SplitAndSanitize(leftExpr.(*Node).Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.(*Node).Place, ",")
		for k, v := range leftExpr {
			n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, v, v, rightExpr[k]))
			if symEntry, ok := Lookup(v); ok && symEntry.kind == ARRAY {
				// The IR notation for assigning an array member to a
				// variable is of the form -
				//	into, destination, destination, index, array-name
				// destination appears twice because of the way register
				// spilling is currently being handled.
				dst := symEntry.symbols[0]
				index := symEntry.symbols[1]
				n.Code = append(n.Code, fmt.Sprintf("into, %s, %s, %s, %s", dst, dst, index, v))
			}
		}
		return n, nil
	case 1:
		n := &Node{"", leftExpr.(*Node).Code}
		n.Code = append(n.Code, rightExpr.(*Node).Code...)
		leftExpr := utils.SplitAndSanitize(leftExpr.(*Node).Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.(*Node).Place, ",")
		if len(leftExpr) != len(rightExpr) {
			return &Node{}, ErrCountMismatch(len(leftExpr), len(rightExpr))
		}
		for k, v := range leftExpr {
			if currScope.symTab[RealName(v)].kind == POINTER {
				if strings.HasPrefix(rightExpr[k], PTR) {
					varName := RealName(rightExpr[k][8:])
					currScope.symTab[RealName(v)].symbols[1] = currScope.symTab[varName].symbols[0]
				} else {
					varName := RealName(rightExpr[k])
					currScope.symTab[RealName(v)].symbols[1] = currScope.symTab[varName].symbols[1]
				}
			} else if strings.HasPrefix(rightExpr[k], DRF) {
				if strings.HasPrefix(v, DRF) {
					leftVar := currScope.symTab[RealName(v[6:])].symbols[1]
					rightVar := currScope.symTab[RealName(rightExpr[k][6:])].symbols[1]
					n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", leftVar, rightVar))
				} else {
					leftVar := currScope.symTab[RealName(v)].symbols[0]
					rightVar := currScope.symTab[RealName(rightExpr[k][6:])].symbols[1]
					n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", leftVar, rightVar))
				}
			} else if strings.HasPrefix(v, DRF) {
				varName := currScope.symTab[RealName(v[6:])].symbols[1]
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", varName, rightExpr[k]))
			} else {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", v, rightExpr[k]))
				if symEntry, ok := Lookup(v); ok && symEntry.kind == ARRAY {
					// The IR notation for assigning an array member to a
					// variable is of the form -
					//	into, destination, destination, index, array-name
					// destination appears twice because of the way register
					// spilling is currently being handled.
					dst := symEntry.symbols[0]
					index := symEntry.symbols[1]
					n.Code = append(n.Code, fmt.Sprintf("into, %s, %s, %s, %s", dst, dst, index, v))
				}
			}
		}
		return n, nil
	case 2:
		n := &Node{"", []string{}}
		// TODO: Structs do not support multiple short declarations in a
		// single statement for now.
		exprName := rightExpr.(*Node).Place
		if strings.HasPrefix(exprName, "struct") {
			return &Node{}, ErrDeclStruct
		} else {
			n.Code = rightExpr.(*Node).Code
			expr := utils.SplitAndSanitize(rightExpr.(*Node).Place, ",")
			if len(leftExpr.(*Node).Code) != len(expr) {
				return &Node{}, ErrCountMismatch(len(leftExpr.(*Node).Code), len(expr))
			}
			for k, v := range leftExpr.(*Node).Code {
				if symEntry, found := Lookup(v); found {
					renamedVar := symEntry.symbols[0]
					if strings.HasPrefix(expr[k], ARR) {
						return &Node{}, ErrDeclArr
					} else if symEntry.kind != POINTER {
						n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
					}
				} else {
					return &Node{}, fmt.Errorf("undefined: %s", v)
				}
			}
		}
		return n, nil
	}
	return &Node{}, nil
}

// NewShortDecl returns a short variable declaration.
func NewShortDecl(identList, exprList Attrib) (*Node, error) {
	n := &Node{"", []string{}}
	// TODO: Structs do not support multiple short declarations in a single
	// statement for now.
	exprName := exprList.(*Node).Place
	if strings.HasPrefix(exprName, "struct") {
		// The following index calculations assume that struct names
		// cannot include a semicolon ':'.
		colonIndex := strings.LastIndexAny(exprName, ":")
		structLen, err := strconv.Atoi(exprName[7:colonIndex])
		if err != nil {
			return &Node{}, err
		}
		// TODO: Multiple struct initializations using short declaration
		// are not handled currently.
		structName := identList.(*Node).Code[0]
		// keeping structName in the symbol table with type as Struct
		InsertSymbol(structName, STRUCT, structName)
		// The individual struct member initializers can contain
		// expressions whose code need to be added before the members
		// are initialized.
		n.Code = append(n.Code, exprList.(*Node).Code[2*structLen:]...)

		// Add code for struct member initializations.
		var varName, varVal string
		for k, v := range exprList.(*Node).Code[:2*structLen] {
			if k%2 == 0 {
				// Member names are located at even locations.
				varName = v
			} else {
				// (Initialized) member values are located at odd locations.
				varVal = v
				renamedVar := RenameVariable(structName + "." + varName)
				InsertSymbol(structName+"."+varName, INTEGER, renamedVar)
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, varVal))
			}
		}
	} else {
		n.Code = exprList.(*Node).Code
		expr := utils.SplitAndSanitize(exprList.(*Node).Place, ",")
		numIdent := len(identList.(*Node).Code) // number of identifiers
		if numIdent != len(expr) {
			return &Node{}, ErrCountMismatch(numIdent, len(expr))
		}
		for k, v := range identList.(*Node).Code {
			renamedVar := RenameVariable(v)
			if _, ok := GetSymbol(v); !ok {
				// TODO: All types are int currently.
				if strings.HasPrefix(expr[k], PTR) {
					InsertSymbol(v, POINTER, renamedVar, expr[k][8:])
				} else if currScope.symTab[RealName(expr[k])].kind == POINTER {
					InsertSymbol(v, POINTER, renamedVar, currScope.symTab[RealName(expr[k])].symbols[1])
				} else if strings.HasPrefix(expr[k], DRF) {
					InsertSymbol(v, INTEGER, renamedVar)
				} else {
					InsertSymbol(v, INTEGER, renamedVar)
				}
			} else {
				return &Node{}, ErrShortDecl
			}
			if strings.HasPrefix(expr[k], ARR) {
				// TODO: rename arrays
				n.Code = append(n.Code, fmt.Sprintf("decl, %s, %s", renamedVar, expr[k][6:]))
			} else if strings.HasPrefix(expr[k], DRF) {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, currScope.symTab[RealName(expr[k][6:])].symbols[1]))
			} else if strings.HasPrefix(expr[k], STR) {
				// Check if the RHS is a string.
				n.Code = append(n.Code, fmt.Sprintf("declStr, %s, %s", renamedVar, expr[k][7:]))
			} else if currScope.symTab[v].kind != POINTER {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, expr[k]))
			}
		}
	}
	return n, nil
}
