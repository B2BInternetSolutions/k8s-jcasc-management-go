package uninstall

import (
	"k8s-management-go/app/models"
	"k8s-management-go/app/utils/helm"
	"k8s-management-go/app/utils/logger"
)

// uninstall Jenkins with Helm
func HelmUninstallNginxIngressController(namespace string) (err error) {
	log := logger.Log()
	log.Info("[Uninstall NginxIngressCtrl] Try to uninstall nginx-ingress-controller in namespace [" + namespace + "]...")

	// prepare Helm command
	helmCmdArgs := []string{
		models.GetConfiguration().Nginx.Ingress.Controller.DeploymentName,
		"-n", namespace,
	}
	// add dry-run flags if necessary
	if models.GetConfiguration().K8sManagement.DryRunOnly {
		helmCmdArgs = append(helmCmdArgs, "--dry-run", "--debug")
	}
	// execute helm command
	if err = helm.ExecutorHelm("uninstall", helmCmdArgs); err != nil {
		return err
	}
	log.Info("[Uninstall NginxIngressCtrl] Uninstall of nginx-ingress-controller in namespace [" + namespace + "] done...")

	return nil
}
