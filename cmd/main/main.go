package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

const (
	version = "0.1.0"
)

type Config struct {
	InputDir       string   `json:"input_dir"`
	OutputFile     string   `json:"output_file"`
	Extensions     []string `json:"extensions"`
	ExcludeHidden  bool     `json:"exclude_hidden"`
	MaxFileSize    int64    `json:"max_file_size"`
	MinFileSize    int64    `json:"min_file_size"`
	ExcludePattern string   `json:"exclude_pattern"`
	IncludePattern string   `json:"include_pattern"`
	OutputFormat   string   `json:"output_format"`
	Compress       bool     `json:"compress"`
	Parallel       int      `json:"parallel"`
	Quiet          bool     `json:"quiet"`
	Verbose        bool     `json:"verbose"`
	DryRun         bool     `json:"dry_run"`
}

type FileInfo struct {
	Path         string `json:"path" xml:"path"`
	Size         int64  `json:"size" xml:"size"`
	Modified     string `json:"modified" xml:"modified"`
	Content      string `json:"content,omitempty" xml:"content,omitempty"`
	RelativePath string `json:"relative_path" xml:"relative_path"`
}

type Stats struct {
	FilesProcessed int     `json:"files_processed"`
	Directories    int     `json:"directories"`
	TotalBytes     int64   `json:"total_bytes"`
	Duration       float64 `json:"duration_seconds"`
	OutputSize     int64   `json:"output_size"`
}

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
)

// Function to check if any flags were provided
func hasFlagsProvided() bool {
	return len(os.Args) > 1
}

// Function to check if any flags were explicitly set
func hasAnyFlagSet() bool {
	anySet := false
	flag.Visit(func(f *flag.Flag) {
		anySet = true
	})
	return anySet
}

// Function to validate directory path
func validateDirectory(dirPath string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("directory does not exist: %s", dirPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}
	return nil
}

// Function to validate file path
func validateFilePath(filePath string) error {
	// Check if the parent directory exists
	dir := filepath.Dir(filePath)
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("parent directory does not exist: %s", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("parent path is not a directory: %s", dir)
	}
	return nil
}

// Function to validate file extensions
func validateExtensions(extStr string) error {
	if extStr == "" {
		return nil
	}

	extensions := strings.Split(extStr, ",")
	for _, ext := range extensions {
		ext = strings.TrimSpace(ext)
		if !strings.HasPrefix(ext, ".") && ext != "*" {
			return fmt.Errorf("extension '%s' should start with a dot (.) or be '*' for all files", ext)
		}
	}
	return nil
}

// Function to prompt user for input with validation
func promptUserWithValidation(prompt string, defaultValue string, validator func(string) error) string {
	for {
		fmt.Printf("%s %s", cyan("?"), prompt)

		if defaultValue != "" {
			fmt.Printf(" (default: %s)", defaultValue)
		}
		fmt.Print(": ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = defaultValue
		}

		if validator != nil {
			if err := validator(input); err != nil {
				fmt.Printf("%s %s\n", red("✗"), err.Error())
				continue
			}
		}

		return input
	}
}

// Function to prompt user for input
func promptUser(prompt string, defaultValue string) string {
	return promptUserWithValidation(prompt, defaultValue, nil)
}

// Function to prompt user for boolean input
func promptBool(prompt string, defaultValue bool) bool {
	fmt.Printf("%s %s (Y/n)", cyan("?"), prompt)
	if defaultValue {
		fmt.Print(" [Y]: ")
	} else {
		fmt.Print(" [n]: ")
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes" || input == "true" || input == "1"
}

// Function to prompt user for selection from options
func promptSelect(prompt string, options []string, defaultValue string) string {
	fmt.Printf("%s %s\n", cyan("?"), prompt)
	for i, option := range options {
		fmt.Printf("  %d) %s", i+1, option)
		if option == defaultValue {
			fmt.Print(" (default)")
		}
		fmt.Println()
	}
	fmt.Print(": ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	// Try to parse as number
	if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= len(options) {
		return options[num-1]
	}

	// Check if input matches any option
	for _, option := range options {
		if strings.EqualFold(option, input) {
			return option
		}
	}

	// Return default if input doesn't match
	return defaultValue
}

func main() {
	// Define command line flags with short versions
	inputDir := flag.String("input", ".", "Input directory path")
	inputShort := flag.String("i", "", "Input directory path (shorthand)")
	outputFile := flag.String("output", "combined.txt", "Output file path")
	outputShort := flag.String("o", "", "Output file path (shorthand)")
	extensions := flag.String("ext", "", "Comma-separated list of file extensions to include")
	excludeHidden := flag.Bool("exclude-hidden", true, "Exclude hidden files and directories")
	excludeShort := flag.Bool("eh", true, "Exclude hidden files (shorthand)")
	maxFileSize := flag.Int64("max-size", 0, "Maximum file size in bytes (0 = unlimited)")
	minFileSize := flag.Int64("min-size", 0, "Minimum file size in bytes")
	excludePattern := flag.String("exclude", "", "Regex pattern to exclude files")
	includePattern := flag.String("include", "", "Regex pattern to include files")
	outputFormat := flag.String("format", "text", "Output format: text, json, xml, markdown")
	compress := flag.Bool("compress", false, "Compress output with gzip")
	dryRun := flag.Bool("dry-run", false, "Show what would be processed without writing")
	quiet := flag.Bool("quiet", false, "Suppress non-essential output")
	verbose := flag.Bool("verbose", false, "Show detailed progress")
	parallel := flag.Int("parallel", 1, "Number of files to process in parallel")
	versionFlag := flag.Bool("version", false, "Show version information")
	versionShort := flag.Bool("v", false, "Show version information (shorthand)")
	configFile := flag.String("config", "", "Load configuration from JSON file")

	// Parse flags early to check if any were provided
	flag.Parse()

	// Handle short flag overrides
	if *inputShort != "" {
		*inputDir = *inputShort
	}
	if *outputShort != "" {
		*outputFile = *outputShort
	}
	if !*excludeShort {
		*excludeHidden = false
	}
	if *versionShort {
		*versionFlag = true
	}

	if *versionFlag {
		fmt.Printf("pecel v%s\n", version)
		os.Exit(0)
	}

	// Check if no flags were provided and enter interactive mode
	if !hasAnyFlagSet() && len(os.Args) == 1 {
		fmt.Printf("%s Welcome to Pecel v%s - Interactive Mode\n\n", cyan("→"), version)

		// Prompt for input directory with validation
		*inputDir = promptUserWithValidation("Enter input directory path", ".", validateDirectory)

		// Prompt for output file with validation
		*outputFile = promptUserWithValidation("Enter output file path", "combined.txt", validateFilePath)

		// Prompt for file extensions with validation
		extInput := promptUserWithValidation("Enter file extensions to include (comma-separated, e.g., .go,.js,.py)", "", validateExtensions)
		if extInput != "" {
			*extensions = extInput
		}

		// Prompt for output format
		formats := []string{"text", "json", "xml", "markdown"}
		*outputFormat = promptSelect("Select output format", formats, "text")

		// Prompt for excluding hidden files
		*excludeHidden = promptBool("Exclude hidden files and directories", true)

		// Prompt for compression
		*compress = promptBool("Compress output with gzip", false)

		// Prompt for max file size
		maxSizeStr := promptUser("Maximum file size in bytes (0 for unlimited)", "0")
		if val, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil && val >= 0 {
			*maxFileSize = val
		} else {
			*maxFileSize = 0
		}

		// Prompt for exclude pattern
		excludePat := promptUser("Regex pattern to exclude files (optional)", "")
		*excludePattern = excludePat

		// Prompt for include pattern
		includePat := promptUser("Regex pattern to include files (optional)", "")
		*includePattern = includePat

		// Prompt for parallel processing with validation
		for {
			parallelStr := promptUser("Number of files to process in parallel", "1")
			if val, err := strconv.Atoi(parallelStr); err == nil && val > 0 {
				*parallel = val
				break
			} else if err != nil || val <= 0 {
				fmt.Printf("%s Parallel value must be a positive integer\n", red("✗"))
				continue
			}
		}

		// Prompt for verbose mode
		*verbose = promptBool("Enable verbose output", false)

		// Prompt for dry run
		*dryRun = promptBool("Perform dry run (show what would be processed without writing)", false)

		fmt.Println()
		fmt.Printf("%s Starting processing with your selections...\n\n", green("✓"))
	}

	// Load config file if specified
	var config Config
	if *configFile != "" {
		cfg, err := loadConfig(*configFile)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("✗"), err)
			os.Exit(1)
		}
		config = cfg
		// Override with command line flags if provided
		if *inputDir != "." {
			config.InputDir = *inputDir
		}
		if *outputFile != "combined.txt" {
			config.OutputFile = *outputFile
		}
		if *extensions != "" {
			config.Extensions = strings.Split(*extensions, ",")
		}
		// Check if the exclude-hidden flag was explicitly set
		if isFlagSet("exclude-hidden") {
			config.ExcludeHidden = *excludeHidden
		}
		if *excludePattern != "" {
			config.ExcludePattern = *excludePattern
		}
		if *includePattern != "" {
			config.IncludePattern = *includePattern
		}
		if *outputFormat != "text" {
			config.OutputFormat = *outputFormat
		}
		if *compress {
			config.Compress = *compress
		}
		if *parallel != 1 {
			config.Parallel = *parallel
		}
		if *quiet {
			config.Quiet = *quiet
		}
		if *verbose {
			config.Verbose = *verbose
		}
		if *dryRun {
			config.DryRun = *dryRun
		}
	} else {
		config = Config{
			InputDir:       *inputDir,
			OutputFile:     *outputFile,
			ExcludeHidden:  *excludeHidden,
			MaxFileSize:    *maxFileSize,
			MinFileSize:    *minFileSize,
			ExcludePattern: *excludePattern,
			IncludePattern: *includePattern,
			OutputFormat:   *outputFormat,
			Compress:       *compress,
			Parallel:       *parallel,
			Quiet:          *quiet,
			Verbose:        *verbose,
			DryRun:         *dryRun,
		}
		if *extensions != "" {
			config.Extensions = strings.Split(*extensions, ",")
		}
	}

	// Validate input directory exists
	if err := validateDirectory(config.InputDir); err != nil {
		fmt.Printf("%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Validate output file path
	if err := validateFilePath(config.OutputFile); err != nil {
		fmt.Printf("%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Validate extensions
	if err := validateExtensions(strings.Join(config.Extensions, ",")); err != nil {
		fmt.Printf("%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	startTime := time.Now()

	// Validate patterns
	var excludeRegex, includeRegex *regexp.Regexp
	if *excludePattern != "" {
		re, err := regexp.Compile(*excludePattern)
		if err != nil {
			fmt.Printf("%s Invalid exclude pattern: %v\n", red("✗"), err)
			os.Exit(1)
		}
		excludeRegex = re
	}
	if *includePattern != "" {
		re, err := regexp.Compile(*includePattern)
		if err != nil {
			fmt.Printf("%s Invalid include pattern: %v\n", red("✗"), err)
			os.Exit(1)
		}
		includeRegex = re
	}

	if !*quiet {
		fmt.Printf("%s Starting Pecel v%s\n", cyan("→"), version)
		fmt.Printf("%s Input directory: %s\n", cyan("→"), config.InputDir)
		fmt.Printf("%s Output file: %s\n", cyan("→"), config.OutputFile)
		if *dryRun {
			fmt.Printf("%s DRY RUN MODE - No files will be written\n", yellow("⚠"))
		}
	}

	// Collect file information
	var fileInfos []FileInfo
	var filePaths []string
	var stats Stats

	// Walk directory to collect files
	err := filepath.Walk(config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if !*quiet {
				fmt.Printf("%s Error accessing %s: %v\n", red("✗"), path, err)
			}
			return nil
		}

		if info.IsDir() {
			stats.Directories++
			if config.ExcludeHidden && isHidden(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply filters
		if !shouldProcessFile(path, info, config, excludeRegex, includeRegex) {
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})

	if err != nil {
		fmt.Printf("%s Error walking directory: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Printf("%s Found %d files to process\n", cyan("→"), len(filePaths))
	}

	// Process files
	if *parallel > 1 {
		fileInfos = processFilesParallel(filePaths, config.InputDir, *parallel, *verbose, *quiet, &stats)
	} else {
		fileInfos = processFilesSequential(filePaths, config.InputDir, *verbose, *quiet, &stats)
	}

	stats.Duration = time.Since(startTime).Seconds()

	// Generate output
	if !*dryRun {
		outputSize, err := writeOutput(fileInfos, config.OutputFile, *outputFormat, *compress, stats)
		if err != nil {
			fmt.Printf("%s Error writing output: %v\n", red("✗"), err)
			os.Exit(1)
		}
		stats.OutputSize = outputSize
	}

	// Print summary
	printSummary(stats, *outputFormat, *compress, *dryRun)

	if *dryRun {
		fmt.Printf("\n%s Dry run completed. %d files would be processed.\n",
			green("✓"), stats.FilesProcessed)
	} else {
		fmt.Printf("\n%s Processing completed successfully!\n", green("✓"))
	}
}

func shouldProcessFile(path string, info os.FileInfo, config Config,
	excludeRegex, includeRegex *regexp.Regexp) bool {

	// Skip hidden files
	if config.ExcludeHidden && isHidden(info.Name()) {
		return false
	}

	// Check file size limits
	if config.MaxFileSize > 0 && info.Size() > config.MaxFileSize {
		return false
	}
	if config.MinFileSize > 0 && info.Size() < config.MinFileSize {
		return false
	}

	// Check extensions
	if len(config.Extensions) > 0 {
		ext := filepath.Ext(path)
		found := false
		for _, allowedExt := range config.Extensions {
			if strings.EqualFold(ext, allowedExt) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check regex patterns
	relPath, _ := filepath.Rel(config.InputDir, path)
	if excludeRegex != nil && excludeRegex.MatchString(relPath) {
		return false
	}
	if includeRegex != nil && !includeRegex.MatchString(relPath) {
		return false
	}

	return true
}

func processFilesSequential(paths []string, baseDir string, verbose, quiet bool, stats *Stats) []FileInfo {
	var fileInfos []FileInfo

	for i, path := range paths {
		if verbose && !quiet {
			fmt.Printf("%s Processing file %d/%d: %s\n",
				cyan("↳"), i+1, len(paths), getRelativePath(path, baseDir))
		} else if !quiet && len(paths) > 10 && (i+1)%int((len(paths)/10)+1) == 0 {
			// Show progress for larger operations
			progress := float64(i+1) / float64(len(paths)) * 100
			fmt.Printf("%s Progress: %d/%d files (%.1f%%)\n",
				cyan("→"), i+1, len(paths), progress)
		}

		info, err := processSingleFile(path, baseDir)
		if err != nil {
			if !quiet {
				fmt.Printf("%s Error processing %s: %v\n", red("✗"), path, err)
			}
			continue
		}

		fileInfos = append(fileInfos, info)
		stats.FilesProcessed++
		stats.TotalBytes += info.Size

		if verbose && !quiet && (i+1)%10 == 0 {
			fmt.Printf("%s Processed %d/%d files\n", cyan("→"), i+1, len(paths))
		}
	}

	return fileInfos
}

func processFilesParallel(paths []string, baseDir string, workers int, verbose, quiet bool, stats *Stats) []FileInfo {
	var wg sync.WaitGroup
	fileChan := make(chan string, len(paths))
	resultChan := make(chan FileInfo, len(paths))
	errorChan := make(chan error, len(paths))

	var processed int32
	totalFiles := len(paths)

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for path := range fileChan {
				info, err := processSingleFile(path, baseDir)
				if err != nil {
					errorChan <- fmt.Errorf("%s: %v", path, err)
					continue
				}
				resultChan <- info

				// Update progress
				curr := atomic.AddInt32(&processed, 1)
				if verbose && !quiet && curr%10 == 0 {
					fmt.Printf("%s Worker %d: Processed %d/%d files\n",
						cyan("→"), workerID, curr, totalFiles)
				} else if !verbose && !quiet && totalFiles > 10 && int(curr)%((totalFiles/10)+1) == 0 {
					// Show overall progress for larger operations
					progress := float64(curr) / float64(totalFiles) * 100
					fmt.Printf("%s Overall progress: %d/%d files (%.1f%%)\n",
						cyan("→"), curr, totalFiles, progress)
				}
			}
		}(i)
	}

	// Send files to workers
	for _, path := range paths {
		fileChan <- path
	}
	close(fileChan)

	// Wait for workers to finish
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var fileInfos []FileInfo
	for info := range resultChan {
		fileInfos = append(fileInfos, info)
		stats.FilesProcessed++
		stats.TotalBytes += info.Size
	}

	// Report errors
	if !quiet {
		for err := range errorChan {
			fmt.Printf("%s %v\n", red("✗"), err)
		}
	}

	return fileInfos
}

func processSingleFile(path, baseDir string) (FileInfo, error) {
	info := FileInfo{
		Path:         path,
		RelativePath: getRelativePath(path, baseDir),
	}

	// Get file stats
	fileInfo, err := os.Stat(path)
	if err != nil {
		return info, err
	}

	info.Size = fileInfo.Size()
	info.Modified = fileInfo.ModTime().Format("2006-01-02 15:04:05")

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return info, err
	}

	info.Content = string(content)
	return info, nil
}

func writeOutput(fileInfos []FileInfo, outputPath, format string, compress bool, stats Stats) (int64, error) {
	var writer io.Writer

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer = file

	// Add compression if requested
	if compress {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
		outputPath += ".gz"
	}

	// Write based on format
	switch strings.ToLower(format) {
	case "json":
		return writeJSONOutput(fileInfos, writer, stats)
	case "xml":
		return writeXMLOutput(fileInfos, writer, stats)
	case "markdown", "md":
		return writeMarkdownOutput(fileInfos, writer, stats)
	default: // text
		return writeTextOutput(fileInfos, writer, stats)
	}
}

func writeTextOutput(fileInfos []FileInfo, writer io.Writer, stats Stats) (int64, error) {
	totalBytes := int64(0)
	bufWriter := bufio.NewWriter(writer)

	header := fmt.Sprintf("Pecel Output\n")
	header += fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("Files: %d | Directories: %d | Total Size: %s\n\n",
		stats.FilesProcessed, stats.Directories, formatBytes(stats.TotalBytes))

	n, _ := bufWriter.WriteString(header)
	totalBytes += int64(n)

	for _, info := range fileInfos {
		section := fmt.Sprintf("\n%s\n%s\n", strings.Repeat("=", 80), info.RelativePath)
		section += fmt.Sprintf("Size: %s | Modified: %s\n", formatBytes(info.Size), info.Modified)
		section += fmt.Sprintf("%s\n", strings.Repeat("-", 80))
		section += info.Content + "\n"
		section += fmt.Sprintf("%s\n", strings.Repeat("=", 80))

		n, _ := bufWriter.WriteString(section)
		totalBytes += int64(n)
	}

	footer := fmt.Sprintf("\n\n=== SUMMARY ===\n")
	footer += fmt.Sprintf("Files processed: %d\n", stats.FilesProcessed)
	footer += fmt.Sprintf("Directories scanned: %d\n", stats.Directories)
	footer += fmt.Sprintf("Total input size: %s\n", formatBytes(stats.TotalBytes))
	footer += fmt.Sprintf("Output size: %s\n", formatBytes(totalBytes))
	footer += fmt.Sprintf("Processing time: %.2f seconds\n", stats.Duration)

	n, _ = bufWriter.WriteString(footer)
	totalBytes += int64(n)

	bufWriter.Flush()
	return totalBytes, nil
}

func writeJSONOutput(fileInfos []FileInfo, writer io.Writer, stats Stats) (int64, error) {
	output := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated":     time.Now().Format(time.RFC3339),
			"version":       version,
			"files_count":   stats.FilesProcessed,
			"directories":   stats.Directories,
			"total_size":    stats.TotalBytes,
			"duration_secs": stats.Duration,
		},
		"files": fileInfos,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(output)
	if err != nil {
		return 0, err
	}

	// Estimate size (not exact but good enough)
	data, _ := json.Marshal(output)
	return int64(len(data)), nil
}

func writeXMLOutput(fileInfos []FileInfo, writer io.Writer, stats Stats) (int64, error) {
	type XMLOutput struct {
		XMLName   xml.Name `xml:"filecombiner_output"`
		Version   string   `xml:"version,attr"`
		Generated string   `xml:"generated,attr"`
		Metadata  struct {
			Files       int     `xml:"files"`
			Directories int     `xml:"directories"`
			TotalSize   int64   `xml:"total_size"`
			Duration    float64 `xml:"duration_seconds"`
		} `xml:"metadata"`
		Files []FileInfo `xml:"file"`
	}

	output := XMLOutput{
		Version:   version,
		Generated: time.Now().Format(time.RFC3339),
	}
	output.Metadata.Files = stats.FilesProcessed
	output.Metadata.Directories = stats.Directories
	output.Metadata.TotalSize = stats.TotalBytes
	output.Metadata.Duration = stats.Duration
	output.Files = fileInfos

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")

	// Write XML header
	writer.Write([]byte(xml.Header))

	err := encoder.Encode(output)
	if err != nil {
		return 0, err
	}

	// Estimate size
	data, _ := xml.MarshalIndent(output, "", "  ")
	return int64(len(data) + len(xml.Header)), nil
}

func writeMarkdownOutput(fileInfos []FileInfo, writer io.Writer, stats Stats) (int64, error) {
	totalBytes := int64(0)
	bufWriter := bufio.NewWriter(writer)

	header := fmt.Sprintf("# Pecel Output\n\n")
	header += fmt.Sprintf("**Generated**: %s  \n", time.Now().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("**Files**: %d | **Directories**: %d | **Total Size**: %s  \n\n",
		stats.FilesProcessed, stats.Directories, formatBytes(stats.TotalBytes))

	n, _ := bufWriter.WriteString(header)
	totalBytes += int64(n)

	for i, info := range fileInfos {
		section := fmt.Sprintf("## File %d: `%s`\n\n", i+1, info.RelativePath)
		section += fmt.Sprintf("**Size**: %s  \n", formatBytes(info.Size))
		section += fmt.Sprintf("**Modified**: %s  \n\n", info.Modified)
		section += "### Content\n```\n"
		section += info.Content + "\n```\n\n"
		section += "---\n\n"

		n, _ := bufWriter.WriteString(section)
		totalBytes += int64(n)
	}

	footer := fmt.Sprintf("## Summary\n\n")
	footer += fmt.Sprintf("- **Files processed**: %d\n", stats.FilesProcessed)
	footer += fmt.Sprintf("- **Directories scanned**: %d\n", stats.Directories)
	footer += fmt.Sprintf("- **Total input size**: %s\n", formatBytes(stats.TotalBytes))
	footer += fmt.Sprintf("- **Processing time**: %.2f seconds\n", stats.Duration)

	n, _ = bufWriter.WriteString(footer)
	totalBytes += int64(n)

	bufWriter.Flush()
	return totalBytes, nil
}

func printSummary(stats Stats, format string, compress, dryRun bool) {
	fmt.Printf("\n%s %s\n", cyan("┌"), strings.Repeat("─", 50))
	fmt.Printf("%s Processing Summary\n", cyan("│"))
	fmt.Printf("%s %s\n", cyan("├"), strings.Repeat("─", 50))
	fmt.Printf("%s Files processed:     %s\n", cyan("│"), green(strconv.Itoa(stats.FilesProcessed)))
	fmt.Printf("%s Directories scanned: %s\n", cyan("│"), green(strconv.Itoa(stats.Directories)))
	fmt.Printf("%s Total size:          %s\n", cyan("│"), green(formatBytes(stats.TotalBytes)))
	fmt.Printf("%s Processing time:     %.2f seconds\n", cyan("│"), stats.Duration)

	if !dryRun {
		fmt.Printf("%s Output format:       %s\n", cyan("│"), green(format))
		if compress {
			fmt.Printf("%s Compression:         %s\n", cyan("│"), green("gzip"))
		}
		fmt.Printf("%s Output size:         %s\n", cyan("│"), green(formatBytes(stats.OutputSize)))
		if stats.OutputSize > 0 {
			ratio := float64(stats.OutputSize) / float64(stats.TotalBytes) * 100
			fmt.Printf("%s Compression ratio:   %.1f%%\n", cyan("│"), ratio)
		}
	}
	fmt.Printf("%s %s\n", cyan("└"), strings.Repeat("─", 50))
}

func loadConfig(filename string) (Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func getRelativePath(path, baseDir string) string {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return path
	}
	return relPath
}

func isHidden(name string) bool {
	return strings.HasPrefix(name, ".") ||
		(strings.HasPrefix(name, "~") && len(name) > 1)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper function to check if a flag was explicitly set
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// Function to display help
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s Pecel v%s - Combine files recursively\n\n", cyan("📁"), version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "%s Basic Options:\n", cyan("📋"))
		fmt.Fprintf(os.Stderr, "  -i, -input string        Input directory path (default \".\")\n")
		fmt.Fprintf(os.Stderr, "  -o, -output string       Output file path (default \"combined.txt\")\n")
		fmt.Fprintf(os.Stderr, "  -ext string              Comma-separated list of file extensions\n")
		fmt.Fprintf(os.Stderr, "  -eh, -exclude-hidden     Exclude hidden files (default true)\n")

		fmt.Fprintf(os.Stderr, "\n%s Filtering Options:\n", cyan("🔍"))
		fmt.Fprintf(os.Stderr, "  -max-size int            Maximum file size in bytes (0 = unlimited)\n")
		fmt.Fprintf(os.Stderr, "  -min-size int            Minimum file size in bytes\n")
		fmt.Fprintf(os.Stderr, "  -include string          Regex pattern to include files\n")
		fmt.Fprintf(os.Stderr, "  -exclude string          Regex pattern to exclude files\n")

		fmt.Fprintf(os.Stderr, "\n%s Output Options:\n", cyan("📄"))
		fmt.Fprintf(os.Stderr, "  -format string           Output format: text, json, xml, markdown (default \"text\")\n")
		fmt.Fprintf(os.Stderr, "  -compress                Compress output with gzip\n")
		fmt.Fprintf(os.Stderr, "  -config string           Load configuration from JSON file\n")

		fmt.Fprintf(os.Stderr, "\n%s Performance Options:\n", cyan("⚡"))
		fmt.Fprintf(os.Stderr, "  -parallel int            Number of files to process in parallel (default 1)\n")

		fmt.Fprintf(os.Stderr, "\n%s Mode Options:\n", cyan("🎯"))
		fmt.Fprintf(os.Stderr, "  -dry-run                 Show what would be processed without writing\n")
		fmt.Fprintf(os.Stderr, "  -quiet                   Suppress non-essential output\n")
		fmt.Fprintf(os.Stderr, "  -verbose                 Show detailed progress\n")

		fmt.Fprintf(os.Stderr, "\n%s Information Options:\n", cyan("ℹ️"))
		fmt.Fprintf(os.Stderr, "  -v, -version             Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, -help                Show this help message\n")

		fmt.Fprintf(os.Stderr, "\n%s Examples:\n", cyan("🚀"))
		fmt.Fprintf(os.Stderr, "  %s -i ./src -o output.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -ext .go,.txt -format json -compress\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -max-size 1000000 -parallel 4 -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -exclude \"\\.git|node_modules\" -dry-run\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config config.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -v\n", os.Args[0])
	}
}
