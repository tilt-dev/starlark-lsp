package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

const SPEC_Bazel = "https://raw.githubusercontent.com/bazelbuild/starlark/master/spec.md"
const SPEC_StarlarkGo = "https://raw.githubusercontent.com/google/starlark-go/master/doc/spec.md"
const ConstantsAndFunctions = "## Built-in constants and functions"
const Methods = "## Built-in methods"

// End of methods section in SPEC_Bazel
const GrammarReference = "## Grammar reference"

// End of methods section in SPEC_StarlarkGo
const DialectDifferences = "## Dialect differences"

type Function struct {
	Name      string
	Desc      string
	Signature string
}

type Method struct {
	Function
	ClassName string
}

type BuiltinsScanner struct {
	scanner   *bufio.Scanner
	hints     *analysis.Builtins
	Functions []Function
	Methods   []Method
}

func (b *BuiltinsScanner) lookingAt(marker string) bool {
	return strings.HasPrefix(string(b.scanner.Bytes()), marker)
}

func (b *BuiltinsScanner) readUntil(marker string) error {
	for {
		if !b.scanner.Scan() {
			return b.scanner.Err()
		}

		if b.lookingAt(marker) {
			return nil
		}
	}
}

func (b *BuiltinsScanner) nextSection() error {
	return b.readUntil("#")
}

func (b *BuiltinsScanner) sectionName() string {
	bytes := b.scanner.Bytes()
	for i, b := range bytes {
		if b != '#' && b != ' ' {
			return string(bytes[i:])
		}
	}
	return ""
}

func (b *BuiltinsScanner) skipBlanks() error {
	for {
		if !b.scanner.Scan() {
			return b.scanner.Err()
		}
		if len(b.scanner.Bytes()) != 0 {
			return nil
		}
	}
}

func (b *BuiltinsScanner) nextParagraph() (string, error) {
	buf := bytes.NewBuffer(make([]byte, 1024))
	buf.Reset()

	err := b.skipBlanks()
	if err != nil {
		return "", err
	}

	for {
		if len(b.scanner.Bytes()) == 0 {
			return buf.String(), nil
		}

		if buf.Len() > 0 {
			err = buf.WriteByte(' ')
			if err != nil {
				return "", err
			}
		}

		_, err = buf.Write(b.scanner.Bytes())
		if err != nil {
			return "", err
		}

		if !b.scanner.Scan() {
			return "", b.scanner.Err()
		}
	}
}

func (b *BuiltinsScanner) loadHints() error {
	contents, err := os.ReadFile("hack/starlark-builtins.py")
	if os.IsNotExist(err) {
		contents, err = os.ReadFile("starlark-builtins.py")
	}
	if err != nil {
		return err
	}
	hints, err := analysis.LoadBuiltinsFromSource(context.Background(), contents, "starlark-builtins.py")
	if err != nil {
		return err
	}
	b.hints = hints
	return nil
}

var signature *regexp.Regexp = regexp.MustCompile("`([^`]+)`")
var altSignature *regexp.Regexp = regexp.MustCompile(`^(\S+)`)
var stripOptional *regexp.Regexp = regexp.MustCompile(`(\w+\.)?(\w+)\(?([^[)]+)?(\[[^]]+\])?\)?`)

func (b *BuiltinsScanner) parseConstantsAndFunctions() error {
	err := b.readUntil(ConstantsAndFunctions)
	if err != nil {
		return err
	}

	for {
		err = b.nextSection()
		if err != nil {
			return err
		}

		if b.lookingAt(Methods) {
			return nil
		}

		if b.lookingAt("### None") || b.lookingAt("### True and False") {
			continue
		} else {
			name := b.sectionName()
			para, err := b.nextParagraph()
			if err != nil {
				return err
			}
			match := signature.FindStringSubmatch(para)
			if len(match) < 2 {
				match = altSignature.FindStringSubmatch(para)
				if len(match) < 2 {
					return fmt.Errorf("expecting signature: %s", para)
				}
			}
			sig := stripOptional.FindStringSubmatch(match[1])
			if len(sig) < 4 {
				return fmt.Errorf("expecting signature: %s", match[1])
			}
			b.Functions = append(b.Functions, Function{
				Name:      name,
				Desc:      para,
				Signature: fmt.Sprintf("%s(%s)", sig[2], sig[3]),
			})
		}
	}
}

func (b *BuiltinsScanner) parseMethods() error {
	for {
		err := b.nextSection()
		if err != nil {
			return err
		}

		if b.lookingAt(GrammarReference) ||
			b.lookingAt(DialectDifferences) {
			return nil
		}

		section := b.sectionName()
		names := strings.Split(section, "·")
		if len(names) < 2 {
			return fmt.Errorf("expecting class·method: %s\n", section)
		}

		para, err := b.nextParagraph()
		if err != nil {
			return err
		}
		match := signature.FindStringSubmatch(para)
		if len(match) < 2 {
			return fmt.Errorf("expecting signature: %s", para)
		}
		sig := stripOptional.FindStringSubmatch(match[1])
		if len(sig) < 4 {
			return fmt.Errorf("expecting signature: %s", match[1])
		}
		methodSig := sig[3]
		if len(methodSig) > 0 {
			methodSig = ", " + methodSig
		}
		methodSig = "self" + methodSig
		b.Methods = append(b.Methods, Method{
			Function: Function{
				Name:      names[1],
				Desc:      para,
				Signature: fmt.Sprintf("%s(%s)", sig[2], methodSig),
			},
			ClassName: names[0],
		})
	}
}

func (b *BuiltinsScanner) outputStubs() {
	for _, f := range b.Functions {
		signature := f.Signature
		if sig, ok := b.hints.Functions[f.Name]; ok {
			signature = sig.Name + sig.Label()
		}
		fmt.Printf("def %s:\n  \"\"\"%s\"\"\"\n  pass\n\n", signature, f.Desc)
	}

	var className string
	var ty query.Type
	for _, m := range b.Methods {
		if className != m.ClassName {
			name := strings.ToTitle(m.ClassName[0:1]) + m.ClassName[1:]
			fmt.Printf("class %s:\n", name)
			className = m.ClassName
			ty = b.hints.Types[name]
		}
		signature := m.Signature
		for _, tm := range ty.Methods {
			if tm.Name == m.Name {
				// Re-insert 'self' parameter
				params := tm.Params
				tm.Params = append([]query.Parameter{{Name: "self", Content: "self"}}, params...)
				signature = tm.Name + tm.Label()
				break
			}
		}
		fmt.Printf("  def %s:\n    \"\"\"%s\"\"\"\n    pass\n\n", signature, m.Desc)
	}
}

func topError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	var bs *bufio.Scanner
	spec := ""
	if len(os.Args) > 1 {
		switch arg := os.Args[1]; arg {
		case "Bazel", "bazel":
			spec = SPEC_Bazel
		case "StarlarkGo", "starlark-go":
			spec = SPEC_StarlarkGo
		default:
			f, err := os.Open(arg)
			topError(fmt.Sprintf("reading %s", os.Args[1]), err)
			defer f.Close()
			bs = bufio.NewScanner(f)
		}
	}

	if bs == nil {
		if spec == "" {
			spec = SPEC_StarlarkGo
		}
		resp, err := http.Get(spec)
		topError("fetching spec", err)
		defer resp.Body.Close()
		bs = bufio.NewScanner(resp.Body)
	}

	scanner := &BuiltinsScanner{scanner: bs}

	topError("loading type hints",
		scanner.loadHints())

	topError("parsing constants and functions",
		scanner.parseConstantsAndFunctions())

	topError("parsing methods",
		scanner.parseMethods())

	if spec != "" {
		fmt.Printf("# This file was generated by `make builtins` based on the spec at:\n# %s\n\n", spec)
	}
	scanner.outputStubs()
}
