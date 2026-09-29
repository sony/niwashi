// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
)

type Provide struct {
	Name       string            `json:"name" yaml:"name"`
	Attributes map[string]string `json:"attrs" yaml:"attrs"`
}

func (p *Provide) Validate() error {
	var err error

	if !provideNamePattern.MatchString(p.Name) {
		err = errors.Join(err, fmt.Errorf("provide name=%q must match pattern %q", p.Name, provideNamePatternString))
	}

	for k, v := range p.Attributes {
		if !provideAttrKeyPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("provide %q: attribute key=%q must match pattern %q", p.Name, k, provideAttrKeyPatternString))
		}
		if !provideAttrValuePattern.MatchString(v) {
			err = errors.Join(err, fmt.Errorf("provide %q: attribute value=%q must match pattern %q", p.Name, v, provideAttrValuePatternString))
		}
	}

	return err
}
