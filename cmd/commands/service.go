package commands

import (
	"io"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

const serviceTemplate = `[Unit]
Description=Watch a list of CRNs
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=30
User={{.User}}
ExecStart={{.Bin}} registration watch

[Install]
WantedBy=multi-user.target
`

func genServiceCmd() *cobra.Command {
	var (
		filename string
		file     io.Writer = os.Stdout
	)
	c := &cobra.Command{
		Use:    "gen-service",
		Short:  "generate a systemd service",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			data := struct {
				User, Bin string
			}{}
			data.Bin, err = os.Executable()
			if err != nil {
				return err
			}
			data.User = os.Getenv("USER")
			tmpl, err := template.New("service").Parse(serviceTemplate)
			if err != nil {
				return err
			}
			if filename != "" {
				osfile, err := os.Open(filename)
				if err != nil {
					return err
				}
				defer osfile.Close()
				file = osfile
			}
			return tmpl.Execute(file, &data)
		},
	}
	c.Flags().StringVarP(&filename, "file", "f", "", "write the service to a file")
	return c
}
