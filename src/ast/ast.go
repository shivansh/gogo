// Package ast implements utility functions for generating abstract syntax
// trees from Go BNF.

package ast

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shivansh/gogo/src/tac"
	"github.com/shivansh/gogo/src/utils"
)

// DeferStackItem is an individual item stored when a call to defer is made. It
// contains the code for the corresponding function call which is placed at the
// end of the function body.
type DeferStackItem []string

var (
	// deferStack stores the deferred function calls which are then called
	// when the surrounding function block ends.
	deferStack *utils.Stack
	re         *regexp.Regexp
	PkgName    string
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

// PrintIR generates the IR instructions accumulated in the "Code" attribute of
// the "SourceFile" non-terminal.
func PrintIR(src *Node) (*Node, error) {
	for _, v := range src.Code {
		if stmt := strings.TrimSpace(v); stmt != "" {
			fmt.Println(stmt)
		}
	}
	return nil, nil
}

// InitNode initializes an AST node with the given "Place" and "Code" attributes.
func InitNode(place string, code []string) (*Node, error) {
	return &Node{place, code}, nil
}

// --- [ Package declarations ] ----==------------------------------------------

// NewPkgDecl initializes package relevant data structures.
func NewPkgDecl(pkgName []byte) (*Node, error) {
	PkgName = string(pkgName)
	return nil, nil
}

// --- [ Top level declarations ] ----------------------------------------------

// NewTopLevelDecl returns a top level declaration.
func NewTopLevelDecl(topDecl, repeatTopDecl *Node) (*Node, error) {
	return &Node{"", append(topDecl.Code, repeatTopDecl.Code...)}, nil
}

// NewTypeDef returns a type definition.
func NewTypeDef(ident string, typ AstNode) (AstNode, error) {
	switch typ.(type) {
	case *StructType:
		node := Node{typ.place(), typ.code()}
		return &StructType{node, ident, 0}, nil
	}
	return &Node{fmt.Sprintf("%s, %s", typ.place(), ident), typ.code()}, nil
}

// NewElementList returns a keyed element list.
func NewElementList(key, keyList *Node) (*Node, error) {
	return &Node{fmt.Sprintf("%s, %s", key.Place, keyList.Place),
		append(key.Code, keyList.Code...)}, nil
}

// AppendKeyedElement appends a keyed element to a list of keyed elements.
func AppendKeyedElement(key, keyList *Node) (*Node, error) {
	return &Node{fmt.Sprintf("%s, %s", key.Place, keyList.Place),
		append(key.Code, keyList.Code...)}, nil
}

// --- [ Variable declarations ] -----------------------------------------------

// NewVarSpec creates a new variable specification.
// The accepted variadic arguments (args) in their order are -
// 	- IdentifierList
// 	- Type
// 	- ExpressionList
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewVarSpec(typ int, args ...*Node) (*Node, error) {
	n := &Node{"", []string{}}
	expr := []string{}
	var vartype, exprtype symkind

	// Evaluate the type of identifier from the declaration.
	switch args[1].Place {
	case INT:
		vartype = INTEGER
	case STR:
		vartype = STRING
	default:
		return &Node{}, fmt.Errorf("unsupported type: %s", args[1].Place)
	}

	// Add the IR instructions for ExpressionList.
	switch typ {
	case 1:
		n.Code = args[2].Code
		expr = utils.SplitAndSanitize(args[2].Place, ",")
	case 2:
		// Infer type of identifier from the expression.
		if re.MatchString(args[1].Place) {
			vartype = INTEGER
		} else {
			vartype = STRING
		}
		n.Code = args[1].Code
		expr = utils.SplitAndSanitize(args[1].Place, ",")
	}

	for k, v := range args[0].Code {
		renamedVar := RenameVariable(v)
		InsertSymbol(v, vartype, renamedVar)

		// Evaluate type of the expression
		if typ == 1 {
			if symEntry, found := Lookup(RealName(expr[k])); found {
				switch symEntry.kind {
				case ARRAYINT:
					exprtype = INTEGER
				case ARRAYSTR:
					exprtype = STRING
				default:
					exprtype = symEntry.kind
				}
			} else if re.MatchString(expr[k]) {
				exprtype = INTEGER
			} else if strings.HasPrefix(expr[k], ARR) {
				switch GetPrefix(expr[k]) {
				case ARRINT:
					exprtype = INTEGER
				case ARRSTR:
					exprtype = STRING
				default:
					panic("type not supported by arrays")
				}
			} else {
				exprtype = STRING
			}
		} else {
			exprtype = STRING
		}

		if typ == 0 {
			// Initialize identifiers to their default values
			// depending on type information.
			switch vartype {
			case INTEGER:
				n.Code = append(n.Code, fmt.Sprintf("declInt, %s, 0", renamedVar))
			case STRING:
				n.Code = append(n.Code, fmt.Sprintf("declStr, %s, \"\"", renamedVar))
			}
		} else if typ == 1 || typ == 2 {
			if vartype == exprtype {
				switch vartype {
				case INTEGER:
					n.Code = append(n.Code, fmt.Sprintf("declInt, %s, %s", renamedVar, StripPrefix(expr[k])))
				case STRING:
					n.Code = append(n.Code, fmt.Sprintf("declStr, %s, %s", renamedVar, StripPrefix(expr[k])))
				}
			} else {
				exprName := RealName(StripPrefix(expr[k]))
				return &Node{}, fmt.Errorf("cannot use %s (type %s) as type %s in assignment",
					exprName, GetType(exprtype), GetType(vartype))
			}
		}
	}

	return n, nil
}

// --- [ Type declarations ] ---------------------------------------------------

// NewTypeDecl returns a type declaration.
func NewTypeDecl(typespec AstNode) (*Node, error) {
	switch t := typespec.(type) {
	case *StructType:
		globalSymTab[t.Name] = SymTabEntry{
			kind:    STRUCT,
			symbols: t.code(),
		}
	default:
		return &Node{}, fmt.Errorf("unknown type %v", t)
	}
	// Member initialization will be done when a new object is instantiated.
	return &Node{"", []string{}}, nil
}

// --- [ Constant declarations ] -----------------------------------------------

// NewConstSpec returns a constant declaration.
func NewConstSpec(typ int, args ...*Node) (*Node, error) {
	n := &Node{"", []string{}}
	expr := []string{}
	if typ == 1 {
		n.Code = append(n.Code, args[1].Code...)
		expr = utils.SplitAndSanitize(args[1].Place, ",")
	}
	for k, v := range args[0].Code {
		renamedVar := RenameVariable(v)
		// TODO: Add remaining types.
		InsertSymbol(v, INTEGER, renamedVar)
		switch typ {
		case 0:
			n.Code = append(n.Code, fmt.Sprintf("declInt, %s, 0", renamedVar))
		case 1:
			n.Code = append(n.Code, fmt.Sprintf("declInt, %s, %s", renamedVar, expr[k]))
		}
	}
	return n, nil
}

// --- [ Expressions ] ---------------------------------------------------------

// AppendExpr appends a list of expressions to a given expression.
func AppendExpr(expr, exprlist *Node) (*Node, error) {
	return &Node{fmt.Sprintf("%s,%s", expr.Place, exprlist.Place),
		append(expr.Code, exprlist.Code...)}, nil
}

// NewBoolExpr returns a new logical expression.
func NewBoolExpr(op string, leftexpr, rightexpr *Node) (*Node, error) {
	n := &Node{"", append(leftexpr.Code, rightexpr.Code...)}
	n.Place = NewTmp()
	afterLabel := NewLabel()
	switch op {
	case OR:
		trueLabel := NewLabel()
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, leftexpr.Place),
			fmt.Sprintf("beq, %s, %s, 1", trueLabel, rightexpr.Place),
			fmt.Sprintf("=, %s, 0", n.Place),
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
			fmt.Sprintf("label, %s", trueLabel),
			fmt.Sprintf("=, %s, 1", n.Place),
			fmt.Sprintf("label, %s", afterLabel),
		)
	case AND:
		falseLabel := NewLabel()
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, leftexpr.Place),
			fmt.Sprintf("beq, %s, %s, 0", falseLabel, rightexpr.Place),
			fmt.Sprintf("=, %s, 1", n.Place),
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
			fmt.Sprintf("label, %s", falseLabel),
			fmt.Sprintf("=, %s, 0", n.Place),
			fmt.Sprintf("label, %s", afterLabel),
		)
	}
	return n, nil
}

// NewRelExpr returns a new relational expression.
func NewRelExpr(op, leftexpr, rightexpr *Node) (*Node, error) {
	n := &Node{"", append(leftexpr.Code, rightexpr.Code...)}
	n.Place = NewTmp()
	branchOp := ""
	falseLabel := NewLabel()
	afterLabel := NewLabel()
	switch op.Place {
	case EQ:
		branchOp = "bne"
	case NEQ:
		branchOp = "beq"
	case LEQ:
		branchOp = "bgt"
	case LT:
		branchOp = "bge"
	case GEQ:
		branchOp = "blt"
	case GT:
		branchOp = "ble"
	}
	n.Code = utils.AppendCode(
		n.Code,
		fmt.Sprintf("%s, %s, %s, %s", branchOp, falseLabel, leftexpr.Place, rightexpr.Place),
		fmt.Sprintf("=, %s, 1", n.Place),
		fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
		fmt.Sprintf("label, %s", falseLabel),
		fmt.Sprintf("=, %s, 0", n.Place),
		fmt.Sprintf("label, %s", afterLabel),
	)
	return n, nil
}

// NewArithExpr returns an arithmetic expression.
func NewArithExpr(op string, leftexpr, rightexpr *Node) (*Node, error) {
	n := &Node{"", append(leftexpr.Code, rightexpr.Code...)}
	if re.MatchString(leftexpr.Place) && re.MatchString(rightexpr.Place) {
		// --- [ Constant folding optimization ] -----------------------
		// Expression is of the form "1 op 2". Such expression can
		// be reduced by evaluating its value during compilation itself.
		leftval, err := strconv.Atoi(leftexpr.Place)
		if err != nil {
			return &Node{}, err
		}
		rightval, err := strconv.Atoi(rightexpr.Place)
		if err != nil {
			return &Node{}, err
		}
		switch op {
		case ADD:
			n.Place = strconv.Itoa(leftval + rightval)
		case SUB:
			n.Place = strconv.Itoa(leftval - rightval)
		case AST:
			n.Place = strconv.Itoa(leftval * rightval)
		case DIV:
			n.Place = strconv.Itoa(leftval / rightval)
		case REM:
			n.Place = strconv.Itoa(leftval % rightval)
		default:
			return &Node{}, fmt.Errorf("Invalid operation %s", op)
		}
	} else if re.MatchString(leftexpr.Place) {
		// Expression is of the form "1 + b", which needs to be
		// converted to the equivalent form "b + 1" to be counted as a
		// valid IR statement.
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, rightexpr.Place, leftexpr.Place))
	} else {
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, n.Place, leftexpr.Place, rightexpr.Place))
	}
	return n, nil
}

// NewUnaryExpr returns a unary expression.
func NewUnaryExpr(op, expr *Node) (*Node, error) {
	n := &Node{"", expr.Code}
	switch op.Place {
	case SUB:
		if re.MatchString(expr.Place) {
			// expression is of the form 1+2, perform constant folding.
			term3val, err := strconv.Atoi(expr.Place)
			if err != nil {
				return &Node{}, err
			}
			n.Place = strconv.Itoa(term3val * -1)
		} else {
			n.Place = NewTmp()
			n.Code = append(n.Code, fmt.Sprintf("*, %s, %s, -1", n.Place, expr.Place))
		}
	case NOT:
		n.Place = NewTmp()
		n.Code = append(n.Code, fmt.Sprintf("not, %s, %s", n.Place, expr.Place))
	case ADD:
		n.Place = expr.Place
	case AMP:
		// Place attribute of a pointer variable starts with "pointer:"
		// followed by its place attribute.
		n.Place = PTR + ":" + expr.Place
	case AST:
		n.Place = DRF + ":" + expr.Place
	default:
		return &Node{}, fmt.Errorf("%s operator not supported", op.Place)
	}
	return n, nil
}

// NewPrimaryExprSel returns an AST node for PrimaryExpr Selector.
func NewPrimaryExprSel(expr, selector *Node) (*Node, error) {
	// The key for a selector's symbol table entry is of the form -
	//	(exprPlace).(selectorPlace)
	varName := fmt.Sprintf("%s.%s", expr.Place, selector.Place)
	if symEntry, found := Lookup(varName); found {
		if _, found := globalSymTab[varName]; found {
			// TODO verify if this is correct.
			return &Node{}, fmt.Errorf("undefined: %s", varName)
		} else {
			return &Node{symEntry.symbols[0], []string{}}, nil
		}
	} else {
		return &Node{}, fmt.Errorf("undefined: %s", varName)
	}
}

// NewPrimaryExprIndex returns an AST node for PrimaryExpr Index.
func NewPrimaryExprIndex(expr, index *Node) (*Node, error) {
	n := &Node{"", []string{}}
	n.Place = NewTmp()
	// NOTE: Indexing is only supported by array types and not pointer types.

	var exprtype symkind
	if symEntry, found := Lookup(RealName(expr.Place)); found {
		switch symEntry.kind {
		case INTEGER:
			exprtype = ARRAYINT
		case STRING:
			exprtype = ARRAYSTR
		default:
			panic("NewPrimaryExprIndex: indexing not supported on type")
		}
	} else {
		return &Node{}, fmt.Errorf("undefined: %s", RealName(expr.Place))
	}
	InsertSymbol(n.Place, exprtype, expr.Place, index.Place)

	n.Code = append(n.Code, index.Code...)
	n.Code = append(n.Code, fmt.Sprintf("from, %s, %s, %s", n.Place, expr.Place, index.Place))
	return n, nil
}

// NewPrimaryExprArgs returns an AST node for PrimaryExpr Arguments.
// NOTE: This is the production rule for a function call.
func NewPrimaryExprArgs(expr, args *Node) (*Node, error) {
	n := &Node{"", args.Code}
	symEntry := globalSymTab[expr.Place]
	returnLen := 0
	if symEntry.kind == FUNCTION {
		if ret, err := strconv.Atoi(symEntry.symbols[0]); err != nil {
			return &Node{}, err
		} else {
			returnLen = ret
		}
	} else {
		return &Node{}, fmt.Errorf("%s is not a function", expr.Place)
	}
	argExpr := utils.SplitAndSanitize(args.Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.Place, k, v))
	}
	n.Code = append(n.Code, fmt.Sprintf("call, %s", expr.Place))
	for k := 0; k < returnLen; k++ {
		n.Place = fmt.Sprintf("%s, return.%d", n.Place, k)
	}
	return n, nil
}

// NewCompositeLit returns a composite literal.
func NewCompositeLit(typ, val *Node) (AstNode, error) {
	n := &Node{"", []string{}}
	// Check if the LiteralType corresponds to ArrayType. This is done
	// because unlike structs it is not required to add a symbol table entry
	// for place values of arrays (which is of the form "array:<length>"),
	// thus returning early.
	if strings.HasPrefix(typ.Place, ARR) {
		n.Place = typ.place()
		return n, nil
	}
	// In case the corresponds to a struct, add the code for its data member
	// initialization.
	if symEntry, found := Lookup(typ.Place); !found {
		return &Node{}, fmt.Errorf("undefined: %s", typ.Place)
	} else {
		switch symEntry.kind {
		case STRUCT:
			litVals := utils.SplitAndSanitize(val.Place, ",")
			litValCodes := val.Code
			structInit := []string{}
			// In case of integral (or any type) initializations, the corresponding
			// lexeme is placed at the place value, justifying the length check which
			// is made on 'litVals' instead of 'litValCodes'. If there are no place
			// values for the data members, then initialize all to their default
			// values. Otherwise initialize them to the corresponding place value.
			if len(litVals) == 0 {
				for k, v := range symEntry.symbols {
					if k%2 == 0 {
						// TODO: Update default values depending on type.
						// The default type is currently assumed to be int.
						structInit = append(structInit, v, "0")
					}
				}
			} else {
				for k, v := range symEntry.symbols {
					if k%2 == 0 {
						structInit = append(structInit, v, litVals[k/2])
					}
				}
			}
			// When these code values will be utilized above, the litValCodes will be
			// placed above the code corresponding to structInit (litVals can be expressions).
			n.Code = utils.AppendCode(n.Code, structInit, litValCodes)
			return &StructType{*n, typ.place(), len(symEntry.symbols) / 2}, nil
		}
	}
	return n, nil
}

// NewIdentifier returns a new identifier.
func NewIdentifier(varName string) (*Node, error) {
	if symEntry, found := Lookup(varName); found {
		if _, found := globalSymTab[varName]; found {
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
func NewFuncDecl(marker, body *Node) (AstNode, error) {
	n := &FuncType{Node{"", append(marker.Code, body.Code...)}}
	funcSymtabCreated = true // end of function block
	// Return statement insertion will be handled when the defer stack is
	// emptied and the code for deferred calls has been inserted.
	if deferStack.Len > 0 {
		defer func() { n.Code = append(n.Code, "ret,") }()
	}
	for deferStack.Len > 0 {
		deferFuncCode := deferStack.Pop().(DeferStackItem)
		n.Code = append(n.Code, deferFuncCode...)
	}
	return n, nil
}

// NewFuncMarker returns a marker non-terminal used in the production rule for
// function declaration.
func NewFuncMarker(name, signature *Node) (*Node, error) {
	n := &Node{name.Place, []string{fmt.Sprintf("func, %s", name.Place)}}
	// Assign values to arguments.
	for k, v := range signature.Code {
		n.Code = append(n.Code, fmt.Sprintf("=, %s, %s.%d", v, name.Place, k))
	}
	if _, found := globalSymTab[name.Place]; !found {
		globalSymTab[name.Place] = SymTabEntry{
			kind:    FUNCTION,
			symbols: []string{signature.Place},
		}
	} else {
		return &Node{}, fmt.Errorf("function %s is already declared\n", name.Place)
	}
	return n, nil
}

// NewSignature returns a function signature.
// The accepted variadic arguments (args) in their order are -
// 	- parameters
//	- result
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewSignature(typ int, args ...*Node) (*Node, error) {
	NewScope()
	for _, v := range args[0].Code {
		if v == "" {
			break
		}
		InsertSymbol(v, INTEGER, v)
	}
	if typ == 0 {
		return &Node{"0", args[0].Code}, nil
	} else {
		return &Node{fmt.Sprintf("%s", args[1].Place), args[0].Code}, nil
	}
}

// NewResult defines the return type of a function.
func NewResult(params *Node) (*Node, error) {
	returnLength := 0
	// Evaluate the number of return values.
	for _, v := range params.Code {
		if v == INT {
			returnLength++
		}
	}
	return &Node{fmt.Sprintf("%d", returnLength), []string{}}, nil
}

// NewParamList returns a list of parameters.
func NewParamList(decl, declList *Node) (*Node, error) {
	n := &Node{"", append(decl.Code, declList.Code...)}
	return n, nil
}

// AppendParam appends a parameter to a list of parameters.
func AppendParam(decl, declList *Node) (*Node, error) {
	n := &Node{"", append(decl.Code, declList.Code...)}
	return n, nil
}

// NewArrayType returns an array.
func NewArrayType(arrLen, arrType string) (*Node, error) {
	n := &Node{"", []string{}}
	switch arrType {
	case INT:
		n.Place = ARRINT + ":" + arrLen
	case STR:
		n.Place = ARRSTR + ":" + arrLen
	default:
		panic("NewArrayType: type not supported by arrays")
	}
	return n, nil
}

// NewStruct returns a struct.
func NewStruct(node *Node) (*StructType, error) {
	n := new(StructType)
	n.Node = *node
	n.Len = len(node.code())
	return n, nil
}

// NewFieldDecl returns a field declaration.
func NewFieldDecl(identList, typ *Node) (*Node, error) {
	n := &Node{"", []string{}}
	// The AST node contains the identifier name and its type one after the
	// another in Code.
	for _, v := range identList.Code {
		n.Code = append(n.Code, v)
		n.Code = append(n.Code, typ.Place)
	}
	return n, nil
}

// AppendFieldDecl appends a field declaration to a list of field declarations.
func AppendFieldDecl(decl, declList *Node) (*Node, error) {
	n := &Node{"", append(decl.Code, declList.Code...)}
	return n, nil
}

// AppendIdent appends an identifier to a list of identifiers.
func AppendIdent(ident string, identList *Node) (*Node, error) {
	// The lexemes corresponding to the individual identifiers are appended
	// to the slice for code to avoid adding comma-separated string in place
	// since the identifiers don't have any IR code to be added.
	n := &Node{"", append([]string{ident}, identList.Code...)}
	return n, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewStmtList returns a statement list.
func NewStmtList(stmt, stmtList *Node) (*Node, error) {
	n := &Node{"", append(stmt.Code, stmtList.Code...)}
	return n, nil
}

// NewLabelStmt returns a labeled statement.
func NewLabelStmt(label, stmt *Node) (AstNode, error) {
	n := &LabeledStmt{Node{"", []string{"label, " + label.Place}}}
	n.Code = append(n.Code, stmt.Code...)
	return n, nil
}

// NewReturnStmt returns a return statement.
// A return statement can be of the following types -
//	- an empty return: In this case the argument expr is empty.
//	- non-empty return: In this case the argument expr contains the return
//	  expression.
func NewReturnStmt(expr ...*Node) (*ReturnStmt, error) {
	n := &ReturnStmt{Node{"", []string{}}}
	if len(expr) == 0 {
		// The return statement is empty.
		// The defer statements need to be inserted before the return stmt (and
		// not at the end of function block as was the previous misconception).
		// When defer stmt is used, the return stmt for main() is also inserted
		// when all the defer calls from stack are popped and inserted in IR.
		if deferStack.Len > 0 {
			// Return statement insertion will be handled when defer
			// stack is emptied and the deferred calls are inserted.
			return n, nil
		} else {
			n.Code = append(n.Code, "ret,")
			return n, nil
		}
	} else {
		if deferStack.Len == 0 {
			retExpr := utils.SplitAndSanitize(expr[0].Place, ",")
			n.Code = append(n.Code, expr[0].Code...)
			for k, v := range retExpr {
				n.Code = append(n.Code, fmt.Sprintf("=, return.%d, %s", k, v))
			}
			n.Code = append(n.Code, "ret,")
		}
		return n, nil
	}
}

// --- [ Blocks ] --------------------------------------------------------------

// NewBlock ends the scope of the previous block and returns a new block.
func NewBlock(stmt *Node) (*Node, error) {
	currScope = currScope.parent // end of the previous block
	return &Node{"", stmt.Code}, nil
}

// NewBlockMarker returns a marker non-terminal used in the production rule for
// block declaration. This marker demarcates the beginning of a new block and
// the corresponding symbol table is instantiated here.
func NewBlockMarker() (*Node, error) {
	if !funcSymtabCreated {
		// The symbol table for functions is created when the rule for
		// Signature is reached so that the arguments can also be added
		// At this point the function scope (if there was any) has ended.
		NewScope()
	} else {
		// Allow creation of symbol table for another function.
		funcSymtabCreated = false
	}
	return nil, nil
}

// --- [ Statements ] ----------------------------------------------------------

// NewIfStmt returns an if statement.
func NewIfStmt(typ int, args ...*Node) (*Node, error) {
	n := &Node{"", args[0].Code}
	afterLabel := NewLabel()
	elseLabel := NewLabel()
	switch typ {
	case 0:
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].Place),
			args[1].Code,
		)

	case 1:
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].Place),
			args[1].Code,
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].Code,
		)

	case 2:
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[0].Place),
			args[1].Code,
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[2].Code,
		)

	case 3:
		n.Code = utils.AppendCode(
			args[1].Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[1].Place),
			args[2].Code,
		)

	case 4, 5:
		n.Code = utils.AppendCode(
			args[1].Code,
			fmt.Sprintf("blt, %s, %s, 1", elseLabel, args[1].Place),
			args[2].Code,
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
			fmt.Sprintf("label, %s", elseLabel),
			args[3].Code,
		)
	}
	n.Code = append(n.Code, fmt.Sprintf("label, %s", afterLabel))
	return n, nil
}

// NewSwitchStmt returns a switch statement.
func NewSwitchStmt(expr, caseClause *Node) (*SwitchStmt, error) {
	n := &SwitchStmt{Node{"", expr.Code}}
	caseLabels := []string{}
	caseStmts := caseClause.Code
	// SplitAndSanitize cannot be used here as removal of empty entries
	// seems to be causing erroneous index calculations.
	// Regression caused in "test/codegen/switch.go".
	caseTemporaries := strings.Split(caseClause.Place, ",")
	afterLabel := NewLabel()
	defaultLabel := afterLabel
	// The last value in caseTemporaries will be the place value returned by
	// Empty (arising from the rule RepeatExprCaseClause -> Empty).
	// This has to be ignored.
	for k, v := range caseTemporaries[:len(caseTemporaries)-1] {
		caseLabel := NewLabel()
		caseLabels = append(caseLabels, caseLabel)
		n.Code = append(n.Code, caseStmts[2*k])
		if strings.TrimSpace(v) == "default" {
			defaultLabel = caseLabel
		} else {
			n.Code = append(n.Code, fmt.Sprintf("beq, %s, %s, %s", caseLabel, expr.Place, v))
		}
	}
	n.Code = append(n.Code, fmt.Sprintf("%s, %s", tac.JMP, defaultLabel))
	for k, v := range caseLabels {
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("label, %s", v),
			caseStmts[2*k+1],
			fmt.Sprintf("%s, %s", tac.JMP, afterLabel),
		)
	}
	n.Code = append(n.Code, fmt.Sprintf("label, %s", afterLabel))
	return n, nil
}

// NewExprCaseClause returns an expression case clause.
func NewExprCaseClause(expr, stmtList *Node) (*Node, error) {
	n := &Node{expr.Place, []string{}}
	var exprCode, stmtCode string
	for _, v := range expr.Code {
		exprCode += fmt.Sprintf("%s\n", v)
	}
	n.Code = append(n.Code, exprCode)
	for _, v := range stmtList.Code {
		stmtCode += fmt.Sprintf("%s\n", v)
	}
	n.Code = append(n.Code, stmtCode)
	return n, nil
}

// AppendExprCaseClause appends an expression case clause to a list of same.
func AppendExprCaseClause(expr, exprList *Node) (*Node, error) {
	n := &Node{"", append(expr.Code, exprList.Code...)}
	n.Place = fmt.Sprintf("%s, %s", expr.Place, exprList.Place)
	return n, nil
}

// NewForStmt returns a for statement.
// The accepted variants of the variadic arguments (args) are -
//	- Block
//	- Condition, Block
//	- ForClause, Block
// The cardinal argument `typ` determines the index of the production rule
// invoked starting from top.
func NewForStmt(typ int, args ...*Node) (*ForStmt, error) {
	n := &ForStmt{Node{"", []string{}}}
	startLabel := NewLabel()
	afterLabel := NewLabel()
	blockCode := []string{}

	switch typ {
	case 0:
		n.Code = append(n.Code, fmt.Sprintf("label, %s", startLabel))
		blockCode = args[0].Code

	case 1:
		n.Code = utils.AppendCode(
			n.Code,
			fmt.Sprintf("label, %s", startLabel),
			args[0].Code,
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].Place),
		)
		blockCode = args[1].Code

	case 2:
		n.Code = utils.AppendCode(
			n.Code,
			args[0].Code[0], // init stmt
			fmt.Sprintf("label, %s", startLabel),
			args[0].Code[1], // condition
			fmt.Sprintf("blt, %s, %s, 1", afterLabel, args[0].Place),
		)
		blockCode = args[1].Code
	}

	for _, v := range blockCode {
		v := strings.TrimSpace(v)
		switch v {
		case "break":
			n.Code = append(n.Code, fmt.Sprintf("%s, %s", tac.JMP, afterLabel))
		case "continue":
			n.Code = append(n.Code, fmt.Sprintf("%s, %s", tac.JMP, startLabel))
		default:
			n.Code = append(n.Code, v)
		}
	}

	if typ == 2 {
		n.Code = append(n.Code, args[0].Code[2]) // post stmt
	}

	n.Code = utils.AppendCode(
		n.Code,
		fmt.Sprintf("%s, %s", tac.JMP, startLabel),
		fmt.Sprintf("label, %s", afterLabel),
	)

	return n, nil
}

// NewForClause returns a for clause.
func NewForClause(typ int, args ...*Node) (*Node, error) {
	var initStmtCode, condStmtCode, postStmtCode string
	switch typ {
	case 0:
		for _, v := range args[0].Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{initStmtCode, "", ""}}, nil
	case 1:
		for _, v := range args[0].Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[0].Place, []string{"", condStmtCode, ""}}, nil
	case 2:
		for _, v := range args[0].Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{"", "", postStmtCode}}, nil
	case 3:
		for _, v := range args[0].Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[1].Place, []string{initStmtCode, condStmtCode, ""}}, nil
	case 4:
		for _, v := range args[0].Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{"1", []string{initStmtCode, "", postStmtCode}}, nil
	case 5:
		for _, v := range args[0].Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[0].Place, []string{"", condStmtCode, postStmtCode}}, nil
	case 6:
		for _, v := range args[0].Code {
			initStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[1].Code {
			condStmtCode += fmt.Sprintf("%s\n", v)
		}
		for _, v := range args[2].Code {
			postStmtCode += fmt.Sprintf("%s\n", v)
		}
		return &Node{args[1].Place, []string{initStmtCode, condStmtCode, postStmtCode}}, nil
	}
	return &Node{}, fmt.Errorf("NewForClause: Invalid type %d", typ)
}

// NewDeferStmt returns a defer statement.
func NewDeferStmt(expr, args *Node) (*DeferStmt, error) {
	// Add code corresponding to the arguments.
	n := &DeferStmt{Node{"", args.Code}}
	argExpr := utils.SplitAndSanitize(args.Place, ",")
	for k, v := range argExpr {
		n.Code = append(n.Code, fmt.Sprintf("=, %s.%d, %s", expr.Place, k, v))
	}
	n.Place = NewTmp()
	// Push the code for the actual function call to the defer stack.
	deferCode := make(DeferStackItem, 0)
	deferCode = append(deferCode, fmt.Sprintf("call, %s", expr.Place))
	deferCode = append(deferCode, fmt.Sprintf("store, %s", n.Place))
	deferStack.Push(deferCode)
	return n, nil
}

// NewIOStmt returns an I/O statement.
func NewIOStmt(typ string, expr *Node) (*Node, error) {
	n := &Node{"", expr.Code}
	switch typ {
	case "printInt":
		// The IR for printInt is supposed to look as following -
		//	printInt, a, a
		n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s", typ, expr.Place, expr.Place))
	case "printStr", "scanInt":
		n.Code = append(n.Code, fmt.Sprintf("%s, %s", typ, expr.Place))
	}
	return n, nil
}

// NewIncDecStmt returns an increment or a decrement statement.
func NewIncDecStmt(op string, expr *Node) (*Node, error) {
	n := &Node{"", expr.Code}
	switch op {
	case INC:
		n.Code = append(n.Code, fmt.Sprintf("+, %s, %s, 1", expr.Place, expr.Place))
	case DEC:
		n.Code = append(n.Code, fmt.Sprintf("-, %s, %s, 1", expr.Place, expr.Place))
	default:
		return &Node{}, fmt.Errorf("Invalid operator %s", op)
	}
	return n, nil
}

// NewAssignStmt returns an assignment statement.
func NewAssignStmt(typ int, op string, leftExpr, rightExpr *Node) (*Node, error) {
	switch typ {
	case 0:
		n := &Node{"", leftExpr.Code}
		n.Code = append(n.Code, rightExpr.Code...)
		leftExpr := utils.SplitAndSanitize(leftExpr.Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.Place, ",")
		for k, v := range leftExpr {
			n.Code = append(n.Code, fmt.Sprintf("%s, %s, %s, %s", op, v, v, rightExpr[k]))
			if symEntry, found := Lookup(v); found {
				switch symEntry.kind {
				case ARRAYINT:
					InsertSymbol(v, INTEGER, symEntry.symbols)
				case ARRAYSTR:
					InsertSymbol(v, STRING, symEntry.symbols)
				default:
					continue
				}
				// The IR notation for assigning an array member to a
				// variable is of the form -
				//	into, destination, destination, index, array-name
				// destination appears twice because of the way register
				// spilling is currently being handled.
				//
				dst := symEntry.symbols[0]
				index := symEntry.symbols[1]
				n.Code = append(n.Code, fmt.Sprintf("into, %s, %s, %s, %s", dst, dst, index, v))
			}
		}
		return n, nil

	case 1:
		n := &Node{"", append(leftExpr.Code, rightExpr.Code...)}
		leftExpr := utils.SplitAndSanitize(leftExpr.Place, ",")
		rightExpr := utils.SplitAndSanitize(rightExpr.Place, ",")
		if len(leftExpr) != len(rightExpr) {
			return &Node{}, ErrCountMismatch(len(leftExpr), len(rightExpr))
		}
		for k, v := range leftExpr {
			if currScope.symTab[RealName(v)].kind == POINTER {
				if strings.HasPrefix(rightExpr[k], PTR) {
					varName := RealName(StripPrefix(rightExpr[k]))
					currScope.symTab[RealName(v)].symbols[1] = currScope.symTab[varName].symbols[0]
				} else {
					varName := RealName(rightExpr[k])
					currScope.symTab[RealName(v)].symbols[1] = currScope.symTab[varName].symbols[1]
				}
				continue
			}
			if strings.HasPrefix(rightExpr[k], DRF) {
				if strings.HasPrefix(v, DRF) {
					symEntry, _ := Lookup(RealName(StripPrefix(v)))
					leftVar := symEntry.symbols[1]
					symEntry, _ = Lookup(RealName(StripPrefix(rightExpr[k])))
					rightVar := symEntry.symbols[1]
					n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", leftVar, rightVar))
				} else {
					symEntry, _ := Lookup(RealName(v))
					leftVar := symEntry.symbols[0]
					symEntry, _ = Lookup(RealName(StripPrefix(rightExpr[k])))
					rightVar := symEntry.symbols[1]
					n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", leftVar, rightVar))
				}
			} else if strings.HasPrefix(v, DRF) {
				symEntry, _ := Lookup(RealName(StripPrefix(v)))
				varName := symEntry.symbols[1]
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", varName, rightExpr[k]))
			} else {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", v, rightExpr[k]))
				if symEntry, found := Lookup(v); found {
					// The IR notation for assigning an array member to a
					// variable is of the form -
					//	into, destination, destination, index, array-name
					// destination appears twice above because of the way
					// register spilling is currently being handled.
					switch symEntry.kind {
					case ARRAYINT:
						InsertSymbol(v, INTEGER, symEntry.symbols)
					case ARRAYSTR:
						InsertSymbol(v, STRING, symEntry.symbols)
					default:
						continue
					}
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
		exprName := rightExpr.Place
		if strings.HasPrefix(exprName, STRCT) {
			return &Node{}, ErrDeclStruct
		} else {
			n.Code = rightExpr.Code
			expr := utils.SplitAndSanitize(rightExpr.Place, ",")
			if len(leftExpr.Code) != len(expr) {
				return &Node{}, ErrCountMismatch(len(leftExpr.Code), len(expr))
			}
			for k, v := range leftExpr.Code {
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
func NewShortDecl(identList *Node, exprList AstNode) (*Node, error) {
	n := &Node{"", []string{}}
	// TODO: Multiple struct initializations using short declaration are not
	// handled currently.
	switch exprList := exprList.(type) {
	case *StructType:
		// The following index calculations assume that struct names
		// cannot include a semicolon ':'.
		structLen := exprList.Len
		structName := identList.Code[0]
		// keeping structName in the symbol table with type as Struct
		InsertSymbol(structName, STRUCT, structName)
		// The individual struct member initializers can contain
		// expressions whose code need to be added before the members
		// are initialized.
		n.Code = append(n.Code, exprList.Code[2*structLen:]...)

		// Add code for struct member initializations.
		var varName, varVal string
		for k, v := range exprList.Code[:2*structLen] {
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

	case *Node:
		n.Code = exprList.Code
		expr := utils.SplitAndSanitize(exprList.Place, ",")
		if numIdent := len(identList.Code); numIdent != len(expr) {
			return &Node{}, ErrCountMismatch(numIdent, len(expr))
		}
		for k, v := range identList.Code {
			renamedVar := RenameVariable(v)
			if _, found := GetSymbol(v); !found {
				if strings.HasPrefix(expr[k], PTR) {
					InsertSymbol(v, POINTER, renamedVar, StripPrefix(expr[k]))
				} else if currScope.symTab[RealName(expr[k])].kind == POINTER {
					InsertSymbol(v, POINTER, renamedVar, currScope.symTab[RealName(expr[k])].symbols[1])
				} else {
					InsertSymbol(v, INTEGER, renamedVar)
				}
			} else {
				return &Node{}, ErrShortDecl
			}
			if strings.HasPrefix(expr[k], ARR) {
				// TODO: rename arrays
				n.Code = append(n.Code, fmt.Sprintf("decl, %s, %s", renamedVar, StripPrefix(expr[k])))
			} else if strings.HasPrefix(expr[k], DRF) {
				n.Code = append(n.Code, fmt.Sprintf("=, %s, %s", renamedVar, currScope.symTab[RealName(StripPrefix(expr[k]))].symbols[1]))
			} else if strings.HasPrefix(expr[k], STR) {
				n.Code = append(n.Code, fmt.Sprintf("declStr, %s, %s", renamedVar, StripPrefix(expr[k])))
			} else if currScope.symTab[v].kind != POINTER {
				// TODO: Add remaining types
				n.Code = append(n.Code, fmt.Sprintf("declInt, %s, %s", renamedVar, expr[k]))
			}
		}
	default:
		panic("NewShortDecl: invalid type")
	}
	return n, nil
}
