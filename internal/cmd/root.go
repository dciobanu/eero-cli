// Package cmd implements CLI commands
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
	"github.com/dorin/eero-cli/internal/config"
)

// App holds the application state
type App struct {
	Config *config.Config
	Client *api.Client
}

// NewApp creates a new application instance
func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	client := api.New(cfg.Token)

	return &App{
		Config: cfg,
		Client: client,
	}, nil
}

// EnsureAuth checks that the user is authenticated
func (a *App) EnsureAuth() error {
	if !a.Config.HasToken() {
		return fmt.Errorf("not logged in. Run 'eero-cli login' first")
	}

	if !a.Client.ValidateToken() {
		return fmt.Errorf("token is invalid or expired. Run 'eero-cli login' to re-authenticate")
	}

	return nil
}

// EnsureNetwork ensures a network ID is available
func (a *App) EnsureNetwork() (string, error) {
	if err := a.EnsureAuth(); err != nil {
		return "", err
	}

	if a.Config.NetworkID != "" {
		return a.Config.NetworkID, nil
	}

	// Fetch account to get network ID
	account, err := a.Client.GetAccount()
	if err != nil {
		return "", fmt.Errorf("getting account: %w", err)
	}

	if len(account.Networks.Data) == 0 {
		return "", fmt.Errorf("no networks found on this account")
	}

	// Use first network, extract ID from URL
	networkID := api.ExtractNetworkID(account.Networks.Data[0].URL)
	a.Config.NetworkID = networkID
	if err := a.Config.Save(); err != nil {
		return "", fmt.Errorf("saving config: %w", err)
	}

	return networkID, nil
}

// Prompt reads a line of input from the user
func Prompt(message string) string {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptSecret reads a line of input without echo (for sensitive data)
func PromptSecret(message string) string {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Confirm asks for a yes/no confirmation
func Confirm(message string) bool {
	response := Prompt(message + " [y/N]: ")
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

// PrintTable prints data in a simple table format
func PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No data to display")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		fmt.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i := range headers {
		fmt.Print(strings.Repeat("-", widths[i]) + "  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}

// Usage prints the help message
func Usage() {
	fmt.Println(`eero-cli - Control your Eero WiFi network

Usage:
  eero-cli <command> [options]

Commands:
  login                     Authenticate with your Eero account
  logout                    Clear saved authentication
  status                    Show current authentication status

  devices [options]           List all devices
    --profile <name|id>       Filter by profile name or ID
    --noprofile               Show only devices without a profile
    --wired                   Show only wired devices
    --wireless                Show only wireless devices
    --online                  Show only online devices
    --offline                 Show only offline devices
    --paused                  Show only paused devices
    --guest                   Show only guest network devices
    --noguest                 Exclude guest network devices
  devices monitor [--interval <sec>]  Monitor devices for state changes
  devices pause <id>          Pause a device's internet access
  devices unpause <id>        Unpause a device
  devices block <id>          Block a device from the network
  devices unblock <id>        Unblock a device
  devices rename <id> <name>  Set a device's nickname

  profiles                  List all profiles
  profiles pause <id>       Pause a profile
  profiles unpause <id>     Unpause a profile

  guest                     Show guest network status
  guest enable              Enable guest network
  guest disable             Disable guest network
  guest password <pass>     Set guest network password

  reboot                    Reboot the network

  help                      Show this help message`)
}
