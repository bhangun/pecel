# Pecel CLI

A powerful CLI tool to recursively combine file contents from directories and subdirectories with an enhanced interactive mode and comprehensive filtering options.

## 🚀 Features

- **Interactive Mode**: When run without arguments, Pecel enters an interactive mode prompting for all options
- **Recursive File Combination**: Combines files from directories and subdirectories
- **Multiple Output Formats**: Text, JSON, XML, and Markdown
- **Flexible Filtering**: Filter by file extensions, size, patterns, and more
- **Parallel Processing**: Process multiple files simultaneously for faster performance
- **Compression Support**: Optional GZIP compression for output
- **Configuration Files**: Load settings from JSON configuration files
- **Progress Indicators**: Real-time progress for large operations
- **Cross-Platform**: Works on Linux, macOS, and Windows

## 📦 Installation

### Automated Installation (Recommended)

The easiest way to install Pecel is using the automated installation scripts:

**macOS/Linux (using curl):**
```bash
curl -sSL https://raw.githubusercontent.com/bhangun/pecel/main/install.sh | bash
```

**Windows (using PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/bhangun/pecel/main/install.ps1 | iex
```

These scripts will automatically:
- Detect your operating system and architecture
- Download the appropriate binary from the latest GitHub release
- Verify the checksum for security
- Install the binary to the correct location
- Add it to your PATH if needed

### Package Managers

For users who prefer package managers:

**Homebrew (macOS/Linux):**
```bash
brew install bhangun/pecel/pecel
```

**Chocolatey (Windows):**
```powershell
choco install pecel
```

### From Source
```bash
git clone https://github.com/bhangun/pecel.git
cd pecel
make build
# Binary will be available at bin/pecel
./bin/pecel --help
```

### Manual Installation
```bash
# Build from source
make build
# The binary will be available at bin/pecel
# Make it executable and move to your PATH
chmod +x bin/pecel
sudo mv bin/pecel /usr/local/bin/
```

## 🛠 Usage

### Interactive Mode (Recommended)
Simply run Pecel without any arguments to enter interactive mode:
```bash
./pecel
```

### Command Line Mode
```bash
# Basic usage
pecel -i ./src -o combined.txt

# Filter by file extensions
pecel -ext .go,.js,.py -o output.txt

# Multiple filters
pecel -i ./src -ext .go,.md --min-size 100 --max-size 1000000

# JSON output with compression
pecel --format json --compress --output output.json.gz

# Parallel processing
pecel --parallel 4 --verbose

# Exclude patterns
pecel --exclude "\.git|node_modules|\.DS_Store"

# Configuration file
pecel --config config.json

# Dry run to see what would be processed
pecel --dry-run --verbose
```


### Available Options

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `--input` | `-i` | Input directory path (default: current directory) |
| `--output` | `-o` | Output file path (default: combined.txt) |
| `--ext` | | Comma-separated list of file extensions to include |
| `--exclude-hidden` | `-eh` | Exclude hidden files and directories (default: true) |
| `--max-size` | | Maximum file size in bytes (0 = unlimited) |
| `--min-size` | | Minimum file size in bytes |
| `--exclude` | | Regex pattern to exclude files |
| `--include` | | Regex pattern to include files |
| `--format` | | Output format: text, json, xml, markdown (default: text) |
| `--compress` | | Compress output with gzip |
| `--parallel` | | Number of files to process in parallel (default: 1) |
| `--dry-run` | | Show what would be processed without writing |
| `--quiet` | | Suppress non-essential output |
| `--verbose` | | Show detailed progress |
| `--config` | | Load configuration from JSON file |
| `--version` | `-v` | Show version information |
| `--help` | `-h` | Show help message |

## 📁 Sample Configuration File (config.json)

```json
{
  "input_dir": "./src",
  "output_file": "combined.txt",
  "extensions": [".go", ".js", ".py"],
  "exclude_hidden": true,
  "max_file_size": 1000000,
  "min_file_size": 0,
  "exclude_pattern": "\\.git|node_modules",
  "include_pattern": "",
  "output_format": "text",
  "compress": false,
  "parallel": 4,
  "quiet": false,
  "verbose": true,
  "dry_run": false
}
```

## 🚀 Deployment

Pecel uses [JReleaser](https://jreleaser.org/) for automated releases and distribution to package managers:

- **Homebrew**: Releases are automatically published to the custom tap at `bhangun/pecel`
- **Chocolatey**: Windows packages are automatically published to the Chocolatey Community Repository
- **GitHub Releases**: Binaries for all platforms are attached to each release with checksums
- **Automatic Signing**: All binaries are signed for security

### Homebrew Installation
After publication, users can install via Homebrew:
```bash
brew tap bhangun/pecel
brew install pecel
```

Or in a single command:
```bash
brew install bhangun/pecel/pecel
```

### Chocolatey Installation
On Windows, users can install via Chocolatey:
```powershell
choco install pecel
```

The deployment workflow is triggered on tagged commits and handles:
1. Cross-compilation for all supported platforms (Linux AMD64/ARM64, macOS AMD64/ARM64, Windows AMD64/ARM64)
2. Binary signing and checksum generation
3. Publication to Homebrew and Chocolatey package managers
4. GitHub release creation with assets
5. Automatic formula/package updates

## 🏗 Development

```bash
# Clone repository
git clone https://github.com/bhangun/pecel.git
cd pecel

# Build
make build

# Run tests
make test

# Build for all platforms
make cross-compile

# Run demo
make demo
```

## 🎯 Use Cases

- **AI Context Gathering**: Combine source code files for LLM context
- **Documentation Aggregation**: Merge documentation files into a single document
- **Code Review Preparation**: Bundle related files for review
- **Backup Operations**: Consolidate files from multiple directories
- **Data Analysis**: Combine structured data files for processing

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT

## 🆘 Support

For support, please open an issue in the GitHub repository.