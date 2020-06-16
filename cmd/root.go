package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
	"github.com/harrybrwn/edu/cmd/commands"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var version string

// Logger for the cmd package
var Logger = &lumberjack.Logger{
	Filename:   filepath.Join(os.TempDir(), "edu.log"),
	MaxSize:    25,  // megabytes
	MaxBackups: 10,  // number of spare files
	MaxAge:     365, //days
	Compress:   false,
}

// Stop will print to stderr and exit with status 1
func Stop(message interface{}) {
	log.Printf("%v", message)
	fmt.Fprintf(os.Stderr, "%v\n", message)
	switch msg := message.(type) {
	case *internal.Error:
		os.Exit(msg.Code)
	default:
		os.Exit(1)
	}
}

// Execute will execute the root comand on the cli
func Execute() (err error) {
	log.SetOutput(Logger)
	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && ok {
		err = createDefaultConfigFile("$HOME/.config/edu", "config.yml")
		if err != nil {
			return err
		}
	} else if err != nil {
		fmt.Fprintf(io.MultiWriter(os.Stderr, Logger), "Error: %v\n", err)
	}

	configfile := viper.ConfigFileUsed()
	if configfile != "" {
		Logger.Filename = filepath.Join(filepath.Dir(configfile), "logs", "edu.log")
	}

	globalFlags := opts.Global{}
	globalFlags.AddToFlagSet(root.PersistentFlags())

	root.AddCommand(commands.All(&globalFlags)...)
	root.AddCommand(completionCmd, testCmd)
	err = root.Execute()
	if err != nil {
		return errors.WithMessage(err, "Error")
	}
	return err
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
	viper.BindEnv("twilio_sid", "TWILIO_SID")
	viper.BindEnv("twilio_token", "TWILIO_TOKEN")

	viper.SetDefault("editor", os.Getenv("EDITOR"))
	viper.SetDefault("basedir", "$HOME/.edu/files")
	viper.SetDefault("notifications", true)
	viper.SetDefault("watch.duration", "12h")

	beeep.DefaultDuration = 800
}

var (
	root = &cobra.Command{
		Use: "edu",
		// SilenceErrors: true,
		// SilenceUsage:  true,
		Version: version,
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
				token = os.ExpandEnv(viper.GetString("token"))
				canvas.SetToken(token)
			}
			if token == "" {
				log.Println("no canvas api token")
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

	testCmd = &cobra.Command{
		Use:    "test",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no arguments")
			}
			return beeep.Notify("edu", args[0], "")
		},
	}
)

func createDefaultConfigFile(dir, file string) error {
	path := os.ExpandEnv("$HOME/.config/edu")
	if err := internal.Mkdir(path); err != nil {
		return fmt.Errorf("couldn't create config dir: %w", err)
	}
	fullpath := filepath.Join(path, file)
	log.Println("setting up config file at", fullpath)
	viper.SetConfigFile(fullpath)
	return nil
}

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
		errmsg(e.Error())
		fmt.Printf("%[1]T %[1]#v\n", e)
		os.Exit(1)
	}
	return nil
}

func errmsg(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
}
