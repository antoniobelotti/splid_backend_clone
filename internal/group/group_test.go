//go:build unit
// +build unit

package group

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInvitationCodeGeneration(t *testing.T) {
	t.Parallel()

	groupName := "test"
	ownerId := 42

	invitationCode, err := getHopefullyUniqueInvitationCode(groupName, ownerId)
	if err != nil {
		t.Errorf("Expected err to be nil, but got %v", err)
	}

	assert.Len(t, invitationCode, 6, "invitation code should be 6 characters long")
}
