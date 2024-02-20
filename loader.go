package loader

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type (
	PackagePath = string
	FileName    = string
	GenBy       = string

	Loader struct {
		Flags *Flags
		Cfg   *packages.Config
		FSet  *token.FileSet
		Init  []*packages.Package
		All   map[PackagePath]*packages.Package
		Gen   map[FileName]GenBy
	}
)

func MustNew(dir string, opts ...Option) *Loader {
	l, err := New(dir, opts...)
	panicIf(err)
	return l
}

func New(dir string, opts ...Option) (l *Loader, err error) {
	flags := &Flags{}
	flags.PrintErrors = true
	flags.Patterns = []string{PatternAll} // default all
	for _, opt := range append(opts, skipTestMain()) {
		opt(flags)
	}

	l = &Loader{
		Flags: flags,
		FSet:  token.NewFileSet(),
		Init:  []*packages.Package{},
		All:   map[PackagePath]*packages.Package{},
		Gen:   map[FileName]GenBy{},
	}
	err = l.load(dir)
	return
}

// dir: run the build system's query tool
func (l *Loader) load(dir string) (err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	l.Cfg = l.mkLoadCfg(dir)

	l.Init, err = packages.Load(l.Cfg, l.Flags.Patterns...)
	if err != nil {
		return err
	}

	if len(l.Init) == 0 {
		errLog("no packages found")
	}

	l.loadAllPkg()
	return nil
}

func (l *Loader) loadAllPkg() {
	packages.Visit(l.Init, nil, func(p *packages.Package) {
		if l.Flags.PrintErrors {
			for _, err := range p.Errors {
				errLog(err)
			}
		}

		// Gather all packages
		l.All[p.PkgPath] = p

		// Gather generated files
		for _, file := range p.Syntax {
			if gen, is := Generator(file); is {
				f := p.Fset.File(file.Pos())
				l.Gen[f.Name()] = gen
			}
		}
	})
}

func (l *Loader) mkLoadCfg(dir string) *packages.Config {
	mode := loadMode
	if l.Flags.LoadDepts {
		mode |= loadDepts
	}

	cfg := &packages.Config{
		Fset:       l.FSet,
		Mode:       mode,
		Tests:      l.Flags.Test,
		Dir:        dir,
		BuildFlags: []string{"-tags=" + l.Flags.BuildTag},
	}
	if l.Flags.Gopath != "" {
		cfg.Env = append(os.Environ(), "GOPATH="+l.Flags.Gopath)
	}

	return cfg
}

// VisitAllFiles Walk all files in all init-pkgs (defined by load pattern)
func (l *Loader) VisitAllFiles(f func(file *File)) {
	for _, pkg := range l.Init {
		if l.Flags.PkgFilter == nil || l.Flags.PkgFilter(pkg) {
			l.visitPkgFiles(pkg, f)
		}
	}
}

// VisitAllPackages in topological order
// walk files order by compiledGoFiles in same package
// walk imports order by declaration in one file
// deep first
func (l *Loader) VisitAllPackages(
	pre func(*packages.Package) bool,
	post func(*packages.Package),
) {
	seen := map[*packages.Package]bool{}
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if seen[pkg] {
			return
		}
		seen[pkg] = true

		if pre == nil || pre(pkg) {
			l.visitPkgFiles(pkg, func(file *File) {
				l.visitFileImports(file, func(name *ast.Ident, path string) {
					impt := l.All[path]
					if impt != nil {
						visit(impt)
					}
				})
			})
		}
		if post != nil {
			post(pkg)
		}
	}

	for _, pkg := range l.Init {
		if l.Flags.PkgFilter == nil || l.Flags.PkgFilter(pkg) {
			visit(pkg)
		}
	}
}

func (l *Loader) visitPkgFiles(pkg *packages.Package, f func(file *File)) {
	for i, filename := range pkg.CompiledGoFiles {
		file := &File{
			File:     pkg.Syntax[i],
			Pkg:      pkg,
			Filename: filename,
			GenBy:    l.Gen[filename],
		}
		if l.Flags.FileFilter == nil || l.Flags.FileFilter(file) {
			f(file)
		}
	}
}

func (l *Loader) visitFileImports(file *File, yield func(name *ast.Ident, path string)) {
	for _, decl := range file.File.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.IMPORT {
			continue
		}
		for _, spec := range d.Specs {
			s := spec.(*ast.ImportSpec)
			name := s.Name
			path := parseImportPath(s)
			yield(name, path)
		}
	}
}

func (l *Loader) ShowPos(n ast.Node) string {
	return l.FSet.Position(n.Pos()).String()
}

func (l *Loader) ShowNode(n ast.Node) string {
	// if IsPseudoNode(n) {
	// 	return showPseudoNode(fset, n)
	// }
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, l.FSet, n)
	return buf.String()
}

func parseImportPath(s *ast.ImportSpec) string {
	path := s.Path.Value
	xs := strings.Split(path, " ")
	if len(xs) > 1 {
		path = xs[1]
	}
	t, _ := strconv.Unquote(path) // ignored illegal path
	return t
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Modes ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

const (
	// loadDepts load all dependencies
	loadDepts = packages.NeedImports | packages.NeedDeps
	loadMode  = packages.NeedTypesInfo |
		packages.NeedName |
		packages.NeedFiles |
		packages.NeedExportFile |
		packages.NeedCompiledGoFiles |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes |
		packages.NeedModule
)
