package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Check struct {
	Name string
	Err  error
}

func Verify(contractsDir string) []Check {
	checks := []Check{{Name: "json-parse", Err: checkJSONParse(contractsDir)}}
	schema, err := loadSchema(contractsDir)
	if err == nil {
		err = checkFixtures(contractsDir, schema)
	}
	checks = append(checks, Check{Name: "schema-fixtures", Err: err})
	checks = append(checks, Check{Name: "boundary", Err: checkBoundary(contractsDir)})
	checks = append(checks, Check{Name: "golden-vectors", Err: checkGoldenVectors(contractsDir)})
	checks = append(checks, Check{Name: "signature", Err: checkSignature(contractsDir)})
	checks = append(checks, Check{Name: "jwks", Err: checkJWKS(contractsDir)})
	return checks
}

func checkJSONParse(contractsDir string) error {
	return filepath.WalkDir(contractsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !json.Valid(contents) {
			return fmt.Errorf("%s: invalid JSON", path)
		}
		return nil
	})
}

func checkBoundary(contractsDir string) error {
	valid, err := loadJSON(filepath.Join(contractsDir, "fixtures", "valid", "offline-artifact.json"))
	if err != nil {
		return err
	}
	if err := verifyBoundary(valid); err != nil {
		return err
	}
	semantic, err := loadJSON(filepath.Join(contractsDir, "fixtures", "semantic-invalid", "lifetime-too-long.json"))
	if err != nil {
		return err
	}
	if err := verifyBoundary(semantic); err == nil {
		return fmt.Errorf("semantic-invalid fixture was accepted")
	}
	return nil
}

func main() {
	contracts := flag.String("contracts", "", "path to the contracts/offline-admission/v1.0 directory")
	flag.Parse()
	if *contracts == "" {
		fmt.Fprintln(os.Stderr, "usage: ossverify -contracts <dir>")
		os.Exit(2)
	}
	failed := false
	for _, check := range Verify(*contracts) {
		if check.Err != nil {
			fmt.Printf("FAIL %s: %v\n", check.Name, check.Err)
			failed = true
			continue
		}
		fmt.Printf("PASS %s\n", check.Name)
	}
	if failed {
		os.Exit(1)
	}
}
