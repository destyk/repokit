package runner

// Builtins returns all built-in runners.
func Builtins() []Runner {
	return []Runner{
		NewExecRunner("make", false),
		NewExecRunner("npm", true),
		NewExecRunner("pnpm", false),
		NewExecRunner("yarn", false),
		NewExecRunner("bun", true),
		NewExecRunner("cargo", false),
		NewExecRunner("uv", true),
		NewExecRunner("poetry", true),
		NewExecRunner("just", false),
		NewExecRunner("task", false),
		NewExecRunner("mise", true),
		NewExecRunner("go", false),
	}
}
