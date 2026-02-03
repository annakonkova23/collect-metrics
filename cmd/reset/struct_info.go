package main

import (
	"fmt"
	"go/types"
	"strings"
)

func extractStructInfo(name string, obj types.Object, currentPackagePath string) StructInfo {
	recv := strings.ToLower(string(name[0]))
	structType := obj.Type().Underlying().(*types.Struct)

	var fields []Field
	info := make(map[string]string)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		ft := field.Type()

		underlyingFT := deref(ft)

		isSlice := false
		if _, ok := underlyingFT.(*types.Slice); ok {
			isSlice = true
		}

		isMap := false
		if _, ok := underlyingFT.(*types.Map); ok {
			isMap = true
		}

		hasReset := hasResetMethod(ft)

		isPointer := ft != underlyingFT

		zero := ""
		pkgAlias := ""

		named := findNamedType(ft)
		if named != nil {
			pkg := named.Obj().Pkg()
			if pkg != nil {
				importPath := pkg.Path()
				if importPath != currentPackagePath {
					pkgAlias = pkg.Name()
					if !isPointer || (isPointer && hasReset) {
						info[importPath] = pkgAlias
					}
				}
			}
		}

		zero = ""
		if !hasReset {
			if isSlice {
			} else if isMap {
			} else if isPointer {
				zero = "nil"
			} else {
				zero = getZeroValue(underlyingFT, pkgAlias)
			}
		}

		fields = append(fields, Field{
			Name:      field.Name(),
			Type:      ft,
			Zero:      zero,
			IsSlice:   isSlice,
			HasReset:  hasReset,
			IsPointer: isPointer,
			IsMap:     isMap,
		})
	}

	return StructInfo{
		Name:    name,
		Fields:  fields,
		Recv:    recv,
		Imports: info,
	}
}

func getZeroValue(t types.Type, pkgAlias string) string {
	underlying := t.Underlying()

	switch u := underlying.(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.Bool:
			return "false"
		case types.String:
			return `""`
		default:
			return "0"
		}
	case *types.Interface, *types.Slice, *types.Map, *types.Chan:
		return "nil"
	case *types.Struct:
		if pkgAlias == "" {
			return fmt.Sprintf("%s{}", t.(*types.Named).Obj().Name())
		}
		return fmt.Sprintf("%s.%s{}", pkgAlias, t.(*types.Named).Obj().Name())
	case *types.Array:
		return fmt.Sprintf("%s{}", t.String())
	default:
		return "0"
	}
}

func findNamedType(t types.Type) *types.Named {
	for {
		switch typ := t.(type) {
		case *types.Named:
			return typ
		case *types.Pointer:
			t = typ.Elem()
		case *types.Slice:
			t = typ.Elem()
		case *types.Array:
			t = typ.Elem()
		case *types.Map:
			t = typ.Elem()
		default:
			return nil
		}
	}
}

func hasResetMethod(t types.Type) bool {

	ut := deref(t)

	named, ok := ut.(*types.Named)
	if !ok {
		return false
	}

	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		if method.Name() == "Reset" {

			sig, ok := method.Type().(*types.Signature)
			if !ok {
				continue
			}

			if sig.Params() == nil && sig.Results() == nil {
				return true
			}
		}
	}
	return false
}

func deref(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}
