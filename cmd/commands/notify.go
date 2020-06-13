package commands

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/harrybrwn/edu/pkg/twilio"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newTextCmd() *cobra.Command {
	var (
		to   string
		from = viper.GetString("twilio_number")
		file string
	)
	c := &cobra.Command{
		Hidden: true,
		Use:    "text [message...]",
		Short:  "Send a text message using the twilio api.",
		Long: "Send a text message using the twilio api.\n" +
			"All arguments will be sent as the text message.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var msg string
			if file != "" {
				contents, err := ioutil.ReadFile(file)
				if err != nil {
					return err
				}
				msg = string(contents)
			} else if len(args) > 1 {
				msg = strings.Join(args, " ")
			} else if len(args) == 1 {
				msg = args[0]
			} else {
				return errors.New("no message")
			}

			if to == "" {
				return errors.New("no number to send to")
			}
			if from == "" {
				return errors.New("no number to send from")
			}
			twilio := twilio.NewClient(
				viper.GetString("twilio_sid"),
				viper.GetString("twilio_token"),
			)
			log.Printf("sending text %s to %s\n", from, to)
			twilio.SetSender(from)
			_, err := twilio.Send(to, msg)
			return err
		},
	}
	flags := c.Flags()
	flags.StringVarP(&to, "to", "t", to, "phone number to send the message to")
	flags.StringVarP(&from, "from", "f", from, "phone number to send the message from")
	flags.StringVar(&file, "file", "", "use the contents of a file as the text message body")
	return c
}
