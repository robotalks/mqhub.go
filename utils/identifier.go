package utils

import (
	"net"
	"os"
	"path/filepath"

	"github.com/rs/xid"
)

// MachineIdentifier retrieves unique identifier from machine (MAC address)
// It returns the same identifiers on the same machine
// If empty string is returned, no MAC address is available
func MachineIdentifier() string {
	netifs, _ := net.Interfaces()
	for _, netif := range netifs {
		info, err := os.Stat(filepath.Join("/sys/class/net", netif.Name, "device"))
		if err == nil && (info.Mode()&os.ModeSymlink) != 0 {
			return netif.HardwareAddr.String()
		}
	}
	return ""
}

// Unique ID returns an ID (not UUID) which can be globally unique in a
// distributed environment (based on MAC address)
func UniqueID() string {
	return xid.New().String()
}
