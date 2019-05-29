// Copyright (c) 2018-2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package crypt

import (
	"fmt"
	"os"
)

// Device is
type Device struct {
	MaxDevices int
}

type errCryptDeviceUnavailable struct {
	message string
}

func newCryptDevUnailable(msg string) *errCryptDeviceUnavailable {
	return &errCryptDeviceUnavailable{
		message: msg,
	}
}

func (e *errCryptDeviceUnavailable) Error() string {
	return e.message
}

// GetCryptDevice returns the next available device in /dev/mapper for encryption/decryption
func (crypt *Device) GetCryptDevice() (string, error) {
	// Return the next available crypt device
	for i := 0; i < crypt.MaxDevices; i++ {
		retStr := fmt.Sprintf("singularity_crypt_%d", i)
		device := fmt.Sprintf("/dev/mapper/%s", retStr)
		if _, err := os.Stat(device); os.IsNotExist(err) {

			return retStr, nil
		}
	}
	return "", newCryptDevUnailable("Crypt Device not available")
}
