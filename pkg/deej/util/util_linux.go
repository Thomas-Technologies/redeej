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

	fmt.Println(lines)
	for _, line := range lines {
		splitLine := strings.Split(line, ":")

		fmt.Println(splitLine[0])
		if strings.TrimSpace(splitLine[0]) == "pid" {
			return getApplicationName_PulseAudio(strings.TrimSpace(splitLine[1]))
		}
	}
	fmt.Println("Error: Hyprctl didn't return a window with a class")

	return "", nil
}

func getApplicationName_PulseAudio(pid string) (string, error) {
	cmd := exec.Command("pactl", "list", "sink-inputs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	list := strings.TrimSpace(string(output))

	sections := strings.Split(list, "\n\n")
	for _, section := range sections {
		if strings.Contains(section, fmt.Sprintf("application.process.id = \"%s\"", pid)) {
			lines := strings.Split(section, "\n")
			for _, line := range lines {
				// TODO: the process binary can be multiple application, like for games its everything wine
				// Figure out a way to get the individual applications
				if strings.Contains(line, "application.process.binary") {
					return strings.ReplaceAll(strings.TrimSpace(strings.Split(line, "=")[1]), "\"", ""), nil
				}
			}
		}
	}
	return "", fmt.Errorf("Window not in audio sink list")
}
