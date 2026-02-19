package statement_test

import (
	"testing"

	"github.com/expr-lang/expr"
)

func TestCompile(t *testing.T) {
	c, err := expr.Compile("Req.Action in ['a']", expr.AllowUndefinedVariables())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(c.Disassemble())
}
