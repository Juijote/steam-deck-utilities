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
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"path/filepath"
	"sort"
	"strconv"
	
	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/font.ttf")
}

type GameStatus struct {
	GameName    string
	IsInstalled bool
}

// Show an error message over the main window.
func presentErrorInUI(err error, win fyne.Window) {
	CryoUtils.ErrorLog.Println(err)
	dialog.ShowError(err, win)
}

// Create a CheckGroup of game data to allow for selection.
func createGameDataList() (*widget.CheckGroup, error) {
	cleanupList := widget.NewCheckGroup([]string{}, func(strings []string) {})
	cleanupList.Enable()
	cleanupList.Refresh()

	localGames, err := getLocalGameList()
	if err != nil {
		return nil, err
	}

	var sortedMap []int

	for i := range localGames {
		// Add each key to the sortedMap slice, so we can sort it afterwards.
		sortedMap = append(sortedMap, i)
	}
	// Sort the slice
	sort.Ints(sortedMap)

	// For each entry in the completed list, add an entry to the check group to return
	for key := range sortedMap {
		// Skips non-game prefixes
		if sortedMap[key] == 0 || sortedMap[key] >= SteamGameMaxInteger {
			continue
		}

		var optionStr string
		var gameStr string

		// If the game name is known, use that, otherwise ???.
		if localGames[sortedMap[key]].GameName != "" {
			gameStr = localGames[sortedMap[key]].GameName
		} else {
			gameStr = "???"
		}

		if localGames[sortedMap[key]].IsInstalled {
			optionStr = fmt.Sprintf("%d - %s - 已安装", sortedMap[key], gameStr)
		} else {
			optionStr = fmt.Sprintf("%d - %s - 未安装", sortedMap[key], gameStr)
		}
		cleanupList.Append(optionStr)
	}

	return cleanupList, nil
}

func getLocalGameList() (map[int]GameStatus, error) {
	// Use the cached API Response if already present
	if CryoUtils.SteamAPIResponse == nil {
		// Make a map of all games stored in the steam API
		steamResponse, err := querySteamAPI()
		if err != nil {
			CryoUtils.ErrorLog.Println(err)
		}
		CryoUtils.SteamAPIResponse = generateGameMap(steamResponse)
	}

	// Get a list of games that Steam classifies as installed
	libraries, err := findDataFolders()
	if err != nil {
		return nil, err
	}
	// Get a list of all the folders in each location
	var storage []string
	for i := range libraries {
		// Get the paths to crawl
		var compat, shader string
		if libraries[i].Path == SteamDataRoot {
			compat = SteamCompatRoot
			shader = SteamShaderRoot
		} else {
			compat = filepath.Join(libraries[i].Path, ExternalCompatRoot)
			shader = filepath.Join(libraries[i].Path, ExternalShaderRoot)
		}

		// Crawl the compat path and append the folders
		// Append if no error, to prevent crashing for users that haven't synced data first.
		dir, _ := getDirectoryList(compat, true)
		if err == nil {
			storage = append(storage, dir...)
		}

		// Crawl the shader path and append the folders
		// Append if no error, to prevent crashing for users that haven't synced data first.
		dir, _ = getDirectoryList(shader, true)
		if err == nil {
			storage = append(storage, dir...)
		}
	}

	// Store all games present on the filesystem, in all compat and shader directories
	localGames := make(map[int]GameStatus)

	for i := range storage {
		intGame, _ := strconv.Atoi(storage[i])
		localGames[intGame] = GameStatus{
			GameName:    CryoUtils.SteamAPIResponse[intGame],
			IsInstalled: false,
		}
	}

	// Loop through each library location's installed games
	for i := range libraries {
		for j := range libraries[i].InstalledGames {
			// If the game has an entry in the list of localGames, then update the isInstalled flag to true for that game.
			val, keyExists := localGames[libraries[i].InstalledGames[j]]
			if keyExists {
				val.IsInstalled = true
				localGames[libraries[i].InstalledGames[j]] = val
			}
		}
	}
	return localGames, nil
}

// Get data to move values as canvas elements.
func getDataToMoveUI(data DataToMove) (*widget.List, *widget.List, error) {
	var leftList, rightList *widget.List

	// Use the cached API Response if already present
	if CryoUtils.SteamAPIResponse == nil {
		// Make a map of all games stored in the steam API
		steamResponse, err := querySteamAPI()
		if err != nil {
			CryoUtils.ErrorLog.Println(err)
		}
		CryoUtils.SteamAPIResponse = generateGameMap(steamResponse)
	}

	// Get lists of data to move
	leftList = widget.NewList(
		func() int {
			return len(data.right)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("左侧")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			gameInt, _ := strconv.Atoi(data.right[i])
			gameName := CryoUtils.SteamAPIResponse[gameInt]

			o.(*widget.Label).SetText(data.right[i] + " - " + gameName)
		})

	rightList = widget.NewList(
		func() int {
			return len(data.left)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("右侧")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			gameInt, _ := strconv.Atoi(data.left[i])
			gameName := CryoUtils.SteamAPIResponse[gameInt]

			o.(*widget.Label).SetText(data.left[i] + " - " + gameName)
		})

	return leftList, rightList, nil
}

func (app *Config) refreshSwapContent() {
	app.InfoLog.Println("正在刷新交换数据...")
	swap, err := getSwapFileSize()
	if err != nil {
		CryoUtils.ErrorLog.Println(err)
		swapStr := "当前交换大小: 未知"
		app.SwapText.Text = swapStr
		app.SwapText.Color = Gray
	} else {
		humanSwapSize := swap / int64(GigabyteMultiplier)
		swapStr := fmt.Sprintf("当前交换大小: %dGB", humanSwapSize)
		app.SwapText.Text = swapStr
		if swap >= RecommendedSwapSizeBytes {
			app.SwapText.Color = Green
		} else {
			app.SwapText.Color = Red
		}
	}

	app.SwapText.Refresh()
}

func (app *Config) refreshSwappinessContent() {
	app.InfoLog.Println("正在刷新交换性数据...")
	swappiness, err := getSwappinessValue()
	if err != nil {
		CryoUtils.ErrorLog.Println(err)
		swappinessStr := "当前交换性: 未知"
		app.SwappinessText.Text = swappinessStr
		app.SwappinessText.Color = Gray
	} else {
		swappinessStr := fmt.Sprintf("当前交换性: %d", swappiness)
		app.SwappinessText.Text = swappinessStr
		if strconv.Itoa(swappiness) == RecommendedSwappiness {
			app.SwappinessText.Color = Green
		} else {
			app.SwappinessText.Color = Red
		}
	}

	app.SwappinessText.Refresh()
}

func (app *Config) refreshHugePagesContent() {
	app.InfoLog.Println("正在刷新大页面数据...")
	if getHugePagesStatus() {
		app.HugePagesButton.Text = "禁用大页面"
		app.HugePagesText.Color = Green
	} else {
		app.HugePagesButton.Text = "启用大页面"
		app.HugePagesText.Color = Red
	}
	app.HugePagesButton.Refresh()
	app.HugePagesText.Refresh()
}

func (app *Config) refreshShMemContent() {
	app.InfoLog.Println("正在刷新共享内存 shmem 数据...")
	if getShMemStatus() {
		app.ShMemButton.Text = "在 THP 中禁用共享内存"
		app.ShMemText.Color = Green
	} else {
		app.ShMemButton.Text = "在 THP 中启用共享内存"
		app.ShMemText.Color = Red
	}
	app.ShMemButton.Refresh()
	app.ShMemText.Refresh()

}

func (app *Config) refreshCompactionProactivenessContent() {
	app.InfoLog.Println("正在刷新压缩主动性数据...")
	if getCompactionProactivenessStatus() {
		app.CompactionProactivenessButton.Text = "恢复压缩主动性"
		app.CompactionProactivenessText.Color = Green
	} else {
		app.CompactionProactivenessButton.Text = "设置压缩主动性"
		app.CompactionProactivenessText.Color = Red
	}
	app.CompactionProactivenessButton.Refresh()
	app.CompactionProactivenessText.Refresh()
}

func (app *Config) refreshDefragContent() {
	app.InfoLog.Println("正在刷新碎片整理数据...")
	if getDefragStatus() {
		app.DefragButton.Text = "启用大页面碎片整理"
		app.DefragText.Color = Green
	} else {
		app.DefragButton.Text = "禁用大页面碎片整理"
		app.DefragText.Color = Red
	}
	app.DefragButton.Refresh()
	app.DefragText.Refresh()
}

func (app *Config) refreshPageLockUnfairnessContent() {
	app.InfoLog.Println("正在刷新页面锁定不公平数据...")
	if getPageLockUnfairnessStatus() {
		app.PageLockUnfairnessButton.Text = "恢复页面锁定不公平"
		app.PageLockUnfairnessText.Color = Green
	} else {
		app.PageLockUnfairnessButton.Text = "设置页面锁定不公平"
		app.PageLockUnfairnessText.Color = Red
	}
	app.PageLockUnfairnessButton.Refresh()
	app.PageLockUnfairnessText.Refresh()
}

func (app *Config) refreshVRAMContent() {
	app.InfoLog.Println("正在刷新显存数据...")
	vram, err := getVRAMValue()
	if err != nil || vram == 0 {
		CryoUtils.ErrorLog.Println(err)
		vramStr := fmt.Sprintf("当前显存: 未知")
		app.VRAMText.Text = vramStr
		app.VRAMText.Color = Gray
	} else {
		humanVRAMSize := getHumanVRAMSize(vram)
		vramStr := fmt.Sprintf("当前显存: %s", humanVRAMSize)
		app.VRAMText.Text = vramStr
		if vram == RecommendedVRAM {
			app.VRAMText.Color = Green
		} else {
			app.VRAMText.Color = Red
		}
	}

	app.VRAMText.Refresh()
}

func (app *Config) refreshAllContent() {
	app.refreshSwapContent()
	app.refreshSwappinessContent()
	app.refreshHugePagesContent()
	app.refreshCompactionProactivenessContent()
	app.refreshShMemContent()
	app.refreshDefragContent()
	app.refreshPageLockUnfairnessContent()
	app.refreshVRAMContent()
}
