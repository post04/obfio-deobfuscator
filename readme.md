# WIP obfuscator.io deobfuscator

## Quick guide
``` 
go get github.com/xkiian/obfio-deobfuscator@main
```

---

## Usage

```go
package main

import (
	"os"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	deobf "github.com/xkiian/obfio-deobfuscator"
)

func main() {
	file, err := os.ReadFile("input.js")
	if err != nil {
		panic(err)
	}
	src := string(file)

	ast, err := parser.ParseFile(src)
	if err != nil {
		panic(err)
	}

	deobf.Deobfuscate(ast)

	os.WriteFile("output.js", []byte(generator.Generate(ast)), 0644)
}
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.