// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

type NotFoundError struct {
	msg string
}

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		msg: msg,
	}
}

func (e *NotFoundError) Error() string {
	return "not found: " + e.msg
}

type ConflictError struct {
	msg string
}

func NewConflictError(msg string) *ConflictError {
	return &ConflictError{
		msg: msg,
	}
}

func (e *ConflictError) Error() string {
	return "conflict: " + e.msg
}
