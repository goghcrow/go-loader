package loader

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type (
	Package    = packages.Package
	Positioner interface{ Pos() token.Pos }
)

type Pkg struct {
	*Package
}

func MkPkg(p *Package) Pkg {
	return Pkg{p}
}

func (p *Pkg) ShowPos(n Positioner) string {
	return p.Fset.Position(n.Pos()).String()
}

func (p *Pkg) ShowNode(n ast.Node) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, p.Fset, n)
	return buf.String()
}

func (p *Pkg) TypeInfo() *types.Info               { return p.TypesInfo }
func (p *Pkg) ObjectOf(id *ast.Ident) types.Object { return p.TypeInfo().ObjectOf(id) }
func (p *Pkg) TypeOf(e ast.Expr) types.Type        { return p.TypeInfo().TypeOf(e) }
func (p *Pkg) Callee(call *ast.CallExpr) types.Object {
	return typeutil.Callee(p.TypeInfo(), call)
}
func (p *Pkg) Fun(fun ast.Expr) types.Object { return p.Callee(&ast.CallExpr{Fun: fun}) }

func (p *Pkg) UpdateType(e ast.Expr, t types.Type) {
	p.TypeInfo().Types[e] = types.TypeAndValue{Type: t}
}

func (p *Pkg) UpdateUses(idOrSel ast.Expr, obj types.Object) {
	info := p.TypeInfo()
	switch x := idOrSel.(type) {
	case *ast.Ident:
		info.Uses[x] = obj
	case *ast.SelectorExpr:
		info.Uses[x.Sel] = obj
	default:
		panic("unreached")
	}
}

func (p *Pkg) UpdateDefs(idOrSel ast.Expr, obj types.Object) {
	info := p.TypeInfo()
	switch x := idOrSel.(type) {
	case *ast.Ident:
		info.Defs[x] = obj
	case *ast.SelectorExpr:
		info.Defs[x.Sel] = obj
	default:
		panic("unreached")
	}
}

func (p *Pkg) CopyTypeInfo(new, old ast.Expr) {
	info := p.TypeInfo()
	//goland:noinspection GoReservedWordUsedAsName
	switch new := new.(type) {
	case *ast.Ident:
		orig := old.(*ast.Ident)
		if obj, ok := info.Defs[orig]; ok {
			info.Defs[new] = obj
		}
		if obj, ok := info.Uses[orig]; ok {
			info.Uses[new] = obj
		}

	case *ast.SelectorExpr:
		orig := old.(*ast.SelectorExpr)
		if sel, ok := info.Selections[orig]; ok {
			info.Selections[new] = sel
		}
	}

	if tv, ok := info.Types[old]; ok {
		info.Types[new] = tv
	}
}

func (p *Pkg) NewIdent(name string, t types.Type) *ast.Ident {
	ident := ast.NewIdent(name)
	p.UpdateType(ident, t)

	obj := types.NewVar(token.NoPos, p.Types, name, t)
	p.UpdateUses(ident, obj)
	return ident
}
