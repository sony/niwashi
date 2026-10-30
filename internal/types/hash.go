// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func MakeHash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}
