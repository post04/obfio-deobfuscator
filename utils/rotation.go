package utils

/*
inspired: https://github.com/ben-sb/obfuscator-io-deobfuscator/blob/443e8855bc75749aa49e59de6f14d571f6f07508/src/deobfuscator/helpers/strings/rotation/rotation.ts

will fix if i actually have to (i've made a "simpler" but cheaty version)
*/
import (
	"fmt"
	"strconv"
	"time"

	"github.com/t14raptor/go-fast/ast"
)

type Operation interface {
	isOperation()
}

func (BinaryOperation) isOperation() {}
func (UnaryOperation) isOperation()  {}
func (NumberOperation) isOperation() {}
func (CallOperation) isOperation()   {}

type NumberOperation struct {
	value int64
}
type BinaryOperation struct {
	operator string
	left     Operation
	right    Operation
}

type UnaryOperation struct {
	operator string
	argument Operation
}

type CallOperation struct {
	args []any
}

func parseOperation(expr *ast.Expression) Operation {
	switch expr.Expr.(type) {
	case *ast.CallExpression:
		return parseCallOperation(expr.Expr.(*ast.CallExpression))
	case *ast.UnaryExpression:
		return parseUnaryOperation(expr.Expr.(*ast.UnaryExpression))
	case *ast.BinaryExpression:
		return parseBinaryOperation(expr.Expr.(*ast.BinaryExpression))
	case *ast.NumberLiteral:
		return &NumberOperation{value: int64(expr.Expr.(*ast.NumberLiteral).Value)}
	}
	return nil
}

func parseCallOperation(expr *ast.CallExpression) *CallOperation {
	if expr.Callee.Expr.(*ast.Identifier).Name != "parseInt" || len(expr.ArgumentList) != 1 {
		panic("callExpr is not parseInt")
	}

	stringCall := expr.ArgumentList[0].Expr.(*ast.CallExpression)

	var args []any
	for _, arg := range stringCall.ArgumentList {
		switch arg.Expr.(type) {
		case *ast.NumberLiteral:
			args = append(args, arg.Expr.(*ast.NumberLiteral).Value)
		case *ast.StringLiteral:
			args = append(args, arg.Expr.(*ast.StringLiteral).Value)
		default:
			panic("invalid arg on rotation")
		}
	}
	return &CallOperation{
		args,
	}
}

func parseUnaryOperation(expr *ast.UnaryExpression) *UnaryOperation {
	argument := parseOperation(expr.Operand)
	return &UnaryOperation{
		expr.Operator.String(), argument,
	}
}

func parseBinaryOperation(expr *ast.BinaryExpression) *BinaryOperation {
	left := parseOperation(expr.Left)
	right := parseOperation(expr.Right)
	return &BinaryOperation{
		expr.Operator.String(), left, right,
	}
}

func applyOperation(op Operation, decoder *Rc4StringDecoder) (any, bool) {
	//godump.Dump(op)
	switch op.(type) {
	case *CallOperation:
		//fmt.Println("CallOperation")
		cOp := op.(*CallOperation)
		if len(cOp.args) != 2 {
			panic("invalid call operation length")
		}
		var val int
		switch cOp.args[0].(type) {
		case string:
			firstVal, err := strconv.ParseInt(cOp.args[0].(string), 10, 32)
			if err != nil {
				panic(err)
			}
			val = int(firstVal)
		case int:
			val = cOp.args[0].(int)
		case float64:
			val = int(cOp.args[0].(float64))
		}
		//fmt.Println(cOp.args[1].(string))
		res, errd := decoder.GetForRotate(val, cOp.args[1].(string))
		if errd {
			return nil, true
		}
		resInt, errr := strconv.ParseInt(res, 10, 64)
		if errr != nil {
			return nil, true
		}
		return resInt, false
	case *UnaryOperation:
		//fmt.Println("UnaryOperation")
		arg, err := applyOperation(op.(*UnaryOperation).argument, decoder)
		if err {
			return nil, true
		}
		switch op.(*UnaryOperation).operator {
		case "+":
			return arg, false
		case "-":
			switch v := arg.(type) {
			case int64:
				return -v, false
			case int:
				return -int64(v), false
			default:
				return nil, true
			}
		}
	case *BinaryOperation:
		leftVal, errd := applyOperation(op.(*BinaryOperation).left, decoder)
		if errd {
			return nil, true
		}
		rightVal, errd := applyOperation(op.(*BinaryOperation).right, decoder)
		if errd {
			return nil, true
		}

		var left, right int64
		switch v := leftVal.(type) {
		case int64:
			left = v
		case int:
			left = int64(v)
		default:
			return nil, true
		}
		switch v := rightVal.(type) {
		case int64:
			right = v
		case int:
			right = int64(v)
		default:
			return nil, true
		}

		switch op.(*BinaryOperation).operator {
		case "+":
			return left + right, false
		case "-":
			return left - right, false
		case "*":
			return left * right, false
		case "/":
			return left / right, false
		case "%":
			return left % right, false
		}
	case *NumberOperation:
		//fmt.Println("NumberOperation")
		return op.(*NumberOperation).value, false
	}
	panic("invalid operation2")
	return nil, true
}

func RotateStringArray(
	array []string,
	expression *ast.Expression,
	decoder *Rc4StringDecoder,
	stopValue int,
) {
	operation := parseOperation(expression)
	i := 0
	for {
		v, errd := applyOperation(operation, decoder)
		fmt.Println(v, errd)
		time.Sleep(100 * time.Millisecond)
		if errd {
			array = append(array[1:], array[0])
			decoder.StringArray = array
		} else {
			if vi, ok := v.(int64); ok && vi == int64(stopValue) {
				break
			}
			array = append(array[1:], array[0])
			decoder.StringArray = array
		}

		i++
		if i > 1e5 {
			panic("invalid rotation")
		}
	}
}
