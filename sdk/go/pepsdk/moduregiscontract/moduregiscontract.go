// Package moduregiscontract ships the published MODUREGIS adapter contract
// (JSON Schema and conformance fixtures) as embedded, versioned Go module
// data so downstream repositories validate the identical artifact set.
//
// Stability: the exported API and the embedded contract follow the
// contracts/moduregis/v1alpha1 compatibility policy; breaking changes
// require a new contract version, never an in-place edit.
package moduregiscontract

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Contract resource identifiers registered with the schema compiler.
const (
	SchemaID      = "https://aegivela.io/contracts/moduregis/v1alpha1/moduregis-adapter.schema.json"
	CommonTypesID = "https://aegivela.io/contracts/common/v1alpha1/types.schema.json"
)

//go:embed files
var files embed.FS

var (
	validFixtureNames = []string{
		"adapter-activate.json",
		"capability-invoke.json",
		"capability-publish.json",
		"capability-read.json",
	}
	invalidFixtureNames = []string{
		"activate-without-version.json",
		"invoke-without-execution.json",
		"publish-without-approval.json",
		"tenant-in-request.json",
	}
)

// ValidFixtureNames lists the positive conformance fixture file names.
func ValidFixtureNames() []string { return append([]string(nil), validFixtureNames...) }

// InvalidFixtureNames lists the negative conformance fixture file names.
func InvalidFixtureNames() []string { return append([]string(nil), invalidFixtureNames...) }

// SchemaJSON returns the MODUREGIS adapter JSON Schema document.
func SchemaJSON() []byte { return mustRead("files/moduregis-adapter.schema.json") }

// CommonTypesJSON returns the shared common-types JSON Schema document.
func CommonTypesJSON() []byte { return mustRead("files/types.schema.json") }

// ValidFixture returns one positive fixture document by file name.
func ValidFixture(name string) ([]byte, error) { return readFixture("valid", name) }

// InvalidFixture returns one negative fixture document by file name.
func InvalidFixture(name string) ([]byte, error) { return readFixture("invalid", name) }

// CompileSchema compiles the adapter schema with the common-types resource
// registered so its external reference resolves offline.
func CompileSchema() (*jsonschema.Schema, error) {
	schema, err := jsonschema.UnmarshalJSON(bytes.NewReader(SchemaJSON()))
	if err != nil {
		return nil, fmt.Errorf("moduregis adapter schema: %w", err)
	}
	common, err := jsonschema.UnmarshalJSON(bytes.NewReader(CommonTypesJSON()))
	if err != nil {
		return nil, fmt.Errorf("common types schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(SchemaID, schema); err != nil {
		return nil, err
	}
	if err := compiler.AddResource(CommonTypesID, common); err != nil {
		return nil, err
	}
	return compiler.Compile(SchemaID)
}

func readFixture(kind, name string) ([]byte, error) {
	if !validName(name, kind) {
		return nil, fmt.Errorf("unknown %s fixture %q", kind, name)
	}
	return files.ReadFile("files/fixtures/" + kind + "/" + name)
}

func validName(name, kind string) bool {
	names := validFixtureNames
	if kind == "invalid" {
		names = invalidFixtureNames
	}
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

func mustRead(path string) []byte {
	contents, err := files.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return contents
}
