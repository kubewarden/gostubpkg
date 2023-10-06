package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
)

func formatType(typ interface{}, importedPackages []string) string {
	switch t := typ.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		// check if it is an allowed import
		if slices.Contains(importedPackages, t.X.(*ast.Ident).Name) {
			return fmt.Sprintf("%s.%s", formatType(t.X, importedPackages), t.Sel.Name)
		} else {
			return "interface{}"
		}
	case *ast.StarExpr:
		// do not add * to interface{}
		ft := formatType(t.X, importedPackages)
		if ft == "interface{}" {
			return ft
		}
		return fmt.Sprintf("*%s", ft)
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", formatType(t.Len, importedPackages), formatType(t.Elt, importedPackages))
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", formatType(t.Elt, importedPackages))
	case *ast.FuncType:
		return fmt.Sprintf("func(%s)%s", formatFields(t.Params, importedPackages), formatFuncResults(t.Results, importedPackages))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key, importedPackages), formatType(t.Value, importedPackages))
	case *ast.ChanType:
		s := "chan"
		if t.Arrow != token.NoPos {
			if t.Begin == token.NoPos {
				s = "<- chan"
			} else {
				s = "chan <-"
			}
		}
		return fmt.Sprintf("%s %s", s, formatType(t.Value, importedPackages))
	case *ast.BasicLit:
		return t.Value
	case *ast.InterfaceType:
		return "interface {}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "interface{}"
	}
}

func formatFields(fields *ast.FieldList, importedPackages []string) string {
	s := ""
	for i, field := range fields.List {
		for j, name := range field.Names {
			s += name.Name
			if j != len(field.Names)-1 {
				s += ","
			}
			s += " "
		}
		s += formatType(field.Type, importedPackages)
		if i != len(fields.List)-1 {
			s += ", "
		}
	}

	return s
}

// formatStructFields formats the fields of a struct.
func formatStructFields(fields *ast.FieldList, importedPackages []string) string {
	s := ""
	for i, field := range fields.List {
		for j, name := range field.Names {
			s += name.Name
			if j != len(field.Names)-1 {
				s += ";"
			}
			s += " "
		}
		ft := formatType(field.Type, importedPackages)

		// quick and dirty, if formatType returns an interface{} it means
		// that the field is an embedded struct of an external package
		// We replace it with Embedme to make the code compilable.
		if ft == "interface{}" && len(field.Names) == 0 {
			s += "Embedme"
		} else {
			s += ft
		}

		if i != len(fields.List)-1 {
			s += "; "
		}
	}

	return s
}

func formatFuncResults(fields *ast.FieldList, importedPackages []string) string {
	s := ""

	// Add brackets anyway. The formatter will remove them if not needed.
	// This is helpful to simplify the code in case of named return values.
	if fields != nil {
		s = fmt.Sprintf("(%s)", formatFields(fields, importedPackages))
	}

	return s
}

func formatFuncDecl(decl *ast.FuncDecl, importedPackages []string) string {
	s := "func "

	if decl.Recv != nil {
		if len(decl.Recv.List) != 1 {
			panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}

		field := decl.Recv.List[0]
		if len(field.Names) != 1 {
			panic(fmt.Errorf("strange receiver field for %s: %#v", decl.Name.Name, field))
		}
		s += fmt.Sprintf("(%s %s) ", field.Names[0], formatType(field.Type, importedPackages))
	}

	// generics
	g := ""
	if decl.Type.TypeParams != nil {
		g += "["
		for _, v := range decl.Type.TypeParams.List {
			g += fmt.Sprintf("%s %s,", v.Names[0].Name, formatType(v.Type, importedPackages))
		}
		g += "]"
	}

	s += fmt.Sprintf("%s%s(%s)", decl.Name.Name, g, formatFields(decl.Type.Params, importedPackages))
	s += formatFuncResults(decl.Type.Results, importedPackages)

	return s
}
