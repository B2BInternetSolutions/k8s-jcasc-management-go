package install

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s-management-go/app/cli/loggingstate"
	"k8s-management-go/app/constants"
	"k8s-management-go/app/models"
	"k8s-management-go/app/utils/files"
	"k8s-management-go/app/utils/kubectl"
	"k8s-management-go/app/utils/logger"
)

type PvcClaimValuesYaml struct {
	Kind       string
	ApiVersion string
	Metadata   struct {
		Name      string
		Namespace string
		Labels    map[string]string
	}
	Spec struct {
		AccessModes      []string
		StorageClassName string
		Resources        struct {
			Requests struct {
				Storage string
			}
		}
	}
}

// install PVC is needed
func PersistenceVolumeClaimInstall(namespace string) (err error) {
	log := logger.Log()
	loggingstate.AddInfoEntry(fmt.Sprintf(" -> Check if PVC should be installed on namespace [%s]", namespace))
	log.Infof("[PVC Install] Check if PVC should be installed on namespace [%s]", namespace)

	// prepare file directories
	projectDir := files.AppendPath(models.GetProjectBaseDirectory(), namespace)
	pvcClaimValuesFilePath := files.AppendPath(projectDir, constants.FilenamePvcClaim)

	// open file
	if files.FileOrDirectoryExists(pvcClaimValuesFilePath) {
		loggingstate.AddInfoEntry(fmt.Sprintf("  -> Kubernetes PVC specification found for namespace [%s]...", namespace))
		log.Infof("[PVC Install] Kubernetes PVC specification found for namespace [%s]...", namespace)
		// variable to check, if pvc already exists
		pvcName, err := readPvcNameFromFile(pvcClaimValuesFilePath)
		if err != nil {
			return err
		}

		// if no name was found, something was wrong here...
		if pvcName == "" {
			loggingstate.AddErrorEntry(fmt.Sprintf("  -> PVC specification was found for namespace [%s], but no name was specified.", namespace))
			err = errors.New(fmt.Sprintf("[PVC Install] PVC specification was found for namespace [%s], but no name was specified.", namespace))
			log.Errorf("%s", err.Error())
			return err
		}

		// check if pvc is already available in namespace
		loggingstate.AddInfoEntry(fmt.Sprintf("  -> Checking if PVC [%s] is already available for namespace [%s].", pvcName, namespace))
		log.Infof("[PVC Install] Checking if PVC [%s] is already available for namespace [%s].", pvcName, namespace)
		pvcExists, err := isPvcAvailableInNamespace(namespace, pvcName)

		// no PVC found, so install it
		if !pvcExists {
			loggingstate.AddInfoEntry(fmt.Sprintf("  -> PVC [%s] does not exist in namespace [%s]. Trying to install it...", pvcName, namespace))
			log.Infof("[PVC Install] PVC [%s] does not exist in namespace [%s]. Trying to install it...", pvcName, namespace)

			// executing command
			kubectlCmdArgs := []string{
				"-n", namespace,
				"-f", pvcClaimValuesFilePath,
			}
			if _, err := kubectl.ExecutorKubectl("apply", kubectlCmdArgs); err != nil {
				loggingstate.AddErrorEntryAndDetails(fmt.Sprintf("  -> Cannot create PVC [%s] for namespace [%s]", pvcName, namespace), err.Error())
				log.Errorf("[PVC Install] Cannot create PVC [%s] for namespace [%s]", pvcName, namespace)
				return err
			}

			loggingstate.AddInfoEntry(fmt.Sprintf("  -> PVC [%s] does not exist in namespace [%s]. Trying to install it...done", pvcName, namespace))
			log.Infof("[PVC Install] PVC [%s] does not exist in namespace [%s]. Trying to install it...done", pvcName, namespace)
		} else {
			loggingstate.AddInfoEntry(fmt.Sprintf("  -> PVC [%s] in namespace [%s] found. No need to install it...", pvcName, namespace))
			log.Infof("[PVC Install] PVC [%s] in namespace [%s] found. No need to install it...", pvcName, namespace)
		}
	}

	return err
}

// read PVC specification and find name
func readPvcNameFromFile(pvcClaimValuesFilePath string) (pvcName string, err error) {
	log := logger.Log()
	// read PVC claim values.yaml file
	yamlFile, err := ioutil.ReadFile(pvcClaimValuesFilePath)
	if err != nil {
		loggingstate.AddErrorEntryAndDetails(fmt.Sprintf("  -> Unable to read pvc file [%s]...", pvcClaimValuesFilePath), err.Error())
		log.Errorf("Unable to read pvc file [%s]...\n%s", pvcClaimValuesFilePath, err.Error())
		return pvcName, err
	}

	// parse YAML
	var pvcClaimValues PvcClaimValuesYaml
	err = yaml.Unmarshal(yamlFile, &pvcClaimValues)
	if err != nil {
		loggingstate.AddErrorEntryAndDetails(fmt.Sprintf("  -> Unable to unmarshal pvc yaml file [%s]...", pvcClaimValuesFilePath), err.Error())
		log.Errorf("Unable to unmarshal pvc file [%s]...\n%s", pvcClaimValuesFilePath, err.Error())
		return pvcName, err
	}
	pvcMetaName := pvcClaimValues.Metadata.Name

	return pvcMetaName, nil
}

// internal function to check if PVC is available in namespace
func isPvcAvailableInNamespace(namespace string, pvcName string) (pvcExists bool, err error) {
	pvcExists = false
	// read all pvc from K8S
	kubectlCmdArgs := []string{
		"-n", namespace,
		"pvc",
	}
	kubectlCmdOutput, err := kubectl.ExecutorKubectl("get", kubectlCmdArgs)
	if err != nil {
		return false, err
	}

	// check if output contains pvcName
	if kubectlCmdOutput != "" {
		pvcExists = kubectl.CheckIfKubectlOutputContainsValueForField(kubectlCmdOutput, constants.KubectlFieldName, pvcName)
	}
	return pvcExists, err
}
