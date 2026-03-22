package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/UnitVectorY-Labs/gowebshot/internal/capture"
	"github.com/UnitVectorY-Labs/gowebshot/internal/cli"
	"github.com/UnitVectorY-Labs/gowebshot/internal/output"
	"github.com/UnitVectorY-Labs/gowebshot/internal/tui"
)

var Version = "dev" // This will be set by the build systems to the release version

func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	cfg, isInteractive, showVersion, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if showVersion {
		fmt.Println(Version)
		return
	}

	if isInteractive {
		fmt.Fprintln(os.Stderr, "Running in interactive TUI mode...")
		if err := tui.Run(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "Running in non-interactive mode...")

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	data, err := capture.Capture(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Capture error: %v\n", err)
		os.Exit(1)
	}

	path, err := output.ResolvePath(cfg.Dir, cfg.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Output path error: %v\n", err)
		os.Exit(1)
	}

	if err := output.WriteFile(path, data); err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Screenshot saved to %s\n", path)
}
