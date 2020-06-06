package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harrybrwn/edu/cmd/commands"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

// Execute will execute the root comand on the cli
func Execute() (err error) {
	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && ok {
		path := os.ExpandEnv("$HOME/.config/edu")
		if err = internal.Mkdir(path); err != nil {
			return fmt.Errorf("couldn't create config dir: %w", err)
		}
		viper.SetConfigFile(filepath.Join(path, "config.yml"))
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	root.AddCommand(commands.All()...)
	root.AddCommand(completionCmd)

	err = root.Execute()
	if err == nil {
		return nil
	}
	errorMessage(err)
	os.Exit(1)
	return nil
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	viper.AddConfigPath("$XDG_CONFIG_HOME/edu")
	viper.AddConfigPath("$HOME/.config/edu")
	viper.AddConfigPath("$HOME/.edu")
	viper.AddConfigPath("$XDG_CONFIG_HOME/canvas")
	viper.AddConfigPath("$HOME/.config/canvas")
	viper.AddConfigPath("$HOME/.canvas")

	viper.SetEnvPrefix("edu")
	viper.BindEnv("host")
	viper.BindEnv("canvas_token", "CANVAS_TOKEN")

	viper.SetDefault("editor", os.Getenv("EDITOR"))
	viper.SetDefault("basedir", "$HOME/.edu/files")
}

var (
	root = &cobra.Command{
		Use:           "edu",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			host := viper.GetString("host")
			if host != "" {
				canvas.DefaultHost = host
			}
			token := viper.GetString("token")
			if token != "" {
				canvas.SetToken(os.ExpandEnv(token))
			} else {
				viper.Set("token", viper.GetString("canvas_token"))
				canvas.SetToken(os.ExpandEnv(viper.GetString("token")))
			}
			canvas.ConcurrentErrorHandler = errorHandler
		},
	}

	completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Print a completion script to stdout.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				return errors.New("no shell type given")
			}
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(out)
			case "ps", "powershell":
				return root.GenPowerShellCompletion(out)
			case "bash":
				return root.GenBashCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, false)
			}
			return errs.New("unknown shell type")
		},
		ValidArgs: []string{"zsh", "bash", "ps", "powershell", "fish"},
		Aliases:   []string{"comp"},
	}
)

func errorMessage(err error) {
	switch err.(type) {
	case *canvas.AuthError:
		fmt.Fprintf(os.Stderr, "Authentication Error: %v\n", err)
	case *canvas.Error:
		fmt.Fprintf(os.Stderr, "Canvas Error: %v\n", err)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func errorHandler(e error) error {
	if e != nil {
		fmt.Println("Error: " + e.Error())
		os.Exit(1)
	}
	return nil
}

func errmsg(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
}
