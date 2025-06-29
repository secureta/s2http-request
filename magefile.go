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
	dispatcherBinary = filepath.Join(binaryDir, "s2req")
	schemaBinary    = filepath.Join(binaryDir, "s2req-schema")
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
	fmt.Println("Building s2req...")
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
	os.Remove("s2req")
	os.Remove("s2req-schema")
	os.Remove("*.prof")
	os.Remove("*.pprof")
	return nil
}

func Install() error {
	if err := Build(); err != nil {
		return err
	}
	fmt.Println("Installing binaries...")

	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			return fmt.Errorf("GOPATH or GOBIN environment variable is not set")
		}
		gobin = filepath.Join(gopath, "bin")
	}

	if err := os.MkdirAll(gobin, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", gobin, err)
	}

	// Install main binary
	dispatcherDest := filepath.Join(gobin, "s2req")
	fmt.Printf("Installing s2req to %s\n", dispatcherDest)
	if err := runCommand("cp", dispatcherBinary, dispatcherDest); err != nil {
		return err
	}

	// Install schema binary
	schemaDest := filepath.Join(gobin, "s2req-schema")
	fmt.Printf("Installing s2req-schema to %s\n", schemaDest)
	if err := runCommand("cp", schemaBinary, schemaDest); err != nil {
		return err
	}

	return nil
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
		{"linux", "amd64", "s2req-linux-amd64"},
		{"linux", "amd64", "s2req-schema-linux-amd64"},
		{"darwin", "amd64", "s2req-darwin-amd64"},
		{"darwin", "amd64", "s2req-schema-darwin-amd64"},
		{"darwin", "arm64", "s2req-darwin-arm64"},
		{"darwin", "arm64", "s2req-schema-darwin-arm64"},
		{"windows", "amd64", "s2req-windows-amd64.exe"},
		{"windows", "amd64", "schema-windows-amd64.exe"},
	}
	for _, p := range platforms {
		if err := runCommand("go", "build", ldflags, "-o", filepath.Join(binaryDir, p.Out), "-GOOS="+p.OS, "-GOARCH="+p.Arch, "./cmd/dispatcher"); err != nil {
			return err
		}
		if err := runCommand("go", "build", ldflags, "-o", filepath.Join(binaryDir, strings.Replace(p.Out, "s2req", "s2req-schema", 1)), "-GOOS="+p.OS, "-GOARCH="+p.Arch, "./cmd/schema"); err != nil {
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
