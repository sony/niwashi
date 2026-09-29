// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

import "fmt"

type PathConstraint interface {
	GetDefaultPrefix() string
	Validate(path string) error
}

type NilPathConstraint struct{}

func (c *NilPathConstraint) GetDefaultPrefix() string {
	return "/"
}

func (c *NilPathConstraint) Validate(path string) error {
	return nil
}

type ErrorPathConstraint struct{}

func (c *ErrorPathConstraint) GetDefaultPrefix() string {
	return "/"
}

func (c *ErrorPathConstraint) Validate(path string) error {
	return fmt.Errorf("invalid patch path: %s", path)
}

type SimplePathConstraint struct {
	DefaultPrefix   string
	AllowedPrefixes []string
}

func (c *SimplePathConstraint) GetDefaultPrefix() string {
	return c.DefaultPrefix
}

func (c *SimplePathConstraint) Validate(path string) error {
	var err error
	for _, prefix := range c.AllowedPrefixes {
		err = PathValidate(path, prefix)
		if err == nil {
			return nil
		}
	}
	return err
}

func PathValidate(path, constraint string) error {
	if len(constraint) == 0 || path == constraint ||
		(len(path) > len(constraint) && path[:len(constraint)] == constraint && path[len(constraint)] == '/') {
		return nil
	}
	return fmt.Errorf("invalid patch path: %s", path)
}
