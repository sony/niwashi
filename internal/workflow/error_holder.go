// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import "errors"

type ErrorHolder interface {
	HasError() bool
	GetError() error
	HoldError(err ...error)
}

type DefaultErrorHolder struct {
	err error
}

func (h *DefaultErrorHolder) HasError() bool {
	return h.err != nil
}

func (h *DefaultErrorHolder) GetError() error {
	return h.err
}

func (h *DefaultErrorHolder) HoldError(err ...error) {
	for _, e := range err {
		if e != nil {
			h.err = errors.Join(h.err, e)
		}
	}
}
