// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import (
	"encoding/json"
	"fmt"

	"github.com/sony/niwashi/internal/patch"
	"gopkg.in/yaml.v3"
)

type Connection struct {
	Data map[string]Transport
}

func (c *Connection) GetTransport(t string) (Transport, error) {
	if c == nil || c.Data == nil {
		return nil, fmt.Errorf("no transports defined in connection")
	}

	if t == "" {
		if len(c.Data) == 1 {
			for _, transport := range c.Data {
				return transport, nil
			}
		} else {
			return nil, fmt.Errorf("multiple transports defined in connection, but no type specified")
		}
	}

	transport, exist := c.Data[t]
	if !exist {
		return nil, fmt.Errorf("unsupported connection type: %s", t)
	}
	return transport, nil
}

func (c *Connection) UnmarshalYAML(node *yaml.Node) error {

	c.Data = make(map[string]Transport)
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		t, err := NewTransport(keyNode.Value)
		if err != nil {
			return err
		}
		if err := valueNode.Decode(t); err != nil {
			return fmt.Errorf("failed to decode transport %s: %w", keyNode.Value, err)
		}
		t.SetDefaults()
		c.Data[keyNode.Value] = t
	}

	return nil
}

func (c *Connection) MarshalYAML() (any, error) {
	return c.Data, nil
}

func (c *Connection) UnmarshalJSON(data []byte) error {

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Data = make(map[string]Transport)
	for k, v := range raw {
		t, err := NewTransport(k)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(v, t); err != nil {
			return fmt.Errorf("failed to decode transport %s: %w", k, err)
		}
		t.SetDefaults()
		c.Data[k] = t
	}
	return nil
}

func (c *Connection) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

func (a *Connection) SetDefaults() {
	for _, v := range a.Data {
		v.SetDefaults()
	}
}

func (a *Connection) Validate() error {
	for k, v := range a.Data {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validation failed for transport %s: %w", k, err)
		}
	}
	return nil
}

func (a *Connection) Clone() *Connection {
	if a == nil {
		return nil
	}
	clone := &Connection{
		Data: make(map[string]Transport),
	}
	for k, v := range a.Data {
		clone.Data[k] = v.Clone()
	}
	return clone
}

func (a *Connection) Merge(other *Connection) (*Connection, error) {
	if other == nil {
		return a.Clone(), nil
	}
	if a == nil {
		return other.Clone(), nil
	}

	clone := &Connection{}
	var err error
	for k, v := range a.Data {
		if ov, exist := other.Data[k]; exist {
			clone.Data[k], err = v.Merge(ov)
			if err != nil {
				return nil, fmt.Errorf("failed to merge transport %s: %w", k, err)
			}
		} else {
			clone.Data[k] = v.Clone()
		}
	}

	return clone, nil
}

func (c *Connection) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	t := c.Data[path[0]]
	if t == nil {
		return fmt.Errorf("transport %s not found in connection", path[0])
	}

	return t.Add(path[1:], value)
}

func (c *Connection) Remove(path []string) error {
	return fmt.Errorf("not implemented")
}

func (*Connection) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Connection")
}
