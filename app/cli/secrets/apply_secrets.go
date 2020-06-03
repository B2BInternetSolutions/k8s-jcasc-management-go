package secrets

import (
	"errors"
	"k8s-management-go/app/cli/dialogs"
	"k8s-management-go/app/models/config"
	"k8s-management-go/app/utils/logger"
	"os"
	"os/exec"
)

func ApplySecrets() (info string, err error) {
	log := logger.Log()
	// select namespace
	namespace, err := dialogs.DialogAskForNamespace()
	if err != nil {
		log.Error(err)
		return info, err
	}
	info, err = ApplySecretsToNamespace(namespace)
	return info, err
}

// apply secrets to one namespace
func ApplySecretsToNamespace(namespace string) (info string, err error) {
	// Decrypt secrets file
	infoLog, err := DecryptSecretsFile()
	info = info + infoLog

	// apply secret to namespace
	secretsFilePath := config.GetGlobalSecretsFile()
	infoLog, nsErr := applySecretsToNamespace(secretsFilePath, namespace)
	info = info + infoLog

	// delete decrypted file
	rmErr := os.Remove(secretsFilePath)

	// Error handling for apply and remove
	if nsErr != nil {
		err = nsErr
	}
	if rmErr != nil {
		if err != nil {
			err = errors.New(err.Error() + rmErr.Error())
		} else {
			err = rmErr
		}
	}

	return info, err
}

// apply secrets to all namespaces
func ApplySecretsToAllNamespaces() (info string, err error) {
	// apply secret to namespaces
	infos := ""
	nsErrs := ""
	secretsFilePath := config.GetGlobalSecretsFile()
	for _, ip := range config.GetIpConfiguration().Ips {
		infoNs, nsErr := applySecretsToNamespace(secretsFilePath, ip.Namespace)
		if infoNs != "" {
			infos = infos + "\n" + infoNs
		}
		if nsErr != nil {
			nsErrs = nsErrs + "\n" + nsErr.Error()
		}
	}

	// delete decrypted file
	rmErr := os.Remove(secretsFilePath)

	// Error handling for apply and remove
	if nsErrs != "" {
		err = errors.New(nsErrs)
	}
	if rmErr != nil {
		if err != nil {
			err = errors.New(err.Error() + rmErr.Error())
		} else {
			err = rmErr
		}
	}

	return info, err
}

// execute the secrets file
func applySecretsToNamespace(secretsFilePath string, namespace string) (info string, err error) {
	log := logger.Log()
	// execute decrypted file
	cmd := exec.Command("sh", "-c", secretsFilePath)
	cmd.Env = append(os.Environ(),
		"NAMESPACE="+namespace,
	)
	err = cmd.Run()
	if err != nil {
		log.Error(err)
	} else {
		info = "Secrets to namespace [" + namespace + "] applied"
	}

	return info, err
}
