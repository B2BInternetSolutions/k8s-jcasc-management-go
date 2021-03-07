package configuration

import (
	"fmt"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s-management-go/app/constants"
	"k8s-management-go/app/utils/files"
	"k8s-management-go/app/utils/loggingstate"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// baseCustomConfig represents the base custom configuration to setup alternative project path and config file.
type baseCustomConfig struct {
	K8SManagement struct {
		ConfigFile string `yaml:"configFile,omitempty"`
		BasePath   string `yaml:"basePath,omitempty"`
	} `yaml:"k8sManagement,omitempty"`
}

// config represents the configuration files
type config struct {
	K8SManagement struct {
		Log struct {
			Level              string `yaml:"level,omitempty"`
			File               string `yaml:"file,omitempty"`
			Encoding           string `yaml:"encoding,omitempty"`
			OverwriteOnRestart bool   `yaml:"overwriteOnRestart,omitempty"`
		} `yaml:"log,omitempty"`
		Ipconfig struct {
			File        string `yaml:"file,omitempty"`
			DummyPrefix string `yaml:"dummyPrefix,omitempty"`
		} `yaml:"ipconfig,omitempty"`
		Project struct {
			BaseDirectory     string `yaml:"baseDirectory,omitempty"`
			TemplateDirectory string `yaml:"templateDirectory,omitempty"`
			SecretFiles       string `yaml:"secretFiles,omitempty"`
		} `yaml:"project,omitempty"`
		VersionCheck bool `yaml:"versionCheck,omitempty"`
		DryRunOnly   bool `yaml:"-"`
		CliOnly      bool `yaml:"-"`
	} `yaml:"k8sManagement,omitempty"`
	Jenkins struct {
		Jcasc struct {
			ConfigurationURL      string `yaml:"configurationUrl,omitempty"`
			AuthorizationStrategy struct {
				AllowAnonymousRead bool `yaml:"allowAnonymousRead,omitempty"`
			} `yaml:"authorizationStrategy,omitempty"`
			CredentialIDs struct {
				Docker string `yaml:"docker,omitempty"`
				Maven  string `yaml:"maven,omitempty"`
				Npm    string `yaml:"npm,omitempty"`
				Vcs    string `yaml:"vcs,omitempty"`
			} `yaml:"credentialIDs,omitempty"`
			SeedJobURL string `yaml:"seedJobURL,omitempty"`
		} `yaml:"jcasc,omitempty"`
		JobDSL struct {
			BaseURL             string `yaml:"baseURL,omitempty"`
			RepoValidatePattern string `yaml:"repoValidatePattern,omitempty"`
		} `yaml:"jobDSL,omitempty"`
		Controller struct {
			Passwords struct {
				AdminUser            string `yaml:"adminUser,omitempty"`
				AdminUserEncrypted   string `yaml:"adminUserEncrypted,omitempty"`
				DefaultUserEncrypted string `yaml:"defaultUserEncrypted,omitempty"`
			} `yaml:"passwords,omitempty"`
			CustomJenkinsLabel string `yaml:"customJenkinsLabel,omitempty"`
			DeploymentName     string `yaml:"deploymentName,omitempty"`
			DefaultURIPrefix   string `yaml:"defaultURIPrefix,omitempty"`
		} `yaml:"controller,omitempty"`
		Persistence struct {
			AccessMode   string `yaml:"accessMode,omitempty"`
			StorageClass string `yaml:"storageClass,omitempty"`
			StorageSize  string `yaml:"storageSize,omitempty"`
		} `yaml:"persistence,omitempty"`
		Container struct {
			Image      string `yaml:"image,omitempty"`
			Tag        string `yaml:"tag,omitempty"`
			PullPolicy string `yaml:"pullPolicy,omitempty"`
			PullSecret string `yaml:"pullSecret,omitempty"`
		} `yaml:"container,omitempty"`
	} `yaml:"jenkins,omitempty"`
	Nginx struct {
		Ingress struct {
			Container struct {
				Image      string `yaml:"image,omitempty"`
				PullSecret string `yaml:"pullSecret,omitempty"`
			} `yaml:"container,omitempty"`
			Deployment struct {
				ForEachNamespace bool   `yaml:"forEachNamespace,omitempty"`
				DeploymentName   string `yaml:"deploymentName,omitempty"`
			} `yaml:"deployment,omitempty"`
			Annotationclass string `yaml:"annotationclass,omitempty"`
		} `yaml:"ingress,omitempty"`
		Loadbalancer struct {
			Enabled bool `yaml:"enabled,omitempty"`
			Ports   struct {
				HTTP        uint64 `yaml:"http,omitempty"`
				HTTPTarget  uint64 `yaml:"httpTarget,omitempty"`
				HTTPS       uint64 `yaml:"https,omitempty"`
				HTTPSTarget uint64 `yaml:"httpsTarget,omitempty"`
			} `yaml:"ports,omitempty"`
			Annotations struct {
				Enabled bool `yaml:"enabled,omitempty"`
			} `yaml:"annotations,omitempty"`
			ExternalDNS struct {
				HostName string `yaml:"hostName,omitempty"`
				TTL      uint64 `yaml:"ttl,omitempty"`
			} `yaml:"externalDNS,omitempty"`
		} `yaml:"loadbalancer,omitempty"`
	} `yaml:"nginx,omitempty"`
	Kubernetes struct {
		Certificates struct {
			Default  string            `yaml:"default,omitempty"`
			Contexts map[string]string `yaml:"contexts,omitempty"`
		} `yaml:"certificates,omitempty"`
	} `yaml:"kubernetes,omitempty"`
	CustomConfig baseCustomConfig `yaml:"-"`
}

var conf *config

// EmptyConfiguration returns an empty instance of config. This should only be used for migration.
func EmptyConfiguration() config {
	return config{}
}

// GetConfiguration returns the current configuration.
func GetConfiguration() *config {
	return conf
}

// GetProjectBaseDirectory : Get project base directory with full path
func (conf *config) GetProjectBaseDirectory() string {
	return conf.calculateFullDirectoryPath(conf.K8SManagement.Project.BaseDirectory)
}

// GetProjectTemplateDirectory : Get project template directory with full path
func (conf *config) GetProjectTemplateDirectory() string {
	return conf.calculateFullDirectoryPath(conf.K8SManagement.Project.TemplateDirectory)
}

// GetIPConfigurationFile is a helper method for IP configuration file
func (conf *config) GetIPConfigurationFile() string {
	return conf.FilePathWithBasePath(conf.K8SManagement.Ipconfig.File)
}

// calculateFullDirectoryPath calculates full directory path
func (conf *config) calculateFullDirectoryPath(targetDir string) string {
	if strings.HasPrefix(targetDir, "./") {
		// if it starts with current directory add base path
		return conf.FilePathWithBasePath(targetDir)
	} else if strings.HasPrefix(targetDir, "../") {
		// if it starts with subdirectory add base path
		return conf.FilePathWithBasePath(targetDir)
	} else {
		// it seems to be an absolute path, use only the project directory
		return targetDir
	}
}

// GetGlobalSecretsPath returns the path for the global secrets files.
func (conf *config) GetGlobalSecretsPath() (secretsFilePath string) {
	var globalSecretsFile = conf.getGlobalSecretsFile()
	var fileName = filepath.Base(globalSecretsFile)
	return strings.Replace(globalSecretsFile, fileName, "", -1)
	// return fmt.Sprintf("%s%s", secretsFilePath, string(os.PathSeparator))
}

// GetSecretsFiles returns a list of secret files to support different environments
func (conf *config) GetSecretsFiles() []string {
	secretsFilePath := conf.GetGlobalSecretsPath()

	if secretsFilePath != "" {
		var filePrefix = "secrets_"
		var fileFilter = files.FileFilter{
			Prefix: &filePrefix,
		}
		var secretFilesWithoutPath, err = files.ListFilesOfDirectoryWithFilter(secretsFilePath, &fileFilter)
		if err != nil {
			loggingstate.AddErrorEntryAndDetails(
				"Unable to filter for secrets files",
				fmt.Sprintf("SearchDirectory: [%s]", secretsFilePath))
		}
		var secretFilesWithPath []string
		if secretFilesWithoutPath != nil && len(*secretFilesWithoutPath) > 0 {
			for _, secretFileOfFilter := range *secretFilesWithoutPath {
				secretFilesWithPath = append(secretFilesWithPath, conf.GetGlobalSecretsPath()+secretFileOfFilter)
			}
		}

		var secretFiles []string

		secretFiles = appendUnique(secretFiles, strings.Replace(conf.getGlobalSecretsFile(), secretsFilePath, "", -1))
		if secretFilesWithPath != nil && len(secretFilesWithPath) > 0 {
			for _, secretFile := range secretFilesWithPath {
				secretFile = strings.Replace(secretFile, secretsFilePath, "", -1)
				secretFile = strings.Replace(secretFile, ".gpg", "", -1)

				secretFiles = appendUnique(secretFiles, secretFile)
			}
		}

		return secretFiles
	}
	return nil
}

// GetGlobalSecretsFile is a helper method for secrets file
func (conf *config) getGlobalSecretsFile() string {
	var globalSecretsFile = conf.K8SManagement.Project.SecretFiles
	if globalSecretsFile == "" {
		panic("The configured secrets file is not a file!")
	}
	return conf.FilePathWithBasePath(globalSecretsFile)
}

// SetDryRun set the dry run option
func (conf *config) SetDryRun(dryRun bool) {
	conf.K8SManagement.DryRunOnly = dryRun
}

// FilePathWithBasePath is a helper method to calculate the correct filepath
func (conf *config) FilePathWithBasePath(configurationFilePath string) string {
	var resultConfigurationFilePath = configurationFilePath
	if conf.K8SManagement.Project.BaseDirectory != "" {
		resultConfigurationFilePath = files.AppendPath(conf.K8SManagement.Project.BaseDirectory, configurationFilePath)
	}

	// check if path exists. else try to check if the path was related to current path.
	// this helps to support to read configuration from other paths and templates from
	// this project path
	if !files.FileOrDirectoryExists(resultConfigurationFilePath) {
		currentDirectory, _ := os.Getwd()
		var currentDirFilePath = files.AppendPath(currentDirectory, configurationFilePath)
		if files.FileOrDirectoryExists(currentDirFilePath) {
			resultConfigurationFilePath = currentDirFilePath
		}
	}
	return resultConfigurationFilePath
}

// LoadConfiguration loads the base configuration and merges it with alternative configurations if defined.
func LoadConfiguration(basePath string, dryRunDebug bool, cliOnly bool) {
	conf = &config{}
	conf.initBaseConfig(basePath)
	conf.K8SManagement.DryRunOnly = dryRunDebug
	conf.K8SManagement.CliOnly = cliOnly
}

func (conf *config) initBaseConfig(basePath string) {
	// read main configuration
	if err := readConfigFromYAMLFile(files.AppendPath(files.AppendPath(basePath, constants.DirConfig), constants.FilenameConfigurationYaml), conf); err != nil {
		log.Panicf("Unable to load base configuration: %v", err.Error())
	}

	// read alternative base config
	var baseCfg = baseCustomConfig{}
	var alternativeFile = files.AppendPath(files.AppendPath(basePath, constants.DirConfig), constants.FilenameConfigurationCustomYaml)
	if files.FileOrDirectoryExists(alternativeFile) {
		if err := readConfigFromYAMLFile(alternativeFile, &baseCfg); err != nil {
			log.Panicf("Unable to load alternative base configuration: %v", err.Error())
		}
		conf.CustomConfig = baseCfg

		// set base path to current path if nothing else was specified
		if conf.CustomConfig.K8SManagement.BasePath == "" {
			conf.CustomConfig.K8SManagement.BasePath = basePath
		}
		conf.K8SManagement.Project.BaseDirectory = conf.CustomConfig.K8SManagement.BasePath
	}

	// load custom configuration if found.
	if conf.CustomConfig.K8SManagement.ConfigFile != "" {
		var customConfig = files.AppendPath(conf.CustomConfig.K8SManagement.BasePath, conf.CustomConfig.K8SManagement.ConfigFile)
		if files.FileOrDirectoryExists(customConfig) {
			var customCfg = config{}
			if err := readConfigFromYAMLFile(customConfig, &customCfg); err != nil {
				log.Panicf("Unable to load custom configuration: %v", err.Error())
			}
			if err := mergo.Merge(conf, customCfg, mergo.WithOverride); err != nil {
				log.Panicf("Unable to merge custom config with config: %v", err)
			}
		} else {
			log.Panicf("Unable to load defined custom config from path [%v]", customConfig)
		}
	}
}

func readConfigFromYAMLFile(file string, target interface{}) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, target)
	if err != nil {
		return err
	}
	return nil
}

func appendUnique(slice []string, element string) []string {
	for _, sliceElement := range slice {
		if sliceElement == element {
			return slice
		}
	}
	return append(slice, element)
}
