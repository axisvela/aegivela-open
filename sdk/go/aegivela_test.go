package aegivela_test

import (
	"errors"
	"testing"

	aegivela "github.com/axisrobo/aegivela-open/sdk/go"
)

func TestRootExposesEvaluateConfigurationAndErrors(t *testing.T) {
	var option aegivela.Option = aegivela.WithHeaders(map[string]string{"X-Aegivela-PEP": "pep"})
	if aegivela.NewClient("http://127.0.0.1", nil, option) == nil {
		t.Fatal("nil client")
	}
	var result aegivela.EvaluateResult
	if result.Outcome != "" {
		t.Fatal("unexpected zero result")
	}
	err := &aegivela.APIError{StatusCode: 502, Err: aegivela.ErrMalformed}
	if !errors.Is(err, aegivela.ErrMalformed) {
		t.Fatal("classification not preserved")
	}
}
