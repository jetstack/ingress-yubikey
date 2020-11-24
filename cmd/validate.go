package cmd

import (
	"fmt"
	"os"

	"github.com/go-piv/piv-go/piv"
	"github.com/spf13/cobra"

	"github.com/jakexks/ingress-yubikey/pkg/util/yubikey"
)

var ValidateCommand = &cobra.Command{
	Use:   "validate",
	Short: "check your Yubikey is set up correctly",
	Long:  "Check your Yubikey is set up correctly",
	Run: func(cmd *cobra.Command, args []string) {
		yk, err := yubikey.Validate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Yubikey setup is invalid: %s\n", err.Error())
			os.Exit(1)
		}
		cert, err := yk.Certificate(piv.SlotSignature)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Yubikey setup is invalid: %s\n", err.Error())
			yk.Close()
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Yubikey setup is valid, slot 9c cert CN: %s\n", cert.Subject.CommonName)
		yk.Close()
	},
}

func init() {
	rootCmd.AddCommand(ValidateCommand)
}
