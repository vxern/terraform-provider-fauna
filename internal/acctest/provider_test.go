package acctest_test

import (
	"testing"

	"github.com/linguition/terraform-provider-fauna/internal/acctest"
)

func TestProvider(t *testing.T) {
	if err := acctest.TestAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
