package replay

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func init() {
	// setup nice looking log formatter
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		DisableQuote:           true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})

	// set default log level
	logrus.SetLevel(logrus.InfoLevel)

	// cleanup the default help template a bit
	cli.AppHelpTemplate = `
DESCRIPTION:
	{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
	{{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

OPTIONS:
	{{range $index, $option := .VisibleFlags}}{{if $index}}
	{{end}}{{$option}}{{end}}

`
}

// App is an instance of the CLI framework that houses our application
//
// docs => https://github.com/urfave/cli
var App = cli.App{
	Name:  "replay",
	Usage: "a CLI for reading the state of remote systems at a point in time",
	UsageText: `./replay --field {fieldOne} ... {dataSource} {dateTime}
	./replay --field ambientTemp --field schedule /tmp/ehub_data 2016-01-01T03:00
	./replay --field ambientTemp --field schedule s3://net.energyhub.assets/public/dev-exercises/audit-data/ 2016-01-01T03:00`,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "field",
			Usage:    "a field to show the state of, can be input multiple times",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "show debug logs on stderr",
		},
	},
	Action: func(c *cli.Context) (err error) {
		// set log level to debug if `--debug` was passed in
		if c.Bool("debug") == true {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// get dataScource arg
		if c.Args().Len() < 1 {
			cli.ShowAppHelp(c)
			err = errors.New("the 1st argument specifying a `dataSource` is required")
			return err
		}
		dataScource := c.Args().Get(0)

		// turn dataScoure arg into a readerFunc
		var readerFunc readerFunc
		if strings.HasPrefix(dataScource, "s3://") {
			readerFunc = s3Reader
		} else {
			readerFunc = localReader
		}

		// get dateTime arg
		if c.Args().Len() < 2 {
			cli.ShowAppHelp(c)
			err = errors.New("the 2nd argument specifying a `dateTime` is required")
			return err
		}
		dateTime := c.Args().Get(1)

		// do business logic
		output, err := getState(getStateInput{
			fields:     c.StringSlice("field"), // <= arg requires no extra validation / conversion
			dataSource: dataScource,
			dateTime:   dateTime,
			readerFunc: readerFunc,
		})
		if err != nil {
			err = fmt.Errorf("error getting state: %w", err)
			return err
		}

		// format output
		jsonOutput, err := json.Marshal(output)
		if err != nil {
			err = fmt.Errorf("error with json.Marshal: %w", err)
			return err
		}
		jsonString := string(jsonOutput)

		// TODO: display all numbers as floats with 1 decimal place

		// show output on stdout
		// this is the only thing allowed to write to stdout!
		fmt.Println(jsonString)

		return nil
	},
}
