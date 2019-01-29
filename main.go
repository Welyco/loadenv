package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/envkey/envkey-fetch/fetch"
	flags "github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	yaml "gopkg.in/yaml.v2"
)

type options struct {
	Envkey     string `short:"e" long:"envkey" description:"ENVKEY variable"`
	SourcePath string `short:"s" long:"source" description:"Source dotenv file path"`
	OutputPath string `short:"o" long:"output" description:"Output file path"`
	Format     string `short:"f" long:"format" description:"Output format. options: [dotenv, json, gaeyaml, export, commonjs]" default:"dotenv"`
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Parse()

	if len(os.Args) == 1 {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	if len(opts.OutputPath) == 0 {
		fmt.Println(`Output file path is not set. Please specify Output path by --output`)
		fmt.Println(``)
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	envkey, err := loadEnvkey(opts)
	if err != nil {
		fmt.Println(`Couldn't load ENVKEY variables. Please specify ENVKEY by --envkey, or --source option`)
		fmt.Println(``)
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	res, err := fetch.Fetch(envkey, fetch.FetchOptions{
		ShouldCache:    false,
		CacheDir:       "",
		ClientName:     "envkeygo",
		ClientVersion:  "",
		VerboseOutput:  false,
		TimeoutSeconds: 5.0,
	})
	if err != nil {
		panic(errors.New("problem while fetching Envkey server"))
	}

	if strings.HasPrefix(res, "error:") {
		panic(errors.New(strings.Split(res, "error:")[1]))
	}

	var resMap map[string]string
	err = json.Unmarshal([]byte(res), &resMap)

	if err != nil {
		panic(errors.New("problem parsing EnvKey's response"))
	}

	writeOutputFile(opts, resMap)
}

func writeOutputFile(opts options, resMap map[string]string) {
	switch opts.Format {
	case "dotenv":
		writeDotenv(opts.OutputPath, resMap)
	case "json":
		writeJSON(opts.OutputPath, resMap)
	case "export":
		writeExport(opts.OutputPath, resMap)
	case "gaeyaml":
		writeGoogleAppEngineYAML(opts.OutputPath, resMap)
	case "commonjs":
		writeJavaScriptCommonModule(opts.OutputPath, resMap)
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

func createFileWithPath(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)

	if dir != "" {
		os.MkdirAll(dir, os.ModePerm)
	}

	f, err := os.Create(filePath)
	return f, err
}

func writeDotenv(filePath string, resMap map[string]string) {
	f, err := createFileWithPath(filePath)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	for k, v := range resMap {
		f.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	f.Sync()
}

func writeExport(filePath string, resMap map[string]string) {
	f, err := createFileWithPath(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	for k, v := range resMap {
		f.WriteString(fmt.Sprintf("export %s=%s\n", k, v))
	}

	f.Sync()
}

func writeJSON(filePath string, resMap map[string]string) {
	f, err := createFileWithPath(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	jsonBytes, _ := json.MarshalIndent(resMap, "", "  ")
	f.Write(jsonBytes)

	f.Sync()
}

func writeJavaScriptCommonModule(filePath string, resMap map[string]string) {
	f, err := createFileWithPath(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	jsonBytes, _ := json.MarshalIndent(resMap, "", "  ")
	jsBytes := append([]byte("module.exports = "), jsonBytes...)

	f.Write(jsBytes)

	f.Sync()
}

func writeGoogleAppEngineYAML(filePath string, resMap map[string]string) {
	f, err := createFileWithPath(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	gaeYamlData := map[string]map[string]string{}
	gaeYamlData["env_variables"] = resMap

	yamlBytes, _ := yaml.Marshal(gaeYamlData)
	f.Write(yamlBytes)

	f.Sync()
}
