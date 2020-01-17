package ign

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/////////////////////////////////////////////////
// Test a bad connection to the database
func TestBadDatabase(t *testing.T) {
	var server Server
	server.Db = nil
	err := server.dbInit()
	assert.Error(t, err, "Should have received an error from the database")
	assert.Nil(t, server.Db, "Database should be nil")
}

/// \todo: Figure out how to test the database without including username
/// and password information in the source code
