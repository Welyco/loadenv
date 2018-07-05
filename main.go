package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/envkey/envkey-fetch/fetch"
	flags "github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
)

type options struct {
	Envkey     string `short:"e" long:"envkey" description:"ENVKEY variable"`
	SourcePath string `short:"s" long:"source" description:"Source dotenv file path" default:".env"`
	OutputPath string `short:"o" long:"output" description:"Output file path" default:"output.env"`
	Format     string `short:"f" long:"format" description:"Output format dotenv, json, export" default:"dotenv"`
}

func main() {
	var opts options
	flags.Parse(&opts)

	envkey, err := loadEnvkey(opts)
	if err != nil {
		panic(err)
	}

	res := fetch.Fetch(envkey, fetch.FetchOptions{false, "", "envkeygo", "", false, 2.0})
	if strings.HasPrefix(res, "error:") {
		panic(errors.New(strings.Split(res, "error:")[1]))
	}

	var resMap map[string]string
	err = json.Unmarshal([]byte(res), &resMap)

	if err != nil {
		panic(errors.New("problem parsing EnvKey's response"))
	}

	switch opts.Format {
	case "dotenv":
		writeDotenv(opts.OutputPath, resMap)
	case "json":
		writeJSON(opts.OutputPath, resMap)
	case "export":
		writeExport(opts.OutputPath, resMap)
	default:
		fmt.Println("Invalid format :", opts.Format)
	}
}

func loadEnvkey(opts options) (string, error) {
	if opts.SourcePath != "" {
		godotenv.Load(opts.SourcePath)
	}

	envkey := os.Getenv("ENVKEY")

	if envkey == "" {
		return "", errors.New("missing ENVKEY")
	}

	return envkey, nil
}

func writeDotenv(fileName string, resMap map[string]string) {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	for k, v := range resMap {
		f.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	f.Sync()
}

func writeExport(fileName string, resMap map[string]string) {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	for k, v := range resMap {
		f.WriteString(fmt.Sprintf("export %s=%s\n", k, v))
	}

	f.Sync()
}

func writeJSON(fileName string, resMap map[string]string) {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	jsonBytes, _ := json.Marshal(resMap)
	f.Write(jsonBytes)

	f.Sync()
}
