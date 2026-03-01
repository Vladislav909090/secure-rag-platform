package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	var serviceDir string
	var outRel string
	flag.StringVar(&serviceDir, "service", ".", "service module directory (contains go.mod)")
	flag.StringVar(&outRel, "out", "internal/transport/grpc", "output directory relative to service dir")
	flag.Parse()

	serviceDir, err := filepath.Abs(serviceDir)
	if err != nil {
		die(err)
	}

	modulePath, err := readModulePath(filepath.Join(serviceDir, "go.mod"))
	if err != nil {
		die(err)
	}

	grpcFiles, err := findGRPCPBFiles(filepath.Join(serviceDir, "gen"))
	if err != nil {
		die(err)
	}
	if len(grpcFiles) == 0 {
		die(errors.New("no *_grpc.pb.go files found under gen/; run proto generation first"))
	}

	services, err := parseServerInterfaces(grpcFiles)
	if err != nil {
		die(err)
	}
	if len(services) == 0 {
		die(errors.New("no *Server interfaces found in *_grpc.pb.go"))
	}

	outDir := filepath.Join(serviceDir, filepath.FromSlash(outRel))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		die(err)
	}

	pbImport := modulePath + "/gen/v1"
	for _, svc := range services {
		// Legacy (single-file) output name from previous generator version.
		legacyName := snakeCase(svc.InterfaceName) + ".gen.go"
		legacyPath := filepath.Join(outDir, legacyName)
		if _, err := os.Stat(legacyPath); err == nil {
			rel, _ := filepath.Rel(serviceDir, legacyPath)
			fmt.Printf("skip %s (legacy combined stub exists; delete to regenerate split stubs)\n", filepath.ToSlash(rel))
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			die(err)
		}

		// Base impl file.
		baseName := snakeCase(svc.InterfaceName) + "_impl.go"
		basePath := filepath.Join(outDir, baseName)
		if _, err := os.Stat(basePath); err == nil {
			rel, _ := filepath.Rel(serviceDir, basePath)
			fmt.Printf("skip %s (already exists)\n", filepath.ToSlash(rel))
		} else if errors.Is(err, os.ErrNotExist) {
			// If legacy base file exists, do not generate a new one to avoid duplicate symbols.
			legacyBasePath := filepath.Join(outDir, snakeCase(svc.InterfaceName)+"_impl.gen.go")
			if _, legacyErr := os.Stat(legacyBasePath); legacyErr == nil {
				rel, _ := filepath.Rel(serviceDir, legacyBasePath)
				fmt.Printf("skip %s (legacy stub exists; delete to regenerate without .gen)\n", filepath.ToSlash(rel))
			} else if !errors.Is(legacyErr, os.ErrNotExist) {
				die(legacyErr)
			} else {
				content, err := renderServiceBase("grpc", pbImport, svc)
				if err != nil {
					die(err)
				}
				if err := os.WriteFile(basePath, content, 0o644); err != nil {
					die(err)
				}
				rel, _ := filepath.Rel(serviceDir, basePath)
				fmt.Printf("wrote %s\n", filepath.ToSlash(rel))
			}
		} else {
			die(err)
		}

		for _, m := range svc.Methods {
			methodName := snakeCase(m.Name) + ".go"
			methodPath := filepath.Join(outDir, methodName)
			if _, err := os.Stat(methodPath); err == nil {
				rel, _ := filepath.Rel(serviceDir, methodPath)
				fmt.Printf("skip %s (already exists)\n", filepath.ToSlash(rel))
				continue
			} else if !errors.Is(err, os.ErrNotExist) {
				die(err)
			}

			// If legacy per-method file exists, do not generate a new one to avoid duplicate symbols.
			legacyMethodPath := filepath.Join(outDir, snakeCase(m.Name)+".gen.go")
			if _, legacyErr := os.Stat(legacyMethodPath); legacyErr == nil {
				rel, _ := filepath.Rel(serviceDir, legacyMethodPath)
				fmt.Printf("skip %s (legacy stub exists; delete to regenerate without .gen)\n", filepath.ToSlash(rel))
				continue
			} else if !errors.Is(legacyErr, os.ErrNotExist) {
				die(legacyErr)
			}

			content, err := renderServiceMethod("grpc", pbImport, svc, m)
			if err != nil {
				die(err)
			}
			if err := os.WriteFile(methodPath, content, 0o644); err != nil {
				die(err)
			}
			rel, _ := filepath.Rel(serviceDir, methodPath)
			fmt.Printf("wrote %s\n", filepath.ToSlash(rel))
		}
	}
}

type serviceIface struct {
	InterfaceName string
	Methods       []ifaceMethod
}

type ifaceMethod struct {
	Name    string
	Params  []param
	Results []ast.Expr
}

type param struct {
	Name string
	Type ast.Expr
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "grpcstubgen:", err)
	os.Exit(1)
}

func readModulePath(goModPath string) (string, error) {
	b, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "module ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
		}
	}
	return "", errors.New("unable to find module path in go.mod")
}

func findGRPCPBFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_grpc.pb.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func parseServerInterfaces(files []string) ([]serviceIface, error) {
	fset := token.NewFileSet()
	var out []serviceIface

	for _, filePath := range files {
		f, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				iface, ok := ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				name := ts.Name.Name
				if !strings.HasSuffix(name, "Server") {
					continue
				}
				if strings.HasPrefix(name, "Unsafe") {
					continue
				}

				methods, err := extractIfaceMethods(iface)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", filePath, err)
				}
				if len(methods) == 0 {
					continue
				}
				out = append(out, serviceIface{InterfaceName: name, Methods: methods})
			}
		}
	}

	seen := map[string]serviceIface{}
	for _, svc := range out {
		seen[svc.InterfaceName] = svc
	}
	out = out[:0]
	for _, svc := range seen {
		out = append(out, svc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].InterfaceName < out[j].InterfaceName })
	return out, nil
}

func extractIfaceMethods(iface *ast.InterfaceType) ([]ifaceMethod, error) {
	if iface.Methods == nil {
		return nil, nil
	}
	var methods []ifaceMethod
	for _, field := range iface.Methods.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		if strings.HasPrefix(name, "mustEmbedUnimplemented") {
			continue
		}
		ft, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		params, err := buildParams(ft.Params)
		if err != nil {
			return nil, err
		}
		results := []ast.Expr{}
		if ft.Results != nil {
			for _, r := range ft.Results.List {
				results = append(results, r.Type)
			}
		}
		methods = append(methods, ifaceMethod{Name: name, Params: params, Results: results})
	}
	return methods, nil
}

func buildParams(fl *ast.FieldList) ([]param, error) {
	if fl == nil {
		return nil, nil
	}
	var params []param
	idx := 0
	for _, f := range fl.List {
		names := f.Names
		if len(names) == 0 {
			params = append(params, param{Name: defaultParamName(f.Type, idx), Type: f.Type})
			idx++
			continue
		}
		for _, n := range names {
			params = append(params, param{Name: n.Name, Type: f.Type})
			idx++
		}
	}
	return params, nil
}

func defaultParamName(t ast.Expr, idx int) string {
	if isContextContext(t) {
		return "ctx"
	}
	if idx == 1 {
		return "req"
	}
	return fmt.Sprintf("arg%d", idx+1)
}

func isContextContext(t ast.Expr) bool {
	se, ok := t.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if se.Sel == nil || se.Sel.Name != "Context" {
		return false
	}
	ident, ok := se.X.(*ast.Ident)
	return ok && ident.Name == "context"
}

func renderServiceBase(pkgName, pbImport string, svc serviceIface) ([]byte, error) {
	implName := svc.InterfaceName + "Impl"
	unimplementedName := "Unimplemented" + svc.InterfaceName

	src := fmt.Sprintf(`// Code generated by grpcstubgen. DO NOT EDIT.

package %s

import (
	pb "%s"
)

type %s struct {
	pb.%s
}

var _ pb.%s = (*%s)(nil)
`, pkgName, pbImport, implName, unimplementedName, svc.InterfaceName, implName)

	return format.Source([]byte(src))
}

func renderServiceMethod(pkgName, pbImport string, svc serviceIface, m ifaceMethod) ([]byte, error) {
	implName := svc.InterfaceName + "Impl"

	var b strings.Builder
	b.WriteString("// Code generated by grpcstubgen. DO NOT EDIT.\n\n")
	b.WriteString("package ")
	b.WriteString(pkgName)
	b.WriteString("\n\n")

	b.WriteString("import (\n")
	b.WriteString("\t\"context\"\n")
	b.WriteString("\tpb \"")
	b.WriteString(pbImport)
	b.WriteString("\"\n")
	b.WriteString("\t\"google.golang.org/grpc/codes\"\n")
	b.WriteString("\t\"google.golang.org/grpc/status\"\n")
	b.WriteString(")\n\n")

	b.WriteString("func (s *")
	b.WriteString(implName)
	b.WriteString(") ")
	b.WriteString(m.Name)
	b.WriteString("(")
	for i, p := range m.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.Name)
		b.WriteString(" ")
		b.WriteString(typeString(p.Type))
	}
	b.WriteString(")")

	if len(m.Results) == 1 {
		b.WriteString(" ")
		b.WriteString(typeString(m.Results[0]))
	} else if len(m.Results) > 1 {
		b.WriteString(" (")
		for i, r := range m.Results {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(typeString(r))
		}
		b.WriteString(")")
	}
	b.WriteString(" {\n")
	b.WriteString("\t")
	b.WriteString(renderUnimplementedReturn(m.Name, m.Results))
	b.WriteString("\n")
	b.WriteString("}\n")

	return format.Source([]byte(b.String()))
}

func renderUnimplementedReturn(methodName string, results []ast.Expr) string {
	errExpr := fmt.Sprintf("status.Error(codes.Unimplemented, %q)", "method "+methodName+" not implemented")
	if len(results) == 0 {
		return "panic(" + fmt.Sprintf("%q", "method "+methodName+" not implemented") + ")"
	}
	if len(results) == 1 {
		return "return " + errExpr
	}
	parts := make([]string, 0, len(results))
	for i := 0; i < len(results)-1; i++ {
		parts = append(parts, zeroValue(results[i]))
	}
	parts = append(parts, errExpr)
	return "return " + strings.Join(parts, ", ")
}

func typeString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		if shouldQualifyIdent(t.Name) {
			return "pb." + t.Name
		}
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + typeString(t.Elt)
		}
		return "[" + typeString(t.Len) + "]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		var buf bytes.Buffer
		_ = printer.Fprint(&buf, token.NewFileSet(), e)
		return buf.String()
	}
}

func zeroValue(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "\"\""
		case "bool":
			return "false"
		case "error":
			return "nil"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128":
			return "0"
		default:
			if shouldQualifyIdent(t.Name) {
				return "pb." + t.Name + "{}"
			}
			return t.Name + "{}"
		}
	case *ast.StarExpr, *ast.MapType, *ast.ArrayType, *ast.InterfaceType, *ast.FuncType, *ast.ChanType:
		return "nil"
	case *ast.SelectorExpr:
		return typeString(t) + "{}"
	default:
		return "nil"
	}
}

func shouldQualifyIdent(name string) bool {
	if name == "" || name == "error" {
		return false
	}
	switch name {
	case "any", "bool", "byte", "comparable", "complex64", "complex128", "float32", "float64", "int", "int8", "int16", "int32", "int64", "rune", "string", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return false
	}
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func snakeCase(s string) string {
	var out []rune
	var prevLower bool
	for i, r := range s {
		isUpper := r >= 'A' && r <= 'Z'
		isLower := r >= 'a' && r <= 'z'
		if i > 0 && isUpper && prevLower {
			out = append(out, '_')
		}
		out = append(out, rune(strings.ToLower(string(r))[0]))
		prevLower = isLower || (r >= '0' && r <= '9')
	}
	return string(out)
}
