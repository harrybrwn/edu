package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

const serviceTemplate = `[Unit]
Description=Watch a list of CRNs
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=10
User={{.User}}
ExecStart={{.Bin}} registration watch -v

[Install]
WantedBy=multi-user.target
`

func genServiceCmd() *cobra.Command {
	var (
		filename string = "./edu.service"
		restart  bool
		install  bool
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

			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("could not create %s: %w", filename, err)
			}
			defer file.Close()
			if err = tmpl.Execute(file, &data); err != nil {
				return err
			}
			if !install {
				return nil
			}
			return createService(filename)
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&restart, "restart", "r", restart, "Restart the systemd service")
	flags.BoolVarP(&install, "install", "i", install, "Install a systemd service.")
	flags.StringVarP(&filename, "file", "f", filename, "Write the service to a file")
	return c
}

func createService(filename string) error {
	_, name := filepath.Split(filename)
	systemdfile := filepath.Join("/etc/systemd/system", name)
	err := system("sudo", "install", filename, systemdfile)
	if err != nil {
		return fmt.Errorf("could not copy service file: %w", err)
	}
	err = system("sudo", "systemctl", "enable", "edu")
	if err != nil {
		return fmt.Errorf("could not enable service: %w", err)
	}
	err = system("sudo", "systemctl", "start", "edu")
	if err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}
	return nil
}

func system(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
