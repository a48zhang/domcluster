package cli

import (
	"fmt"

	"d8rctl/auth"
)

// Password 显示或重置密码
func Password(args []string) error {
	if len(args) > 0 && args[0] == "reset" {
		password, err := auth.GetPasswordManager().ResetPassword()
		if err != nil {
			return fmt.Errorf("failed to reset password: %w", err)
		}

		fmt.Println("========================================")
		fmt.Printf("NEW PASSWORD: %s\n", password)
		fmt.Println("========================================")
		fmt.Println("Please save this password to access the web interface")
		return nil
	}

	// Password cannot be retrieved from hash
	fmt.Println("Password cannot be retrieved from hash.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Check the d8rctl.log file for the initial password")
	fmt.Println("  2. Use 'd8rctl password reset' to generate a new password")
	fmt.Println()
	fmt.Println("Note: The password is only displayed when the daemon runs for the first time.")

	return nil
}