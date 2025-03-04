package config

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Name        string `yaml:"name"`
	Summary     string `yaml:"summary,omitempty"`
	Description string `yaml:"description,omitempty"`
	Output      string `yaml:"output,omitempty" jsonschema_description:"output directory for the generated code"`

	Repository  Repository   `yaml:"repository"`
	Maintainers []Maintainer `yaml:"maintainers"`

	Generators []Generator `yaml:"generators"` // Generators can be used to fully customize the generation process
	Presets    Presets     `yaml:"presets"`    // Presets are pre-configured generators for specific languages

	Spec Spec `yaml:"spec"`
}

func (c Configuration) HasGenerator() bool {
	return (c.Presets.EnabledCount() + len(c.Generators)) > 0
}

func (c Configuration) MultiLanguage() bool {
	return (c.Presets.EnabledCount() + len(c.Generators)) > 1
}

type Repository struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	URL           string `yaml:"url"`
	InceptionYear int    `yaml:"inceptionYear"`
	LicenseName   string `yaml:"licenseName"`
	LicenseURL    string `yaml:"licenseURL"`
}

type Maintainer struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
	URL   string `yaml:"url"`
}

type Generator struct {
	Enabled   bool                   `yaml:"enabled"`   // Enable the generator
	Name      string                 `yaml:"name"`      // Name of the generator
	Type      GeneratorType          `yaml:"type"`      // Type of the generator
	Arguments []string               `yaml:"arguments"` // Arguments that are passed to the generator command
	Config    map[string]interface{} `yaml:"config"`    // Config that is passed to the generator
}

// Presets are pre-configured generators for specific languages
type Presets struct {
	Go         GoLanguageOptions         `yaml:"go"`
	Java       JavaLanguageOptions       `yaml:"java"`
	Python     PythonLanguageOptions     `yaml:"python"`
	CSharp     CSharpLanguageOptions     `yaml:"csharp"`
	Typescript TypescriptLanguageOptions `yaml:"typescript"`
}

func (c Presets) EnabledCount() int {
	enabledCount := 0

	if c.Go.Enabled {
		enabledCount++
	}
	if c.Java.Enabled {
		enabledCount++
	}
	if c.Python.Enabled {
		enabledCount++
	}
	if c.Typescript.Enabled {
		enabledCount++
	}

	return enabledCount
}

type OpenApiGeneratorOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type PrimeCodeGenOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type GoLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	ModuleName string `yaml:"module"`
}

type JavaLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	GroupId    string `yaml:"groupId"`
	ArtifactId string `yaml:"artifactId"`
}

type PythonLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	PypiPackageName string `yaml:"pypiPackageName"`
}

type CSharpLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type TypescriptLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	NpmOrg  string `yaml:"npmOrg"`
	NpmName string `yaml:"npmName"`
}

type Spec struct {
	// File is the path to the openapi specification file
	File string `yaml:"file" default:"openapi.yaml" required:"true"`
	// SourcesDir is the directory where specifications are stored
	SourcesDir string `yaml:"sourcesDir"`
	// Sources contains one or multiple sources to specifications
	Sources []SpecSource `yaml:"sources" required:"true"`
	// Type is the format of the api specification
	Type SpecType `yaml:"type" required:"true"`
	// Customization allows overwriting certain parts of the specification
	Customization Customization `yaml:"customization"`
	// InputPatches are applied to the source specifications before merging
	InputPatches []string `yaml:"inputPatches"`
	// Patches are the patches that are applied to the specification
	Patches []string `yaml:"patches"`
}

func (s Spec) UrlSlice() []string {
	urls := make([]string, len(s.Sources))
	for i, u := range s.Sources {
		urls[i] = u.URL
	}
	return urls
}

func (s Spec) GetSourcesDir(rootDir string) string {
	if s.SourcesDir == "" {
		return rootDir
	}

	if filepath.IsAbs(s.SourcesDir) {
		return s.SourcesDir
	}

	return filepath.Join(rootDir, s.SourcesDir)
}

type SpecSource struct {
	File   string     `yaml:"file"` // File path to the openapi specification
	URL    string     `yaml:"url"`  // URL to the openapi specification
	Format SourceType `yaml:"format" default:"spec"`
	Type   SpecType   `yaml:"type"`
}

type Customization struct {
	Title       string                `yaml:"title"`
	Summary     string                `yaml:"summary"`
	Description string                `yaml:"description"`
	Version     string                `yaml:"version"`
	Contact     CustomizationContact  `yaml:"contact"`
	License     CustomizationLicense  `yaml:"license"`
	Servers     []CustomizationServer `yaml:"servers"`

	// Prune operations, tags and schemas
	PruneOperations []string `yaml:"pruneOperations"`
	PruneTags       []string `yaml:"pruneTags"`
	PruneSchemas    []string `yaml:"pruneSchemas"`
}

type CustomizationContact struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Email string `yaml:"email"`
}

type CustomizationLicense struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	Identifier string `yaml:"identifier"`
}

type CustomizationServer struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
}

type GeneratorConfig struct {
	GeneratorName         string                 `json:"generatorName" yaml:"generatorName"`
	InvokerPackage        string                 `json:"invokerPackage" yaml:"invokerPackage"`
	ApiPackage            string                 `json:"apiPackage" yaml:"apiPackage"`
	ModelPackage          string                 `json:"modelPackage" yaml:"modelPackage"`
	EnablePostProcessFile bool                   `json:"enablePostProcessFile" yaml:"enablePostProcessFile"`
	GlobalProperty        map[string]interface{} `json:"globalProperty" yaml:"globalProperty"`
	AdditionalProperties  map[string]interface{} `json:"additionalProperties" yaml:"additionalProperties"`
}

type GeneratorArgs struct {
	// UserArgs are the arguments that are passed to the generator
	OpenAPIGeneratorArgs []string `yaml:"openapi_generator"`
}

func FromString(content string) (Configuration, error) {
	var config Configuration
	err := yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// repository defaults
	if config.Repository.Name == "" {
		config.Repository.Name = config.Name
	}
	if config.Repository.Description == "" {
		config.Repository.Description = config.Summary
	}

	// spec defaults
	for i, _ := range config.Spec.Sources {
		if config.Spec.Sources[i].Format == "" {
			config.Spec.Sources[i].Format = SourceTypeSpec
		}
	}
	if config.Spec.Customization.Title == "" {
		config.Spec.Customization.Title = config.Name
	}
	if config.Spec.File == "" {
		config.Spec.File = "openapi.yaml"
	}

	return config, nil
}
