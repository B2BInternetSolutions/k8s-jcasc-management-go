package secrets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"k8s-management-go/app/actions/namespaceactions"
	"k8s-management-go/app/actions/secretsactions"
	"k8s-management-go/app/configuration"
	"k8s-management-go/app/events"
	"k8s-management-go/app/gui/uielements"
	"k8s-management-go/app/utils/logger"
	"time"
)

// Namespace
var namespaceErrorLabel = widget.NewLabel("")
var namespaceSelectEntry = widget.NewSelectEntry([]string{})

// ScreenApplySecretsToAllNamespace shows the apply to all namespaces screen
func ScreenApplySecretsToAllNamespace(window fyne.Window) fyne.CanvasObject {
	// secrets password
	var secretsFiles = uielements.CreateSecretsFileEntry()
	var passwordEntry = widget.NewPasswordEntry()
	namespaceSelectEntry = uielements.CreateNamespaceSelectEntry(namespaceErrorLabel)

	var form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Secrets file", Widget: secretsFiles},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() {
			// first try to decrypt the file
			if err := secretsactions.ActionDecryptSecretsFile(passwordEntry.Text, secretsFiles.Selected); err == nil {
				// execute the file and apply to all namespaces
				var bar = uielements.ProgressBar{
					Bar:        widget.NewProgressBar(), //NewProgress("Apply secrets to all namespaces", "Progress", window),
					CurrentCnt: 0,
					MaxCount:   float64(len(configuration.GetConfiguration().K8SManagement.IPConfig.Deployments)),
				}
				bar.Bar.Show()
				_ = secretsactions.ActionApplySecretsToAllNamespaces(secretsFiles.Selected, bar.AddCallback)
				bar.Bar.Hide()
			}

			uielements.ShowLogOutput(window)
		},
	}

	return container.NewVBox(
		widget.NewLabel(""),
		form,
	)
}

// ScreenApplySecretsToNamespace shows the apply to one selected namespace screen
func ScreenApplySecretsToNamespace(window fyne.Window) fyne.CanvasObject {
	var secretsFiles = uielements.CreateSecretsFileEntry()
	var passwordEntry = widget.NewPasswordEntry()
	namespaceSelectEntry = uielements.CreateNamespaceSelectEntry(namespaceErrorLabel)

	var form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Secrets file", Widget: secretsFiles},
			{Text: "Namespace", Widget: namespaceSelectEntry},
			{Text: "", Widget: namespaceErrorLabel},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() {
			// first try to decrypt the file
			var err = secretsactions.ActionDecryptSecretsFile(passwordEntry.Text, secretsFiles.Selected)
			if err == nil {
				// execute the file
				_ = secretsactions.ActionApplySecretsToNamespace(namespaceSelectEntry.Text, secretsFiles.Selected)
			}
			// show output
			uielements.ShowLogOutput(window)
		},
	}

	return container.NewVBox(
		widget.NewLabel(""),
		form,
	)
}

func init() {
	createNamespaceNotifier := namespaceCreatedNotifier{}
	events.NamespaceCreated.Register(createNamespaceNotifier)
}

type namespaceCreatedNotifier struct {
	namespace string
}

func (notifier namespaceCreatedNotifier) Handle(payload events.NamespaceCreatedPayload) {
	logger.Log().Info("[secrets_gui] -> Retrieved event to that new namespace was created")
	if namespaceSelectEntry != nil {
		namespaceSelectEntry.SetOptions(namespaceactions.ActionReadNamespaceWithFilter(nil))
	}

	events.RefreshTabs.Trigger(events.RefreshTabsPayload{
		Time: time.Now(),
	})
}
