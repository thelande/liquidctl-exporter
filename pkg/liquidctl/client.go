// Package liquidctl provides functionality to interact with the liquidctl CLI tool.
package liquidctl

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

// baseAttributes contains common attributes for liquidctl devices.
type baseAttributes struct {
	Bus         string `json:"bus"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// LiquidCtlDevice represents a device detected by liquidctl.
type LiquidCtlDevice struct {
	baseAttributes
	VendorID      int    `json:"vendor_id"`
	ProductID     int    `json:"product_id"`
	ReleaseNumber int    `json:"release_number"`
	SerialNumber  string `json:"serial_number"`
	Driver        string `json:"driver"`
	Experimental  bool   `json:"experimental"`
}

// LiquidCtlStatusField represents a single status field returned by liquidctl.
type LiquidCtlStatusField struct {
	Key   string `json:"key"`
	Value *Value `json:"value"`
	Unit  string `json:"unit"`
}

// LiquidCtlStatus represents the status of a liquidctl device.
type LiquidCtlStatus struct {
	baseAttributes
	Status []LiquidCtlStatusField `json:"status"`
}

// LiquidCtl is the main client for interacting with the liquidctl CLI tool.
type LiquidCtl struct {
	Logger *slog.Logger
}

// NewLiquidCtl creates a new LiquidCtl client instance with the default logger.
func NewLiquidCtl() *LiquidCtl {
	return &LiquidCtl{
		Logger: slog.Default(),
	}
}

// runCommand executes the liquidctl CLI with the given arguments and returns the JSON output.
func (l *LiquidCtl) runCommand(args []string) ([]byte, error) {
	args = append(args, "--json")
	cmd := exec.Command("liquidctl", args...)
	output, err := cmd.Output()
	if err != nil {
		l.Logger.Error("failed to run command", "args", args)
		return nil, err
	}
	return output, nil
}

// ListDevices returns a list of all devices detected by liquidctl.
func (l *LiquidCtl) ListDevices() ([]LiquidCtlDevice, error) {
	output, err := l.runCommand([]string{"list"})
	if err != nil {
		return nil, err
	}

	var devices []LiquidCtlDevice
	json.Unmarshal(output, &devices)
	return devices, nil
}

// GetDevice retrieves the status of a specific device identified by bus and address.
func (l *LiquidCtl) GetDevice(bus, address string) (*LiquidCtlStatus, error) {
	args := []string{"--bus", bus, "--address", address, "status"}
	output, err := l.runCommand(args)
	if err != nil {
		return nil, err
	}

	var status []LiquidCtlStatus
	json.Unmarshal(output, &status)
	if len(status) != 1 {
		return nil, fmt.Errorf("incorrect number of devices returned: %d", len(status))
	}

	return &status[0], nil
}

// GetAttributes initializes a device and retrieves its attributes.
func (l *LiquidCtl) GetAttributes(bus, address string) (*LiquidCtlStatus, error) {
	args := []string{"--bus", bus, "--address", address, "initialize"}
	output, err := l.runCommand(args)
	if err != nil {
		return nil, err
	}

	var status []LiquidCtlStatus
	json.Unmarshal(output, &status)
	if len(status) != 1 {
		return nil, fmt.Errorf("incorrect number of devices returned: %d", len(status))
	}

	return &status[0], nil
}
