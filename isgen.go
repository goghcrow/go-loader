package go_loader

import (
	"go/ast"
	"strings"
)

// copy from go1.21 ast.IsGenerated

// IsGenerated reports whether the file was generated by a program,
// not handwritten, by detecting the special comment described
// at https://go.dev/s/generatedcode.
//
// The syntax tree must have been parsed with the ParseComments flag.
// Example:
//
//	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments|parser.PackageClauseOnly)
//	if err != nil { ... }
//	gen := ast.IsGenerated(f)
func IsGenerated(file *ast.File) bool {
	_, ok := Generator(file)
	return ok
}

func Generator(file *ast.File) (by string, is bool) {
	var (
		cutPrefix = func(s, prefix string) (after string, found bool) {
			if !strings.HasPrefix(s, prefix) {
				return s, false
			}
			return s[len(prefix):], true
		}
		cutSuffix = func(s, suffix string) (before string, found bool) {
			if !strings.HasSuffix(s, suffix) {
				return s, false
			}
			return s[:len(s)-len(suffix)], true
		}
	)

	for _, group := range file.Comments {
		for _, comment := range group.List {
			if comment.Pos() > file.Package {
				break // after package declaration
			}
			// opt: check Contains first to avoid unnecessary array allocation in Split.
			const prefix = "// Code generated "
			if strings.Contains(comment.Text, prefix) {
				for _, line := range strings.Split(comment.Text, "\n") {
					if rest, ok := cutPrefix(line, prefix); ok {
						if gen, ok := cutSuffix(rest, " DO NOT EDIT."); ok {
							return gen, true
						}
					}
				}
			}
		}
	}
	return "", false
}
