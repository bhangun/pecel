# Pecel CLI - AI Agent Instructions

This document provides instructions for AI agents (like GitHub Copilot, Cursor AI, Claude, etc.) to effectively work with the Pecel CLI project.

## Project Overview

The Pecel CLI is a Go-based command-line tool that recursively combines file contents from directories and subdirectories, separating each file with metadata headers. It features an interactive mode, advanced filtering, and multiple output formats.

## Project Structure

```
pecel/
├── cmd/
│   └── main/
│       └── main.go        # Main application logic
├── go.mod                 # Go module definition
├── Makefile               # Build automation
├── install.sh             # Linux/macOS installer
├── install.ps1            # Windows installer
├── scripts/
│   ├── install-online.sh  # Online installer script
│   └── install.ps1        # Windows installer script
├── .github/
│   └── workflows/
│       └── release.yaml   # CI/CD pipeline with JReleaser
├── config/
│   └── config.json        # Sample configuration
├── completions/
│   ├── bash/
│   │   └── pecel          # Bash completion
│   └── zsh/
│       └── _pecel         # Zsh completion
├── AGENTS.md              # This file
├── README.md              # User documentation
└── Dockerfile             # Container build
```

## Development Guidelines for AI Agents

### 1. Code Style & Standards

- **Go Code**: Follow Go standard library conventions
- **Error Handling**: Use explicit error handling, avoid panics
- **Logging**: Use fmt for CLI output, structured logging for internal
- **Naming**: Use descriptive names, camelCase for variables/functions
- **Comments**: Document exported functions and complex logic
- **User Experience**: Prioritize clear, helpful error messages

### 2. Testing Requirements

When modifying code:
- Write unit tests for new functions
- Test edge cases (empty directories, large files, permission issues)
- Verify cross-platform compatibility
- Test installation scripts on target platforms
- Ensure interactive mode works correctly

### 3. CLI Interface Standards

- Use `flag` package for CLI arguments
- Provide clear help text with examples
- Include `--version` flag
- Support both short and long flags where appropriate
- Validate input parameters early
- Implement interactive mode when no flags are provided

### 4. Security Considerations

- Validate file paths to prevent traversal attacks
- Set appropriate file permissions
- Sanitize user input in install scripts
- Use HTTPS for all downloads
- Verify checksums when downloading binaries
- Limit file size to prevent memory exhaustion

### 5. Cross-Platform Compatibility

- Use `filepath` package instead of `path`
- Handle different line endings (CRLF vs LF)
- Consider path length limitations on Windows
- Support both POSIX and Windows environments
- Test on all target platforms

## JReleaser Deployment Configuration

The project uses JReleaser for automated releases and distribution to package managers:

### Configuration File
- Located at `.jreleaser/config.yaml` (or similar)
- Defines release workflow for GitHub, Homebrew, and Chocolatey
- Specifies artifact packaging, signing, and publishing

### JReleaser Configuration Example
```yaml
project:
  name: pecel
  description: A powerful CLI tool to recursively combine file contents from directories
  links:
    homepage: https://github.com/bhangun/pecel
  authors:
    - Bhangun
  license: MIT

release:
  github:
    owner: bhangun
    name: pecel
    overwrite: true
    prerelease: false

distributions:
  pecel:
    type: BINARY
    artifacts:
      - path: ./pecel
        platform: 'linux_amd64'
        transform: 'pecel -> {{distributionName}}'
      - path: ./pecel
        platform: 'linux_arm64'
        transform: 'pecel -> {{distributionName}}'
      - path: ./pecel
        platform: 'darwin_amd64'
        transform: 'pecel -> {{distributionName}}'
      - path: ./pecel
        platform: 'darwin_arm64'
        transform: 'pecel -> {{distributionName}}'
      - path: ./pecel.exe
        platform: 'windows_amd64'
        transform: 'pecel.exe -> {{distributionName}}.exe'

publishers:
  brew:
    active: ALWAYS
    commitAuthor:
      name: bhangun
      email: bhangun@example.com
    tap:
      active: ALWAYS
      owner: bhangun
      name: pecel
      branch: main
    formulaName: pecel
    skipTemplates: []

  chocolatey:
    active: ALWAYS
    commitAuthor:
      name: bhangun
      email: bhangun@example.com
    repository:
      active: ALWAYS
      owner: bhangun
      name: pecel-chocolatey
      branch: main
    packageName: pecel
    project:
      tags:
        license: MIT
        projectUrl: https://github.com/bhangun/pecel
        packageSourceUrl: https://github.com/bhangun/pecel
        projectSourceUrl: https://github.com/bhangun/pecel
        docsUrl: https://github.com/bhangun/pecel/blob/main/README.md
        mailingListUrl: https://github.com/bhangun/pecel/issues
        bugTrackerUrl: https://github.com/bhangun/pecel/issues
        tags: cli,go,golang,utility,tool,file-combiner
        summary: Pecel CLI - File combination utility
        description: |
          A powerful CLI tool to recursively combine file contents from directories and subdirectories.
        releaseNotesUrl: https://github.com/bhangun/pecel/releases/tag/{{tagName}}
```

### Deployment Process
1. **Tag Creation**: Create a Git tag (e.g., `v1.2.3`)
2. **CI/CD Trigger**: GitHub Actions workflow triggers on tag push
3. **Build & Test**: Cross-compile for all platforms
4. **Sign Artifacts**: Sign binaries for security
5. **Package Distribution**: Create packages for Homebrew, Chocolatey
6. **Publish**: Upload to GitHub Releases, package managers

### Homebrew Tap
- Custom tap: `bhangun/homebrew-pecel`
- Formula automatically generated by JReleaser
- Published to GitHub Releases assets
- Install command: `brew install bhangun/pecel/pecel`

### Chocolatey Package
- Package automatically generated by JReleaser
- Published to Chocolatey community repository
- Install command: `choco install pecel`

## Automated Installation

### Quick Install Scripts
The project provides one-line installation commands:

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/bhangun/pecel/main/install.sh | bash
```

**Windows:**
```powershell
iwr -useb https://raw.githubusercontent.com/bhangun/pecel/main/install.ps1 | iex
```

### Installation Script Features
- Detects OS and architecture automatically
- Downloads appropriate binary from GitHub Releases
- Verifies checksums for security
- Installs to appropriate system location
- Adds to PATH if needed

## Common Tasks & Instructions

### Adding New Features

When implementing new features:

1. **Start with user story**: Define the problem being solved
2. **Update CLI interface**: Add appropriate flags in `main.go`
3. **Implement core logic**: Keep business logic separate from CLI handling
4. **Update documentation**: Update help text, README, and install scripts
5. **Add tests**: Include unit and integration tests
6. **Update JReleaser config**: If new assets or platforms are added

### Modifying Existing Code

1. **Check dependencies**: Review `go.mod` for required changes
2. **Update tests**: Ensure existing tests pass, add new ones
3. **Verify installation**: Test that install scripts still work
4. **Cross-compile**: Build for all target platforms
5. **Update version**: Increment version in relevant files

### Debugging Issues

Common issues to watch for:

1. **File permissions**: Especially with install scripts
2. **Path handling**: Cross-platform path differences
3. **Memory usage**: When processing large files
4. **Concurrency**: If adding parallel processing
5. **Error messages**: Ensure they're user-friendly
6. **Interactive mode**: Verify prompts work correctly

## Build & Release Process

### Local Development Build

```bash
# Build for current platform
make build
# Binary will be available at bin/pecel

# Build for all platforms
make cross-compile

# Run tests
make test

# Create demo output
make demo

# Clean build artifacts
make clean
```

### Release Process

1. **Update version** in `cmd/main/main.go` and other relevant files
2. **Update JReleaser config** if needed
3. **Create tag**: `git tag v1.x.x`
4. **Push tag**: `git push origin v1.x.x`
5. **JReleaser** will automatically:
   - Build binaries for all platforms
   - Sign binaries
   - Create GitHub release
   - Publish to Homebrew and Chocolatey
   - Attach binaries to release

### Installation Script Updates

When modifying install scripts:

1. **Test on target OS**: Verify script works on Linux, macOS, Windows
2. **Check dependencies**: Ensure required tools are available
3. **Handle errors**: Provide clear error messages
4. **Support rollback**: Allow users to revert if installation fails
5. **Verify binaries**: Check SHA256 checksums

## Common Commands for AI Agents

### Code Generation Patterns

```go
// Pattern for adding new CLI flag
var newFlag = flag.String("new-flag", "default", "description")

// Pattern for recursive directory processing
err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
    // Process each file
})

// Pattern for buffered file reading
file, err := os.Open(path)
reader := bufio.NewReader(file)
buffer := make([]byte, 1024*1024) // 1MB buffer
```

### Error Handling Patterns

```go
// For file operations
if err != nil {
    return fmt.Errorf("failed to process %s: %w", filename, err)
}

// For CLI arguments
if *inputDir == "" {
    return errors.New("input directory is required")
}

// For user-friendly messages
if os.IsNotExist(err) {
    fmt.Printf("File not found: %s\n", path)
}
```

## Testing Guidelines

### Unit Tests

```go
func TestProcessFile(t *testing.T) {
    // Create temp file
    // Call function
    // Verify results
    // Cleanup
}
```

### Integration Tests

```go
func TestCombineFiles(t *testing.T) {
    // Create test directory structure
    // Run CLI command
    // Verify output file
    // Cleanup
}
```

### Installation Tests

```bash
# Test install script
./install.sh install
./install.sh update
./install.sh uninstall

# Test on different platforms if possible
```

## Performance Considerations

1. **Memory**: Use buffers for large files (currently 1MB)
2. **Concurrency**: Consider `sync.WaitGroup` for parallel processing if needed
3. **I/O**: Use buffered I/O (already implemented)
4. **CPU**: Profile if performance issues arise

## User Experience Guidelines

1. **Clear output**: Show progress for long operations
2. **Error messages**: Provide actionable error messages
3. **Help text**: Include examples in help output
4. **Defaults**: Sensible defaults for all options
5. **Validation**: Validate inputs before processing
6. **Interactive mode**: Guide users through options when no args provided

## Version Management

- Update version constant in `cmd/main/main.go`
- Update version in install scripts
- Update version in README if needed
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Update JReleaser configuration if needed

## Contributing Guidelines for AI Agents

When suggesting or implementing changes:

1. **Check existing issues**: Avoid duplicating work
2. **Follow project structure**: Maintain consistency
3. **Document changes**: Update relevant documentation
4. **Test thoroughly**: Ensure no regression
5. **Consider backwards compatibility**: Don't break existing functionality
6. **Update JReleaser config**: If new assets or platforms are added

## Emergency Fix Procedures

For critical bugs:

1. **Hotfix branch**: Create from main
2. **Minimal changes**: Fix only the critical issue
3. **Test immediately**: Verify fix works
4. **Release quickly**: Create patch release
5. **Document**: Update issue tracker and changelog

## Environment Variables

The tool supports these environment variables (if implemented):

- `PECEL_INPUT_DIR`: Default input directory
- `PECEL_OUTPUT_FILE`: Default output file
- `PECEL_EXTENSIONS`: Default file extensions

## Dependencies

Current dependencies (from go.mod):
- `github.com/fatih/color` v1.16.0: For colored output

When adding dependencies:
1. Check license compatibility
2. Consider maintenance status
3. Test on all platforms
4. Update install scripts if needed
5. Update JReleaser configuration if needed

## Platform-Specific Notes

### Linux
- Install to `/usr/local/bin/`
- Use `sudo` for system-wide installation
- Check for existing installations in PATH

### macOS
- Similar to Linux
- Consider Homebrew compatibility
- Handle SIP restrictions if any

### Windows
- Install to `%USERPROFILE%\bin\`
- Add to user PATH
- Handle `.exe` extension
- Consider PowerShell vs CMD differences
- Chocolatey package management

## Common Issues & Solutions

### Installation Fails
- Check internet connection
- Verify GitHub API rate limits not exceeded
- Check directory permissions
- Verify OS/architecture compatibility

### Tool Doesn't Run
- Check file permissions: `chmod +x pecel`
- Verify in PATH: `which pecel`
- Check dependencies: `ldd pecel` (Linux)
- Try rebuilding: `go build -o pecel .`

### Permission Errors
- Use `sudo` for system-wide install
- Check ownership of install directory
- Verify user has write permissions

## Future Enhancement Areas

Keep these in mind for potential improvements:

1. **Additional filters**: File size, date modified, content patterns
2. **Output formats**: YAML, TOML, CSV
3. **Compression**: Support for bzip2, xz output
4. **Remote sources**: Combine files from URLs/S3
5. **Database output**: Store combined content in SQLite
6. **Progress indicators**: More detailed progress for large operations
7. **Dry run mode**: Show what would be processed
8. **Exclude patterns**: Regex patterns to exclude files
9. **Parallel processing**: Process files concurrently
10. **Plugins/extensions**: Allow custom processors
11. **JReleaser integration**: Improve deployment automation

## Maintenance Tasks

Regular maintenance tasks:

1. **Update dependencies**: `go mod tidy`
2. **Run linters**: `golangci-lint run`
3. **Update install scripts**: Check for new OS versions
4. **Test on latest Go version**: Verify compatibility
5. **Review security**: Check for vulnerabilities
6. **Update JReleaser config**: Keep deployment updated

This document should be updated whenever significant changes are made to the project structure, build process, deployment configuration, or development guidelines.