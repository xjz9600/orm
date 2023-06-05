package ast

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestPrintVisitor(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", `
package main
 
type Article struct{}
 
// @GET("/get")
func (a Article) Get() {
 
}
 
// @POST("/save")
func (a Article) Save() {
 
}`, parser.ParseComments)
	require.NoError(t, err)
	v := &PrintVisitor{}
	ast.Walk(v, f)

}
