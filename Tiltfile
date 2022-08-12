src_dirs = ['cmd', 'pkg']

def make(target, deps=src_dirs, resource_deps=[], **kwargs):
    cmd = ['make', target]
    if type(target) == 'list':
        cmd = ['make']
        cmd.extend(target)
        target = '-'.join(target)
    local_resource(target, cmd, deps=deps, resource_deps=resource_deps)

local_resource(
    'run',
    cmd="tilt dump api-docs",
    serve_cmd="go run ./cmd/starlark-lsp --debug --verbose start --address=127.0.0.1:8760 --builtin-paths=api",
    deps=src_dirs
)

make('test', resource_deps=['run'])

make(
    ['fmt', 'lint', 'tidy', 'install'],
    deps=src_dirs + ['go.mod', 'go.sum'],
    resource_deps=['run']
)
