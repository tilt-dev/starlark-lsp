package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func (f *fixture) builtinSymbols() {
	_ = WithStarlarkBuiltins()(f.a)
}

func (f *fixture) osSysSymbols() {
	f.Symbols("os", "sys")
	f.builtins.Symbols[0].Children = []query.Symbol{
		f.Symbol("environ"),
		f.Symbol("name"),
	}
	f.builtins.Symbols[1].Children = []query.Symbol{
		f.Symbol("argv"),
		f.Symbol("executable"),
	}
}

func assertCompletionResult(t *testing.T, names []string, result *protocol.CompletionList) {
	labels := make([]string, len(result.Items))
	for i, item := range result.Items {
		labels[i] = item.Label
	}
	assert.ElementsMatch(t, names, labels)
}

func TestSimpleCompletion(t *testing.T) {
	f := newFixture(t)

	f.Symbols("foo", "bar", "baz")

	doc := f.MainDoc("")
	result := f.a.Completion(doc, protocol.Position{})
	assertCompletionResult(t, []string{"foo", "bar", "baz"}, result)

	doc = f.MainDoc("ba")
	result = f.a.Completion(doc, protocol.Position{Character: 2})
	assertCompletionResult(t, []string{"bar", "baz"}, result)
}

const docWithMultiplePlaces = `
def f1():
    pass

s = "a string"

def f2():
    # <- position 2
	return False

# <- position 1

if True:
    # position 3
	pass

t = 1234
`

const docWithErrorNode = `
def foo():
  pass

f(

def quux():
  pass
`

func TestCompletions(t *testing.T) {
	tests := []struct {
		doc            string
		line, char     uint32
		expected       []string
		osSys, builtin bool
	}{
		{doc: "", expected: []string{"os", "sys"}, osSys: true},
		{doc: "os.", char: 3, expected: []string{"environ", "name"}, osSys: true},
		{doc: "os.e", char: 4, expected: []string{"environ"}, osSys: true},

		// position 1
		{doc: docWithMultiplePlaces, line: 10, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},
		// position 2
		{doc: docWithMultiplePlaces, line: 7, char: 4, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},
		// position 3
		{doc: docWithMultiplePlaces, line: 13, char: 4, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},
		{doc: docWithErrorNode, line: 4, char: 1, expected: []string{"foo"}, osSys: true},
		// inside string
		{doc: `f = "abc123"`, char: 5, expected: []string{}, osSys: true},
		// inside comment
		{doc: `f = true # abc123`, char: 12, expected: []string{}, osSys: true},
		// builtins
		{doc: `f`, char: 1, expected: []string{"float", "fail"}, builtin: true},
		{doc: `N`, char: 1, expected: []string{"None"}, builtin: true},
		{doc: `T`, char: 1, expected: []string{"True"}, builtin: true},
		{doc: `F`, char: 1, expected: []string{"False"}, builtin: true},
		// inside function body
		{doc: "def fn():\n  \nx = True", line: 1, char: 2, expected: []string{"fn", "x", "os", "sys"}, osSys: true},
		{doc: "def fn():\n  a = 1\n  \n  \b  b = 2\n  return b\nx = True", line: 2, char: 2, expected: []string{"a", "fn", "os", "sys", "x"}, osSys: true},
		// inside a list
		{doc: "x = [os.]", char: 8, expected: []string{"environ", "name"}, osSys: true},
		// inside a binary expression
		{doc: "x = 'foo' + \nprint('')", char: 15, expected: []string{"x", "os", "sys"}, osSys: true},
		{doc: "x = 'foo' + os.\nprint('')", char: 15, expected: []string{"environ", "name"}, osSys: true},
		// inside function argument lists
		{doc: `foo()`, char: 4, expected: []string{"os", "sys"}, osSys: true},
		{doc: `foo(1, )`, char: 7, expected: []string{"os", "sys"}, osSys: true},
		// inside condition of a conditional
		{doc: "if :\n  pass\n", char: 3, expected: []string{"os", "sys"}, osSys: true},
		{doc: "if os.:\n  pass\n", char: 6, expected: []string{"environ", "name"}, osSys: true},
		{doc: "if flag and os.:\n  pass\n", char: 15, expected: []string{"environ", "name"}, osSys: true},
		// other edge cases
		// - because this gets parsed as an ERROR node at the top level, there's
		//   no assignment expression and the variable `flag` will not be in
		//   scope
		{doc: "flag = ", char: 7, expected: []string{"os", "sys"}, osSys: true},
		{doc: "flag = os.", char: 10, expected: []string{"environ", "name"}, osSys: true},
		// These should not trigger completion since the attribute expression is
		// anchored to a function call
		{doc: "flag = len(os).", char: 15, expected: []string{}, osSys: true},
		{doc: "flag = len(os).sys", char: 15, expected: []string{}, osSys: true},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newFixture(t)
			if tt.builtin {
				f.builtinSymbols()
			}
			if tt.osSys {
				f.osSysSymbols()
			}
			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, protocol.Position{Line: tt.line, Character: tt.char})
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestIdentifierCompletion(t *testing.T) {
	tests := []struct {
		doc      string
		col      uint32
		expected []string
	}{
		{doc: "", col: 0, expected: []string{""}},
		{doc: "os", col: 2, expected: []string{"os"}},
		{doc: "os.", col: 3, expected: []string{"os", ""}},
		{doc: "os.e", col: 4, expected: []string{"os", "e"}},
		{doc: "os.path.", col: 8, expected: []string{"os", "path", ""}},
		{doc: "os.path.e", col: 9, expected: []string{"os", "path", "e"}},
		{doc: "[os]", col: 3, expected: []string{"os"}},
		{doc: "[os.]", col: 4, expected: []string{"os", ""}},
		{doc: "[os.e]", col: 5, expected: []string{"os", "e"}},
		{doc: "x = [os]", col: 7, expected: []string{"os"}},
		{doc: "x = [os.]", col: 8, expected: []string{"os", ""}},
		{doc: "x = [os.e]", col: 9, expected: []string{"os", "e"}},
		{doc: "x = [os.path.]", col: 13, expected: []string{"os", "path", ""}},
		{doc: "x = [os.path.e]", col: 14, expected: []string{"os", "path", "e"}},
		{doc: "x = ", col: 4, expected: []string{""}},
		{doc: "if x and : pass", col: 9, expected: []string{""}},
		{doc: "if x and os.: pass", col: 12, expected: []string{"os", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newFixture(t)
			doc := f.MainDoc(tt.doc)
			pt := sitter.Point{Column: tt.col}
			nodes, ok := f.a.nodesAtPointForCompletion(doc, pt)
			assert.True(t, ok)
			ids := query.ExtractIdentifiers(doc, nodes, nil)
			assert.ElementsMatch(t, tt.expected, ids)
		})
	}
}

const functionFixture = `
def docker_build(ref: str,
                 context: str,
                 build_args: Dict[str, str] = {},
                 dockerfile: str = "Dockerfile",
                 dockerfile_contents: Union[str, Blob] = "",
                 live_update: List[LiveUpdateStep]=[],
                 match_in_env_vars: bool = False,
                 ignore: Union[str, List[str]] = [],
                 only: Union[str, List[str]] = [],
                 entrypoint: Union[str, List[str]] = [],
                 target: str = "",
                 ssh: Union[str, List[str]] = "",
                 network: str = "",
                 secret: Union[str, List[str]] = "",
                 extra_tag: Union[str, List[str]] = "",
                 container_args: List[str] = None,
                 cache_from: Union[str, List[str]] = [],
                 pull: bool = False,
                 platform: str = "") -> None:
    pass

def local(command: Union[str, List[str]],
          quiet: bool = False,
          command_bat: Union[str, List[str]] = "",
          echo_off: bool = False,
          env: Dict[str, str] = {},
          dir: str = "") -> Blob:
    pass
`

const customFn = `
def fn(a, b, c):
  pass

fn()
fn(b=1,)
`

func TestKeywordArgCompletion(t *testing.T) {
	tests := []struct {
		doc        string
		line, char uint32
		expected   []string
	}{
		{doc: "local(c)", char: 7, expected: []string{"command=", "command_bat="}},
		{doc: "local(c", char: 7, expected: []string{"command=", "command_bat="}},
		{doc: "local()", char: 6, expected: []string{"command=", "quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		{doc: "local(", char: 6, expected: []string{"command=", "quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		{doc: "docker_build()", char: 13, expected: []string{"ref=", "context=", "build_args=", "dockerfile=", "dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"}},

		// past first arg, exclude `command`
		{doc: "local('echo',", char: 13, expected: []string{"quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		// past second arg, exclude `ref` and `context`
		{doc: "docker_build(ref, context,)", char: 26, expected: []string{"build_args=", "dockerfile=", "dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"}},
		// used several kwargs
		{
			doc: "docker_build(ref='image:latest', context='.', dockerfile='Dockerfile.test', build_args={'DEBUG':'1'},)", char: 101,
			expected: []string{"dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"},
		},

		// used `command` by position, `env` by keyword
		{doc: "local('echo $MESSAGE', env={'MESSAGE':'HELLO'},)", char: 47, expected: []string{"quiet=", "command_bat=", "echo_off=", "dir=", "docker_build", "local"}},

		// didn't use any positional arguments, but `quiet` is used
		{doc: "local(quiet=True,)", char: 17, expected: []string{"command=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},

		// started to complete a keyword argument
		{doc: "local(quiet=True,command)", char: 24, expected: []string{"command=", "command_bat="}},

		// not in an argument context
		{doc: "local(quiet=True,command=)", char: 25, expected: []string{"docker_build", "local"}},
		{doc: "local(quiet=True,command=c)", char: 25, expected: []string{}},

		{doc: customFn, line: 4, char: 3, expected: []string{"a=", "b=", "c=", "fn", "docker_build", "local"}},
		{doc: customFn, line: 5, char: 7, expected: []string{"a=", "c=", "fn", "docker_build", "local"}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newFixture(t)
			f.ParseBuiltins(functionFixture)

			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, protocol.Position{Line: tt.line, Character: tt.char})
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestMemberCompletion(t *testing.T) {
	f := newFixture(t)
	_ = WithStarlarkBuiltins()(f.a)

	tests := []struct {
		doc        string
		line, char uint32
		expected   []string
	}{
		{doc: "pr", char: 2, expected: []string{"print"}},
		{doc: "pr.end", char: 6, expected: []string{"endswith"}},
		{doc: `"".isa`, char: 5, expected: []string{"isalnum", "isalpha"}},
		{doc: `[].ex`, char: 5, expected: []string{"extend"}},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, protocol.Position{Line: tt.line, Character: tt.char})
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestTypedMemberCompletion(t *testing.T) {
	f := newFixture(t)
	_ = WithStarlarkBuiltins()(f.a)

	f.builtins.Functions["foo"] = query.Signature{
		Name:       "foo",
		ReturnType: "str",
	}
	f.builtins.Functions["bar"] = query.Signature{
		Name:       "bar",
		ReturnType: "None",
	}

	tests := []struct {
		doc        string
		line, char uint32
		expected   []string
	}{
		{doc: `"".c`, char: 4, expected: []string{"capitalize", "codepoint_ords", "count", "codepoints"}},
		{doc: `[].c`, char: 4, expected: []string{"clear"}},
		{doc: `{}.i`, char: 4, expected: []string{"items"}},
		{doc: `s = ""
s.c`, line: 1, char: 3, expected: []string{"capitalize", "codepoint_ords", "count", "codepoints"}},
		{doc: `s = []
s.c`, line: 1, char: 3, expected: []string{"clear"}},
		{doc: `s = {}
s.i`, line: 1, char: 3, expected: []string{"items"}},
		{doc: `foo().c`, char: 7, expected: []string{"capitalize", "codepoint_ords", "count", "codepoints"}},
		{doc: `bar().`, char: 6, expected: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, protocol.Position{Line: tt.line, Character: tt.char})
			assertCompletionResult(t, tt.expected, result)
		})
	}
}
