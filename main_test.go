package gz

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This function applies to ALL tests in the application.
// It will run the test and then clean the database.
func TestMain(m *testing.M) {
	code := m.Run()
	packageTearDown()
	log.Println("Cleaned database tables after all tests")
	os.Exit(code)
}

// Clean up our mess
func packageTearDown() {
	cleanDBTables()
}

func cleanDBTables() {
}

// TestInitTestModeIgnoresBinaryName checks that test mode comes from the
// binary being built by go test, not from its name.
func TestInitTestModeIgnoresBinaryName(t *testing.T) {
	origArg0, origServer := os.Args[0], gServer
	t.Cleanup(func() { os.Args[0], gServer = origArg0, origServer })
	os.Args[0] = "/srv/fuel-server"
	t.Setenv("TEST_RSA256_PUBLIC_KEY", "test-key")

	s, err := Init("auth0-key", "", nil)
	require.NoError(t, err)
	assert.True(t, s.IsTest)
	assert.True(t, strings.HasSuffix(s.DbConfig.Name, "_test"), s.DbConfig.Name)
	assert.Equal(t, "test-key", s.Auth0RsaPublicKey())
}
