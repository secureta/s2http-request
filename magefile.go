// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	binaryDir       = "bin"
	dispatcherBinary = filepath.Join(binaryDir, "dispatcher")
	schemaBinary    = filepath.Join(binaryDir, "schema")
	version         = getVersion()
	ldflags         = "-ldflags=-X main.version=" + version
)

func getVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

func ensureBinDir() error {
	return os.MkdirAll(binaryDir, 0755)
}

// Default target
func Default() error {
	return Build()
}

// Help prints available mage targets
func Help() {
	fmt.Println("Simple Request Dispatcher - Magefile")
	fmt.Println("")
	fmt.Println("Available targets:")
	fmt.Println("  Default       Build")
	fmt.Println("  Help          Show this help message")
	fmt.Println("  Build         Build all binaries")
	fmt.Println("  Deps          Install Go dependencies")
	fmt.Println("  Test          Run all tests")
	fmt.Println("  TestCoverage  Run tests with coverage report")
	fmt.Println("  Clean         Clean build artifacts")
	fmt.Println("  Install       Install binaries to GOPATH/bin")
	fmt.Println("  Schema        Generate JSON schema file")
	fmt.Println("  RunExample    Run dispatcher with example request")
	fmt.Println("  DevBuild      Build with race detector for development")
	fmt.Println("  Lint          Run golangci-lint")
	fmt.Println("  Fmt           Format Go code")
	fmt.Println("  Security      Run gosec security scanner")
	fmt.Println("  BuildAll      Build for multiple platforms")
	fmt.Println("  Version       Show version information")
}

func Build() error {
	if err := ensureBinDir(); err != nil {
		return err
	}
	fmt.Println("Building dispatcher...")
	if err := runCommand("go", "build", ldflags, "-o", dispatcherBinary, "./cmd/dispatcher"); err != nil {
		return err
	}
	fmt.Println("Building schema generator...")
	return runCommand("go", "build", ldflags, "-o", schemaBinary, "./cmd/schema")
}

func Deps() error {
	fmt.Println("Installing dependencies...")
	if err := runCommand("go", "mod", "download"); err != nil {
		return err
	}
	return runCommand("go", "mod", "tidy")
}

func Test() error {
	fmt.Println("Running tests...")
	return runCommand("go", "test", "-v", "./...")
}

func TestCoverage() error {
	fmt.Println("Running tests with coverage...")
	if err := runCommand("go", "test", "-v", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	if err := runCommand("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return err
	}
	fmt.Println("Coverage report generated: coverage.html")
	return nil
}

func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	os.RemoveAll(binaryDir)
	os.Remove("coverage.out")
	os.Remove("coverage.html")
	os.Remove("dispatcher")
	os.Remove("schema")
	os.Remove("*.prof")
	os.Remove("*.pprof")
	return nil
}

func Install() error {
	if err := Build(); err != nil {
		return err
	}
	fmt.Println("Installing binaries...")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("GOPATH environment variable is not set")
	}
	if err := runCommand("cp", dispatcherBinary, filepath.Join(gopath, "bin", "dispatcher")); err != nil {
		return err
	}
	return runCommand("cp", schemaBinary, filepath.Join(gopath, "bin", "schema"))
}

func Schema() error {
	if _, err := os.Stat(schemaBinary); os.IsNotExist(err) {
		if err := Build(); err != nil {
			return err
		}
	}
	fmt.Println("Generating JSON schema...")
	return runCommand(schemaBinary, "--output", "request-schema.json")
}

func RunExample() error {
	if _, err := os.Stat(dispatcherBinary); os.IsNotExist(err) {
		if err := Build(); err != nil {
			return err
		}
	}
	fmt.Println("Running example...")
	return runCommand(dispatcherBinary, "--help")
}

func DevBuild() error {
	if err := ensureBinDir(); err != nil {
		return err
	}
	fmt.Println("Building with race detector...")
	if err := runCommand("go", "build", "-race", ldflags, "-o", dispatcherBinary, "./cmd/dispatcher"); err != nil {
		return err
	}
	return runCommand("go", "build", "-race", ldflags, "-o", schemaBinary, "./cmd/schema")
}

func Lint() error {
	fmt.Println("Running linter...")
	return runCommand("golangci-lint", "run")
}

func Fmt() error {
	fmt.Println("Formatting code...")
	return runCommand("go", "fmt", "./...")
}

func Security() error {
	fmt.Println("Running security scanner...")
	return runCommand("gosec", "./...")
}

func BuildAll() error {
	if err := ensureBinDir(); err != nil {
		return err
	}
	fmt.Println("Building for multiple platforms...")
	platforms := []struct {
		OS   string
		Arch string
		Out  string
	}{
		{"linux", "amd64", "dispatcher-linux-amd64"},
		{"linux", "amd64", "schema-linux-amd64"},
		{"darwin", "amd64", "dispatcher-darwin-amd64"},
		{"darwin", "amd64", "schema-darwin-amd64"},
		{"darwin", "arm64", "dispatcher-darwin-arm64"},
		{"darwin", "arm64", "schema-darwin-arm64"},
		{"windows", "amd64", "dispatcher-windows-amd64.exe"},
		{"windows", "amd64", "schema-windows-amd64.exe"},
	}
	for _, p := range platforms {
		if err := runCommand("go", "build", ldflags, "-o", filepath.Join(binaryDir, p.Out), "-GOOS="+p.OS, "-GOARCH="+p.Arch, "./cmd/dispatcher"); err != nil {
			return err
		}
		if err := runCommand("go", "build", ldflags, "-o", filepath.Join(binaryDir, strings.Replace(p.Out, "dispatcher", "schema", 1)), "-GOOS="+p.OS, "-GOARCH="+p.Arch, "./cmd/schema"); err != nil {
			return err
		}
	}
	return nil
}

func Version() error {
	fmt.Println("Version:", version)
	return runCommand("go", "version")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
