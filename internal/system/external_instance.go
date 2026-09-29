// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package system

import _ "embed"

//go:embed recipe/external-instance.yaml
var externalInstanceRecipeYaml []byte

//go:embed recipe/external-instance-from-file.yaml
var externalInstanceFromFile []byte

//go:embed recipe/external-instance-from-ssh_config.yaml
var externalInstanceFromSshConfig []byte

func init() {
	RegisterSystemRecipe("external-instance", externalInstanceRecipeYaml)
	RegisterSystemRecipe("external-instance-from-file", externalInstanceFromFile)
	RegisterSystemRecipe("external-instance-from-ssh_config", externalInstanceFromSshConfig)
}
