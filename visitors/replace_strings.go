package visitors

import (
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/xkiian/obfio-deobfuscator/visitors/utils"
)

type stringReplacerGather struct {
	ast.NoopVisitor
	stopValue   float64
	ShuffleExpr *ast.Expression
	DecoderFunc *ast.FunctionLiteral
	offset      int
	stringArray []string
}

func (v *stringReplacerGather) VisitExpressionStatement(n *ast.ExpressionStatement) {
	n.VisitChildrenWith(v)

	cexpr, ok := n.Expression.Expr.(*ast.CallExpression)
	if !ok || len(cexpr.ArgumentList) != 2 {
		return
	}

	if _, ok = cexpr.ArgumentList[0].Expr.(*ast.Identifier); !ok {
		return
	}
	num, ok := cexpr.ArgumentList[1].Expr.(*ast.NumberLiteral)
	if !ok {
		return
	}
	v.stopValue = num.Value

	callee, ok := cexpr.Callee.Expr.(*ast.FunctionLiteral)
	if !ok {
		return
	}
	if len(callee.Body.List) < 2 {
		return
	}

	body, ok := callee.Body.List[1].Stmt.(*ast.WhileStatement)
	if !ok {
		return
	}

	block, ok := body.Body.Stmt.(*ast.BlockStatement)
	if !ok {
		return
	}
	if len(block.List) != 1 {
		return
	}

	try, ok := block.List[0].Stmt.(*ast.TryStatement)
	if !ok {
		return
	}
	if len(try.Body.List) != 2 {
		return
	}

	varDecl, ok := try.Body.List[0].Stmt.(*ast.VariableDeclaration)
	if !ok {
		return
	}

	_, ok = varDecl.List[0].Initializer.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	v.ShuffleExpr = varDecl.List[0].Initializer
	//fmt.Println(generator.Generate(varDeclor))
}

func (v *stringReplacerGather) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	n.VisitChildrenWith(v)
	code := generator.Generate(n)
	if !strings.Contains(code, "return decodeURIComponent") && v.DecoderFunc == nil {
		return
	}
	v.DecoderFunc = n

	if len(n.Body.List) != 2 {
		return
	}
	returnStmt, ok := n.Body.List[1].Stmt.(*ast.ReturnStatement)
	if !ok {
		return
	}

	seqExpr, ok := returnStmt.Argument.Expr.(*ast.SequenceExpression)
	if !ok {
		return
	}

	aExpr, ok := seqExpr.Sequence[0].Expr.(*ast.AssignExpression)
	if !ok {
		return
	}

	fLit, ok := aExpr.Right.Expr.(*ast.FunctionLiteral)
	if !ok {
		return
	}
	if len(fLit.Body.List) < 2 {
		return
	}

	ExprStmt, ok := fLit.Body.List[0].Stmt.(*ast.ExpressionStatement)
	if !ok {
		return
	}

	aExpr, ok = ExprStmt.Expression.Expr.(*ast.AssignExpression)
	if !ok {
		return
	}

	right, ok := aExpr.Right.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	val := int(right.Right.Expr.(*ast.NumberLiteral).Value)

	op := right.Operator.String()
	switch op {
	case "+":
		v.offset = val
	case "-":
		v.offset = -val
	default:
		panic("unsupported array offset | op: " + op)
	}

}

func (v *stringReplacerGather) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	n.VisitChildrenWith(v)
	if len(n.List) != 1 {
		return
	}
	if n.List[0].Initializer == nil {
		return
	}
	varDecl, ok := n.List[0].Initializer.Expr.(*ast.ArrayLiteral)
	if !ok {
		return
	}
	var values []string
	for _, val := range varDecl.Value {
		strLit, ok := val.Expr.(*ast.StringLiteral)
		if !ok {
			return
		}
		values = append(values, strLit.Value)
	}

	v.stringArray = values
}

type stringReplacer struct {
	ast.NoopVisitor
	gather  *stringReplacerGather
	decoder *utils.Rc4StringDecoder
}

func ReplaceStrings(p *ast.Program) {
	f := &stringReplacerGather{}
	f.V = f
	p.VisitWith(f)

	decoder := utils.NewRc4StringDecoder(f.stringArray, f.offset)

	utils.RotateStringArray(f.stringArray, f.ShuffleExpr, decoder, int(f.stopValue))

	/*f2 := &stringReplacer{
		gather:  f,
		decoder: decoder,
	}
	f2.V = f2
	p.VisitWith(f2)*/

}
