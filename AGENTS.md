# Agent Guidelines for Betty File Manager

## Build/Lint/Test Commands

### Build
- `make` - Build the application with version info
- `go build -ldflags="-X main.version=$(VERSION)"` - Direct build command
- `go mod tidy` - Download and organize dependencies

### Test
No test framework configured. No test files present in codebase.

### Lint
No linting tools configured. Code follows standard Go formatting.

## Code Style Guidelines

### Formatting
- Use `gofmt` for code formatting (standard Go tool)
- Follow standard Go formatting conventions
- Preserve parameter alignment in `doc.WriteString` calls (e.g., help text formatting)
  - Align description parameters to consistent column positions
  - Maintain visual spacing between keys, separators, and descriptions

### Imports
- Standard library imports first
- Third-party imports second (alphabetically sorted)
- Blank line between standard library and third-party imports
- Use aliases for imports when needed (e.g., `tea "github.com/charmbracelet/bubbletea"`)

### Naming Conventions
- **Functions/Methods**: camelCase (e.g., `writeLastd`, `ClearLastd`)
- **Types/Structs**: PascalCase (e.g., `model`, `tabData`, `selectedFile`)
- **Variables**: camelCase (e.g., `termWidth`, `currentTab`)
- **Constants**: PascalCase or ALL_CAPS as appropriate
- **Package**: lowercase single word (`main`)

### Types and Structs
- Use meaningful struct field names
- Add comments for struct fields when purpose isn't obvious
- Use appropriate Go types (`int`, `string`, `bool`, etc.)

### Error Handling
- Return errors from functions rather than panicking
- Use `log.Print` for logging errors (logs to `~/.local/log/bfm.log`)
- Check for errors with `if err != nil`
- Use `errors.Is()` for error type checking

### Comments
- Add package comments for exported functions/types
- Use `//` for single-line comments
- Document struct fields with inline comments when needed

### Dependencies
- Uses Charm ecosystem: `bubbletea`, `lipgloss`, `bubbles`
- Standard library: `os`, `path/filepath`, `strings`, etc.
- TOML config parsing with `github.com/BurntSushi/toml`

### Architecture Patterns
- Bubbletea TUI framework with Model-Update-View pattern
- Message passing for state updates
- Plugin system for extensibility
- Tab-based interface with multiple directory views</content>
<parameter name="filePath">/Users/chad/working/betty-file-manager/AGENTS.md