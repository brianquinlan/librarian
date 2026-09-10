// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dart

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/googleapis/librarian/internal/config"
	"github.com/googleapis/librarian/internal/serviceconfig"
	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/parser"
	"github.com/googleapis/librarian/internal/sources"
)

var errInvalidSpecificationFormat = errors.New("dart generation requires protobuf or discovery specification format")

func toModelConfig(library *config.Library, ch *config.API, srcs *sources.Sources, pc *config.Protoc) (*parser.ModelConfig, error) {
	specFormat := config.SpecProtobuf
	if library.SpecificationFormat != "" {
		specFormat = library.SpecificationFormat
	}
	if specFormat != config.SpecProtobuf && specFormat != config.SpecDiscovery {
		return nil, fmt.Errorf("%w, got %q", errInvalidSpecificationFormat, library.SpecificationFormat)
	}

	src := sources.NewSourceConfig(srcs, library.Roots)

	if library.Dart != nil && library.Dart.IncludeList != nil {
		src.IncludeList = library.Dart.IncludeList
	}
	root := srcs.Googleapis
	if ch.Path == "schema/google/showcase/v1beta1" {
		root = srcs.Showcase
		src.ActiveRoots = append(src.ActiveRoots, "showcase")
	}
	if specFormat == config.SpecDiscovery && !slices.Contains(src.ActiveRoots, "discovery") {
		src.ActiveRoots = append(src.ActiveRoots, "discovery")
	}
	svcConfig, err := serviceconfig.Find(root, ch.Path, config.LanguageDart)
	if err != nil {
		return nil, err
	}

	var specSource string
	switch specFormat {
	case config.SpecDiscovery:
		if svcConfig.Discovery != "" {
			specSource = svcConfig.Discovery
		} else {
			specSource = ch.Path
		}
	default:
		specSource = ch.Path
	}

	title := svcConfig.Title
	var name string
	var skippedIDs []string
	if library.Dart != nil {
		name = library.Dart.NameOverride
		if library.Dart.TitleOverride != "" {
			title = library.Dart.TitleOverride
		}
		skippedIDs = library.Dart.SkippedIds
	}

	modelConfig := &parser.ModelConfig{
		Language:            config.LanguageDart,
		SpecificationFormat: specFormat,
		ServiceConfig:       svcConfig.ServiceConfig,
		SpecificationSource: specSource,
		Source:              src,
		Protoc:              pc,
		Codec:               buildCodec(library),
		Override: api.ModelOverride{
			Name:        name,
			Description: svcConfig.Description,
			Title:       title,
			SkippedIDs:  skippedIDs,
		},
	}
	return modelConfig, nil
}

func buildCodec(library *config.Library) map[string]string {
	codec := make(map[string]string)
	if library.CopyrightYear != "" {
		codec["copyright-year"] = library.CopyrightYear
	}
	if library.Version != "" {
		codec["version"] = library.Version
	}
	if library.SkipRelease {
		codec["not-for-publication"] = "true"
	}
	if library.Name != "" {
		codec["package-name-override"] = library.Name
	}
	if library.Dart == nil {
		return codec
	}

	dart := library.Dart
	if dart.APIKeysEnvironmentVariables != "" {
		codec["api-keys-environment-variables"] = dart.APIKeysEnvironmentVariables
	}
	if dart.Dependencies != "" {
		codec["dependencies"] = dart.Dependencies
	}
	if dart.DevDependencies != "" {
		codec["dev-dependencies"] = dart.DevDependencies
	}
	if dart.ExtraImports != "" {
		codec["extra-imports"] = dart.ExtraImports
	}
	if dart.IssueTrackerURL != "" {
		codec["issue-tracker-url"] = dart.IssueTrackerURL
	}
	if dart.LibraryPathOverride != "" {
		codec["library-path-override"] = dart.LibraryPathOverride
	}
	if dart.PartFile != "" {
		codec["part-file"] = dart.PartFile
	}
	if dart.ReadmeAfterTitleText != "" {
		codec["readme-after-title-text"] = dart.ReadmeAfterTitleText
	}
	if dart.ReadmeQuickstartText != "" {
		codec["readme-quickstart-text"] = dart.ReadmeQuickstartText
	}
	if dart.RepositoryURL != "" {
		codec["repository-url"] = dart.RepositoryURL
	}
	if dart.SupportsSSE {
		codec["supports-sse"] = "true"
	}
	maps.Copy(codec, dart.Packages)
	maps.Copy(codec, dart.Prefixes)
	maps.Copy(codec, dart.Protos)
	return codec
}
