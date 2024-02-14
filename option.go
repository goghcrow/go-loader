package loader

import (
	"golang.org/x/tools/go/packages"
)

type (
	Option  func(*Flags)
	Pattern = string
)

type Flags struct {
	Patterns    []Pattern // go list patterns
	Gopath      string
	BuildTag    string // comma-separated list of extra build tags (see: go help buildconstraint)
	Test        bool   // load includes tests packages
	LoadDepts   bool   // load all dependencies, may heavily slow
	PrintErrors bool

	PkgFilter  func(*Package) bool
	FileFilter func(*File) bool
}

const (
	PatternAll Pattern = "./..."
	PatternStd Pattern = "std"
)

// WithPatterns go list patterns
// e.g.,
// "file=path/to/file.go"
// all "./...",
// "std"
// "bytes", "unicode..."
// "pattern=github.com/goghcrow/go-co/..."
// "github.com/goghcrow/go-co/..."
func WithPatterns(patterns ...Pattern) Option { return func(opts *Flags) { opts.Patterns = patterns } }

// WithLoadDepts is slow, but you can match external types
func WithLoadDepts() Option { return func(opts *Flags) { opts.LoadDepts = true } }

// WithBuildTag comma-separated list of extra build tags (see: go help buildconstraint)
func WithBuildTag(tag string) Option { return func(opts *Flags) { opts.BuildTag = tag } }

func WithGopath(gopath string) Option { return func(opts *Flags) { opts.Gopath = gopath } }

func WithLoadTest() Option { return func(opts *Flags) { opts.Test = true } }

func WithSuppressErrors() Option { return func(opts *Flags) { opts.PrintErrors = false } }

func WithPkgFilter(f func(pkg *packages.Package) bool) Option {
	return func(opts *Flags) {
		pre := opts.PkgFilter
		if pre == nil {
			pre = func(*packages.Package) bool { return true }
		}
		opts.PkgFilter = func(pkg *packages.Package) bool {
			return pre(pkg) && f(pkg)
		}
	}
}

func WithFileFilter(f func(file *File) bool) Option {
	return func(opts *Flags) {
		pre := opts.FileFilter
		if pre == nil {
			pre = func(*File) bool { return true }
		}
		opts.FileFilter = func(file *File) bool {
			return pre(file) && f(file)
		}
	}
}

func WithSkipGenerated() Option {
	return WithFileFilter(func(file *File) bool {
		return file.GenBy == ""
	})
}

func skipTestMain() Option {
	return WithFileFilter(func(file *File) bool {
		return file.GenBy != "by 'go test'."
	})
}
