package go_loader

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

func (l *Loader) LookupPackage(pkg string) *packages.Package {
	return l.All[pkg]
}

// Lookup builtin | qualified ident
// e.g. "error", "string", "encoding/json.Marshal"
func (l *Loader) Lookup(qualifiedName string) types.Object {
	idx := strings.LastIndex(qualifiedName, ".")
	if idx == -1 {
		return types.Universe.Lookup(qualifiedName)
	}
	pkg := qualifiedName[:idx]
	id := qualifiedName[idx+1:]
	p := l.LookupPackage(pkg)
	if p == nil {
		return nil
	}
	return p.Types.Scope().Lookup(id)
}

func (l *Loader) MustLookup(qualifiedName string) types.Object {
	obj := l.Lookup(qualifiedName)
	assert(obj != nil, "object not found: "+qualifiedName)
	return obj
}

func (l *Loader) MustLookupType(qualified string) types.Type {
	obj := l.Lookup(qualified)
	assert(obj != nil, "type not found: "+qualified)
	return obj.Type()
}

// Lookups name in current package and all imported packages
func (l *Loader) Lookups(pkg *types.Package, name string) []types.Object {
	if o := pkg.Scope().Lookup(name); o != nil {
		return []types.Object{o}
	}

	var ret []types.Object
	for _, imp := range pkg.Imports() {
		if obj := imp.Scope().Lookup(name); obj != nil {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (l *Loader) MustLookupFieldOrMethod(pkg, typ, fieldOrMethod string) types.Object {
	qualified := pkg + "." + typ
	tyObj := l.Lookup(qualified)
	assert(tyObj != nil, qualified+" not found")

	p := l.LookupPackage(pkg)
	assert(p != nil, pkg+" not found")
	obj, _, indirect := types.LookupFieldOrMethod(tyObj.Type(), false, p.Types, fieldOrMethod)
	if obj == nil && indirect {
		obj, _, _ = types.LookupFieldOrMethod(tyObj.Type(), true, p.Types, fieldOrMethod)
	}
	return obj
}
