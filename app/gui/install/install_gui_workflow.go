package install

import (
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"k8s-management-go/app/actions/install_actions"
	"k8s-management-go/app/actions/namespace_actions"
	"k8s-management-go/app/gui/ui_elements"
	"k8s-management-go/app/models"
	"k8s-management-go/app/utils/logger"
	"k8s-management-go/app/utils/loggingstate"
	"time"
)

// execute the workflow
func ExecuteInstallWorkflow(window fyne.Window, state models.StateData) (err error) {
	log := logger.Log()

	// Progress Bar
	progressCnt := 1
	progressMaxCnt := install_actions.CalculateBarCounter(state)
	bar := dialog.NewProgress(state.HelmCommand, "Installing on namespace "+state.Namespace, window)
	bar.Show()

	// it is not a dry-run -> install_actions required stuff
	if !models.GetConfiguration().K8sManagement.DryRunOnly {
		// check if namespace is available or create a new one if not
		err = namespace_actions.ProcessNamespaceCreation(state)
		bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
		progressCnt++
		if err != nil {
			bar.Hide()
			ui_elements.ShowLogOutput(window)
			return err
		}

		// check if PVC was specified and install_actions it if needed
		err = install_actions.ProcessCheckAndCreatePvc(state)
		bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
		progressCnt++
		if err != nil {
			bar.Hide()
			ui_elements.ShowLogOutput(window)
			return err
		}

		// Jenkins exists and it is not a dry-run install_actions secrets
		if state.JenkinsHelmValuesExist {
			// apply secrets
			err = install_actions.ProcessCreateSecrets(state)
			bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
			progressCnt++
			if err != nil {
				bar.Hide()
				ui_elements.ShowLogOutput(window)
				return err
			}
		}
	} else {
		loggingstate.AddInfoEntry("-> Dry run. Skipping namespace creation, pvc installation and secrets apply...")
		log.Infof("[DoUpgradeOrInstall] Dry run only, skipping namespace [%s] creation, pvc installation and secrets apply...", state.Namespace)
	}

	// install_actions Jenkins
	err = install_actions.ProcessInstallJenkins(state.HelmCommand, state)
	bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
	progressCnt++
	if err != nil {
		bar.Hide()
		ui_elements.ShowLogOutput(window)
		return err
	}

	// install_actions Nginx ingress controller
	err = install_actions.ProcessNginxController(state.HelmCommand, state)
	bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
	progressCnt++
	if err != nil {
		bar.Hide()
		log.Errorf("[DoUpgradeOrInstall] Unable to install_actions nginx-ingress-controller.\n%s", err.Error())
		ui_elements.ShowLogOutput(window)
		return err
	}

	// last but not least execute install_actions scripts if it is not dry-run only
	err = install_actions.ProcessScripts(state)
	bar.SetValue(float64(1) / float64(progressMaxCnt) * float64(progressCnt))
	time.Sleep(time.Duration(1) * time.Second)
	bar.Hide()

	ui_elements.ShowLogOutput(window)

	return err
}
