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
	"strconv"
	"strings"

	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/NotoSansSC.ttf")
}

// ChangeSwapSizeCLI Change the swap file size to the specified size in GB
func ChangeSwapSizeCLI(size int, isUI bool) error {
	// Refresh creds if running with UI
	if isUI {
		renewSudoAuth()
	}
	// Disable swap temporarily
	err := disableSwap()
	if err != nil {
		return err
	}

	// Resize the file
	err = resizeSwapFile(size)
	if err != nil {
		return err
	}

	// Refresh creds if running with UI
	// Prevents long-running swap resized from causing issues
	if isUI {
		renewSudoAuth()
	}
	// Set permissions on file
	err = setSwapPermissions()
	if err != nil {
		return err
	}

	// Initialize new swap file
	err = initNewSwapFile()
	if err != nil {
		return err
	}
	return nil
}

func UseRecommendedSettings() error {
	// Change swap
	CryoUtils.InfoLog.Println("开始调整交换文件大小...")
	availableSpace, err := getFreeSpace("/home")
	if err != nil {
		return err
	}
	if availableSpace < RecommendedSwapSizeBytes {
		size := 1
		var availableSizes []string
		availableSizes, err = getAvailableSwapSizes()
		if err != nil {
			return err
		}
		if len(availableSizes) != 1 {
			// Get the last entry in the availableSizes list
			size, err = strconv.Atoi(strings.Fields(availableSizes[len(availableSizes)-1])[0])
			if err != nil {
				return err
			}
			// Never create a swap file larger than 16GB automatically.
			if size > 16 {
				size = 16
			}
		}
		err = ChangeSwapSizeCLI(size, true)
		if err != nil {
			return err
		}
	} else {
		err = ChangeSwapSizeCLI(RecommendedSwapSize, true)
		if err != nil {
			return err
		}
	}
	CryoUtils.InfoLog.Println("调整交换文件大小，改变交换性能...")
	err = ChangeSwappiness(RecommendedSwappiness)
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("交换功能已更改，启用大页面...")
	err = SetHugePages()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("启用大页面，设置主动压缩...")
	err = SetCompactionProactiveness()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("主动压缩已更改，禁用大页面碎片整理...")
	err = SetDefrag()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("禁用大页面碎片整理，设置页面锁非公平性...")
	err = SetPageLockUnfairness()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("页面锁不公平已更改，启用共享内存...")
	err = SetShMem()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("所有设置已配置！")
	return nil
}

func UseStockSettings() error {
	CryoUtils.InfoLog.Println("将交换文件大小调整为 1GB...")
	// Revert swap file size
	err := ChangeSwapSizeCLI(DefaultSwapSize, true)
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("将交换性设置为 100...")
	// Revert swappiness
	err = ChangeSwappiness(DefaultSwappiness)
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("禁用大页面...")
	// Enable HugePages
	err = RevertHugePages()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("恢复主动压缩...")
	err = RevertCompactionProactiveness()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("启用大页面碎片整理...")
	err = RevertDefrag()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("恢复页面锁定的不公正性...")
	err = RevertPageLockUnfairness()
	if err != nil {
		return err
	}

	CryoUtils.InfoLog.Println("禁用大页面中的共享内存...")
	err = RevertShMem()
	if err != nil {
		CryoUtils.InfoLog.Println("所有设置恢复为默认值！")
	}

	return nil
}
