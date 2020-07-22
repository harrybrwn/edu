package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const serviceTemplate = `[Unit]
Description=Watch a list of CRNs
StartLimitIntervalSec=0
After=network.target

[Service]
Type=simple
Restart=on-failure
RestartSec=10
ExecStart={{.Bin}} registration watch -v
WorkingDirectory=/home/{{.User}}
User={{.User}}
Group={{.User}}
Environment=XDG_CONFIG_HOME=/home/{{.User}}/.config
Environment=CANVAS_TOKEN={{.CanvasToken}}

[Install]
WantedBy=multi-user.target
`

func genServiceCmd() *cobra.Command {
	var (
		filename string = "./edu.service"

		restart, stop, start, install bool
	)
	c := &cobra.Command{
		Use:    "service",
		Short:  "Generate and install a systemd service that runs 'edu' registration watch",
		Hidden: false,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if restart {
				fmt.Println("restarting service")
				return systemSudo("systemctl", "restart", "edu")
			}
			if stop {
				return systemSudo("systemctl", "stop", "edu")
			}
			if start {
				return systemSudo("systemctl", "start", "edu")
			}
			if len(args) == 1 {
				filename = args[0]
			}

			data := struct {
				User, Bin   string
				CanvasToken string
			}{
				CanvasToken: os.ExpandEnv(viper.GetString("token")),
				User:        os.Getenv("USER"),
			}
			data.Bin, err = os.Executable()
			if err != nil {
				return err
			}
			tmpl, err := template.New("service").Parse(serviceTemplate)
			if err != nil {
				return err
			}

			var file io.Writer
			if filename == "--" {
				file = os.Stdout
			} else {
				file, err := os.Create(filename)
				if err != nil {
					return fmt.Errorf("could not create %s: %w", filename, err)
				}
				defer file.Close()
			}
			if err = tmpl.Execute(file, &data); err != nil {
				return err
			}
			if !install {
				return nil
			}
			return installService(filename)
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&restart, "restart", "r", restart, "Restart the systemd service")
	flags.BoolVar(&stop, "stop", stop, "Stop the systemd service")
	flags.BoolVar(&start, "start", start, "Start the systemd service")
	flags.BoolVarP(&install, "install", "i", install, "Install a systemd service.")
	flags.StringVarP(&filename, "file", "f", filename, "Write the service to a file")
	return c
}

func installService(filename string) (err error) {
	_, name := filepath.Split(filename)
	serviceName := strings.Replace(name, ".service", "", 1)
	systemdfile := filepath.Join("/etc/systemd/system", name)

	statusCmd := sudoCommand("systemctl", "status", serviceName)
	// if the service exists then we are going to stop it
	if err = statusCmd.Run(); err == nil {
		log.Printf("stoping/disabling current service '%s' (%s)\n", name, systemdfile)
		if err = systemSudo("systemctl", "stop", serviceName); err != nil {
			return fmt.Errorf("could not stop existing service %s: %w", serviceName, err)
		}
		if err = systemSudo("systemctl", "disable", serviceName); err != nil {
			return fmt.Errorf("could not disable existing service %s: %w", serviceName, err)
		}
		if err = systemSudo("rm", systemdfile); err != nil {
			log.Printf("could not remove existing service %s: %v\n", systemdfile, err)
		}
	}

	err = systemSudo("install", "-m", "644", filename, systemdfile)
	if err != nil {
		return fmt.Errorf("could not copy service file: %w", err)
	}
	err = systemSudo("systemctl", "enable", "edu")
	if err != nil {
		return fmt.Errorf("could not enable service: %w", err)
	}
	err = systemSudo("systemctl", "start", "edu")
	if err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}
	return os.Remove(filename)
}

func system(command string, args ...string) error {
	fmt.Printf("%s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

var sudoMsgOnce = sync.Once{}

func sudoCommand(args ...string) *exec.Cmd {
	sudoMsgOnce.Do(func() {
		fmt.Print("Running some commands with sudo...\n\n")
	})
	fmt.Printf("sudo %s\n", strings.Join(args, " "))
	return exec.Command("sudo", args...)
}

func systemSudo(args ...string) error {
	cmd := sudoCommand(args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
