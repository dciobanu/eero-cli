package cmd

import (
	"fmt"

	"github.com/dorin/eero-cli/internal/api"
	"github.com/dorin/eero-cli/internal/config"
)

// Login handles the login command
func (a *App) Login() error {
	identity := Prompt("Enter your email or phone number: ")
	if identity == "" {
		return fmt.Errorf("email or phone number is required")
	}

	fmt.Println("Requesting verification code...")

	loginResp, err := a.Client.Login(identity)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("A verification code has been sent to your email/phone.")
	code := Prompt("Enter verification code: ")
	if code == "" {
		return fmt.Errorf("verification code is required")
	}

	fmt.Println("Verifying...")

	if err := a.Client.LoginVerify(loginResp.UserToken, code); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Save the token
	a.Config.Token = loginResp.UserToken
	a.Client.SetToken(loginResp.UserToken)

	// Fetch and save network ID
	account, err := a.Client.GetAccount()
	if err != nil {
		// Token is saved, but couldn't get network
		if err := a.Config.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Println("Login successful! (Warning: couldn't fetch network info)")
		return nil
	}

	if len(account.Networks) > 0 {
		a.Config.NetworkID = api.ExtractNetworkID(account.Networks[0].URL)
		fmt.Printf("Logged in to network: %s\n", account.Networks[0].Name)
	}

	if err := a.Config.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("Login successful! Token saved.")
	return nil
}

// Logout handles the logout command
func (a *App) Logout() error {
	if err := a.Config.Clear(); err != nil {
		return fmt.Errorf("clearing config: %w", err)
	}
	fmt.Println("Logged out. Token cleared.")
	return nil
}

// Status shows the current authentication status
func (a *App) Status() error {
	path, _ := config.ConfigPath()

	if !a.Config.HasToken() {
		fmt.Println("Status: Not logged in")
		fmt.Printf("Config: %s\n", path)
		return nil
	}

	fmt.Println("Status: Checking token...")

	if !a.Client.ValidateToken() {
		fmt.Println("Status: Token is invalid or expired")
		fmt.Printf("Config: %s\n", path)
		return nil
	}

	account, err := a.Client.GetAccount()
	if err != nil {
		fmt.Println("Status: Authenticated (couldn't fetch account details)")
		fmt.Printf("Config: %s\n", path)
		return nil
	}

	fmt.Println("Status: Authenticated")
	if account.Email != "" {
		fmt.Printf("Email: %s\n", account.Email)
	}
	if account.Phone != "" {
		fmt.Printf("Phone: %s\n", account.Phone)
	}
	if account.Name != "" {
		fmt.Printf("Name: %s\n", account.Name)
	}
	if len(account.Networks) > 0 {
		fmt.Println("Networks:")
		for _, n := range account.Networks {
			premium := ""
			if n.Premium {
				premium = " (Premium)"
			}
			fmt.Printf("  - %s%s\n", n.Name, premium)
		}
	}
	fmt.Printf("Config: %s\n", path)

	return nil
}
