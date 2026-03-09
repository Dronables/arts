package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

const aurBase = "https://aur.archlinux.org/rpc/v5"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
		case "draw":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: arts draw <package>")
				os.Exit(1)
			}
			if err := draw(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
				os.Exit(1)
			}

		case "erase":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: arts erase <package>")
				os.Exit(1)
			}
			if err := erase(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
				os.Exit(1)
			}

		case "redraw":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: arts redraw <package>")
				os.Exit(1)
			}
			if err := redraw(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
				os.Exit(1)
			}

		case "repaint":
			pkg := ""
			if len(os.Args) >= 3 {
				pkg = os.Args[2]
			}
			if err := repaint(pkg); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
				os.Exit(1)
			}

		case "try":
			if len(os.Args) < 4 {
				fmt.Fprintln(os.Stderr, "usage: arts try <command> <package> [times]")
				fmt.Fprintln(os.Stderr, "  commands: drawing, erasing, repainting, redrawing")
				os.Exit(1)
			}
			sub := os.Args[2]
			pkg := os.Args[3]
			times := 0
			if len(os.Args) >= 5 {
				fmt.Sscanf(os.Args[4], "%d", &times)
			}
			if err := try(sub, pkg, times); err != nil {
				fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
				os.Exit(1)
			}

		default:
			fmt.Fprintf(os.Stderr, "unknown brush '%s'\n", command)
			printHelp()
			os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("ARTS Project 1.0 'Another Ridiculous Time Saver'")
	fmt.Println("Usage:")
	fmt.Println("  arts draw <package>              — install a package from AUR")
	fmt.Println("  arts erase <package>             — remove a package")
	fmt.Println("  arts repaint [package]           — update a package (or all)")
	fmt.Println("  arts redraw <package>            — reinstall a package")
	fmt.Println("  arts try <command> <package> [n] — retry command n times (unlimited if omitted)")
	fmt.Println("    commands: drawing, erasing, repainting, redrawing")
}

func draw(name string) error {
	fmt.Printf("\033[35m🎨 Drawing '%s' onto your canvas...\033[0m\n\n", name)
	if err := checkExists(name); err != nil {
		return err
	}
	dir := filepath.Join(os.TempDir(), "arts-build", name)
	_ = os.RemoveAll(dir)
	fmt.Println("\033[34m  Drawing from AUR...\033[0m")
	if err := run("git", "clone", "https://aur.archlinux.org/"+name+".git", dir); err != nil {
		return fmt.Errorf("drawing failed: %w", err)
	}
	fmt.Println("\033[34m  Crafting and sculpting...\033[0m")
	cmd := exec.Command("makepkg", "-si", "--noconfirm")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("crafting failed: %w", err)
	}
	fmt.Printf("\n\033[32m✓ '%s' drawn successfully!\033[0m\n", name)
	return nil
}

func erase(name string) error {
	fmt.Printf("\033[35m🎨 Erasing '%s' from the canvas...\033[0m\n\n", name)
	if err := run("sudo", "pacman", "-R", "--noconfirm", name); err != nil {
		return fmt.Errorf("erase failed: %w", err)
	}
	fmt.Printf("\n\033[32m✓ '%s' erased.\033[0m\n", name)
	return nil
}

func redraw(name string) error {
	fmt.Printf("\033[35m🎨 Redrawing '%s' onto your canvas...\033[0m\n\n", name)
	if err := erase(name); err != nil {
		return err
	}
	return draw(name)
}

func repaint(name string) error {
	if name == "" {
		fmt.Println("\033[35m🎨 Repainting the whole canvas...\033[0m")
		return run("sudo", "pacman", "-Syu")
	}
	fmt.Printf("\033[35m🎨 Repainting '%s'...\033[0m\n\n", name)
	return draw(name)
}

func try(sub, name string, times int) error {
	var action func(string) error
	switch sub {
		case "drawing":
			action = draw
		case "erasing":
			action = erase
		case "repainting":
			action = repaint
		case "redrawing":
			action = redraw
		default:
			return fmt.Errorf("unknown sub-brush '%s' — use: drawing, erasing, repainting, redrawing", sub)
	}

	for attempt := 1; ; attempt++ {
		if times > 0 && attempt > times {
			return fmt.Errorf("gave up after %d attempt(s)", times)
		}

		if times > 0 {
			fmt.Printf("\033[34m🖌  Attempt %d/%d...\033[0m\n", attempt, times)
		} else {
			fmt.Printf("\033[34m🖌  Attempt %d (unlimited)...\033[0m\n", attempt)
		}

		if err := action(name); err != nil {
			fmt.Printf("\033[33m  ✗ Failed: %s — retrying immediately...\033[0m\n\n", err)
			continue
		}

		// success
		fmt.Printf("\033[32m✓ '%s' succeeded on attempt %d!\033[0m\n", name, attempt)
		return nil
	}
}

func checkExists(name string) error {
	endpoint := fmt.Sprintf("%s/info?arg[]=%s", aurBase, url.QueryEscape(name))
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("couldn't reach AUR Gallery: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		ResultCount int `json:"resultcount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("bad AUR response: %w", err)
	}
	if result.ResultCount == 0 {
		return fmt.Errorf("painting '%s' not found in the AUR", name)
	}
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
