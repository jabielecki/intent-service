package neutron_test

import (
	"testing"

	"github.com/tungstenfabric-preview/intent-service/pkg/testutil/integration"
)

var server *integration.APIServer

func TestMain(m *testing.M) {
	integration.TestMain(m, &server)
}
