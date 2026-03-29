package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	log "github.com/sirupsen/logrus"
)

// Engine evaluates drift events against OPA/Rego policies.
type Engine struct {
	mu       sync.RWMutex
	compiler *ast.Compiler
	store    storage.Store
	modules  map[string]string // filename -> rego source
}

// NewEngine creates a policy engine with no policies loaded.
func NewEngine() *Engine {
	return &Engine{
		store:   inmem.New(),
		modules: make(map[string]string),
	}
}

// LoadDir loads all .rego files from a directory (non-recursive).
func (e *Engine) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read policy dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rego") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read policy file %s: %w", path, err)
		}
		e.modules[entry.Name()] = string(data)
		log.WithField("file", entry.Name()).Debug("Loaded policy file")
	}

	return e.compile()
}

// LoadModule loads a single Rego module from source.
func (e *Engine) LoadModule(name, source string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.modules[name] = source
	return e.compile()
}

// compile compiles all loaded modules into an ast.Compiler.
func (e *Engine) compile() error {
	mods := make(map[string]*ast.Module, len(e.modules))
	for name, src := range e.modules {
		parsed, err := ast.ParseModuleWithOpts(name, src, ast.ParserOptions{RegoVersion: ast.RegoV1})
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		mods[name] = parsed
	}
	compiler := ast.NewCompiler()
	compiler.Compile(mods)
	if compiler.Failed() {
		return fmt.Errorf("compile policies: %v", compiler.Errors)
	}
	e.compiler = compiler
	return nil
}

// Evaluate evaluates a drift input against loaded policies and returns
// the decision. The default query path is "data.tfdrift.decision".
//
// Policy output contract (Rego):
//
//	decision := "allow" | "alert" | "remediate" | "deny"
//	reason   := <string>           (optional)
//	severity := <string>           (optional, overrides input severity)
//	labels   := {<string>:<string>} (optional)
func (e *Engine) Evaluate(ctx context.Context, input *DriftInput) (*EvalResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.compiler == nil || len(e.modules) == 0 {
		// No policies loaded → default to alert (pass-through)
		return &EvalResult{Decision: DecisionAlert, Reason: "no policy loaded"}, nil
	}

	r := rego.New(
		rego.Query("data.tfdrift"),
		rego.Compiler(e.compiler),
		rego.Store(e.store),
		rego.Input(input),
	)

	rs, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("policy eval: %w", err)
	}

	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return &EvalResult{Decision: DecisionAlert, Reason: "policy returned no result"}, nil
	}

	// The result is data.tfdrift which should be an object
	raw, ok := rs[0].Expressions[0].Value.(map[string]interface{})
	if !ok {
		return &EvalResult{Decision: DecisionAlert, Reason: "policy result is not an object"}, nil
	}

	return parseResult(raw), nil
}

// parseResult converts the raw Rego output map into an EvalResult.
func parseResult(raw map[string]interface{}) *EvalResult {
	result := &EvalResult{
		Decision: DecisionAlert, // default
	}

	if d, ok := raw["decision"].(string); ok {
		switch Decision(d) {
		case DecisionAllow, DecisionAlert, DecisionRemediate, DecisionDeny:
			result.Decision = Decision(d)
		}
	}

	if r, ok := raw["reason"].(string); ok {
		result.Reason = r
	}

	if s, ok := raw["severity"].(string); ok {
		result.Severity = s
	}

	if labels, ok := raw["labels"].(map[string]interface{}); ok {
		result.Labels = make(map[string]string, len(labels))
		for k, v := range labels {
			if sv, ok := v.(string); ok {
				result.Labels[k] = sv
			}
		}
	}

	if suppressors, ok := raw["suppressors"].([]interface{}); ok {
		for _, s := range suppressors {
			if sv, ok := s.(string); ok {
				result.Suppressors = append(result.Suppressors, sv)
			}
		}
	}

	return result
}

// ModuleCount returns the number of loaded policy modules.
func (e *Engine) ModuleCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.modules)
}
