//go:build unit

package group

import (
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func FuzzInvitationCodeGeneration(f *testing.F) {
	f.Add("test", 42)
	f.Add("test", 43)
	f.Fuzz(func(t *testing.T, groupName string, ownerId int) {
		invitationCode, err := getHopefullyUniqueInvitationCode(groupName, ownerId)
		if err != nil {
			t.Errorf("Expected err to be nil, but got %v", err)
		}
		assert.Len(t, invitationCode, 6, "invitation code should be 6 characters long")
	})
}

func TestCalculateGroupBalance(t *testing.T) {
	// people in the group
	componentIds := []int{1, 2, 3}

	// p1 spends 10€ then 2,50 then 5, then 0.9    	= 18,4
	// p2 spends 7,77€ then 50						= 57.77
	// p3 spends nothing							= 0
	expenses := []expense.Expense{
		{
			AmountInCents: 1000,
			PersonId:      1,
		},
		{
			AmountInCents: 250,
			PersonId:      1,
		},
		{
			AmountInCents: 500,
			PersonId:      1,
		},
		{
			AmountInCents: 90,
			PersonId:      1,
		},
		{
			AmountInCents: 777,
			PersonId:      2,
		},
		{
			AmountInCents: 5000,
			PersonId:      2,
		},
	}
	var transfers []transfer.Transfer

	// avg is (10+2.5+5+0.9+7.77+50)/6 = 12.695

	expectedBalance := map[int]int{
		1: 1840 - 1269,
		2: 5777 - 1269,
		3: 0000 - 1269,
	}

	for p, b := range calculateGroupBalance(componentIds, expenses, transfers) {
		assert.Equal(t, expectedBalance[p], b, fmt.Sprintf("for person %d epected balance %d but got %d instead", p, expectedBalance[p], b))
	}

	// let's say p3 transfers 5€ to p2
	transfers = append(transfers, transfer.Transfer{
		AmountInCents: 500,
		SenderId:      3,
		ReceiverId:    2,
	})
	expectedBalance[3] -= 500
	expectedBalance[2] += 500

	for p, b := range calculateGroupBalance(componentIds, expenses, transfers) {
		assert.Equal(t, expectedBalance[p], b, fmt.Sprintf("for person %d epected balance %d but got %d instead", p, expectedBalance[p], b))
	}
}
