# edu
A command line interface for automating school interactions. This is a work in progress and it will be impossible to support the system that all schools use but this project will do its best to support the more popular systems (i.e. canvas)

## Configuration
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

#### Replacments
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

#### Token
Token for the canvas api
