// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types

import "regexp"

var EnvVarNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
var TemplateVarNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
var LabelKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
var ParamsNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

var (
	CommonIdPatternString    = "[a-zA-Z0-9][a-zA-Z0-9_-]*"
	NodeIdPatternString      = CommonIdPatternString
	ClusterIdPatternString   = CommonIdPatternString
	GeneratorIdPatternString = CommonIdPatternString
	InstanceIdPatternString  = CommonIdPatternString
	commonIdPattern          = regexp.MustCompile("^" + CommonIdPatternString + "$")
	NodeIdPattern            = commonIdPattern
	ClusterIdPattern         = commonIdPattern
	GeneratorIdPattern       = commonIdPattern
	InstanceIdPattern        = commonIdPattern
)
