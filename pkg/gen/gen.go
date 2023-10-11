package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

func GenerateStubs(inputDir string, patterns []string, outputDir string, generateGoMod bool, allowImports []string, functionBodies map[string]string) error {
	if generateGoMod {
		log.Debugf("generating go.mod file")
		goModFile, err := os.ReadFile(filepath.Join(inputDir, "go.mod"))
		if err != nil {
			return err
		}

		goMod, err := modfile.Parse("go.mod", goModFile, nil)
		if err != nil {
			return err
		}

		genGoModPath := filepath.Join(outputDir, goMod.Module.Mod.Path)
		err = os.MkdirAll(genGoModPath, 0755)
		if err != nil {
			return err
		}

		genGoModFile, err := os.Create(filepath.Join(genGoModPath, "go.mod"))
		if err != nil {
			return err
		}

		_, err = genGoModFile.WriteString("module " + goMod.Module.Mod.Path + "\n\n")
		if err != nil {
			return err
		}

		_, err = genGoModFile.WriteString("go " + goMod.Go.Version + "\n")
		if err != nil {
			return err
		}
	}

	pkgs, err := loadPackages(inputDir, patterns)
	if err != nil {
		return err
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in %s", strings.Join(patterns, ", "))
	}

	for _, pkg := range pkgs {
		log.Debugf("generating stubs for package %s", pkg.PkgPath)

		err := os.MkdirAll(filepath.Join(outputDir, pkg.PkgPath), 0755)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)

		_, err = buf.WriteString("package " + pkg.Name + "\n\n")
		if err != nil {
			return err
		}
		// Get all the imports from the package and add it to the file
		// A the end we will programmatically use "goimports" on the generated file to fix the imports
		importedPackagesSet := make(map[string]struct{})
		for _, astFile := range pkg.Syntax {
			if ast.IsGenerated(astFile) {
				continue
			}

			for _, o := range astFile.Imports {
				if isThirdParty(o.Path.Value, allowImports) && !isLocalImport(o.Path.Value, pkgs) {
					continue
				}

				if o.Name != nil {
					if _, ok := importedPackagesSet[o.Name.Name]; ok {
						continue
					}

					_, err := buf.WriteString("import " + o.Name.Name + " " + o.Path.Value + "\n\n")
					if err != nil {
						return err
					}
					importedPackagesSet[o.Name.Name] = struct{}{}
				} else {
					name := o.Path.Value[strings.LastIndex(o.Path.Value, "/")+1:]
					name = strings.ReplaceAll(name, "\"", "")
					if _, ok := importedPackagesSet[name]; ok {
						continue
					}

					_, err := buf.WriteString("import " + o.Path.Value + "\n\n")
					if err != nil {
						return err
					}
					importedPackagesSet[name] = struct{}{}
				}
			}
		}

		importedPackages := []string{}
		for k := range importedPackagesSet {
			importedPackages = append(importedPackages, k)
		}

		for _, astFile := range pkg.Syntax {
			if ast.IsGenerated(astFile) {
				continue
			}

			err = stubConstsVars(astFile, buf, importedPackages)
			if err != nil {
				return err
			}

			err = stubTypes(astFile, buf, importedPackages)
			if err != nil {
				return err
			}

			err = stubFunctions(astFile, buf, pkg.Name, functionBodies, importedPackages)
			if err != nil {
				return err
			}

		}

		_, err = buf.WriteString("type Embedme interface{}\n\n")
		if err != nil {
			return (err)
		}

		// The file is created before since the imports.Process() function
		// requires to know the file path.
		outFile, err := os.Create(filepath.Join(outputDir, pkg.PkgPath, pkg.Name+".go"))
		if err != nil {
			return err
		}

		// Programmatically use "goimports"
		res, err := imports.Process(outFile.Name(), buf.Bytes(), nil)
		if err != nil {
			return err
		}

		_, err = outFile.Write(res)
		if err != nil {
			return err
		}
	}

	return nil
}

// isThirdParty checks if the given import path is a third party package. (no standard library)
func isThirdParty(importPath string, allowImports []string) bool {
	if slices.Contains(allowImports, strings.Replace(importPath, "\"", "", -1)) {
		return false
	}
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isLocalImport checks if the given import path is local to the given packages.
func isLocalImport(importPath string, pkgs []*packages.Package) bool {
	for _, pkg := range pkgs {
		if "\""+pkg.PkgPath+"\"" == importPath {
			return true
		}
	}
	return false
}

// loadPackages loads packages from patterns.
func loadPackages(inputDir string, patterns []string) ([]*packages.Package, error) {
	config := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedSyntax,
		Dir: inputDir,
	}

	return packages.Load(config, patterns...)
}

func stubConstsVars(astFile *ast.File, buf *bytes.Buffer, importedPackages []string) error {
	for _, xdecl := range astFile.Decls {
		decl, ok := xdecl.(*ast.GenDecl)
		if !ok {
			continue
		}

		t := ""
		if decl.Tok == token.CONST {
			t = "const"
		} else if decl.Tok == token.VAR {
			t = "var"
		} else {
			continue
		}

		for _, spec := range decl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range valueSpec.Names {
				log.Tracef("stubbing %s %s", t, name)

				v := fmt.Sprintf("%s %s", t, name)
				if len(valueSpec.Values) > 0 {
					value, ok := valueSpec.Values[0].(*ast.BasicLit)
					// TODO: handle other types
					if !ok {
						continue
					}
					v += fmt.Sprintf(" = %s", value.Value)
				} else {
					v += fmt.Sprintf(" %s", formatType(valueSpec.Type, importedPackages))
				}
				v += "\n\n"

				_, err := buf.WriteString(v)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func stubTypes(astFile *ast.File, buf *bytes.Buffer, importedPackages []string) error {
	for n, o := range astFile.Scope.Objects {
		// private types are create too
		// this is needed for private embedded types in structs
		node := o.Decl

		switch ts := node.(type) {
		case *ast.TypeSpec:
			switch t := ts.Type.(type) {
			case *ast.StructType:
				log.Tracef("stubbing struct %s", n)
				field := formatStructFields(t.Fields, importedPackages)
				_, err := buf.WriteString("type " + n + " struct " + "{" + field + "}\n\n")
				if err != nil {
					return err
				}
			case *ast.InterfaceType:
				log.Tracef("stubbing interface %s", n)
				i := "type " + n + " interface {\n"
				for _, method := range t.Methods.List {
					m, ok := method.Type.(*ast.FuncType)
					if !ok {
						// TODO: handle embedded interfaces
						log.Debugf("skipping embedded interface %s", method.Names[0].Name)
						continue
					}
					i += fmt.Sprintf("%s(%s) %s\n", method.Names[0].Name, formatFields(m.Params, importedPackages), formatFuncResults(m.Results, importedPackages))
				}
				i += "}\n\n"
				_, err := buf.WriteString(i)
				if err != nil {
					return err
				}

			default:
				log.Tracef("stubbing type %s", n)
				_, err := buf.WriteString("type " + n + " " + formatType(ts.Type, importedPackages) + "\n\n")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func stubFunctions(astFile *ast.File, buf *bytes.Buffer, pkgName string, functionsBodies map[string]string, importedPackages []string) error {
	for _, xdecl := range astFile.Decls {
		decl, ok := xdecl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if isInterfaceDecl(decl) {
			continue
		}

		if !ast.IsExported(decl.Name.Name) {
			continue
		}

		foo := formatFuncDecl(decl, importedPackages)

		// check if function body is provided
		recv := getRecvType(decl)
		if recv != "" {
			recv = fmt.Sprintf(".%s", recv)
		}

		key := fmt.Sprintf("%s%s.%s", pkgName, recv, decl.Name.Name)

		log.Tracef("stubbing function %s", key)
		if body, ok := functionsBodies[key]; ok {
			log.Tracef("using stub body for %s", key)
			foo += "{" + body + "\n}\n\n"
		} else {
			foo += " {\n panic(\"stub\")\n}\n\n"
		}

		_, err := buf.WriteString(foo)
		if err != nil {
			return err
		}
	}

	return nil
}

// check if it's an interface method declaration
func isInterfaceDecl(decl *ast.FuncDecl) bool {
	if decl.Recv != nil {
		if len(decl.Recv.List) != 1 {
			panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}

		field := decl.Recv.List[0]
		if len(field.Names) == 0 {
			return true
		}
	}
	return false
}

// getRecvType get the name of a method receiver
// Examples:
// func (s *Struct) Foo() {} -> (*Struct)
// func (s Struct) Foo() {}  -> (Struct)
func getRecvType(decl *ast.FuncDecl) string {
	if decl.Recv == nil {
		return ""
	}

	if len(decl.Recv.List) != 1 {
		panic(fmt.Errorf("multiple receivers for %s: %#v", decl.Name.Name, decl.Recv))
	}

	field := decl.Recv.List[0]

	switch t := field.Type.(type) {
	case *ast.Ident:
		return fmt.Sprintf(".(%s)", t.Name)
	case *ast.StarExpr:
		switch xType := t.X.(type) {
		case *ast.Ident:
			return fmt.Sprintf("(*%s)", xType.Name)
		default:
			// not an identificator?
			return ""
		}
	default:
		// some new syntax?
		return ""
	}
}
