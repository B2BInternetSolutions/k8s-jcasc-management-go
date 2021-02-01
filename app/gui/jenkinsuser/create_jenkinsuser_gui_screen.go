package jenkinsuser

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	"k8s-management-go/app/utils/encryption"
	"k8s-management-go/app/utils/loggingstate"
	"k8s-management-go/app/utils/validator"
)

// ScreenJenkinsUserPasswordCreate shows the Jenkins user psasword creation screen
func ScreenJenkinsUserPasswordCreate(window fyne.Window) fyne.CanvasObject {
	// UI elements
	var passwordErrorLabel = widget.NewLabel("")
	// secrets password
	var userPasswordEntry = widget.NewPasswordEntry()
	var userConfirmPasswordEntry = widget.NewPasswordEntry()

	var form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password", Widget: userPasswordEntry},
			{Text: "Confirm password", Widget: userConfirmPasswordEntry},
			{Text: "", Widget: passwordErrorLabel},
		},
		OnSubmit: func() {
			isValid, errMessage := validator.ValidateConfirmPasswords(userPasswordEntry.Text, userConfirmPasswordEntry.Text)
			passwordErrorLabel.SetText(errMessage)
			if isValid {
				// Encrypt password
				hashedPassword, err := encryption.EncryptJenkinsUserPassword(userPasswordEntry.Text)
				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Prepare dialog to show result
					var encPassEntry = widget.NewEntry()
					encPassEntry.Text = hashedPassword

					var encPassBox = container.NewVBox(
						container.NewHBox(layout.NewSpacer()),
						encPassEntry,
						container.NewHBox(layout.NewSpacer()),
						widget.NewLabelWithStyle("Do you want to copy the password to clipboard?", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
					)
					dialog.ShowCustomConfirm("Your encrypted password",
						"Copy it!",
						"No thanks!",
						encPassBox,
						func(wantCopy bool) {
							if wantCopy {
								err = clipboard.WriteAll(hashedPassword)
								if err != nil {
									loggingstate.AddErrorEntryAndDetails("-> Unable to copy password to clipboard", err.Error())
								}
							}
							loggingstate.LogLoggingStateEntries()
						},
						window)
				}
			}
		},
	}

	return container.NewVBox(
		widget.NewLabel(""),
		form,
	)
}
