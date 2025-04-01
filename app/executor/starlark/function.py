def platform_system():
    runtime_goos = _runtimeGOOS()
    result = "Unknown"
    if runtime_goos == "windows":
        result = "Windows"
    if runtime_goos == "linux":
        result = "Linux"
    if runtime_goos == "darwin":
        result = "Darwin"
    return result


def subprocess_run(cmd, cwd=None, env=None):
    if not env:
        env = _osEnvironment()
    if not cwd:
        cwd = ""
    return _commandRun(args=cmd, cwd=cwd, env=env)


subprocess = struct(run=subprocess_run)
platform = struct(system=platform_system)
os = struct(environment=_osEnvironment())
