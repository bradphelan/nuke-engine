# CLAUDE.md

## Rules

- Never use compound commands (e.g., `cd dir && command`) or multiline commands. They stall Claude Code and trigger permission prompts. Use separate tool calls or absolute paths instead.
- For Go commands, use `go -C <dir>` to set the working directory instead of `cd`.
- Always set PATH in a single-line command if needed: use the full path to the executable instead.
