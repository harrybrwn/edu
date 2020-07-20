# Configuration

Configuration for this program uses [yaml](https://yaml.org/). To find the configuration file, run `edu config -f` and that will output the path to the config file currently being used. Or use `edu config -e` to edit it with your favorite text editor (hint use the `$EDITOR` environment variable).

#### Token
This is the API token for the canvas api. For directions on getting a student api token look [here](https://community.canvaslms.com/docs/DOC-16005-42121018197).
You also have the option of setting this as an environment variable called `$CANVAS_TOKEN`.

#### Host
The `host` config variable will set the host used by the canvas api.
```yaml
host: canvas.instructure.com
```

#### Base Dir
The `basedir` config variable will set the base directory used for downloading course files.
```yaml
basedir: $HOME/school
```

#### Replacements
The `replacements` config variable is an array of regex patterns and replacement strings.
```yaml
replacements:
  - pattern: "S20-([a-zA-Z]+) (0){0,1}([0-9]+) .*?/"
    replacement: "$1$3/" # replace using group 1 and group 3
    lower: true # convert the replacement to lowercase
  - pattern: " "
    replacements: "_"
  - pattern: \.text$ # use a literal '.'
    replacements: ".txt"
```

#### watch
The `watch` config field is an object that houses configuration data for the `edu registration watch` command.
* crns - an array of crn IDs that will be watched for open seats
* duration - tells the `watch` command how often to repeat (default is '12h')
```yaml
watch:
  duration: '1h35m100ms'
  crns: [123, 234, 345, 456, 567]
```