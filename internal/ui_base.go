// CryoUtilities
// Copyright (C) 2023 CryoByte33 and contributors to the CryoUtilities project

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package internal

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func InitUI() {
	// Create a Fyne application
	screenSizer := NewScreenSizer()
	screenSizer.UpdateScaleForActiveMonitor()
	fyneApp := app.NewWithID("io.cryobyte.cryoutilities")
	CryoUtils.App = fyneApp
	CryoUtils.App.SetIcon(ResourceIconPng)

	// Set system font
	systemFont := fyne.SystemFont()
	fyneApp.Settings().SetTheme(theme.DefaultTheme())
	fyneApp.Settings().SetTheme(theme.CustomThemeWithParams("SystemFont", map[string]interface{}{"DefaultFont": systemFont}))

	// Show and run the app
	title := "CryoUtilities " + CurrentVersionNumber
	CryoUtils.MainWindow = fyneApp.NewWindow(title)
	CryoUtils.makeUI()
	CryoUtils.MainWindow.CenterOnScreen()
	CryoUtils.MainWindow.ShowAndRun()
}

func (app *Config) makeUI() {
	app.authUI()

	// Show a disclaimer that I'm not responsible for damage.
	dialog.ShowConfirm("免责声明",
		"此脚本由 CryoByte33 制作，Juij 汉化，用于调整 Steam Deck 上交换文件的大小。\n\n"+
			"免责声明：对任何人造成的损害不承担任何责任\n"+
			"执行此操作的设备，所有责任均由用户承担。\n\n"+
			"接受这些条款吗？",
		func(b bool) {
			if !b {
				presentErrorInUI(errors.New("不接受条款"), CryoUtils.MainWindow)
				CryoUtils.MainWindow.Close()
			} else {
				CryoUtils.InfoLog.Println("接受条款，继续...")
			}
		},
		CryoUtils.MainWindow,
	)

	// Create and size a Fyne window
	CryoUtils.MainWindow.Resize(fyne.NewSize(700, 410))
	CryoUtils.MainWindow.SetFixedSize(true)
	CryoUtils.MainWindow.SetMaster()
}

func (app *Config) mainUI() {
	// Create heading section
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("主页", theme.HomeIcon(), app.homeTab()),
		container.NewTabItemWithIcon("交换文件", theme.MailReplyAllIcon(), app.swapTab()),
		container.NewTabItemWithIcon("内存", theme.ComputerIcon(), app.memoryTab()),
		container.NewTabItemWithIcon("存储", theme.StorageIcon(), app.storageTab()),
		container.NewTabItemWithIcon("显存", theme.ViewFullScreenIcon(), app.vramTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	finalContent := container.NewVBox(tabs)
	app.MainWindow.SetContent(finalContent)
}

func (app *Config) authUI() {
	// Refactor this, duplicated code.
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.OnSubmitted = func(s string) {
		CryoUtils.InfoLog.Println("检测密码...")
		err := testAuth(s)
		if err != nil {
			CryoUtils.InfoLog.Println("密码无效，请重新输入...")
			dialog.ShowInformation("密码错误", "密码错误，请重试。",
				CryoUtils.MainWindow)
		} else {
			CryoUtils.InfoLog.Println("密码有效，继续...")
			CryoUtils.UserPassword = s
			app.mainUI()
		}
	}
	passwordButton := widget.NewButton("提交", func() {
		CryoUtils.InfoLog.Println("检测密码...")
		err := testAuth(passwordEntry.Text)
		if err != nil {
			CryoUtils.InfoLog.Println("密码无效，请重新输入...")
			dialog.ShowInformation("密码错误", "密码错误，请重试。",
				CryoUtils.MainWindow)
		} else {
			CryoUtils.InfoLog.Println("密码有效，继续...")
			CryoUtils.UserPassword = passwordEntry.Text
			app.mainUI()
		}
	})
	passwordVBox := container.NewVBox(passwordEntry, passwordButton)
	passwordContainer := widget.NewCard("输入你的 sudo/deck 密码", "输入你的 sudo/deck 密码", passwordVBox)

	//  Add container to window

	app.MainWindow.SetContent(passwordContainer)
}
