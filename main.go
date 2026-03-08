package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const aurBase = "https://aur.archlinux.org/rpc/v5"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("ARTS Project 1.0 'Another Ridiculous Time Saver'")
		fmt.Println("Usage:")
		fmt.Println("  arts draw <package>   — install a package from AUR")
		fmt.Println("  arts erase <package>  — remove a package")
		fmt.Println("  arts repaint <package> - updates a package")
		fmt.Println("  arts redraw <package> -reinstalls a package")
		fmt.Println("  arts try <Command> <package> <number> -tries to install package n.times")
		os.Exit(1)
	}

	command := os.Args[1]
	pkg := os.Args[2]

	switch command {
	case "draw":
		if err := draw(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
			os.Exit(1)
		}
	case "erase":
		if err := erase(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", err)
			os.Exit(1)
		}
	case "redraw":
		if err := redraw(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "\033{31m✗ %s\033[0m\n", err)
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
			fmt.Fprint(os.Stderr, "usage: arts try <command> <package> [times]\n")
			os.Exit(1)
		}
		sub := os.Args[2]
		pkg := os.Args[3]
		times := 0
		if len(os.Args) >= 5 {
			fmt.Sscanf(os.Args[4], "%d", &times)
		}
		if err := try(sub, pkg, times); err != nil {
			fmt.Fprintf(os.Stderr, "\033[32m✗ %s\033[0m\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown brush '%s'\n", command)
		os.Exit(1)
	}
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

func checkExists(name string) error {
	endpoint := fmt.Sprintf("%s/info?arg[]=%s", aurBase, url.QueryEscape(name))
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("couldn't get in AUR Gallery: %w", err)
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
	attempt := 1
	for {
		if times > 0 && attempt > times {
			return fmt.Errorf("gave up after %d attempts", times)
		}
		fmt.Printf("\033[34m Attempt %d...\033[0m\n", attempt)
		var err error
		switch sub {
		case "drawing":
			err = draw(name)
		case "erasing":
			err = erase(name)
		case "repainting":
			err = repaint(name)
		case "redrawing":
			err = redraw(name)
		default:
			return fmt.Errorf("unknown sub-brush '%s'", sub)
		}
		fmt.Printf("\033[32m failed: %s - retrying in 3s...\033[0m\n", err)
		attempt++
		time.Sleep(3 * time.Second)
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
