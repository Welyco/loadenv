# loadenv

CLI tool to load environment variables from ENVKEY and export them in portable file format.

# Install

```
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
  -f, --format= Output format. options: [dotenv, json, export] (default: dotenv)

Help Options:
  -h, --help    Show this help message
```
