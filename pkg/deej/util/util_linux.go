package util

import (
	"fmt"
	"os/exec"
	"strings"
)

// Wayland doesn't have a universal way of getting window names
// Logic has to be implemented per desktop environment
// I use hyprland currently so thats the one I implemented
func getCurrentWindowProcessNames() ([]string, error) {
	winName, err := getCurrentWindow_Hyprland()
	if err != nil {
		return []string{}, err
	}
	return []string{winName}, nil
}

func getCurrentWindow_Hyprland() (string, error) {
	cmd := exec.Command("hyprctl", "activewindow")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	windowInfo := strings.TrimSpace(string(output))
	lines := strings.Split(windowInfo, "\n")

	for _, line := range lines {
		splitLine := strings.Split(line, ":")

		fmt.Println(splitLine[0])
		if strings.TrimSpace(splitLine[0]) == "class" {
			return strings.TrimSpace(splitLine[1]), nil
		}
	}
	fmt.Println("Error: Hyprctl didn't return a window with a class")

	return "", nil
}
