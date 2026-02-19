package group

import "github.com/expr-thoth/expr-thoth/core"

type PolicyGroup struct {
	Name        string
	policyFiles []*core.PolicyFile
}
