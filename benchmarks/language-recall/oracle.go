package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// oracle type-checks one repository and reports the errors it raises. The
// baseline is the errors of the untouched source: a repository that does not
// check cleanly still works as long as only the errors a rename adds are read.
type oracle struct {
	repository corpusRepository
	baseline   map[string]bool
	lines      map[string][]string
}

// oracleError is one diagnostic, keyed without its column so a rename that
// lengthens a line does not make an old error look new.
type oracleError struct {
	File    string
	Line    int
	Code    string
	Message string
}

func (e oracleError) key() string {
	return fmt.Sprintf("%s:%d:%s:%s", e.File, e.Line, e.Code, e.Message)
}

func newOracle(ctx context.Context, repository corpusRepository) (*oracle, error) {
	o := &oracle{repository: repository, baseline: map[string]bool{}, lines: map[string][]string{}}
	errs, err := o.check(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, e := range errs {
		o.baseline[e.key()] = true
	}
	return o, nil
}

// dependents renames the declaration called name in file (a path relative to the
// repository) by replacing the file's content, type-checks, and returns the
// files that gained an error because of the rename, with how many each gained.
// The source on disk is restored before returning.
//
// An error counts only when the line it is reported on mentions the old name.
// Renaming a type can change what the checker infers elsewhere, and those
// cascading errors are on lines that never name the declaration: counting
// them would make files that do not use it look like dependents.
//
// Errors on the declaration's own line are left out too. They are about the
// declaration no longer matching something -- an `override` with nothing to
// override, a method a structural interface expected -- which is a relation,
// not a use.
func (o *oracle) dependents(ctx context.Context, file string, line int, name string, renamed []byte) (map[string]int, error) {
	var errs []oracleError
	var err error
	switch o.repository.Language {
	case "go":
		errs, err = o.check(ctx, file, renamed)
	default:
		errs, err = o.checkOnDisk(ctx, file, renamed)
	}
	if err != nil {
		return nil, err
	}
	mention := regexp.MustCompile(`(?:^|[^\w$#])` + regexp.QuoteMeta(name) + `(?:$|[^\w$])`)
	files := map[string]int{}
	for _, e := range errs {
		if o.baseline[e.key()] || isRelationError(o.repository.Language, e) {
			continue
		}
		if e.File == file && e.Line == line {
			continue
		}
		if !mention.MatchString(o.line(e.File, e.Line)) {
			continue
		}
		files[e.File]++
	}
	return files, nil
}

// line returns a line of the original source. The disk holds the original
// again by the time errors are read, and a rename only lengthens a line.
func (o *oracle) line(file string, number int) string {
	lines, ok := o.lines[file]
	if !ok {
		content, err := os.ReadFile(filepath.Join(o.repository.Path, file))
		if err == nil {
			lines = strings.Split(string(content), "\n")
		}
		o.lines[file] = lines
	}
	if number < 1 || number > len(lines) {
		return ""
	}
	return lines[number-1]
}

// isRelationError reports whether an error is about a type no longer satisfying
// an interface or a base class, not about the renamed name being used. Those
// files depend on the declaration through an implementation relation, which
// `find_references` is not the tool for, so they are not part of the truth.
func isRelationError(language string, e oracleError) bool {
	switch language {
	case "go":
		return strings.Contains(e.Message, "does not implement") || strings.Contains(e.Message, "missing method") ||
			strings.Contains(e.Message, "wrong type for method")
	default:
		switch e.Code {
		case "TS2415", "TS2416", "TS2420", "TS2515", "TS2559":
			return true
		}
		return false
	}
}

func (o *oracle) checkOnDisk(ctx context.Context, file string, renamed []byte) ([]oracleError, error) {
	path := filepath.Join(o.repository.Path, file)
	original, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if renamed != nil {
		if err := os.WriteFile(path, renamed, 0o644); err != nil {
			return nil, err
		}
		defer func() { _ = os.WriteFile(path, original, 0o644) }()
	}
	return o.check(ctx, "", nil)
}

// check runs the type checker. For Go the rename travels as an overlay and the
// disk is never written; for TypeScript the caller has already written it.
func (o *oracle) check(ctx context.Context, file string, content []byte) ([]oracleError, error) {
	if o.repository.Language == "go" {
		return o.checkGo(ctx, file, content)
	}
	return o.checkTypeScript(ctx)
}

// checkGo loads every package of the module with its tests, type-checking from
// source all the way down. Loading from source matters: a package whose
// dependency no longer compiles would otherwise not be checked at all, and the
// uses of the renamed name in it would be missed.
func (o *oracle) checkGo(ctx context.Context, file string, content []byte) ([]oracleError, error) {
	cfg := &packages.Config{
		Context: ctx,
		Dir:     o.repository.Path,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: true,
		Env:   toolEnvironment(),
	}
	if content != nil {
		cfg.Overlay = map[string][]byte{filepath.Join(o.repository.Path, file): content}
	}
	loaded, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var errs []oracleError
	packages.Visit(loaded, nil, func(pkg *packages.Package) {
		for _, packageError := range pkg.Errors {
			e, ok := parseGoError(o.repository.Path, packageError)
			if !ok || seen[e.key()] {
				continue
			}
			seen[e.key()] = true
			errs = append(errs, e)
		}
	})
	sort.Slice(errs, func(i, j int) bool { return errs[i].key() < errs[j].key() })
	return errs, nil
}

var goPosition = regexp.MustCompile(`^(.*):(\d+):(\d+)$`)

func parseGoError(root string, e packages.Error) (oracleError, bool) {
	match := goPosition.FindStringSubmatch(e.Pos)
	if match == nil {
		return oracleError{}, false
	}
	relative, err := filepath.Rel(root, match[1])
	if err != nil || strings.HasPrefix(relative, "..") {
		return oracleError{}, false
	}
	line, _ := strconv.Atoi(match[2])
	return oracleError{File: filepath.ToSlash(relative), Line: line, Message: e.Msg}, true
}

var tscError = regexp.MustCompile(`^(.+?)\((\d+),\d+\): error (TS\d+): (.*)$`)

// checkTypeScript runs the repository's own compiler on its own project, so the
// files it reads are the files `tsc` is configured to read.
func (o *oracle) checkTypeScript(ctx context.Context) ([]oracleError, error) {
	process := exec.CommandContext(ctx, "npx", "--no-install", "tsc", "--noEmit", "-p", o.repository.Project, "--pretty", "false")
	process.Dir = o.repository.Path
	process.Env = toolEnvironment()
	var out bytes.Buffer
	process.Stdout, process.Stderr = &out, &out
	// A type error is a nonzero exit and the whole point. Only a failure that
	// printed no diagnostic is an error here.
	runErr := process.Run()
	var errs []oracleError
	for _, line := range strings.Split(out.String(), "\n") {
		match := tscError.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if match == nil {
			continue
		}
		number, _ := strconv.Atoi(match[2])
		errs = append(errs, oracleError{File: filepath.ToSlash(match[1]), Line: number, Code: match[3], Message: match[4]})
	}
	if runErr != nil && len(errs) == 0 {
		return nil, fmt.Errorf("tsc: %w: %s", runErr, strings.TrimSpace(out.String()))
	}
	return errs, nil
}
