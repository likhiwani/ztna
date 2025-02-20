package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/identity"
	"github.com/spf13/cobra"
)

type IdentityConfigFile struct {
	ZtAPI       string          `json:"ztAPI"`
	ID          identity.Config `json:"id"`
	ConfigTypes []string        `json:"configTypes"`
}

func NewUnwrapIdentityFileCommand(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	outCertFile := ""
	outKeyFile := ""
	outCaFile := ""

	cmd := &cobra.Command{
		Use:   "unwrap <identity_file>",
		Short: "unwrap a Ziti Identity file into its separate pieces (supports PEM only)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			identityFile := args[0]

			rootFileName := strings.TrimSuffix(identityFile, ".json")

			if outCertFile == "" {
				outCertFile = rootFileName + ".cert"
			}

			if outKeyFile == "" {
				outKeyFile = rootFileName + ".key"
			}

			if outCaFile == "" {
				outCaFile = rootFileName + ".ca"
			}

			identityJson, err := os.ReadFile(identityFile)

			if err != nil {
				_, _ = fmt.Fprintf(errOut, "error opening file %s: %v\n", args[0], err)
				return
			}

			config := &IdentityConfigFile{}
			if err := json.Unmarshal(identityJson, config); err != nil {
				_, _ = fmt.Fprintf(errOut, "error unmarshaling identity config JSON: %v\n", err)
				return
			}

			if strings.HasPrefix(config.ID.Cert, "pem:") {
				data := strings.TrimPrefix(config.ID.Cert, "pem:")
				if err := os.WriteFile(outCertFile, []byte(data), 0); err != nil {
					_, _ = fmt.Fprintf(errOut, "error writing certificate to file [%s]: %v\n", outCertFile, err)
					return
				}
			} else {
				_, _ = fmt.Fprintf(errOut, "error writing certificate to file [%s]: missing pem prefix, type is unsupported\n", outCertFile)
			}

			if strings.HasPrefix(config.ID.Key, "pem:") {
				data := strings.TrimPrefix(config.ID.Key, "pem:")
				if err := os.WriteFile(outKeyFile, []byte(data), 0); err != nil {
					_, _ = fmt.Fprintf(errOut, "error writing private key to file [%s]: %v\n", outKeyFile, err)
					return
				}
			} else {
				_, _ = fmt.Fprintf(errOut, "error writing private key to file [%s]: missing pem prefix, type is unsupported\n", outKeyFile)
			}

			if strings.HasPrefix(config.ID.Key, "pem:") {
				data := strings.TrimPrefix(config.ID.CA, "pem:")
				if err := os.WriteFile(outCaFile, []byte(data), 0); err != nil {
					_, _ = fmt.Fprintf(errOut, "error writing CAs to file [%s]: %v\n", outCaFile, err)
					return
				}
			} else {
				_, _ = fmt.Fprintf(errOut, "error writing CAs to file [%s]: missing pem prefix, type is unsupported\n", outKeyFile)
			}
		},
	}

	cmd.Flags().StringVarP(&outCertFile, "cert", "", "", "output certificate file, defaults to ./<root>.cert")
	cmd.Flags().StringVarP(&outKeyFile, "key", "", "", "output private key file, defaults to ./<root>.key")
	cmd.Flags().StringVarP(&outCaFile, "ca", "", "", "output ca bundle file, defaults to ./<root>.ca")

	return cmd
}
