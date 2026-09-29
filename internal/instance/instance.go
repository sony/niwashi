// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/types"
)

type Instance struct {
	Status     string                `json:"status,omitempty" yaml:"status,omitempty"`
	Connection *transport.Connection `json:"connection,omitempty" yaml:"connection,omitempty"`
	Addresses  []string              `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	NodeRef    string                `json:"nodeRef,omitempty" yaml:"nodeRef,omitempty"`
	SystemInfo *platform.SystemInfo  `json:"system,omitempty" yaml:"system,omitempty"`
}

func (pi *Instance) SetDefaults() {
	if pi.Connection != nil {
		pi.Connection.SetDefaults()
	}
}

func (pi *Instance) Validate() error {
	if pi == nil {
		return fmt.Errorf("instance cannot be null")
	}

	var err error

	if pi.NodeRef != "" {
		if e := types.NodeIdPattern.MatchString(pi.NodeRef); !e {
			err = errors.Join(err, fmt.Errorf("invalid node reference: %q", pi.NodeRef))
		}
	}

	if pi.Connection != nil {
		if e := pi.Connection.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid connection: %w", e))
		}
	}

	if pi.SystemInfo != nil {
		e := pi.SystemInfo.Validate()
		if e != nil {
			err = errors.Join(err, fmt.Errorf("invalid system info: %w", e))
		}
	}

	return err
}

func (pi *Instance) Clone() *Instance {
	if pi == nil {
		return nil
	}
	clone := &Instance{}
	clone.Status = strings.Clone(pi.Status)
	clone.Connection = pi.Connection.Clone()
	clone.Addresses = append([]string{}, pi.Addresses...)
	clone.NodeRef = strings.Clone(pi.NodeRef)
	clone.SystemInfo = pi.SystemInfo.Clone()
	return clone
}

func (pi *Instance) Merge(other *Instance) (*Instance, error) {
	if pi == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return pi.Clone(), nil
	}

	clone := &Instance{}
	var err error
	clone.Status = merge.String(pi.Status, other.Status)
	clone.Connection, err = pi.Connection.Merge(other.Connection)
	if err != nil {
		return nil, err
	}
	clone.Addresses = merge.StringArray(pi.Addresses, other.Addresses)
	clone.NodeRef = merge.String(pi.NodeRef, other.NodeRef)
	clone.SystemInfo, err = pi.SystemInfo.Merge(other.SystemInfo)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (p *Instance) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "status":
		return patch.AddSimpleValue(path[1:], value, func(v string) {
			p.Status = v
		})
	case "connection":
		return patch.AddValue(
			path[1:],
			value,
			func() (*transport.Connection, bool) {
				return p.Connection, p.Connection != nil
			},
			func(v *transport.Connection) {
				p.Connection = v
			})
	case "addresses":
		return patch.AddListValue(
			path[1:],
			value,
			func(v []string) {
				p.Addresses = v
			},
			func(v []string) {
				p.Addresses = append(p.Addresses, v...)
			},
		)
	case "nodeRef":
		return patch.AddSimpleValue(path[1:], value, func(v string) {
			p.NodeRef = v
		})
	case "system":
		v, err := patch.ConvertValue[*platform.SystemInfo](value)
		if err != nil {
			return err
		}
		p.SystemInfo = v
		return nil
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (i *Instance) Remove(path []string) error {

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "status":
		return patch.RemoveSimpleValue(path[1:], func() {
			i.Status = ""
		})
	case "connection":
		return patch.RemoveComposite(
			path[1:],
			func() (*transport.Connection, bool) {
				return i.Connection, i.Connection != nil
			},
			func() {
				i.Connection = nil
			})
	case "addresses":
		return patch.RemoveListValue(
			path[1:],
			func(index int) error {
				if index == -1 {
					i.Addresses = i.Addresses[:0]
				} else {
					i.Addresses = append(i.Addresses[:index], i.Addresses[index+1:]...)
				}
				return nil
			},
			func() {
				i.Addresses = []string{}
			},
		)
	case "nodeRef":
		return patch.RemoveSimpleValue(path[1:], func() {
			i.NodeRef = ""
		})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (i *Instance) Update(o patch.Patchable) error {
	other := o.(*Instance)

	// deny NodeRef update
	if other.NodeRef != "" {
		return fmt.Errorf("nodeRef cannot be updated")
	}

	i.Status = other.Status
	if !cmp.Equal(i.Addresses, other.Addresses) {
		logger.Warn("Addresses are different, updating addresses")
		i.Addresses = other.Addresses
	}
	if !cmp.Equal(i.Connection, other.Connection) {
		logger.Warn("Connections are different, updating connection")
		i.Connection = other.Connection
	}
	i.SystemInfo = other.SystemInfo

	return nil
}
