package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"starshell/star"
)

// ANSI color codes
const (
	Reset        = "\033[0m"
	BoldRed      = "\033[1;31m"
	DefaultRed   = "\033[0;31m"
	DefaultGreen = "\033[0;32m"
)

// Config structure for customization
type Config struct {
	PromptColors map[string]string `json:"prompt_colors"` // Colors for prompt components
	FileColors   map[string]string `json:"file_colors"`   // Colors for file types
	PromptFormat string            `json:"prompt_format"` // Format of the prompt
	Aliases      map[string]string `json:"aliases"`       // Command aliases
}

// Global configuration variable
var config Config

// LoadConfig reads the config.json file
func LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

// Get color from config or use default
func getColor(name, fallback string) string {
	color, exists := config.PromptColors[name]
	if exists {
		return color
	}
	return fallback
}

// Get current working directory
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// Get current time
func getCurrentTime() string {
	return time.Now().Format("15:04:05")
}

// Get username and hostname
func getUserAndHost() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME") // Windows support
	}
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	return fmt.Sprintf("%s@%s", user, host)
}

// Replace home directory with ~
func replaceHomeWithTilde(path string) string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows support
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// Generate the prompt
func generatePrompt() string {
	userHost := getUserAndHost()
	currentDir := replaceHomeWithTilde(getCurrentDirectory())
	currentTime := getCurrentTime()

	// Add colors to components
	user := getColor("user", "\033[1;33m") + strings.Split(userHost, "@")[0] + Reset
	host := getColor("host", "\033[1;34m") + strings.Split(userHost, "@")[1] + Reset
	path := getColor("path", "\033[1;32m") + currentDir + Reset
	time := getColor("time", "\033[1;35m") + currentTime + Reset

	// Replace placeholders in the prompt format
	prompt := config.PromptFormat
	prompt = strings.ReplaceAll(prompt, "{user}", user)
	prompt = strings.ReplaceAll(prompt, "{host}", host)
	prompt = strings.ReplaceAll(prompt, "{path}", path)
	prompt = strings.ReplaceAll(prompt, "{time}", time)

	return prompt + " $ "
}

// Color files based on type
func getFileColor(file os.DirEntry) string {
	if file.IsDir() {
		return getColor("directory", "\033[1;34m") // Directories
	}

	// Check for symlinks
	if fileInfo, err := file.Info(); err == nil && (fileInfo.Mode()&os.ModeSymlink != 0) {
		return getColor("symlink", "\033[1;33m") // Symlinks
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Name()))
	switch ext {
	case ".zip", ".tar", ".gz":
		return getColor("compressed", "\033[1;35m") // Compressed files
	case ".mp4", ".jpg", ".png":
		return getColor("media", "\033[1;36m") // Media files
	default:
		if isExecutable(file.Name()) {
			return getColor("executable", "\033[1;32m") // Executables
		}
		return "" // Default file color
	}
}

// Custom ls command
func customLs() {
	dir := getCurrentDirectory()
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(BoldRed + "[ERROR] Error reading directory: " + err.Error() + Reset)
		return
	}

	for _, file := range files {
		color := getFileColor(file)
		fmt.Println(color + file.Name() + Reset)
	}
}

// Check if a file is executable
func isExecutable(filename string) bool {
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(filename, ".exe") || strings.HasSuffix(filename, ".bat")
	}
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// Change directory
func changeDirectory(path string) error {
	err := os.Chdir(path)
	if err != nil {
		return fmt.Errorf("[ERROR] Cannot change directory: %v", err)
	}
	return nil
}
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default: // Unix-like systems
		fmt.Print("\033[H\033[2J")
	}
}

// Get the home directory based on the OS
func getHomeDirectory() string {
	home := os.Getenv("HOME")
	if home == "" {
		// For Windows, use USERPROFILE
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		// If no HOME or USERPROFILE found, return the current directory
		return "."
	}
	return home
}

func executeCommand(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	// Check for alias
	if alias, exists := config.Aliases[args[0]]; exists {
		args = append(strings.Fields(alias), args[1:]...)
	}

	switch args[0] {
	case "cd":
		dir := "."
		if len(args) > 1 {
			dir = args[1]
		}
		if dir == "~" {
			dir = getHomeDirectory()
		}

		err := changeDirectory(dir)
		if err != nil {
			fmt.Println(BoldRed + err.Error() + Reset)
		}
		return

	case "ls", "dir":
		customLs()
		return

	case "clear", "cls":
		clearScreen()
		return

	case "star":
		if len(args) < 2 {
			fmt.Println(BoldRed + "[ERROR] Missing subcommand. Use 'star install user/repo' or other commands." + Reset)
			return
		}

		switch args[1] {
		case "install":
			if len(args) < 3 {
				fmt.Println(BoldRed + "[ERROR] Missing repository argument. Use 'star install user/repo'." + Reset)
				return
			}
			repo := args[2]
			fmt.Println(DefaultGreen + "Installing " + repo + "..." + Reset)
			err := star.Install(star.Package{User: strings.Split(repo, "/")[0], Repo: strings.Split(repo, "/")[1]})
			if err != nil {
				fmt.Println(BoldRed + "[ERROR] Installation failed: " + err.Error() + Reset)
			} else {
				fmt.Println(DefaultGreen + "[SUCCESS] " + repo + " installed successfully!" + Reset)
			}

		case "list":
			installed, err := star.ListInstalledStars()
			if err != nil {
				fmt.Println(BoldRed + "[ERROR] Could not list installed packages: " + err.Error() + Reset)
				return
			}
			for _, pkg := range installed {
				fmt.Printf(DefaultGreen+"- %s/%s@%s\n"+Reset, pkg.User, pkg.Repo, pkg.Version)
			}

		case "uninstall":
			if len(args) < 3 {
				fmt.Println(BoldRed + "[ERROR] Missing repository argument. Use 'star uninstall user/repo'." + Reset)
				return
			}
			repo := args[2]
			fmt.Println(DefaultRed + "Uninstalling " + repo + "..." + Reset)
			err := star.Uninstall(star.Package{User: strings.Split(repo, "/")[0], Repo: strings.Split(repo, "/")[1]})
			if err != nil {
				fmt.Println(BoldRed + "[ERROR] Uninstallation failed: " + err.Error() + Reset)
			} else {
				fmt.Println(DefaultGreen + "[SUCCESS] " + repo + " uninstalled successfully!" + Reset)
			}

		default:
			fmt.Println(BoldRed + "[ERROR] Unknown 'star' subcommand." + Reset)
		}
		return
	}

	// External commands
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Println(BoldRed+"[ERROR] "+Reset, err)
	}
}

// Main function
func main() {
	// Load configuration
	err := LoadConfig("config.json")
	if err != nil {
		fmt.Println(BoldRed + "[ERROR] Failed to load config: " + err.Error() + Reset)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		// Display the prompt
		fmt.Print(generatePrompt())

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, BoldRed+"[ERROR] "+Reset, err)
			continue
		}
		input = strings.TrimSpace(input)

		// Exit on "exit"
		if input == "exit" {
			break
		}

		// Execute the command
		executeCommand(input)
	}
}
