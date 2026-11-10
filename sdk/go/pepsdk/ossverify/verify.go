package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type vectorManifest struct {
	ProtectedHeader struct {
		Members         map[string]any `json:"members"`
		CanonicalSha256 string         `json:"canonicalSha256"`
		Base64URL       string         `json:"base64url"`
	} `json:"protectedHeader"`
	ArtifactPayload struct {
		CanonicalSha256 string `json:"canonicalSha256"`
		Base64URL       string `json:"base64url"`
	} `json:"artifactPayload"`
	SigningInput struct {
		Value  string `json:"value"`
		Sha256 string `json:"sha256"`
	} `json:"signingInput"`
	Parent struct {
		CompactToken string `json:"compactToken"`
		Sha256       string `json:"sha256"`
	} `json:"parent"`
	BindingDigest struct {
		CanonicalSha256 string `json:"canonicalSha256"`
	} `json:"bindingDigest"`
	PayloadDigest struct {
		CanonicalSha256 string `json:"canonicalSha256"`
	} `json:"payloadDigest"`
}

func loadSchema(dir string) (*jsonschema.Schema, error) {
	contents, err := os.ReadFile(filepath.Join(dir, "offline-admission.schema.json"))
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(contents, &document); err != nil {
		return nil, err
	}
	schemaID, _ := document["$id"].(string)
	if schemaID == "" {
		return nil, fmt.Errorf("schema has no $id")
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaID, document); err != nil {
		return nil, err
	}
	return compiler.Compile(schemaID)
}

func checkFixtures(dir string, schema *jsonschema.Schema) error {
	valid, err := filepath.Glob(filepath.Join(dir, "fixtures", "valid", "*.json"))
	if err != nil || len(valid) == 0 {
		return fmt.Errorf("list valid fixtures: files=%d err=%v", len(valid), err)
	}
	for _, file := range valid {
		document, err := loadJSON(file)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if err := schema.Validate(document); err != nil {
			return fmt.Errorf("valid fixture %s rejected: %w", filepath.Base(file), err)
		}
	}
	invalid, err := filepath.Glob(filepath.Join(dir, "fixtures", "invalid", "*.json"))
	if err != nil || len(invalid) == 0 {
		return fmt.Errorf("list invalid fixtures: files=%d err=%v", len(invalid), err)
	}
	for _, file := range invalid {
		document, err := loadJSON(file)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if err := schema.Validate(document); err == nil {
			return fmt.Errorf("invalid fixture %s unexpectedly conformed", filepath.Base(file))
		}
	}
	return nil
}

func loadJSON(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(contents, &document); err != nil {
		return nil, err
	}
	return document, nil
}

func verifyBoundary(document map[string]any) error {
	if purpose, _ := document["purpose"].(string); purpose != "offline-admission" {
		return fmt.Errorf("purpose must be offline-admission")
	}
	if delegation, ok := document["offlineDelegation"].(bool); !ok || delegation {
		return fmt.Errorf("offline delegation must be explicitly false")
	}
	device, _ := document["device"].(map[string]any)
	epoch, ok := device["sourceEpoch"].(float64)
	if !ok || epoch < 1 {
		return fmt.Errorf("sourceEpoch must be at least 1")
	}
	validity, _ := document["validity"].(map[string]any)
	notBefore, err := parseTime(validity["notBefore"])
	if err != nil {
		return err
	}
	expiresAt, err := parseTime(validity["expiresAt"])
	if err != nil {
		return err
	}
	if !expiresAt.After(notBefore) {
		return fmt.Errorf("expiry must be after notBefore")
	}
	if expiresAt.Sub(notBefore) > 900*time.Second {
		return fmt.Errorf("lifetime exceeds the approved 900-second maximum")
	}
	if skew, ok := validity["maxClockSkewMs"].(float64); !ok || skew < 0 || skew > 5000 {
		return fmt.Errorf("maxClockSkewMs must be between 0 and 5000")
	}
	revocation, _ := document["revocation"].(map[string]any)
	switch revocation["policy"] {
	case "connectedOnly":
		if _, present := revocation["maxOfflineStalenessMs"]; present {
			return fmt.Errorf("connectedOnly must not declare offline staleness")
		}
	case "boundedDisconnected":
		staleness, ok := revocation["maxOfflineStalenessMs"].(float64)
		if !ok || staleness < 1 {
			return fmt.Errorf("boundedDisconnected requires positive staleness")
		}
	default:
		return fmt.Errorf("unknown revocation policy")
	}
	return nil
}

func parseTime(value any) (time.Time, error) {
	text, ok := value.(string)
	if !ok || text == "" {
		return time.Time{}, fmt.Errorf("missing RFC3339 timestamp")
	}
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid RFC3339 timestamp: %w", err)
	}
	return parsed, nil
}

func readManifest(vectorDir string) (vectorManifest, error) {
	var manifest vectorManifest
	contents, err := os.ReadFile(filepath.Join(vectorDir, "manifest.json"))
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func checkGoldenVectors(dir string) error {
	vectorDir := filepath.Join(dir, "vectors")
	manifest, err := readManifest(vectorDir)
	if err != nil {
		return err
	}
	artifact, err := loadJSONNumber(filepath.Join(dir, "fixtures", "valid", "offline-artifact.json"))
	if err != nil {
		return err
	}

	if err := assertCanonicalFile(vectorDir, "protected-header.canonical.json", manifest.ProtectedHeader.Members); err != nil {
		return err
	}
	headerBytes, err := Marshal(manifest.ProtectedHeader.Members)
	if err != nil {
		return err
	}
	if got := sha256Hex(headerBytes); got != manifest.ProtectedHeader.CanonicalSha256 {
		return fmt.Errorf("protected header sha256 = %s, want %s", got, manifest.ProtectedHeader.CanonicalSha256)
	}
	if got := b64url(headerBytes); got != manifest.ProtectedHeader.Base64URL {
		return fmt.Errorf("protected header base64url mismatch")
	}

	if err := assertCanonicalFile(vectorDir, "artifact-payload.canonical.json", artifact); err != nil {
		return err
	}
	payloadBytes, err := Marshal(artifact)
	if err != nil {
		return err
	}
	if got := sha256Hex(payloadBytes); got != manifest.ArtifactPayload.CanonicalSha256 {
		return fmt.Errorf("artifact payload sha256 = %s, want %s", got, manifest.ArtifactPayload.CanonicalSha256)
	}
	if got := b64url(payloadBytes); got != manifest.ArtifactPayload.Base64URL {
		return fmt.Errorf("artifact payload base64url mismatch")
	}

	signingInput := manifest.ProtectedHeader.Base64URL + "." + manifest.ArtifactPayload.Base64URL
	if signingInput != manifest.SigningInput.Value {
		return fmt.Errorf("signing input does not concatenate the manifest base64url parts")
	}
	if got := sha256Hex([]byte(signingInput)); got != manifest.SigningInput.Sha256 {
		return fmt.Errorf("signing input sha256 = %s, want %s", got, manifest.SigningInput.Sha256)
	}

	if got := sha256Hex([]byte(manifest.Parent.CompactToken)); got != manifest.Parent.Sha256 {
		return fmt.Errorf("parent digest = %s, want %s", got, manifest.Parent.Sha256)
	}
	if got, _ := artifact["parentGrantDigest"].(string); got != manifest.Parent.Sha256 {
		return fmt.Errorf("fixture parentGrantDigest = %s, want %s", got, manifest.Parent.Sha256)
	}

	binding := bindingVector(artifact)
	if err := assertCanonicalFile(vectorDir, "binding.canonical.json", binding); err != nil {
		return err
	}
	bindingBytes, err := Marshal(binding)
	if err != nil {
		return err
	}
	if got := sha256Hex(bindingBytes); got != manifest.BindingDigest.CanonicalSha256 {
		return fmt.Errorf("binding digest = %s, want %s", got, manifest.BindingDigest.CanonicalSha256)
	}

	payloadInput := map[string]any{
		"operationType": "ros2.nav2.navigate_to_pose@1",
		"payload": map[string]any{
			"frameId": "map",
			"goal": map[string]any{
				"xMm":         json.Number("12400"),
				"yMm":         json.Number("-3100"),
				"yawMilliRad": json.Number("1571"),
			},
		},
	}
	if err := assertCanonicalFile(vectorDir, "payload-input.canonical.json", payloadInput); err != nil {
		return err
	}
	payloadInputBytes, err := Marshal(payloadInput)
	if err != nil {
		return err
	}
	if got := sha256Hex(payloadInputBytes); got != manifest.PayloadDigest.CanonicalSha256 {
		return fmt.Errorf("payload digest = %s, want %s", got, manifest.PayloadDigest.CanonicalSha256)
	}
	return nil
}

func assertCanonicalFile(dir, name string, value any) error {
	want, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	if len(want) == 0 || want[len(want)-1] == '\n' {
		return fmt.Errorf("%s must be non-empty exact bytes without a trailing newline", name)
	}
	got, err := Marshal(value)
	if err != nil {
		return err
	}
	if string(got) != string(want) {
		return fmt.Errorf("%s canonical bytes differ: got %s want %s", name, got, want)
	}
	return nil
}

func bindingVector(artifact map[string]any) map[string]any {
	device, _ := artifact["device"].(map[string]any)
	return map[string]any{
		"namespace":                artifact["namespace"],
		"audience":                 artifact["audience"],
		"action":                   artifact["action"],
		"resourceRef":              artifact["resourceRef"],
		"structuredResourceDigest": artifact["structuredResourceDigest"],
		"taskBinding":              artifact["taskBinding"],
		"scope":                    artifact["scope"],
		"device": map[string]any{
			"deviceRef":   device["deviceRef"],
			"deviceKeyId": device["deviceKeyId"],
			"sourceEpoch": device["sourceEpoch"],
		},
		"parentGrantRef":    artifact["parentGrantRef"],
		"parentGrantDigest": artifact["parentGrantDigest"],
		"validity":          artifact["validity"],
	}
}

func loadJSONNumber(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return fmt.Sprintf("%x", digest)
}

func b64url(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

type signatureVector struct {
	PublicKeyJWK struct {
		Kty string `json:"kty"`
		Crv string `json:"crv"`
		X   string `json:"x"`
	} `json:"publicKeyJwk"`
	SignatureBase64URL string `json:"signature"`
	SigningInputSha256 string `json:"signingInputSha256"`
}

func readSignature(vectorDir string) (signatureVector, error) {
	var vector signatureVector
	contents, err := os.ReadFile(filepath.Join(vectorDir, "signature.json"))
	if err != nil {
		return vector, err
	}
	if err := json.Unmarshal(contents, &vector); err != nil {
		return vector, err
	}
	return vector, nil
}

func checkSignature(dir string) error {
	vectorDir := filepath.Join(dir, "vectors")
	manifest, err := readManifest(vectorDir)
	if err != nil {
		return err
	}
	vector, err := readSignature(vectorDir)
	if err != nil {
		return err
	}
	if vector.PublicKeyJWK.Kty != "OKP" || vector.PublicKeyJWK.Crv != "Ed25519" {
		return fmt.Errorf("test public key is not Ed25519: kty=%q crv=%q", vector.PublicKeyJWK.Kty, vector.PublicKeyJWK.Crv)
	}
	if vector.SigningInputSha256 != manifest.SigningInput.Sha256 {
		return fmt.Errorf("signature vector signing-input digest disagrees with the manifest")
	}
	if err := verifyEd25519(vector.PublicKeyJWK.X, manifest.SigningInput.Value, vector.SignatureBase64URL); err != nil {
		return err
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(manifest.ProtectedHeader.Base64URL)
	if err != nil {
		return err
	}
	var header map[string]string
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return err
	}
	if header["alg"] != "EdDSA" || header["typ"] != "aegivela-offline-admission+jws" || header["kid"] == "" {
		return fmt.Errorf("protected header does not use the fixed strings: %+v", header)
	}
	return nil
}

func verifyEd25519(publicKeyX, signingInput, signature string) error {
	publicKey, err := base64.RawURLEncoding.DecodeString(publicKeyX)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length %d", len(publicKey))
	}
	signatureBytes, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}
	if len(signatureBytes) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length %d", len(signatureBytes))
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), []byte(signingInput), signatureBytes) {
		return fmt.Errorf("signature does not verify")
	}
	return nil
}

func checkJWKS(dir string) error {
	vectorDir := filepath.Join(dir, "vectors")
	manifest, err := readManifest(vectorDir)
	if err != nil {
		return err
	}
	vector, err := readSignature(vectorDir)
	if err != nil {
		return err
	}
	contents, err := os.ReadFile(filepath.Join(vectorDir, "jwks.json"))
	if err != nil {
		return err
	}
	var document struct {
		Keys []struct {
			Kty     string `json:"kty"`
			Crv     string `json:"crv"`
			Kid     string `json:"kid"`
			Use     string `json:"use"`
			Alg     string `json:"alg"`
			Purpose string `json:"purpose"`
			X       string `json:"x"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(contents, &document); err != nil {
		return err
	}
	if len(document.Keys) != 1 {
		return fmt.Errorf("JWKS must contain exactly one purpose-separated key, got %d", len(document.Keys))
	}
	key := document.Keys[0]
	if key.Kty != "OKP" || key.Crv != "Ed25519" || key.Kid != "offline-key-1" ||
		key.Use != "sig" || key.Alg != "EdDSA" || key.Purpose != "offline-admission" || key.X == "" {
		return fmt.Errorf("JWKS entry violates the purpose-separated key contract: %+v", key)
	}
	if key.X != vector.PublicKeyJWK.X {
		return fmt.Errorf("JWKS key does not match the signature vector")
	}
	if err := verifyEd25519(key.X, manifest.SigningInput.Value, vector.SignatureBase64URL); err != nil {
		return fmt.Errorf("JWKS key does not verify the signature vector: %w", err)
	}
	return nil
}
