package namespace

import (
	"k8s-management-go/app/actions/install"
	"k8s-management-go/app/actions/namespaceactions"
	"k8s-management-go/app/cli/createproject"
	"k8s-management-go/app/utils/loggingstate"
)

// WorkflowCreateNamespace is the workflow to create a namespace
func WorkflowCreateNamespace() (err error) {
	var projectConfig = install.NewInstallProjectConfig()
	namespace, err := createproject.NamespaceWorkflow()
	projectConfig.Project.SetNamespace(namespace)

	if err != nil {
		loggingstate.AddErrorEntryAndDetails("-> AskForNamespace dialog aborted...", err.Error())
	}

	err = namespaceactions.ProcessNamespaceCreation(projectConfig)

	return nil
}
