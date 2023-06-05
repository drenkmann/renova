package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type manager struct {
	name    string
	command string
	local   bool
}

var managers = []manager{
	{
		name:    "dnf",
		command: "dnf update -y",
	},
	{
		name:    "apt",
		command: "apt update && apt upgrade",
	},
	{
		name:    "pacman",
		command: "pacman -Syu",
	},
	{
		name:    "zypper",
		command: "zypper dup",
	},
	{
		name:    "flatpak",
		command: "flatpak update -y",
	},
	{
		name:    "npm",
		command: "npm update -g",
	},
	{
		name:    "pip",
		command: "python -m pip install --upgrade pip",
	},
	{
		name:    "rustup",
		command: "rustup update",
		local:   true,
	},
}

var fail = color.RedString("[\uf00d]")
var success = color.GreenString("[\uf00c]")
var hint = color.BlackString("[?]")

func UpdateSpinner(s *spinner.Spinner, suffix string) {
	s.Suffix = fmt.Sprintf("%s %s", color.MagentaString("]"), suffix)
}

func UpdateAll() error {
	var failed []manager
	var succeeded []manager
	var found []manager

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithColor("magenta"))
	s.Prefix = color.MagentaString("[")
	s.Suffix = fmt.Sprintf("%s Loading...", color.MagentaString("]"))
	s.Start()
	s.Stop()

	local := os.Geteuid() != 0

	color.Set(color.Bold)
	if local {
		fmt.Printf("\n[?] Discovering local managers.\n\n")
	} else {
		fmt.Printf("\n[?] Discovering managers.\n\n")
	}
	color.Unset()

	for _, m := range managers {
		if local == m.local {
			UpdateSpinner(s, fmt.Sprintf("Checking if %s is installed...", m.name))
			s.Restart()

			out, err := exec.Command("which", m.name).Output()

			s.Stop()
			if err != nil {
				fmt.Printf("%s %s\n", hint, color.BlackString(fmt.Sprintf("%s was not found.", m.name)))
			} else {
				fmt.Printf("%s %s was found at %s.\n", success, m.name, strings.TrimSpace(string(out)))
				found = append(found, m)
			}
		}
	}

	fmt.Printf("\n%s", success)
	color.Set(color.Bold)
	if local {
		fmt.Printf(" Found %d local managers.\n[?] Updating locally.\n\n", len(found))
	} else {
		fmt.Printf(" Found %d managers.\n[?] Updating.\n\n", len(found))
	}
	color.Unset()

	shell := os.Getenv("SHELL")
	for _, m := range found {
		UpdateSpinner(s, fmt.Sprintf("Updating using %s (%s)...", m.name, m.command))
		s.Restart()

		out, err := exec.Command(shell, "-c", m.command).CombinedOutput()

		s.Stop()
		if err != nil {
			fmt.Printf("%s %s failed with %s: %s\n", fail, m.name, err, out)
			failed = append(failed, m)
		} else {
			fmt.Printf("%s %s succeeded.\n", success, m.name)
			succeeded = append(succeeded, m)
		}
	}

	fmt.Println()

	fmt.Print(success)

	color.Set(color.Bold)
	fmt.Printf(" Successfully updated %d/%d.\n", len(succeeded), len(found))
	color.Unset()

	if len(failed) != 0 {
		fmt.Printf("\n%s %d updates failed:\n", fail, len(found))
		for _, m := range failed {
			fmt.Printf("    - %s\n", color.RedString(m.name))
		}
	}

	return nil
}

func install() error {
	filepath, err := os.Executable()
	if err != nil {
		return err
	}

	_, err = exec.Command(os.Getenv("SHELL"), "-c", fmt.Sprintf("sudo cp %s /usr/local/bin", filepath)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s Failed to install renova: %s", fail, err)
	}

	fmt.Printf("%s Successfully installed renova under /usr/local/bin.", success)
	return err
}

func uninstall() error {
	filepath := "/usr/local/bin/renova"

	if os.Geteuid() != 0 {
		if _, err := os.Stat("/usr/local/bin/renova"); err == nil {
			return cli.Exit(fmt.Sprintf("%s renova is still installed under /usr/local/bin. Please remove that installation first using sudo renova uninstall", fail), 1)
		} else if errors.Is(err, os.ErrNotExist) {
			filepath, err := os.Executable()
			if err != nil {
				return err
			}

			_, err = exec.Command(os.Getenv("SHELL"), "-c", fmt.Sprintf("rm %s", filepath)).CombinedOutput()

			return err
		}
	}

	if _, err := os.Stat("/usr/local/bin/renova"); err != nil {
		fmt.Printf("%s No install of renova was found under /usr/local/bin.", fail)
		return err
	}

	_, err := exec.Command(os.Getenv("SHELL"), "-c", fmt.Sprintf("rm %s", filepath)).CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s Successfully removed renova from /usr/local/bin. To completely uninstall, run renova uninstall now.\n", success)

	return nil
}

func main() {
	app := &cli.App{
		Name:        "renova",
		Usage:       "Update all your packages",
		Description: "renova updates packages for the current user. To update global packages, run \"sudo renova\", to update local packages, run \"renova\".",
		Version:     "v1.1.1",
		Suggest:     true,
		Action: func(ctx *cli.Context) error {
			return UpdateAll()
		},
		Commands: []*cli.Command{
			{
				Name:    "install",
				Usage:   "Install renova",
				Aliases: []string{"setup"},
				Action: func(ctx *cli.Context) error {
					return install()
				},
			},
			{
				Name:    "uninstall",
				Usage:   "Uninstall renova",
				Aliases: []string{"remove"},
				Action: func(ctx *cli.Context) error {
					return uninstall()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
