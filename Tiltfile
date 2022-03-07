src_dirs = ['cmd', 'pkg']

def make(target, deps=src_dirs, resource_deps=[], **kwargs):
    cmd = ['make', target]
    if type(target) == 'list':
        cmd = ['make']
        cmd.extend(target)
        target = '-'.join(target)
    local_resource(target, cmd, deps=deps, resource_deps=resource_deps)

builtins = []
if os.path.exists('../tilt.build'):
    builtins.append('../tilt.build/api/api.py')
    builtins.append('../tilt.build/api/modules')

def lsp_args():
    args = ['--address=127.0.0.1:8760']
    args.extend(['--builtin-paths='+b for b in builtins])
    return ' '.join(args)

local_resource(
    'run',
    serve_cmd="go run ./cmd/starlark-lsp --debug --verbose start %s" % lsp_args(),
    deps=src_dirs
)

make('test', resource_deps=['run'])

make(
    ['fmt', 'lint', 'tidy', 'install'],
    deps=src_dirs + ['go.mod', 'go.sum'],
    resource_deps=['test']
)
