package util

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"runtime"
	"slices"
	"syscall"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// EnsureDirExists creates the given directory path if it doesn't already exist
func EnsureDirExists(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("ensure directory exists (%s): %w", path, err)
	}

	return nil
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Linux returns true if we're running on Linux
func Linux() bool {
	return runtime.GOOS == "linux"
}

func Windows() bool {
	return runtime.GOOS == "windows"
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS
func SetupCloseHandler() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	return c
}

// GetCurrentWindowProcessNames returns the process names (including extension, if applicable)
// of the current foreground window. This includes child processes belonging to the window.
// This is currently only implemented for Windows
func GetCurrentWindowProcessNames() ([]string, error) {
	return getCurrentWindowProcessNames()
}

// OpenExternal spawns a detached window with the provided command and argument
func OpenExternal(logger *zap.SugaredLogger, cmd string, arg string) error {

	// use cmd for windows, bash for linux
	execCommandArgs := []string{"cmd.exe", "/C", "start", "/b", cmd, arg}
	if Linux() {
		execCommandArgs = []string{"/bin/bash", "-c", fmt.Sprintf("%s %s", cmd, arg)}
	}

	command := exec.Command(execCommandArgs[0], execCommandArgs[1:]...)

	if err := command.Run(); err != nil {
		logger.Warnw("Failed to spawn detached process",
			"command", cmd,
			"argument", arg,
			"error", err)

		return fmt.Errorf("spawn detached proc: %w", err)
	}

	return nil
}

// NormalizeScalar "trims" the given float32 to 2 points of precision (e.g. 0.15442 -> 0.15)
// This is used both for windows core audio volume levels and for cleaning up slider level values from serial
func NormalizeScalar(v float32) float32 {
	return float32(math.Floor(float64(v)*100) / 100.0)
}

// SignificantlyDifferent returns true if there's a significant enough volume difference between two given values
func SignificantlyDifferent(old float32, new float32, noiseReductionLevel string) bool {

	const (
		noiseReductionHigh = "high"
		noiseReductionLow  = "low"
	)

	// this threshold is solely responsible for dealing with hardware interference when
	// sliders are producing noisy values. this value should be a median value between two
	// round percent values. for instance, 0.025 means volume can move at 3% increments
	var significantDifferenceThreshold float64

	// choose our noise reduction level based on the config-provided value
	switch noiseReductionLevel {
	case noiseReductionHigh:
		significantDifferenceThreshold = 0.035
		break
	case noiseReductionLow:
		significantDifferenceThreshold = 0.015
		break
	default:
		significantDifferenceThreshold = 0.025
		break
	}

	if math.Abs(float64(old-new)) >= significantDifferenceThreshold {
		return true
	}

	// special behavior is needed around the edges of 0.0 and 1.0 - this makes it snap (just a tiny bit) to them
	if (almostEquals(new, 1.0) && old != 1.0) || (almostEquals(new, 0.0) && old != 0.0) {
		return true
	}

	// values are close enough to not warrant any action
	return false
}

// a helper to make sure volume snaps correctly to 0 and 100, where appropriate
func almostEquals(a float32, b float32) bool {
	return math.Abs(float64(a-b)) < 0.000001
}

// Handles the logic for adding the window to the slider in the config
func AddWindowToSlider(windowTitle string, index int) {
	// Read the YAML file
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal the YAML data into a generic map
	var genericData map[string]interface{}
	err = yaml.Unmarshal(data, &genericData)
	if err != nil {
		fmt.Errorf("error unmarshaling file: %v", err)
	}

	// Extract the slider_mapping section
	var sliderMapping map[int]interface{}
	if sm, exists := genericData["slider_mapping"]; exists {
		// Convert to the expected type
		smBytes, err := yaml.Marshal(sm)
		if err != nil {
			fmt.Errorf("error marshaling slider_mapping: %v", err)
		}
		err = yaml.Unmarshal(smBytes, &sliderMapping)
		if err != nil {
			fmt.Errorf("error unmarshaling slider_mapping: %v", err)
		}
	} else {
		sliderMapping = make(map[int]interface{})
	}

	for index, val := range sliderMapping {
		valArr := makeStringSlice(val)
		if slices.Contains(valArr, windowTitle) {
			valArr = remove(valArr, windowTitle)
		}
		sliderMapping[index] = valArr
	}

	// Check the value type and append windowTitle accordingly
	if val, exists := sliderMapping[index]; exists {
		switch v := val.(type) {
		case string:
			// Convert string to list of strings
			sliderMapping[index] = []string{v, windowTitle}
		case []any:
			// Convert interface{} slice to string slice and append the new value
			newVal := make([]string, len(v))
			for i, elem := range v {
				newVal[i] = fmt.Sprint(elem)
			}
			sliderMapping[index] = append(newVal, windowTitle)
		case []string:
			// If it's already a list of strings, simply append the new value
			sliderMapping[index] = append(v, windowTitle)
		default:
			fmt.Println(fmt.Errorf("unexpected type for slider_mapping: %v", reflect.TypeOf(val)))
		}
	} else {
		sliderMapping[index] = []string{windowTitle}
	}

	// Update the genericData map with the modified slider_mapping
	genericData["slider_mapping"] = sliderMapping

	// Marshal the modified genericData back to YAML
	modifiedData, err := yaml.Marshal(&genericData)
	if err != nil {
		fmt.Println(fmt.Errorf("error marshaling modified data: %v", err))
	}

	// Write the modified YAML back to the file
	err = os.WriteFile("config.yaml", modifiedData, 0644)
	if err != nil {
		fmt.Println(fmt.Errorf("error writing file: %v", err))
	}

	// Print the modified configuration
	fmt.Println("Modified configuration:")
	fmt.Println(string(modifiedData))
}

func makeStringSlice(s any) []string {
	// Convert s to a slice of strings
	vals, ok := s.([]any)
	if !ok {
		return nil
	}

	var newVals []string
	for _, v := range vals {
		newVals = append(newVals, fmt.Sprint(v))
	}
	return newVals
}

func remove(s []string, str string) []string {
	// Convert s to a slice of strings
	var newVals []string
	for _, v := range s {
		if v != str {
			newVals = append(newVals, v)
		}
	}
	return newVals
}
