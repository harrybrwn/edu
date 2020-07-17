package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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
ExecStart={{.Bin}} registration watch -v

[Install]
WantedBy=multi-user.target
`

func genServiceCmd() *cobra.Command {
	var (
		filename string
		file     io.Writer = os.Stdout
		restart  bool
	)
	c := &cobra.Command{
		Use:    "service",
		Short:  "generate a systemd service",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if restart {
				fmt.Println("restarting service")
				return exec.Command("sudo", "systemctl", "restart", "edu").Run()
			}
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
			} else {
				osfile, err := os.Open("./edu.service")
				// osfile, err := os.Open("/etc/systemd/system/edu.service")
				// osfile, err := os.Create("/etc/systemd/system/edu.service")
				// osfile, err := os.OpenFile("/etc/systemd/system/edu.service",
				if err != nil {
					return err
				}
				defer osfile.Close()
				file = osfile
			}

			if err = tmpl.Execute(file, &data); err != nil {
				return err
			}
			return exec.Command("sudo", "systemctl", "enable", "edu").Run()
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&restart, "restart", "r", restart, "restart the systemd service")
	flags.StringVarP(&filename, "file", "f", "", "write the service to a file")
	return c
}
