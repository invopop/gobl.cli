package internal

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/flimzy/testy"
)

// These tests depend on the build_test.go for some of the basics.

func TestCorrect(t *testing.T) {
	type tt struct {
		opts *CorrectOptions
		err  string
	}

	tests := testy.NewTable()
	tests.Add("success", func(t *testing.T) interface{} {
		return tt{
			opts: &CorrectOptions{
				ParseOptions: &ParseOptions{
					Data: testFileReader(t, "testdata/success.json"),
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		opts := tt.opts
		got, err := Correct(context.Background(), opts)
		if tt.err == "" {
			assert.Nil(t, err)
		} else {
			assert.EqualError(t, err, tt.err)
		}
		if err != nil {
			return
		}
		replacements := []testy.Replacement{
			{
				Regexp:      regexp.MustCompile(`(?s)"sigs": \[.*\]`),
				Replacement: `"sigs": ["signature data"]`,
			},
			{
				Regexp:      regexp.MustCompile(`"uuid":.?"[^\"]+"`),
				Replacement: `"uuid":"00000000-0000-0000-0000-000000000000"`,
			},
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), got, replacements...); d != nil {
			t.Error(d)
		}
	})
}
