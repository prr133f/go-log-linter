package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
)

// ---------- helpers ----------

// makeLitCall creates *ast.CallExpr whose first argument is a *ast.BasicLit.
func makeLitCall(kind token.Token, value string) *ast.CallExpr {
	return &ast.CallExpr{
		Args: []ast.Expr{
			&ast.BasicLit{Kind: kind, Value: value},
		},
	}
}

// makeIdentCall creates *ast.CallExpr whose first argument is an *ast.Ident.
func makeIdentCall(name string) *ast.CallExpr {
	return &ast.CallExpr{
		Args: []ast.Expr{
			&ast.Ident{Name: name},
		},
	}
}

// collectDiagnostics returns a *analysis.Pass with a Report that appends
// diagnostics to the provided slice, plus a getter for the collected messages.
func collectDiagnostics() (*analysis.Pass, *[]analysis.Diagnostic) {
	var diags []analysis.Diagnostic
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}
	return pass, &diags
}

// messages extracts Message strings from a slice of diagnostics.
func messages(diags []analysis.Diagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Message
	}
	return out
}

// ---------- TestGetStringLiteral ----------

func TestGetStringLiteral(t *testing.T) {
	tests := []struct {
		name   string
		node   *ast.CallExpr
		want   string
		wantOk bool
	}{
		{
			name:   "normal string",
			node:   makeLitCall(token.STRING, `"hello world"`),
			want:   "hello world",
			wantOk: true,
		},
		{
			name:   "raw string",
			node:   makeLitCall(token.STRING, "`raw`"),
			want:   "raw",
			wantOk: true,
		},
		{
			name:   "empty string returns false",
			node:   makeLitCall(token.STRING, `""`),
			want:   "",
			wantOk: false,
		},
		{
			name:   "integer literal returns false",
			node:   makeLitCall(token.INT, "42"),
			want:   "",
			wantOk: false,
		},
		{
			name:   "first arg is ident returns false",
			node:   makeIdentCall("variable"),
			want:   "",
			wantOk: false,
		},
		{
			name:   "string with unicode",
			node:   makeLitCall(token.STRING, `"café"`),
			want:   "café",
			wantOk: true,
		},
		{
			name:   "string with escape sequences",
			node:   makeLitCall(token.STRING, `"line1\nline2"`),
			want:   "line1\nline2",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getStringLiteral(tt.node)
			if ok != tt.wantOk {
				t.Fatalf("getStringLiteral() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.want {
				t.Errorf("getStringLiteral() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------- TestCollectIdents ----------

func TestCollectIdents(t *testing.T) {
	t.Run("single ident", func(t *testing.T) {
		id := &ast.Ident{Name: "x"}

		idents := collectIdents(id)

		if len(idents) != 1 || idents[0].Name != "x" {
			t.Fatalf("expected [x], got %v", identNames(idents))
		}
	})

	t.Run("binary add with two idents", func(t *testing.T) {
		expr := &ast.BinaryExpr{
			Op: token.ADD,
			X:  &ast.Ident{Name: "a"},
			Y:  &ast.Ident{Name: "b"},
		}

		idents := collectIdents(expr)

		names := identNames(idents)
		if len(names) != 2 || names[0] != "a" || names[1] != "b" {
			t.Fatalf("expected [a b], got %v", names)
		}
	})

	t.Run("nested binary add", func(t *testing.T) {
		// ("a" + "b") + "c"
		expr := &ast.BinaryExpr{
			Op: token.ADD,
			X: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.Ident{Name: "a"},
				Y:  &ast.Ident{Name: "b"},
			},
			Y: &ast.Ident{Name: "c"},
		}

		idents := collectIdents(expr)

		names := identNames(idents)
		if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
			t.Fatalf("expected [a b c], got %v", names)
		}
	})

	t.Run("non ADD binary op stops recursion", func(t *testing.T) {
		expr := &ast.BinaryExpr{
			Op: token.MUL,
			X:  &ast.Ident{Name: "a"},
			Y:  &ast.Ident{Name: "b"},
		}

		idents := collectIdents(expr)

		if len(idents) != 0 {
			t.Fatalf("expected [], got %v", identNames(idents))
		}
	})

	t.Run("basic lit is ignored", func(t *testing.T) {
		expr := &ast.BinaryExpr{
			Op: token.ADD,
			X:  &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
			Y:  &ast.Ident{Name: "x"},
		}

		idents := collectIdents(expr)

		if len(idents) != 1 || idents[0].Name != "x" {
			t.Fatalf("expected [x], got %v", identNames(idents))
		}
	})

	t.Run("nil expression", func(t *testing.T) {

		idents := collectIdents(nil)
		if len(idents) != 0 {
			t.Fatalf("expected [], got %v", identNames(idents))
		}
	})
}

func identNames(idents []*ast.Ident) []string {
	names := make([]string, len(idents))
	for i, id := range idents {
		names[i] = id.Name
	}
	return names
}

// ---------- TestCheckStartsWithUpper ----------

func TestCheckStartsWithUpper(t *testing.T) {
	tests := []struct {
		name      string
		node      *ast.CallExpr
		wantDiags int
	}{
		{
			name:      "lowercase start no report",
			node:      makeLitCall(token.STRING, `"hello world"`),
			wantDiags: 0,
		},
		{
			name:      "uppercase start reports",
			node:      makeLitCall(token.STRING, `"Hello world"`),
			wantDiags: 1,
		},
		{
			name:      "digit start no report",
			node:      makeLitCall(token.STRING, `"123 hello"`),
			wantDiags: 0,
		},
		{
			name:      "non string arg no report",
			node:      makeIdentCall("someVar"),
			wantDiags: 0,
		},
		{
			name:      "uppercase unicode letter reports",
			node:      makeLitCall(token.STRING, `"Über"`),
			wantDiags: 1,
		},
		{
			name:      "lowercase unicode letter no report",
			node:      makeLitCall(token.STRING, `"über"`),
			wantDiags: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diags := collectDiagnostics()
			checkStartsWithUpper(pass, tt.node)
			if len(*diags) != tt.wantDiags {
				t.Errorf("got %d diagnostics, want %d: %v", len(*diags), tt.wantDiags, messages(*diags))
			}
			if tt.wantDiags > 0 && len(*diags) > 0 {
				want := "log messages must start with lowercase letter"
				if (*diags)[0].Message != want {
					t.Errorf("message = %q, want %q", (*diags)[0].Message, want)
				}
			}
		})
	}
}

// ---------- TestCheckNotAllowedSymbols ----------

func TestCheckNotAllowedSymbols(t *testing.T) {
	tests := []struct {
		name         string
		node         *ast.CallExpr
		wantNonLatin bool
		wantSpecial  bool
	}{
		{
			name:         "only latin letters and digits",
			node:         makeLitCall(token.STRING, `"hello world 123"`),
			wantNonLatin: false,
			wantSpecial:  false,
		},
		{
			name:         "cyrillic letters",
			node:         makeLitCall(token.STRING, `"привет"`),
			wantNonLatin: true,
			wantSpecial:  false,
		},
		{
			name:         "special symbols only",
			node:         makeLitCall(token.STRING, `"hello!"`),
			wantNonLatin: false,
			wantSpecial:  true,
		},
		{
			name:         "both non latin and special",
			node:         makeLitCall(token.STRING, `"привет!"`),
			wantNonLatin: true,
			wantSpecial:  true,
		},
		{
			name:         "dots are special",
			node:         makeLitCall(token.STRING, `"hello..."`),
			wantNonLatin: false,
			wantSpecial:  true,
		},
		{
			name:         "non string arg no report",
			node:         makeIdentCall("someVar"),
			wantNonLatin: false,
			wantSpecial:  false,
		},
		{
			name:         "mixed latin and cyrillic",
			node:         makeLitCall(token.STRING, `"hello мир"`),
			wantNonLatin: true,
			wantSpecial:  false,
		},
		{
			name:         "emoji is special",
			node:         makeLitCall(token.STRING, "\"hello\\u2764\""),
			wantNonLatin: false,
			wantSpecial:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diags := collectDiagnostics()
			checkNotAllowedSymbols(pass, tt.node)

			msgs := messages(*diags)
			hasNonLatin := containsMsg(msgs, "log messages must only contains latin letters")
			hasSpecial := containsMsg(msgs, "log messages must not contains any special symbols")

			if hasNonLatin != tt.wantNonLatin {
				t.Errorf("non-latin diagnostic: got %v, want %v (diags: %v)", hasNonLatin, tt.wantNonLatin, msgs)
			}
			if hasSpecial != tt.wantSpecial {
				t.Errorf("special diagnostic: got %v, want %v (diags: %v)", hasSpecial, tt.wantSpecial, msgs)
			}
		})
	}
}

func containsMsg(msgs []string, target string) bool {
	for _, m := range msgs {
		if m == target {
			return true
		}
	}
	return false
}

// ---------- TestCheckSensitiveData ----------

func TestCheckSensitiveData(t *testing.T) {
	tests := []struct {
		name      string
		expr      ast.Expr
		wantDiags int
		wantIdent string
	}{
		{
			name: "token variable detected",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
				Y:  &ast.Ident{Name: "myToken"},
			},
			wantDiags: 1,
			wantIdent: "myToken",
		},
		{
			name: "password variable detected",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"data"`},
				Y:  &ast.Ident{Name: "userPassword"},
			},
			wantDiags: 1,
			wantIdent: "userPassword",
		},
		{
			name: "safe variable no report",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
				Y:  &ast.Ident{Name: "username"},
			},
			wantDiags: 0,
		},
		{
			name:      "non binary expr no report",
			expr:      &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
			wantDiags: 0,
		},
		{
			name: "non ADD binary op no report",
			expr: &ast.BinaryExpr{
				Op: token.MUL,
				X:  &ast.Ident{Name: "secret"},
				Y:  &ast.Ident{Name: "count"},
			},
			wantDiags: 0,
		},
		{
			name: "apiKey in nested concat detected",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X: &ast.BinaryExpr{
					Op: token.ADD,
					X:  &ast.BasicLit{Kind: token.STRING, Value: `"prefix"`},
					Y:  &ast.Ident{Name: "apiKey"},
				},
				Y: &ast.BasicLit{Kind: token.STRING, Value: `"suffix"`},
			},
			wantDiags: 1,
			wantIdent: "apiKey",
		},
		{
			name: "multiple sensitive vars reports each",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.Ident{Name: "secretValue"},
				Y:  &ast.Ident{Name: "authHeader"},
			},
			wantDiags: 2,
		},
		{
			name: "credential variable detected",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"cred: "`},
				Y:  &ast.Ident{Name: "credential"},
			},
			wantDiags: 1,
			wantIdent: "credential",
		},
		{
			name: "private variable detected",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"val: "`},
				Y:  &ast.Ident{Name: "privateKey"},
			},
			// "privateKey" matches both "private" and "key"
			wantDiags: 1,
		},
		{
			name: "case insensitive match TOKEN",
			expr: &ast.BinaryExpr{
				Op: token.ADD,
				X:  &ast.BasicLit{Kind: token.STRING, Value: `"data"`},
				Y:  &ast.Ident{Name: "AccessTOKEN"},
			},
			wantDiags: 1,
			wantIdent: "AccessTOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diags := collectDiagnostics()
			checkSensitiveData(pass, tt.expr)
			if len(*diags) != tt.wantDiags {
				t.Fatalf("got %d diagnostics, want %d: %v", len(*diags), tt.wantDiags, messages(*diags))
			}
			if tt.wantIdent != "" && len(*diags) > 0 {
				wantMsg := "potentially sensitive data \"" + tt.wantIdent + "\" is concatenated into log message"
				if (*diags)[0].Message != wantMsg {
					t.Errorf("message = %q, want %q", (*diags)[0].Message, wantMsg)
				}
			}
		})
	}
}

// ---------- TestIsLinted ----------

func TestIsLinted(t *testing.T) {
	// Builds a minimal pass with TypesInfo so isLinted can resolve package paths.
	makeSlogPass := func(methodName string) (*analysis.Pass, *ast.SelectorExpr) {
		pkg := types.NewPackage("log/slog", "slog")
		pkgName := types.NewPkgName(token.NoPos, nil, "slog", pkg)

		ident := &ast.Ident{Name: "slog"}
		sel := &ast.SelectorExpr{
			X:   ident,
			Sel: &ast.Ident{Name: methodName},
		}

		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: pkgName,
				},
			},
		}
		return pass, sel
	}

	makeZapPass := func(methodName string) (*analysis.Pass, *ast.SelectorExpr) {
		pkg := types.NewPackage("go.uber.org/zap", "zap")
		pkgName := types.NewPkgName(token.NoPos, nil, "zap", pkg)

		ident := &ast.Ident{Name: "zap"}
		sel := &ast.SelectorExpr{
			X:   ident,
			Sel: &ast.Ident{Name: methodName},
		}

		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: pkgName,
				},
			},
		}
		return pass, sel
	}

	t.Run("slog Info is linted", func(t *testing.T) {
		pass, sel := makeSlogPass("Info")
		if !isLinted(pass, sel) {
			t.Error("expected slog.Info to be linted")
		}
	})

	t.Run("slog Debug is linted", func(t *testing.T) {
		pass, sel := makeSlogPass("Debug")
		if !isLinted(pass, sel) {
			t.Error("expected slog.Debug to be linted")
		}
	})

	t.Run("slog Warn is linted", func(t *testing.T) {
		pass, sel := makeSlogPass("Warn")
		if !isLinted(pass, sel) {
			t.Error("expected slog.Warn to be linted")
		}
	})

	t.Run("slog Error is linted", func(t *testing.T) {
		pass, sel := makeSlogPass("Error")
		if !isLinted(pass, sel) {
			t.Error("expected slog.Error to be linted")
		}
	})

	t.Run("slog Fatal is linted", func(t *testing.T) {
		pass, sel := makeSlogPass("Fatal")
		if !isLinted(pass, sel) {
			t.Error("expected slog.Fatal to be linted")
		}
	})

	t.Run("zap Info is linted", func(t *testing.T) {
		pass, sel := makeZapPass("Info")
		if !isLinted(pass, sel) {
			t.Error("expected zap.Info to be linted")
		}
	})

	t.Run("zap Error is linted", func(t *testing.T) {
		pass, sel := makeZapPass("Error")
		if !isLinted(pass, sel) {
			t.Error("expected zap.Error to be linted")
		}
	})

	t.Run("slog non linted method", func(t *testing.T) {
		pass, sel := makeSlogPass("With")
		if isLinted(pass, sel) {
			t.Error("expected slog.With not to be linted")
		}
	})

	t.Run("unknown package not linted", func(t *testing.T) {
		pkg := types.NewPackage("fmt", "fmt")
		pkgName := types.NewPkgName(token.NoPos, nil, "fmt", pkg)

		ident := &ast.Ident{Name: "fmt"}
		sel := &ast.SelectorExpr{
			X:   ident,
			Sel: &ast.Ident{Name: "Info"},
		}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: pkgName,
				},
			},
		}
		if isLinted(pass, sel) {
			t.Error("expected fmt.Info not to be linted")
		}
	})

	t.Run("unresolved ident not linted", func(t *testing.T) {
		ident := &ast.Ident{Name: "unknown"}
		sel := &ast.SelectorExpr{
			X:   ident,
			Sel: &ast.Ident{Name: "Info"},
		}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{},
			},
		}
		if isLinted(pass, sel) {
			t.Error("expected unresolved.Info not to be linted")
		}
	})
}

// ---------- TestGetPackagePath ----------

func TestGetPackagePath(t *testing.T) {
	t.Run("package name returns import path", func(t *testing.T) {
		pkg := types.NewPackage("log/slog", "slog")
		pkgName := types.NewPkgName(token.NoPos, nil, "slog", pkg)

		ident := &ast.Ident{Name: "slog"}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: pkgName,
				},
			},
		}

		got := getPackagePath(pass, ident)
		if got != "log/slog" {
			t.Errorf("getPackagePath() = %q, want %q", got, "log/slog")
		}
	})

	t.Run("non ident expr returns empty", func(t *testing.T) {
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{},
			},
		}
		got := getPackagePath(pass, &ast.BasicLit{Kind: token.STRING, Value: `"x"`})
		if got != "" {
			t.Errorf("getPackagePath() = %q, want empty", got)
		}
	})

	t.Run("unresolved ident returns empty", func(t *testing.T) {
		ident := &ast.Ident{Name: "missing"}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{},
			},
		}
		got := getPackagePath(pass, ident)
		if got != "" {
			t.Errorf("getPackagePath() = %q, want empty", got)
		}
	})

	t.Run("named type returns package path", func(t *testing.T) {
		pkg := types.NewPackage("go.uber.org/zap", "zap")
		typeName := types.NewTypeName(token.NoPos, pkg, "Logger", nil)
		named := types.NewNamed(typeName, types.Typ[types.Bool], nil)
		varObj := types.NewVar(token.NoPos, pkg, "logger", named)

		ident := &ast.Ident{Name: "logger"}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: varObj,
				},
			},
		}

		got := getPackagePath(pass, ident)
		if got != "go.uber.org/zap" {
			t.Errorf("getPackagePath() = %q, want %q", got, "go.uber.org/zap")
		}
	})

	t.Run("pointer to named type returns package path", func(t *testing.T) {
		pkg := types.NewPackage("go.uber.org/zap", "zap")
		typeName := types.NewTypeName(token.NoPos, pkg, "Logger", nil)
		named := types.NewNamed(typeName, types.Typ[types.Bool], nil)
		ptrType := types.NewPointer(named)
		varObj := types.NewVar(token.NoPos, pkg, "logger", ptrType)

		ident := &ast.Ident{Name: "logger"}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: varObj,
				},
			},
		}

		got := getPackagePath(pass, ident)
		if got != "go.uber.org/zap" {
			t.Errorf("getPackagePath() = %q, want %q", got, "go.uber.org/zap")
		}
	})

	t.Run("basic type variable returns empty", func(t *testing.T) {
		varObj := types.NewVar(token.NoPos, nil, "x", types.Typ[types.Int])

		ident := &ast.Ident{Name: "x"}
		pass := &analysis.Pass{
			TypesInfo: &types.Info{
				Uses: map[*ast.Ident]types.Object{
					ident: varObj,
				},
			},
		}

		got := getPackagePath(pass, ident)
		if got != "" {
			t.Errorf("getPackagePath() = %q, want empty", got)
		}
	})
}
