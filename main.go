package obfio_deobfuscator

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/transform/simplifier"
	"github.com/xkiian/obfio-deobfuscator/visitors"
)

func Deobfuscate(a *ast.Program) {
	simplifier.Simplify(a, true)
	visits := []func(p *ast.Program){
		visitors.ReplaceReassignments,
		visitors.ReplaceStrings,
		visitors.ConcatStrings,
		visitors.UnrollProxyFunctions,
	}

	for _, fn := range visits {
		fn(a)
	}
}

func DeobfuscateFromString(s *string) error {
	parsed, err := parser.ParseFile(*s)
	if err != nil {
		return err
	}

	Deobfuscate(parsed)
	return nil
}
