// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sony/niwashi/internal/types"
)

var regexInstanceRef = regexp.MustCompile(
	`^` +
		types.GeneratorIdPatternString +
		`:` +
		types.InstanceIdPatternString +
		`$`)

func ValidateInstanceRef(s string) error {
	if !regexInstanceRef.MatchString(s) {
		return fmt.Errorf("invalid instance reference: %s", s)
	}
	return nil
}

func ParseInstanceRef(s string) (string, string) {
	parts := strings.SplitN(s, ":", 2)
	return parts[0], parts[1]
}

func MakeInstanceRef(genName, instName string) string {
	return genName + ":" + instName
}
