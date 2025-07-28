/*
Copyright © 2025 Curt Self <curtself.cs@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"ssl-tools/internal/certsvc"
	"ssl-tools/internal/options"
)

var infoOpts options.InfoOptions

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Gets information about certificates and CSRs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := infoOpts.Validate(); err != nil {
			return err
		}
		//fmt.Printf("Arguments given %+v\n",infoOpts)
		svc := certsvc.New()
		err := svc.GetInfo(infoOpts)
		if err != nil {
			return fmt.Errorf("Error getting info: %w", err)
		}
		return nil
	},
}

func init() {
	infoCmd.Flags().StringArrayVarP(&infoOpts.Certificates, "cert", "c", []string{}, "Certificate file list (optional)")
	infoCmd.Flags().StringArrayVarP(&infoOpts.URLs, "url", "u", []string{}, "URL list (optional)")
	infoCmd.Flags().StringVarP(&infoOpts.CSR, "csr", "r", "", "CSR file (optional)")
	infoCmd.Flags().StringVarP(&infoOpts.Password, "password", "p", "", "Password (optional, used with pkcs12/pfx files)")
	infoCmd.Flags().BoolVarP(&infoOpts.ShortSummary, "short-summary", "s", false, "Show short summary (optional)")
	rootCmd.AddCommand(infoCmd)

}
