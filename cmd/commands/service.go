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

	"github.com/harrybrwn/config"
	"github.com/harrybrwn/errs"
	"github.com/spf13/cobra"
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
		name         string
		filename     string = ""
		templateFile string
	)
	c := &cobra.Command{
		Use:    "service",
		Short:  "Generate and install a systemd service that runs 'edu' registration watch",
		Hidden: false,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) == 1 {
				filename = args[0]
			}
			if filename == "-" {
				return execTemplate(os.Stdout, templateFile)
			}

			if filename != "" {
				file, err := os.Create(filename)
				if err != nil {
					return fmt.Errorf("could not create %s: %w", filename, err)
				}
				defer file.Close()
				defer os.Remove(filename)
				return execTemplate(file, templateFile)
			}
			return system("systemctl", "--no-pager", "--lines", "15", "-l", "status", "edu")
		},
	}

	c.AddCommand(&cobra.Command{
		Use: "install", Short: "Create a service and install it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename == "" {
				filename = filepath.Join(os.TempDir(), "edu.service")
			}
			file, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer file.Close()
			defer os.Remove(filename)
			if err = execTemplate(file, templateFile); err != nil {
				return err
			}
			return installService(filename)
		},
	}, &cobra.Command{
		Use: "delete", Short: "Delete the service",
		RunE: func(*cobra.Command, []string) error {
			return errs.Chain(
				sudo("systemctl", "stop", "edu"),
				sudo("rm", "/etc/systemd/system/edu.service"),
				sudo("systemctl", "daemon-reload"))
		},
	}, &cobra.Command{
		Use: "restart", Short: "Restart the service",
		RunE: sudoCmd("systemctl", "restart", "edu"),
	}, &cobra.Command{
		Use: "stop", Short: "Stop the service",
		RunE: sudoCmd("systemctl", "stop", "edu"),
	}, &cobra.Command{
		Use: "start", Short: "Start the service",
		RunE: sudoCmd("systemctl", "start", "edu"),
	})

	pflags := c.PersistentFlags()
	pflags.StringVarP(&filename, "file", "f", filename, "Write the service to a file")
	pflags.StringVar(&templateFile, "template-file", "", "Read the systemd service template from a file")
	pflags.StringVar(&name, "name", name, "name of the systemd service")
	return c
}

func sudoCmd(args ...string) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		return sudo(args...)
	}
}

func execTemplate(w io.Writer, tmplFile string) (err error) {
	data := struct {
		User, Bin   string
		CanvasToken string
	}{
		CanvasToken: os.ExpandEnv(config.GetString("token")),
		User:        os.Getenv("USER"),
	}
	data.Bin, err = os.Executable()
	if err != nil {
		return err
	}

	var tmpl *template.Template
	if tmplFile != "" {
		tmpl, err = template.ParseFiles(tmplFile)
	} else {
		tmpl, err = template.New("service").Parse(serviceTemplate)
	}
	return tmpl.Execute(w, &data)
}

func installService(filename string) (err error) {
	_, name := filepath.Split(filename)
	serviceName := strings.Replace(name, ".service", "", 1)
	systemdfile := filepath.Join("/etc/systemd/system", name)

	statusCmd := sudoCommand("systemctl", "status", serviceName)
	// if the service exists then we are going to stop it
	if err = statusCmd.Run(); err == nil {
		log.Printf("stoping/disabling current service '%s' (%s)\n", name, systemdfile)
		if err = sudo("systemctl", "stop", serviceName); err != nil {
			return fmt.Errorf("could not stop existing service %s: %w", serviceName, err)
		}
		if err = sudo("systemctl", "disable", serviceName); err != nil {
			return fmt.Errorf("could not disable existing service %s: %w", serviceName, err)
		}
		if err = sudo("rm", systemdfile); err != nil {
			log.Printf("could not remove existing service %s: %v\n", systemdfile, err)
		}
	}

	err = sudo("install", "-m", "644", filename, systemdfile)
	if err != nil {
		return fmt.Errorf("could not copy service file: %w", err)
	}
	err = sudo("systemctl", "enable", "edu")
	if err != nil {
		return fmt.Errorf("could not enable service: %w", err)
	}
	err = sudo("systemctl", "start", "edu")
	if err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}
	return nil
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

func sudo(args ...string) error {
	cmd := sudoCommand(args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
