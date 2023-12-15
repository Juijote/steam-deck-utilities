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
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/font.ttf")
}

func syncGameDataWindow() {
	var selectionContainer *fyne.Container
	// Create a new window
	w := CryoUtils.App.NewWindow("同步游戏数据")

	driveList, err := getListOfAttachedDrives()
	if err != nil {
		presentErrorInUI(err, w)
	}

	if len(driveList) > 1 {
		// Place a prompt near the top of the window
		prompt := canvas.NewText("请选择你想要同步数据的设备。", nil)
		prompt.TextSize, prompt.TextStyle = 18, fyne.TextStyle{Bold: true}

		// Make the list widgets with initial contents, excluding what the other has pre-selected
		// This is simply to make it more "one click" for users.
		leftList := widget.NewSelect(removeElementFromStringSlice(driveList[1], driveList), func(s string) {})
		rightList := widget.NewSelect(removeElementFromStringSlice(driveList[0], driveList), func(s string) {})

		// Pre-define each direction with the default values
		leftSelected := removeElementFromStringSlice(driveList[1], driveList)[0]
		leftList.Selected = leftSelected
		rightSelected := removeElementFromStringSlice(driveList[0], driveList)[0]
		rightList.Selected = rightSelected
		// Define the OnChanged functions after, so both are aware of the other's existence.
		leftList.OnChanged = func(s string) {
			leftSelected = s
			// Remove the selected option from the other list to prevent both sides being the same.
			rightList.Options = removeElementFromStringSlice(s, driveList)
		}
		rightList.OnChanged = func(s string) {
			rightSelected = s
			// Remove the selected option from the other list to prevent both sides being the same.
			leftList.Options = removeElementFromStringSlice(s, driveList)
		}

		cancelButton := widget.NewButton("取消", func() {
			w.Close()
		})
		submitButton := widget.NewButton("提交", func() {
			selectionContainer.Hide()
			w.CenterOnScreen()
			populateGameDataWindow(w, leftSelected, rightSelected)
		})
		buttonBar := container.NewHSplit(cancelButton, submitButton)
		selectionContainer = container.NewVBox(prompt, leftList, rightList, buttonBar)
	} else {
		// Place a prompt near the top of the window
		prompt := canvas.NewText("连接的硬盘空间不足，无法同步数据", Red)
		prompt.TextSize, prompt.TextStyle = 18, fyne.TextStyle{Bold: true}

		cancelButton := widget.NewButton("取消", func() {
			w.Close()
		})

		selectionContainer = container.NewVBox(prompt, cancelButton)
	}

	w.SetContent(selectionContainer)
	w.CenterOnScreen()
	w.RequestFocus()
	w.Show()
}

func populateGameDataWindow(w fyne.Window, left string, right string) {
	var leftCard, rightCard *widget.Card
	var syncDataButton *widget.Button
	var data DataToMove

	p := widget.NewProgressBarInfinite()
	d := dialog.NewCustom("寻找要移动的数据...", "解除", p, w)
	d.Show()

	// Get a list of data to move
	err := data.getDataToMove(left, right)
	if err != nil {
		CryoUtils.ErrorLog.Println(err)
		d.Hide()
		presentErrorInUI(err, w)
	}

	// Get the storage totals necessary for each side
	data.getSpaceNeeded(left, right)

	leftSpaceAvailable, err := getFreeSpace(left)
	if err != nil {
		presentErrorInUI(err, w)
	}
	rightSpaceAvailable, err := getFreeSpace(right)
	if err != nil {
		presentErrorInUI(err, w)
	}

	// Place a prompt near the top of the window
	prompt := canvas.NewText("请确认可以移动此数据:", nil)
	prompt.TextSize, prompt.TextStyle = 18, fyne.TextStyle{Bold: true}

	// User-presentable strings
	leftDataStr := fmt.Sprintf("数据要移动到 %s", right)
	leftSizeStr := fmt.Sprintf("总大小: %.2fGB", float64(data.leftSize)/float64(GigabyteMultiplier))
	rightDataStr := fmt.Sprintf("数据要移动到 %s", left)
	rightSizeStr := fmt.Sprintf("总大小: %.2fGB", float64(data.rightSize)/float64(GigabyteMultiplier))

	// Deal with lack of space on left
	if leftSpaceAvailable < data.rightSize {
		leftCard = widget.NewCard(leftDataStr, "",
			canvas.NewText("错误：目标硬盘上没有足够的可用空间。", Red))
		// Provide a button to close the window
		syncDataButton = widget.NewButton("Close", func() {
			w.Close()
		})
	}

	// Deal with lack of space on right
	if rightSpaceAvailable < data.leftSize {
		rightCard = widget.NewCard(rightDataStr, "",
			canvas.NewText("错误：目标硬盘上没有足够的可用空间。", Red))
		// Provide a button to close the window
		syncDataButton = widget.NewButton("Close", func() {
			w.Close()
		})
	}

	leftList, rightList, err := getDataToMoveUI(data)
	// Deal with error
	if err != nil {
		// Create an error in each card if directories can't be listed
		leftCard = widget.NewCard(leftDataStr, "",
			canvas.NewText("错误：无法获取要移动的目录列表。", Red))
		rightCard = widget.NewCard(rightDataStr, "",
			canvas.NewText("错误：无法获取要移动的目录列表。", Red))
		// Provide a button to close the window
		syncDataButton = widget.NewButton("关闭", func() {
			w.Close()
		})
	}

	// If there's anything to move left
	if len(data.right) != 0 {
		leftCard = widget.NewCard(rightDataStr, rightSizeStr, leftList)
	} else {
		leftCard = widget.NewCard(rightDataStr, "",
			canvas.NewText("没有其它！所有数据都往这个方向同步。", Green))
	}

	// If there's anything to move right
	if len(data.left) != 0 {
		rightCard = widget.NewCard(leftDataStr, leftSizeStr, rightList)
	} else {
		rightCard = widget.NewCard(leftDataStr, "",
			canvas.NewText("没有其它！所有数据都往这个方向同步。", Green))
	}

	// Create button if something can sync
	if len(data.right) != 0 || len(data.left) != 0 {
		syncDataButton = widget.NewButton("确认", func() {
			// Do the actual sync
			CryoUtils.InfoLog.Println("同步数据已确认")
			progress := widget.NewProgressBar()
			CryoUtils.MoveDataProgressBar = progress
			progress.Resize(fyne.NewSize(500, 50))
			tempVBox := container.NewVBox(canvas.NewText("正在同步，请稍候...", nil), progress)
			widget.ShowModalPopUp(tempVBox, w.Canvas())
			err = moveGameData(data, left, right)
			if err != nil {
				presentErrorInUI(err, w)
			} else {
				_, err := data.confirmDirectoryStatus(left, right)
				if err != nil {
					presentErrorInUI(err, w)
				} else {
					CryoUtils.InfoLog.Println("所有数据移动正常，复制成功！")
					dialog.ShowInformation(
						"成功!",
						"数据移动完成，所有游戏数据同步到相应设备。",
						CryoUtils.MainWindow,
					)
					w.Close()
				}
			}
		})
	} else {
		// Otherwise, provide a button to close the window
		syncDataButton = widget.NewButton("关闭", func() {
			w.Close()
		})
	}
	cancelButton := widget.NewButton("取消", func() {
		w.Close()
	})

	d.Hide()

	// Format the window
	syncMain := container.NewGridWithColumns(1, leftCard, rightCard)
	syncButtonBorder := container.NewGridWithColumns(2, cancelButton, syncDataButton)
	syncLayout := container.NewBorder(nil, syncButtonBorder, nil, nil, syncMain)
	w.SetContent(syncLayout)
	w.Resize(fyne.NewSize(300, 450))
	w.CenterOnScreen()
	w.RequestFocus()
	w.Show()
}

func cleanupDataWindow() {
	var cleanupCard *widget.Card
	var cleanupButton, cancelButton *widget.Button

	// Create a new window
	w := CryoUtils.App.NewWindow("清理游戏数据")

	var removeList []string
	cleanupList, err := createGameDataList()
	if err != nil {
		presentErrorInUI(err, CryoUtils.MainWindow)
	}
	cleanupList.OnChanged = func(s []string) {
		var tempList []string

		for i := range s {
			// Get only the game ID for the selected games
			tempList = append(tempList, strings.Split(s[i], " ")[0])
		}
		removeList = tempList
	}
	cleanupScroll := container.NewVScroll(cleanupList)

	// Create an error in each card if directories can't be listed
	cleanupCard = widget.NewCard("清理过时的游​​戏数据",
		"选择要删除的游戏的前缀和着色器缓存。",
		cleanupScroll)
	cancelButton = widget.NewButton("取消", func() {
		w.Close()
	})
	cleanupButton = widget.NewButton("删除所选", func() {
		dialog.ShowConfirm("你确定吗?", "确定要删除这些文件吗?\n\n"+
			"请务必先备份所有非 Steam 云的游戏存档\n"+
			"使用此工具删除它们，任何选定的内容都会丢失。",
			func(b bool) {
				if b {
					possibleLocations, err := getListOfDataAllDataLocations()
					if err != nil {
						CryoUtils.ErrorLog.Println(err)
						presentErrorInUI(err, CryoUtils.MainWindow)
					}

					removeGameData(removeList, possibleLocations)

					dialog.ShowInformation(
						"Success!",
						"Process completed!",
						CryoUtils.MainWindow,
					)
					w.Close()
				} else {
					w.Close()
				}
			}, w)
	})

	cleanAllUninstalled := widget.NewButton("删除所有已卸载的", func() {
		dialog.ShowConfirm("你确定吗?", "确定要删除这些文件吗?\n\n"+
			"请务必先备份所有非 Steam 云的游戏存档\n"+
			"使用此工具删除它们，任何选定的内容都会丢失。",
			func(b bool) {
				if !b {
					w.Close()
				}

				locations, err := getListOfDataAllDataLocations()
				if err != nil {
					CryoUtils.ErrorLog.Println(err)
					presentErrorInUI(err, CryoUtils.MainWindow)
				}

				removeGameData(getUninstalledGamesData(), locations)

				dialog.ShowInformation(
					"成功!",
					"操作完成!",
					CryoUtils.MainWindow,
				)
				w.Close()

			}, w)

	})

	// Format the window
	cleanupMain := container.NewGridWithColumns(1, cleanupCard)
	cleanupButtonsGrid := container.NewGridWithColumns(2, cancelButton, cleanupButton)
	extraButtonGrid := container.NewGridWithColumns(1, cleanAllUninstalled)
	footerButtons := container.NewGridWithColumns(1, cleanupButtonsGrid, extraButtonGrid)
	cleanupLayout := container.NewBorder(nil, footerButtons, nil, nil, cleanupMain)
	w.SetContent(cleanupLayout)
	w.Resize(fyne.NewSize(300, 450))
	w.CenterOnScreen()
	w.RequestFocus()
	w.Show()
}

func swapSizeWindow() {
	// Create a new window
	w := CryoUtils.App.NewWindow("更改交换文件大小")

	// Place a prompt near the top of the window
	prompt := canvas.NewText("请选择新的交换文件大小（以 GB 为单位）:", nil)
	prompt.TextSize, prompt.TextStyle = 18, fyne.TextStyle{Bold: true}

	// Determine maximum available space for a swap file and construct a list of available sizes based on it
	availableSwapSizes, err := getAvailableSwapSizes()
	if err != nil {
		presentErrorInUI(err, w)
	}

	// Give the user a choice in swap file sizes
	var chosenSize int
	choice := widget.NewRadioGroup(availableSwapSizes, func(value string) {
		// Only grab the number at the beginning of the string, allows for suffixes.
		chosenSize, err = strconv.Atoi(strings.Split(value, " ")[0])
		if err != nil {
			presentErrorInUI(err, w)
		}
	})

	// Provide a button to submit the choice
	swapResizeButton := widget.NewButton("调整交换文件大小", func() {
		progress := widget.NewProgressBarInfinite()
		d := dialog.NewCustom("正在调整交换文件大小，请耐心等待..."+
			"(这最多可能需要 30 分钟)", "退出", progress,
			w,
		)
		d.Show()
		err = changeSwapSizeGUI(chosenSize)
		if err != nil {
			d.Hide()
			presentErrorInUI(err, w)
		} else {
			d.Hide()
			dialog.ShowInformation(
				"成功!",
				"操作完成！你可以验证文件是否已调整大小\n"+
					"在终端中运行 “ls -lash /home/swapfile” 或 “swapon -s”",
				CryoUtils.MainWindow,
			)
			CryoUtils.refreshSwapContent()
			w.Close()
		}
	})

	// Make a progress bar and hide it
	progress := widget.NewProgressBar()
	CryoUtils.SwapResizeProgressBar = progress
	progress.Hide()

	// Format the window
	swapVBox := container.NewVBox(prompt, choice, swapResizeButton)
	w.SetContent(swapVBox)
	w.Resize(fyne.NewSize(400, 300))
	w.CenterOnScreen()
	w.RequestFocus()
	w.Show()
}

// Note: Having a separate function for this is hacky, but necessary for progress bar functionality
func changeSwapSizeGUI(size int) error {
	// Disable swap temporarily
	renewSudoAuth()
	CryoUtils.InfoLog.Println("暂时禁用交换...")
	err := disableSwap()
	if err != nil {
		return err
	}
	// Resize the file
	renewSudoAuth()
	err = resizeSwapFile(size)
	if err != nil {
		return err
	}
	// Set permissions on file
	renewSudoAuth()
	err = setSwapPermissions()
	if err != nil {
		return err
	}
	// Initialize new swap file
	renewSudoAuth()
	err = initNewSwapFile()
	if err != nil {
		return err
	}
	return nil
}

func swappinessWindow() {
	// Create a new window
	w := CryoUtils.App.NewWindow("改变交换性")

	// Place a prompt near the top of the window
	prompt := canvas.NewText("请选择新的交换值。", nil)
	prompt.TextSize, prompt.TextStyle = 18, fyne.TextStyle{Bold: true}

	// Give the user a choice in swap file sizes
	var chosenSwappiness string
	choice := widget.NewRadioGroup(AvailableSwappinessOptions, func(value string) {
		chosenSwappiness = strings.Fields(value)[0]
	})

	// Provide a button to submit the choice
	swappinessChangeButton := widget.NewButton("改变交换性", func() {
		renewSudoAuth()
		err := ChangeSwappiness(chosenSwappiness)
		if err != nil {
			presentErrorInUI(err, w)
		} else {
			dialog.ShowInformation(
				"成功!",
				"交换性变更完成!",
				CryoUtils.MainWindow,
			)
			CryoUtils.refreshSwappinessContent()
			w.Close()
		}
	})

	// Format the window
	swapVBox := container.NewVBox(prompt, choice, swappinessChangeButton)
	w.SetContent(swapVBox)
	w.CenterOnScreen()
	w.RequestFocus()
	w.Show()
}
