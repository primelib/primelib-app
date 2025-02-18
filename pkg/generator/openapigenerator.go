package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

type OpenAPIGenerator struct {
	Directory string   `json:"-" yaml:"-"`
	APISpec   string   `json:"-" yaml:"-"`
	Args      []string `json:"-" yaml:"-"`
	Config    OpenAPIGeneratorConfig
}

type OpenAPIGeneratorConfig struct {
	GeneratorName         string                 `json:"generatorName" yaml:"generatorName"`
	InvokerPackage        string                 `json:"invokerPackage" yaml:"invokerPackage"`
	ApiPackage            string                 `json:"apiPackage" yaml:"apiPackage"`
	ModelPackage          string                 `json:"modelPackage" yaml:"modelPackage"`
	EnablePostProcessFile bool                   `json:"enablePostProcessFile" yaml:"enablePostProcessFile"`
	GlobalProperty        map[string]interface{} `json:"globalProperty" yaml:"globalProperty"`
	AdditionalProperties  map[string]interface{} `json:"additionalProperties" yaml:"additionalProperties"`
}

// openApiGeneratorArgumentAllowList is a list of arguments that are allowed to be passed to the openapi generator
var openApiGeneratorArgumentAllowList = []string{
	// spec validation
	"--skip-validate-spec",
	// normalizer - see https://openapi-generator.tech/docs/customization/#openapi-normalizer
	"--openapi-normalizer",
	"SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=true",
	"SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=false",
	"SIMPLIFY_BOOLEAN_ENUM=true",
	"SIMPLIFY_BOOLEAN_ENUM=false",
	"SIMPLIFY_ONEOF_ANYOF=true",
	"SIMPLIFY_ONEOF_ANYOF=false",
	"ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=true",
	"ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=false",
	"REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=true",
	"REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=false",
	"REF_AS_PARENT_IN_ALLOF=true",
	"REF_AS_PARENT_IN_ALLOF=false",
	"REMOVE_ANYOF_ONEOF_AND_KEEP_PROPERTIES_ONLY=true",
	"REMOVE_ANYOF_ONEOF_AND_KEEP_PROPERTIES_ONLY=false",
	"KEEP_ONLY_FIRST_TAG_IN_OPERATION=true",
	"KEEP_ONLY_FIRST_TAG_IN_OPERATION=false",
	"SET_TAGS_FOR_ALL_OPERATIONS=true",
	"SET_TAGS_FOR_ALL_OPERATIONS=false",
	"DISABLE_ALL=true",
}

// Name returns the name of the task
func (n OpenAPIGenerator) Name() string {
	return "openapi-generator"
}

func (n OpenAPIGenerator) Generate() error {
	// cleanup
	err := n.deleteGeneratedFiles()
	if err != nil {
		return fmt.Errorf("failed to delete generated files: %w", err)
	}

	// generate
	err = n.generateCode()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	return nil
}

func (n OpenAPIGenerator) deleteGeneratedFiles() error {
	// check if .openapi-generator/FILES exists
	filesDir := filepath.Join(n.Directory, ".openapi-generator", "FILES")
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		return nil
	}

	// read file list
	bytes, err := os.ReadFile(filesDir)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filesDir, err)
	}

	// iterate over files
	files := strings.Split(string(bytes), "\n")
	log.Info().Int("count", len(files)).Msg("deleting generated files")
	for _, file := range files {
		// skip empty lines
		if file == "" {
			continue
		}

		// skip if file does not exist
		if _, err := os.Stat(filepath.Join(n.Directory, file)); os.IsNotExist(err) {
			continue
		}

		// delete file
		log.Trace().Str("path", filepath.Join(n.Directory, file)).Msg("deleting file")
		err = os.Remove(filepath.Join(n.Directory, file))
		if err != nil {
			return fmt.Errorf("failed to delete file %s: %w", file, err)
		}
	}

	return nil
}

func (n OpenAPIGenerator) generateCode() error {
	// auto generate config
	tempConfigFile, tmpErr := os.CreateTemp("", "openapi-generator.json")
	if tmpErr != nil {
		return fmt.Errorf("failed to create temporary config openapi-generator.json: %w", tmpErr)
	}
	defer tempConfigFile.Close()

	// config
	configFile := path.Join(n.Directory, "openapi-generator.json")
	if _, fileErr := os.Stat(configFile); os.IsNotExist(fileErr) {
		// set defaults and missing properties
		n.Config.EnablePostProcessFile = true

		// marshal config
		bytes, err := json.MarshalIndent(n.Config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// write to temp file
		err = os.WriteFile(tempConfigFile.Name(), bytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		configFile = tempConfigFile.Name()
	}

	// default user args
	if len(n.Args) == 0 {
		n.Args = []string{
			"--openapi-normalizer", "SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=true",
			"--openapi-normalizer", "SIMPLIFY_BOOLEAN_ENUM=true",
			"--openapi-normalizer", "SIMPLIFY_ONEOF_ANYOF=true",
			"--openapi-normalizer", "ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=true",
			"--openapi-normalizer", "REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=true",
		}
	}

	// all user args must be present in the allow list
	for _, arg := range n.Args {
		if !slices.Contains(openApiGeneratorArgumentAllowList, arg) {
			return fmt.Errorf("openapi generator argument not allowed: %s", arg)
		}
	}

	generateSubCommand := "generate"
	if strings.HasPrefix(n.Config.GeneratorName, "primecodegen-") {
		generateSubCommand = "prime-generate"
	}

	// primecodegen bin and args
	executable := []string{"primecodegen"}
	if binPath := os.Getenv("PRIMECODEGEN_BIN"); binPath != "" {
		executable = strings.Fields(binPath)
	}
	args := []string{
		generateSubCommand,
		"-e", "auto",
		"-i", n.APISpec,
		"-o", n.Directory,
		"-c", configFile,
		"--skip-validate-spec",
	}
	command := append(executable, args...)
	command = append(command, n.Args...)

	cmd := exec.Command("bash", "-c", strings.Join(command, " "))
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute code generation: %w", err)
	}

	return nil
}
