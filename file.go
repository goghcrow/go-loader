package loader

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type (
	Package = packages.Package
	File    struct {
		File     *ast.File
		Pkg      *Package
		Filename FileName
		GenBy    string
	}
)
