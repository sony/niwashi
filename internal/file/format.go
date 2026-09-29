// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

type EncDec struct {
	encode func(f io.Writer, v any) error
	decode func(f io.Reader, v any) error
}

var encDec = map[string]EncDec{
	".json": {
		encode: encodeJSON,
		decode: decodeJSON,
	},
	".jsonc": {
		encode: encodeJSON,
		decode: decodeJSON,
	},
	".yaml": {
		encode: encodeYAML,
		decode: decodeYAML,
	},
	".yml": {
		encode: encodeYAML,
		decode: decodeYAML,
	},
}

func encodeJSON(file io.Writer, v any) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func decodeJSON(file io.Reader, v any) error {
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func encodeYAML(file io.Writer, v any) error {
	encoder := yaml.NewEncoder(file)
	return encoder.Encode(v)
}

func decodeYAML(file io.Reader, v any) error {
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	return decoder.Decode(v)
}
