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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/font.ttf")
}

// Get swap file location from the system (/proc/swaps)
// Sample output:
// Filename				Type		Size	Used	Priority
// /home/swapfile			file		8388604	0	-2
func getSwapFileLocation() (string, error) {
	file, err := os.Open("/proc/swaps")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// skip the first line (header)
	scanner.Scan()

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[0] != "Filename" {
			location := fields[0]
			// If swapfile is a partition then return no swapfile found
			if strings.HasPrefix(location, "/dev/") {
				return "", fmt.Errorf("没有找到交换文件")
			}
			return location, nil
		}
	}

	if doesFileExist(DefaultSwapFileLocation) {
		return DefaultSwapFileLocation, nil
	}

	return "", fmt.Errorf("没有找到交换文件")
}

// Get the current swap and swappiness values
func getSwappinessValue() (int, error) {
	cmd, err := exec.Command("sysctl", "vm.swappiness").Output()
	if err != nil {
		return 100, fmt.Errorf("获取当前交换度时出错")
	}
	output := strings.Fields(string(cmd))
	CryoUtils.InfoLog.Println("找到交换度", output[2])
	swappiness, _ := strconv.Atoi(output[2])

	return swappiness, nil
}

// Get current swap file size, in bytes.
func getSwapFileSize() (int64, error) {
	location, err := getSwapFileLocation()
	if err != nil {
		return DefaultSwapSizeBytes, fmt.Errorf("获取交换文件位置时出错: %v", err)
	}

	CryoUtils.SwapFileLocation = location

	info, err := os.Stat(CryoUtils.SwapFileLocation)
	if err != nil {
		// Don't crash the program, just report the default size
		return DefaultSwapSizeBytes, fmt.Errorf("获取当前交换文件大小时出错")
	}
	CryoUtils.InfoLog.Println("发现一个交换文件，其大小为", info.Size())
	return info.Size(), nil
}

// Get the available space for a swap file and return a slice of strings
func getAvailableSwapSizes() ([]string, error) {
	// Get the free space in /home
	currentSwapSize, _ := getSwapFileSize()
	availableSpace, err := getFreeSpace("/home")
	if err != nil {
		return nil, fmt.Errorf("在 /home 中获取可用空间时出错")
	}

	// Loop through the range of available sizes and create a list of viable options for the current Deck.
	// This will always leave 1 as an available option, just in case.
	validSizes := []string{"1 - Default"}
	for _, size := range AvailableSwapSizes {
		intSize, _ := strconv.Atoi(size)
		byteSize := intSize * GigabyteMultiplier
		if int64(byteSize+SpaceOverhead) < (availableSpace + currentSwapSize) {
			if byteSize == int(currentSwapSize) {
				currentSizeString := fmt.Sprintf("%s -当前大小", size)
				validSizes = append(validSizes, currentSizeString)
			} else {
				validSizes = append(validSizes, size)
			}
		}
	}

	CryoUtils.InfoLog.Println("可用的交换大小:", validSizes)
	return validSizes, nil
}

// Disable swapping completely
func disableSwap() error {
	CryoUtils.InfoLog.Println("暂时禁用交换...")
	_, err := exec.Command("sudo", "swapoff", "-a").Output()
	if err != nil {
		return fmt.Errorf("禁用交换时出错")
	}
	return err
}

// Resize the swap file to the provided size, in GB.
func resizeSwapFile(size int) error {
	locationArg := fmt.Sprintf("of=%s", CryoUtils.SwapFileLocation)
	countArg := fmt.Sprintf("count=%d", size)

	CryoUtils.InfoLog.Println("将交换大小调整为", size, "GB...")
	// Use dd to write zeroes, reevaluate using Go directly in the future
	_, err := exec.Command("sudo", "dd", "if=/dev/zero", locationArg, "bs=1G", countArg, "status=progress").Output()
	if err != nil {
		return fmt.Errorf("调整大小时出错 %s", CryoUtils.SwapFileLocation)
	}
	return nil
}

// Set swap permissions to a valid value.
func setSwapPermissions() error {
	CryoUtils.InfoLog.Println("设置权限", CryoUtils.SwapFileLocation, "to 0600...")
	_, err := exec.Command("sudo", "chmod", "600", CryoUtils.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("设置权限时出错 %s", CryoUtils.SwapFileLocation)
	}
	return nil
}

// Enable swapping on the newly resized file.
func initNewSwapFile() error {
	CryoUtils.InfoLog.Println("启用交换", CryoUtils.SwapFileLocation, "...")
	_, err := exec.Command("sudo", "mkswap", CryoUtils.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("创建交换时出错 %s", CryoUtils.SwapFileLocation)
	}
	_, err = exec.Command("sudo", "swapon", CryoUtils.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("启用交换时出错 %s", CryoUtils.SwapFileLocation)
	}
	return nil
}

// ChangeSwappiness Set swappiness to the provided integer.
func ChangeSwappiness(value string) error {
	CryoUtils.InfoLog.Println("设置交换性...")
	// Remove old swappiness file while we're at it
	_ = removeFile(OldSwappinessUnitFile)
	err := setUnitValue("swappiness", value)
	if err != nil {
		return err
	}

	if value == DefaultSwappiness {
		CryoUtils.InfoLog.Println("删除交换性单元以恢复默认行为...")
		err = removeUnitFile("swappiness")
		if err != nil {
			return err
		}
	} else {
		err = writeUnitFile("swappiness", value)
		if err != nil {
			return err
		}
		return nil
	}

	// Return no error if everything went as planned
	return nil
}
