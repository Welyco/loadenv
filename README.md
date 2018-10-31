# loadenv

CLI tool to load environment variables from ENVKEY and export them in portable file format.

# Install

```
go get github.com/Welyco/loadenv
go install github.com/Welyco/loadenv
```

# Usage

```
> loadenv
Usage:
  loadenv [OPTIONS]

Application Options:
  -e, --envkey= ENVKEY variable
  -s, --source= Source dotenv file path
  -o, --output= Output file path
  -f, --format= Output format. options: [dotenv, json, gaeyaml, export, commonjs] (default: dotenv)

Help Options:
  -h, --help    Show this help message
```

# format

## Basic Format

### dotenv

```
# loadenv -s .env -o credentials/dotenv
# cat credentials/dotenv
ENV1=foo
ENV2=bar
```

### json

```
# loadenv -s .env -o credentials/json -f json
# cat credentials/json
{
  "ENV1":"foo",
  "ENV2":"bar"
}
```

### commonjs

```
# loadenv -s .env -o credentials/js -f commonjs
# cat credentials/js
module.exports = {
  "ENV1":"foo",
  "ENV2":"bar"
}
```

## Advanced Format

### gaeyaml

```
# loadenv -s .env -o credentials/env.yaml -f gaeyaml
# cat credentials/env.yaml
env_variables:
  ENV1: foo
  ENV2: bar
```

This file format can be included in the Google App Engine configuration file.
https://cloud.google.com/appengine/docs/standard/python/config/appref

```
runtime: nodejs
env: flex
service: graphql
includes:
- credentials/env.yaml
```

### export

```
# loadenv -s .env -o credentials/env.sh -f export
# cat credentials/env.sh
export ENV1=foo
export ENV2=bar
```

Load environment variables in the terminal
```
eval $(cat credentials/env.sh)
```
