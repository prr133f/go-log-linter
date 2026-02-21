package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		node, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}
		sel, ok := node.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if ok := isLinted(pass, sel); !ok {
			return
		}
		if len(node.Args) == 0 {
			return
		}

		checkStartsWithUpper(pass, node)

		checkNotAllowedSymbols(pass, node)

		checkSensitiveData(pass, node.Args[0])
	})
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
		// pass.Reportf(expr.Args[0].Pos(), "log messages must start with lowercase letter")
		pass.Report(analysis.Diagnostic{
			Pos:     expr.Args[0].Pos(),
			End:     expr.Args[0].End(),
			Message: "log messages must start with lowercase letter",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("letter %s must be lowercase", string(r)),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     expr.Args[0].Pos() + 1,
							End:     expr.Args[0].Pos() + 1 + token.Pos(utf8.RuneLen(r)),
							NewText: []byte(string(unicode.ToLower(r))),
						},
					},
				},
			},
		})
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
		pass.Report(analysis.Diagnostic{
			Pos:     expr.Args[0].Pos(),
			End:     expr.Args[0].End(),
			Message: "log messages must only contains latin letters",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "remove non-latin characters",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     expr.Args[0].Pos(),
							End:     expr.Args[0].End(),
							NewText: []byte(strconv.Quote(removeNonLatin(lit))),
						},
					},
				},
			},
		})
	}
	if hasSpecial {
		pass.Report(analysis.Diagnostic{
			Pos:     expr.Args[0].Pos(),
			End:     expr.Args[0].End(),
			Message: "log messages must not contains any special symbols",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "remove special symbols",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     expr.Args[0].Pos(),
							End:     expr.Args[0].End(),
							NewText: []byte(strconv.Quote(removeSpecialSymbols(lit))),
						},
					},
				},
			},
		})
	}
}

// removeNonLatin удаляет из строки все символы, не являющиеся
// латинскими буквами, цифрами или пробелами.
func removeNonLatin(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// removeSpecialSymbols удаляет из строки все символы, не являющиеся
// буквами (любого алфавита), цифрами или пробелами.
func removeSpecialSymbols(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var sensitivePatterns = []string{
	"token",
	"password",
	"passwd",
	"secret",
	"apikey",
	"credential",
	"auth",
	"private",
}

// checkSensitiveData проверяет, не конкатенируется ли в лог-сообщение
// переменная с потенциально чувствительным именем
func checkSensitiveData(pass *analysis.Pass, expr ast.Expr) {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.ADD {
		return
	}

	idents := collectIdents(binExpr)

	for _, ident := range idents {
		name := strings.ToLower(ident.Name)
		for _, pattern := range sensitivePatterns {
			if strings.Contains(name, pattern) {
				pass.Reportf(ident.Pos(),
					"potentially sensitive data %q is concatenated into log message",
					ident.Name,
				)
				break
			}
		}
	}
}

// collectIdents рекурсивно собирает все идентификаторы переменных
// из дерева конкатенации (вложенных BinaryExpr с token.ADD)
func collectIdents(expr ast.Expr) []*ast.Ident {
	var idents []*ast.Ident
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch e := e.(type) {
		case *ast.BinaryExpr:
			if e.Op == token.ADD {
				walk(e.X)
				walk(e.Y)
			}
		case *ast.Ident:
			idents = append(idents, e)
		}
	}
	walk(expr)
	return idents
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
