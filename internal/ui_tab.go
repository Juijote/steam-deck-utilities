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
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
	
	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/NotoSansSC.ttf")
}

// Home tab for "recommended" and "default" buttons
func (app *Config) homeTab() *fyne.Container {
	welcomeText := canvas.NewText("欢迎来到 CryoUtilities 冻结实用程序！", White)
	welcomeText.TextSize = HeaderTextSize
	welcomeText.TextStyle.Bold = true

	subheadingText := canvas.NewText("快速设置，使用窗口顶部的选项卡来使用 "+
		"单独设置", White)
	subheadingText.TextSize = SubHeadingTextSize

	availableSpace, err := getFreeSpace("/home")
	if err != nil {
		presentErrorInUI(err, app.MainWindow)
	}
	var chosenSize string
	if availableSpace < RecommendedSwapSizeBytes {
		availableSizes, _ := getAvailableSwapSizes()
		chosenSize = strings.Fields(availableSizes[len(availableSizes)-1])[0]
	} else {
		chosenSize = strconv.Itoa(RecommendedSwapSize)
	}

	actionText := widget.NewLabel(
		"交换大小: " + chosenSize + "GB\n" +
			"交换性: " + RecommendedSwappiness + "\n" +
			"大页面: 启用\n" +
			"主动压缩: " + RecommendedCompactionProactiveness + "\n" +
			"大页碎片整理: 禁用\n" +
			"页锁不公平: " + RecommendedPageLockUnfairness + "\n" +
			"大页面中的共享内存: 启用")

	recommendedButton := widget.NewButton("Recommended", func() {
		progressGroup := container.NewVBox(
			canvas.NewText("正在应用推荐设置...", White),
			actionText,
			widget.NewProgressBarInfinite())
		modal := widget.NewModalPopUp(progressGroup, CryoUtils.MainWindow.Canvas())
		modal.Show()
		renewSudoAuth()
		err := UseRecommendedSettings()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		modal.Hide()
		app.refreshAllContent()
		dialog.ShowInformation(
			"成功!",
			"应用推荐设置!",
			CryoUtils.MainWindow,
		)
	})
	stockButton := widget.NewButton("Stock", func() {
		progressText := canvas.NewText("恢复到默认设置...", White)
		progressBar := widget.NewProgressBarInfinite()
		progressGroup := container.NewVBox(progressText, progressBar)
		modal := widget.NewModalPopUp(progressGroup, CryoUtils.MainWindow.Canvas())
		modal.Show()
		renewSudoAuth()
		err := UseStockSettings()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		modal.Hide()
		app.refreshAllContent()
		dialog.ShowInformation(
			"成功!",
			"已恢复到默认设置!",
			CryoUtils.MainWindow,
		)
	})

	recommendedSettings := widget.NewCard("推荐设置", "将所有设置设置为 "+
		"CryoByte33（作者） 的建议。", recommendedButton)
	stockSettings := widget.NewCard("默认设置", "将所有设置重置为 V社 默认值，不包含 "+
		"“游戏数据” 选项卡/位置。", stockButton)

	homeVBox := container.NewVBox(
		welcomeText,
		subheadingText,
		recommendedSettings,
		stockSettings,
	)
	app.HomeContainer = homeVBox

	return homeVBox
}

// Swap tab for all swap-related tasks.
func (app *Config) swapTab() *fyne.Container {
	app.SwapText = canvas.NewText("交换文件大小: 未知", Gray)
	app.SwappinessText = canvas.NewText("交换性: 未知", Gray)
	// Main content including buttons to resize swap and change swappiness
	swapResizeButton := widget.NewButton("调整大小", func() {
		swapSizeWindow()
		app.refreshSwapContent()
	})
	swappinessChangeButton := widget.NewButton("变更", func() {
		swappinessWindow()
		app.refreshSwappinessContent()
	})

	swapCard := widget.NewCard("交换文件", "调整交换文件的大小。", swapResizeButton)
	swappinessCard := widget.NewCard("交换性", "调整交换值。", swappinessChangeButton)

	// Swap info gathering
	app.refreshSwapContent()
	app.refreshSwappinessContent()

	app.SwapBar = container.NewGridWithColumns(2,
		container.NewCenter(app.SwapText),
		container.NewCenter(app.SwappinessText))

	topBar := container.NewVBox(
		container.NewGridWithRows(1),
		container.NewGridWithRows(1, container.NewCenter(canvas.NewText("当前交换状态:", White))),
		app.SwapBar,
	)

	swapVBox := container.NewVBox(
		swapCard,
		swappinessCard,
	)

	full := container.NewBorder(topBar, nil, nil, nil, swapVBox)

	return full
}

// Game Data tab to move and delete prefixes and shadercache.
func (app *Config) storageTab() *fyne.Container {
	// These can take a minute to come up, so create a loading bar to show things are happening.
	syncDataButton := widget.NewButton("同步", func() {
		progressText := canvas.NewText("正在计算设备状态...", White)
		progressBar := widget.NewProgressBarInfinite()
		progressGroup := container.NewVBox(progressText, progressBar)
		modal := widget.NewModalPopUp(progressGroup, CryoUtils.MainWindow.Canvas())
		modal.Show()
		syncGameDataWindow()
		modal.Hide()
	})
	cleanupDataButton := widget.NewButton("清除", func() {
		progressText := canvas.NewText("正在计算设备状态...", White)
		progressBar := widget.NewProgressBarInfinite()
		progressGroup := container.NewVBox(progressText, progressBar)
		modal := widget.NewModalPopUp(progressGroup, CryoUtils.MainWindow.Canvas())
		modal.Show()
		cleanupDataWindow()
		modal.Hide()
	})

	syncData := widget.NewCard("同步游戏数据", "将前缀和着色器同步到游戏所在的设备上 "+
		"已安装", syncDataButton)
	cleanStaleData := widget.NewCard("删除游戏数据", "删除选定游戏的前缀和着色器。",
		cleanupDataButton)

	gameDataVBox := container.NewVBox(
		syncData,
		cleanStaleData,
	)
	app.GameDataContainer = gameDataVBox

	return gameDataVBox
}

// Tab for non-swap, memory-related tweaks.
func (app *Config) memoryTab() *fyne.Container {
	app.HugePagesText = canvas.NewText("大页面 (THP)", Red)
	app.ShMemText = canvas.NewText("THP 中的共享内存", Red)
	app.CompactionProactivenessText = canvas.NewText("主动压缩", Red)
	app.DefragText = canvas.NewText("碎片整理", Red)
	app.PageLockUnfairnessText = canvas.NewText("页面锁不公平", Red)

	CryoUtils.HugePagesButton = widget.NewButton("启用大页面", func() {
		renewSudoAuth()
		err := ToggleHugePages()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		app.refreshHugePagesContent()
	})

	CryoUtils.ShMemButton = widget.NewButton("在 THP 中启用共享内存", func() {
		renewSudoAuth()
		err := ToggleShMem()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		app.refreshShMemContent()
	})

	CryoUtils.CompactionProactivenessButton = widget.NewButton("设置压缩主动性", func() {
		renewSudoAuth()
		err := ToggleCompactionProactiveness()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		app.refreshCompactionProactivenessContent()
	})

	CryoUtils.DefragButton = widget.NewButton("禁用大页面碎片整理", func() {
		renewSudoAuth()
		err := ToggleDefrag()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		app.refreshDefragContent()
	})

	CryoUtils.PageLockUnfairnessButton = widget.NewButton("设置页面锁定不公平", func() {
		renewSudoAuth()
		err := TogglePageLockUnfairness()
		if err != nil {
			presentErrorInUI(err, CryoUtils.MainWindow)
		}
		app.refreshPageLockUnfairnessContent()
	})

	app.refreshHugePagesContent()
	app.refreshCompactionProactivenessContent()
	app.refreshShMemContent()
	app.refreshDefragContent()
	app.refreshPageLockUnfairnessContent()

	app.MemoryBar = container.NewGridWithColumns(5,
		container.NewCenter(app.HugePagesText),
		container.NewCenter(app.ShMemText),
		container.NewCenter(app.CompactionProactivenessText),
		container.NewCenter(app.DefragText),
		container.NewCenter(app.PageLockUnfairnessText))
	topBar := container.NewVBox(
		container.NewGridWithRows(1),
		container.NewGridWithRows(1, container.NewCenter(canvas.NewText("当前调整状态:", White))),
		app.MemoryBar,
	)

	hugePagesCard := widget.NewCard("大页面", "切换大页面", app.HugePagesButton)
	shMemCard := widget.NewCard("THP 中的共享内存", "在 THP 中切换共享内存", app.ShMemButton)
	compactionProactivenessCard := widget.NewCard("主动压缩", "设置主动压缩", app.CompactionProactivenessButton)
	defragCard := widget.NewCard("大页面碎片整理", "切换大页面碎片整理", app.DefragButton)
	pageLockUnfairnessCard := widget.NewCard("页面锁不公平", "设置页面锁定不公平", app.PageLockUnfairnessButton)

	memoryVBox := container.NewVBox(
		hugePagesCard,
		shMemCard,
		compactionProactivenessCard,
		defragCard,
		pageLockUnfairnessCard,
	)
	scroll := container.NewScroll(memoryVBox)
	full := container.NewBorder(topBar, nil, nil, nil, scroll)

	return full
}

func (app *Config) vramTab() *fyne.Container {
	app.VRAMText = canvas.NewText("当前显存大小: 未知", Gray)

	// Get VRAM value
	app.refreshVRAMContent()

	textHowTo := widget.NewLabel("1. 关闭 Steam Deck\n\n" +
			"2. 按住音量增大按钮，按下电源按钮，然后松开两个按钮\n\n" +
			"3. Setup Utility -> 进阶 -> UMA Frame Buffer Size")
	
	textRecommended := widget.NewLabelWithStyle("大多数情况下建议使用 4G 设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	textWarning := widget.NewLabel("请注意，某些游戏 (RDR2) 可能会出现性能下降的情况。")


	textVBox := container.NewVBox(
		textHowTo,
		textRecommended,
		textWarning,
	)

	vramCard := widget.NewCard("低显存", "如何更改最小 VRAM:", textVBox)

	vramBAR := container.NewGridWithColumns(1,
		container.NewCenter(app.VRAMText))
	topBar := container.NewVBox(
		container.NewGridWithRows(1),
		container.NewGridWithRows(1, container.NewCenter(canvas.NewText("当前调整状态:", White))),
		vramBAR,
	)

	vramVBOX := container.NewVBox(
		vramCard,
	)
	scroll := container.NewScroll(vramVBOX)
	full := container.NewBorder(topBar, nil, nil, nil, scroll)

	return full
}
