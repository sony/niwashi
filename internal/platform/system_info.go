// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package platform

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/types"
)

type SystemInfo struct {
	Os   string     `json:"os" yaml:"os"`
	Arch string     `json:"arch" yaml:"arch"`
	Data types.Dict `json:"data" yaml:"data"`
}

func (si *SystemInfo) Clone() *SystemInfo {
	if si == nil {
		return nil
	}
	clone := &SystemInfo{}
	clone.Os = si.Os
	clone.Arch = si.Arch
	clone.Data = si.Data.Clone()
	for k, v := range si.Data {
		clone.Data[k] = v
	}
	return clone
}

func (si *SystemInfo) Validate() error {
	err := si.Data.ValidateAsLabelKey()
	for k := range si.Data {
		if !types.LabelKeyPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid system info data key: %q", k))
		}
	}
	return err
}

func (si *SystemInfo) Merge(other *SystemInfo) (*SystemInfo, error) {
	if si == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return si.Clone(), nil
	}

	var err error

	clone := &SystemInfo{}
	clone.Os = merge.String(si.Os, other.Os)
	clone.Arch = merge.String(si.Arch, other.Arch)
	clone.Data, err = types.Merge(si.Data, other.Data)
	if err != nil {
		return nil, err
	}

	return clone, nil
}
