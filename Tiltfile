src_dirs = ['cmd', 'internal', 'pkg']

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
    # builtins.extend(listdir('../tilt.build/api/modules', recursive=True))

def lsp_args():
    args = ['--address=127.0.0.1:8760']
    args.extend(['--builtin-paths='+b for b in builtins])
    return ' '.join(args)

serve_cmd = """while [ $? -eq 0 ]; do
  go run ./cmd/starlark-lsp --debug --verbose start %s
done""" % lsp_args()
local_resource('run', serve_cmd=serve_cmd, deps=src_dirs)

make('test', resource_deps=['run'])

make(['fmt', 'lint', 'tidy', 'install'],
     src_dirs + ['go.mod', 'go.sum'],
     resource_deps=['test'])