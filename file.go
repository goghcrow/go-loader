package loader

import (
	"go/ast"
)

type File struct {
	File     *ast.File
	Pkg      *Package
	Filename FileName
	GenBy    string
}

func (f *File) Package() Pkg {
	return MkPkg(f.Pkg)
}
