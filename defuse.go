package loader

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type (
	Def = *ast.Ident
	Use = *ast.Ident
)

func DefUses(l *Loader) map[Def][]Use {
	obj2uses := map[types.Object][]*ast.Ident{}
	def2Objs := map[*ast.Ident][]types.Object{} // test and non-test

	l.VisitAllPackages(nil, func(pkg *packages.Package) {
		info := pkg.TypesInfo
		for id, obj := range info.Uses {
			obj2uses[obj] = append(obj2uses[obj], id)
		}
		for id, obj := range info.Defs {
			def2Objs[id] = append(def2Objs[id], obj)
		}
	})

	useMap := map[Def][]Use{}
	for def, objs := range def2Objs {
		for _, obj := range objs {
			for _, use := range obj2uses[obj] {
				useMap[def] = append(useMap[def], use)
			}
		}
	}
	return useMap
}
