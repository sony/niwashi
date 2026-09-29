// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

import (
	"fmt"
	"maps"
	"strings"
)

type Alias struct {
	Entries map[string][]*AliasEntry
}

type AliasEntry struct {
	Id      string
	Version string
	Attrs   map[string]string
}

func NewAlias() *Alias {
	return &Alias{
		Entries: make(map[string][]*AliasEntry),
	}
}

func (a *Alias) Add(k, id, v string, at map[string]string) {
	entry := a.Entries[k]
	if entry == nil {
		entry = []*AliasEntry{}
	}

	entry = append(entry, &AliasEntry{
		Id:      id,
		Version: v,
		Attrs:   maps.Clone(at),
	})
	a.Entries[k] = entry
}

func ParseAlias(n string) (string, map[string]string, error) {

	var names []string
	attrs := make(map[string]string)

	for _, v := range strings.Split(n, ".") {
		if len(v) == 0 {
			return "", nil, fmt.Errorf("empty segment found alias=%q", n)
		}

		if strings.Contains(v, "=") {
			if len(names) == 0 {
				return "", nil, fmt.Errorf("attribute specified before identifier alias=%q", n)
			}

			a := strings.Split(v, "=")
			if len(a[0]) == 0 {
				return "", nil, fmt.Errorf("empty attribute key alias=%q", n)
			}

			attrs[a[0]] = v[len(a[0])+1:]
		} else {
			if len(attrs) > 0 {
				return "", nil, fmt.Errorf("identifier found after attributes alias=%q", n)
			}

			names = append(names, v)
		}
	}

	return strings.Join(names, "."), attrs, nil
}

func (a *Alias) Find(k string) (string, string, error) {

	// Parse alias key
	name, attrs, err := ParseAlias(k)
	if err != nil {
		return "", "", err
	}

	// Look for the name in the alias entries
	var exist bool
	var entries []*AliasEntry
	if entries, exist = a.Entries[name]; !exist {
		return "", "", NewNotFoundError(fmt.Sprintf("no recipe provides %q", k))
	}

	candidates := []*AliasEntry{}
	for _, e := range entries {

		// Complete match
		if maps.Equal(e.Attrs, attrs) {
			return e.Id, e.Version, nil
		}

		// Partial match(or no attrs specified)
		match := true
		for key, value := range attrs {
			if ev, ok := e.Attrs[key]; !ok || ev != value {
				match = false
				break
			}
		}
		if match {
			candidates = append(candidates, e)
		}
	}

	if len(candidates) == 1 {
		return candidates[0].Id, candidates[0].Version, nil
	} else if len(candidates) > 1 {
		return "", "", NewNotFoundError(fmt.Sprintf("ambiguous alias: multiple candidates match the given attributes key=%q", k))
	}

	return "", "", NewNotFoundError(fmt.Sprintf("no matching alias found for key=%q", k))
}
