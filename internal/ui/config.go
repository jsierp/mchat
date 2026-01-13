package ui

import (
	"mchat/internal/config"

	"github.com/rivo/tview"
)

func (a *App) openMChatForm(page *tview.Flex) {
	form := tview.NewForm()
	username := tview.NewInputField().SetLabel("Username").SetFieldWidth(20)
	password := tview.NewInputField().SetLabel("Password").SetFieldWidth(20).SetMaskCharacter('*')
	form.AddFormItem(username).AddFormItem(password)
	form.AddButton("Save", func() {
		a.cfg = &config.Config{
			Login:    username.GetText(),
			Password: password.GetText(),
		}
		config.SaveConfig(a.cfg)
		a.pages.SwitchToPage("inbox")
	})
	page.Clear()
	page.AddItem(form, 0, 1, true)
	a.app.SetFocus(form)
}

func (a *App) openGooglePage(page *tview.Flex) {
	svc := config.NewGoogleAuthService()
	url := svc.GetGoogleUrl()
	urlView := tview.NewTextView().SetText("Go to the URL:\n\n" + url)
	page.Clear()
	page.AddItem(urlView, 0, 1, true)

	go func() {
		code := svc.WaitForAuthCode()
		cfg, _ := svc.ExchangeCode(code)
		a.cfg = cfg
		config.SaveConfig(cfg)

		a.app.QueueUpdateDraw(func() {
			a.pages.SwitchToPage("inbox")
			a.app.SetFocus(a.pages.GetPage("inbox"))
		})
	}()
}

func (a *App) initConfigPage() *tview.Flex {
	menu := tview.NewList()

	page := tview.NewFlex()
	menu.AddItem("MChat", "Use MChat Account", 'm', func() {
		a.openMChatForm(page)
	}).AddItem("Google", "Connect with Gmail Account", 'g', func() {
		a.openGooglePage(page)
	}).AddItem("Quit", "Exit the Application", 'q', func() {
		a.app.Stop()
	}).ShowSecondaryText(true)
	page.AddItem(menu, 0, 1, true)

	return page
}
