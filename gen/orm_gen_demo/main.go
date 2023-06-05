package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed tpl.gohtml
var genOrm string

func gen(w io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, Data{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}

type Data struct {
	*File
	Ops []string
}

func main() {
	src := os.Args[1]
	dstDir := filepath.Dir(src)
	fileName := filepath.Base(src)
	idx := strings.LastIndexByte(fileName, '.')
	dst := filepath.Join(dstDir, fileName[:idx]+".gen.go")
	f, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gen(f, src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("生成成功")
}
