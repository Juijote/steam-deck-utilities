package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"os"
)

func init() {
	//设置中文字体
	os.Setenv("FYNE_FONT", "/home/deck/.cryo_utilities/font.ttf")
}

// Get the current VRAM
func getVRAMValue() (int, error) {
	cmd, err := exec.Command("glxinfo", "-B").Output()

	// Extract video memory
	re := regexp.MustCompile(`显存: [0-9]+`)
	match := re.FindStringSubmatch(string(cmd))

	if err != nil || match == nil {
		return 100, fmt.Errorf("获取当前显存时出错")
	}

	output := strings.Split(match[0], " ")[2]
	CryoUtils.InfoLog.Println("找到显存", output)
	vram, _ := strconv.Atoi(output)

	return vram, nil
}
