package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			node, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if ok := isLinted(pass, sel); !ok {
				return true
			}
			if len(node.Args) == 0 {
				return true
			}

			checkStartsWithUpper(pass, node)

			checkNotAllowedSymbols(pass, node)

			return true
		})
	}
	return nil, nil
}

// Определяет родительский пакет логгера, а также вызванный у него метод.
// На основе этого принимается решение, линтить ли вызов или нет.
//
// Текущие логгеры подлежащие линту:
//   - log/slog
//   - go.uber.org/zap
//
// Методы:
//   - Info
//   - Debug
//   - Warn
//   - Error
//   - Fatal
func isLinted(pass *analysis.Pass, expr *ast.SelectorExpr) bool {
	switch expr.Sel.Name {
	case "Info", "Debug", "Warn", "Error", "Fatal":
	default:
		return false
	}

	pkgPath := getPackagePath(pass, expr.X)

	return pkgPath == "log/slog" || pkgPath == "go.uber.org/zap"
}

// getPackagePath возвращает путь к пакету, в котором определен логгер.
func getPackagePath(pass *analysis.Pass, expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return ""
	}

	if pkgName, ok := obj.(*types.PkgName); ok {
		return pkgName.Imported().Path()
	}

	typ := obj.Type()
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}
	if named, ok := typ.(*types.Named); ok {
		if pkg := named.Obj().Pkg(); pkg != nil {
			return pkg.Path()
		}
	}

	return ""
}

// checkStartsWithUpper проверяет что лог-сообещние не начинается
// с заглавной буквы.
func checkStartsWithUpper(pass *analysis.Pass, expr *ast.CallExpr) {
	lit, ok := getStringLiteral(expr)
	if !ok {
		return
	}

	r, _ := utf8.DecodeRuneInString(lit)
	if unicode.IsUpper(r) {
		pass.Reportf(expr.Args[0].Pos(), "log messages must start with lowercase letter")
	}
}

// checkNotAllowedSymbols проверяет что лог-сообщение не содержит
// нелатинских и специальных символов.
func checkNotAllowedSymbols(pass *analysis.Pass, expr *ast.CallExpr) {
	lit, ok := getStringLiteral(expr)
	if !ok {
		return
	}

	var hasNonLatin, hasSpecial bool
	for _, r := range lit {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == ' ' {
			continue
		}
		if unicode.IsLetter(r) {
			hasNonLatin = true
		} else {
			hasSpecial = true
		}
	}
	if hasNonLatin {
		pass.Reportf(expr.Args[0].Pos(), "log messages must only contains latin letters")
	}
	if hasSpecial {
		pass.Reportf(expr.Args[0].Pos(), "log messages must not contains any special symbols")
	}
}

// Возвращает строковый литерал первого аргумента функции без кавычек
func getStringLiteral(node *ast.CallExpr) (string, bool) {
	lit, ok := node.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	unq, err := strconv.Unquote(lit.Value)
	if err != nil || unq == "" {
		return "", false
	}

	return unq, true
}
