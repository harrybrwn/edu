package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
	"github.com/harrybrwn/config"
	"github.com/harrybrwn/edu/cmd/commands"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

var version string

// Logger for the cmd package
var Logger = &lumberjack.Logger{
	Filename:   filepath.Join(os.TempDir(), "edu.log"),
	MaxSize:    25,  // megabytes
	MaxBackups: 10,  // number of spare files
	MaxAge:     365, // days
	Compress:   false,
}

// Stop will print to stderr and exit with status 1
func Stop(message interface{}) {
	log.Printf("%v\n", message)
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

	config.SetFilename("config.yml")
	config.SetType("yaml")
	config.AddPath("$EDU_CONFIG")
	config.AddDefaultDirs("edu")
	config.SetConfig(commands.Conf)

	err = config.ReadConfigFile()
	switch err {
	case nil:
		break
	case config.ErrNoConfigDir, config.ErrNoConfigFile:
		log.Println(err)
	default:
		return err
	}

	configfile := config.FileUsed()
	if configfile != "" {
		Logger.Filename = filepath.Join(filepath.Dir(configfile), "logs", "edu.log")
	}

	beeep.DefaultDuration = 800
	root := &cobra.Command{
		Use:           "edu <command>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		Short:         "Command line tool for managing online school with canvas.",
		Long:          ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(*cobra.Command, []string) {
			rootPreRun()
		},
	}

	globalFlags := opts.Global{}
	globalFlags.AddToFlagSet(root.PersistentFlags())

	root.SetUsageTemplate(commandTemplate)
	root.AddCommand(append(
		commands.All(&globalFlags),
		completionCmd,
		testCmd,
		canvasHelp,
	)...)
	err = root.Execute()
	if err != nil {
		return errors.WithMessage(err, "Error")
	}
	return err
}

func rootPreRun() {
	host := config.GetString("host")
	if host != "" {
		canvas.DefaultHost = host
	}
	token := config.GetString("token")
	if token == "" {
		log.Println("no canvas api token")
	}
	canvas.SetToken(token)
	canvas.ConcurrentErrorHandler = errorHandler
}

var (
	completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Print a completion script to stdout.",
		Long: `Use the completion command to generate a script for shell
completion. Note: for zsh you will need to use the command
'compdef _edu edu' after you source the generated script.`,
		Example:   "$ source <(edu completion zsh)",
		ValidArgs: []string{"zsh", "bash", "ps", "powershell", "fish"},
		Aliases:   []string{"comp"},
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
	}

	testCmd = &cobra.Command{
		Use:    "test",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("test command doesn't do anything")
		},
	}

	canvasHelp = &cobra.Command{
		Use:   "canvas",
		Short: "interfacing with the Canvas api",
		Long: `
In order to interface with the Canvas API, you need to first
obtain a token from you account. You can find a tutorial for
obtaining a token from this url:

    https://community.canvaslms.com/docs/DOC-16005-42121018197

Once you have the token, either set the $CANVAS_TOKEN environment
variable or set it in the config file (see 'edu help config').`,
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
		errmsg(e.Error())
		os.Exit(2)
	}
	return nil
}

func errmsg(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
}

var commandTemplate = `Usage:
{{if .Runnable}}
	{{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
	{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
	{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:

{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
