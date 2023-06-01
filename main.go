package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type manager struct {
	name    string
	command string
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
}

var fail = color.RedString("[\uf00d]")
var success = color.GreenString("[\uf00c]")
var warn = color.YellowString("[!]")
var hint = color.BlackString("[?]")

func UpdateSpinner(s *spinner.Spinner, suffix string) {
	s.Suffix = fmt.Sprintf("%s %s", color.MagentaString("]"), suffix)
}

func main() {
	var failed []manager
	var succeeded []manager
	var found []manager

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithColor("magenta"))
	s.Prefix = color.MagentaString("[")
	s.Suffix = fmt.Sprintf("%s Loading...", color.MagentaString("]"))
	s.Start()

	s.Stop()

	if os.Geteuid() != 0 {
		fmt.Printf("%s Updating requires elevated privilages! Please rerun the program with sudo.\n", warn)
		os.Exit(1)
	}

	color.Set(color.Bold)
	fmt.Printf("\n[?] Discovering managers.\n\n")
	color.Unset()

	for _, m := range managers {
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

	fmt.Printf("\n%s", success)
	color.Set(color.Bold)
	fmt.Printf(" Found %d managers.\n[?] Updating.\n\n", len(found))
	color.Unset()

	shell := os.Getenv("SHELL")
	for _, m := range found {
		UpdateSpinner(s, fmt.Sprintf("Updating using %s (%s)", m.name, m.command))
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
}
