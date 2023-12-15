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
	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/NotoSansSC.ttf")
}

func getHugePagesStatus() bool {
	status, err := getUnitStatus("hugepages")
	if err != nil {
		CryoUtils.ErrorLog.Println("无法获取当前的大页面值")
		return false
	}
	if status == RecommendedHugePages {
		return true
	}
	return false
}

func getCompactionProactivenessStatus() bool {
	status, err := getUnitStatus("compaction_proactiveness")
	if err != nil {
		CryoUtils.ErrorLog.Println("无法获取当前的压缩主动性")
		return false
	}
	if status == RecommendedCompactionProactiveness {
		return true
	}
	return false
}

func getPageLockUnfairnessStatus() bool {
	status, err := getUnitStatus("page_lock_unfairness")
	if err != nil {
		CryoUtils.ErrorLog.Println("无法获取当前的页面锁定不公平值")
		return false
	}
	if status == RecommendedPageLockUnfairness {
		return true
	}
	return false
}

func getShMemStatus() bool {
	status, err := getUnitStatus("shmem_enabled")
	if err != nil {
		CryoUtils.ErrorLog.Println("无法获取当前已启用的内存管理 shmem_enabled")
		return false
	}
	if status == RecommendedShMem {
		return true
	}
	return false
}

func getDefragStatus() bool {
	status, err := getUnitStatus("defrag")
	if err != nil {
		CryoUtils.ErrorLog.Println("无法获取当前碎片整理")
		return false
	}
	if status == RecommendedHugePageDefrag {
		return true
	}
	return false
}

// ToggleHugePages Simple one-function toggle for the button to use
func ToggleHugePages() error {
	if getHugePagesStatus() {
		err := RevertHugePages()
		if err != nil {
			return err
		}
	} else {
		err := SetHugePages()
		if err != nil {
			return err
		}
	}
	return nil
}

// ToggleShMem Simple one-function toggle for the button to use
func ToggleShMem() error {
	if getShMemStatus() {
		err := RevertShMem()
		if err != nil {
			return err
		}
	} else {
		err := SetShMem()
		if err != nil {
			return err
		}
	}
	return nil
}

// ToggleCompactionProactiveness Simple one-function toggle for the button to use
func ToggleCompactionProactiveness() error {
	if getCompactionProactivenessStatus() {
		err := RevertCompactionProactiveness()
		if err != nil {
			return err
		}
	} else {
		err := SetCompactionProactiveness()
		if err != nil {
			return err
		}
	}
	return nil
}

// ToggleDefrag Simple one-function toggle for the button to use
func ToggleDefrag() error {
	if getDefragStatus() {
		err := RevertDefrag()
		if err != nil {
			return err
		}
	} else {
		err := SetDefrag()
		if err != nil {
			return err
		}
	}
	return nil
}

// TogglePageLockUnfairness Simple one-function toggle for the button to use
func TogglePageLockUnfairness() error {
	if getPageLockUnfairnessStatus() {
		err := RevertPageLockUnfairness()
		if err != nil {
			return err
		}
	} else {
		err := SetPageLockUnfairness()
		if err != nil {
			return err
		}
	}
	return nil
}

func SetHugePages() error {
	CryoUtils.InfoLog.Println("启用大页面...")
	// Remove a file accidentally included in a beta for testing
	_ = removeFile(NHPTestingFile)
	err := setUnitValue("hugepages", RecommendedHugePages)
	if err != nil {
		return err
	}
	err = writeUnitFile("hugepages", RecommendedHugePages)
	if err != nil {
		return err
	}
	return nil
}

func RevertHugePages() error {
	CryoUtils.InfoLog.Println("禁用大页面...")
	err := setUnitValue("hugepages", DefaultHugePages)
	if err != nil {
		return err
	}
	err = removeUnitFile("hugepages")
	if err != nil {
		return err
	}
	return nil
}

func SetCompactionProactiveness() error {
	CryoUtils.InfoLog.Println("设置压缩主动性...")
	err := setUnitValue("compaction_proactiveness", RecommendedCompactionProactiveness)
	if err != nil {
		return err
	}
	err = writeUnitFile("compaction_proactiveness", RecommendedCompactionProactiveness)
	if err != nil {
		return err
	}
	return nil
}

func RevertCompactionProactiveness() error {
	CryoUtils.InfoLog.Println("禁用压缩主动性...")
	err := setUnitValue("compaction_proactiveness", DefaultCompactionProactiveness)
	if err != nil {
		return err
	}
	err = removeUnitFile("compaction_proactiveness")
	if err != nil {
		return err
	}
	return nil
}

func SetPageLockUnfairness() error {
	CryoUtils.InfoLog.Println("启用页面锁定不公平...")
	err := setUnitValue("page_lock_unfairness", RecommendedPageLockUnfairness)
	if err != nil {
		return err
	}
	err = writeUnitFile("page_lock_unfairness", RecommendedPageLockUnfairness)
	if err != nil {
		return err
	}
	return nil
}

func RevertPageLockUnfairness() error {
	CryoUtils.InfoLog.Println("禁用页面锁定不公平...")
	err := setUnitValue("page_lock_unfairness", DefaultPageLockUnfairness)
	if err != nil {
		return err
	}
	err = removeUnitFile("page_lock_unfairness")
	if err != nil {
		return err
	}
	return nil
}

func SetShMem() error {
	CryoUtils.InfoLog.Println("启用内存管理 shmem_enabled")
	err := setUnitValue("shmem_enabled", RecommendedShMem)
	if err != nil {
		return err
	}
	err = writeUnitFile("shmem_enabled", RecommendedShMem)
	if err != nil {
		return err
	}
	return nil
}

func RevertShMem() error {
	CryoUtils.InfoLog.Println("禁用内存管理 shmem_enabled")
	err := setUnitValue("shmem_enabled", DefaultShMem)
	if err != nil {
		return err
	}
	err = removeUnitFile("shmem_enabled")
	if err != nil {
		return err
	}
	return nil
}

func SetDefrag() error {
	CryoUtils.InfoLog.Println("启用内存管理 shmem_enabled")
	err := setUnitValue("defrag", RecommendedHugePageDefrag)
	if err != nil {
		return err
	}
	err = writeUnitFile("defrag", RecommendedHugePageDefrag)
	if err != nil {
		return err
	}
	return nil
}

func RevertDefrag() error {
	CryoUtils.InfoLog.Println("禁用内存管理 shmem_enabled")
	err := setUnitValue("defrag", DefaultHugePageDefrag)
	if err != nil {
		return err
	}
	err = removeUnitFile("defrag")
	if err != nil {
		return err
	}
	return nil
}
