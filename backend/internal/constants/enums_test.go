package constants

import "testing"

func TestRoleValid(t *testing.T) {
	tests := []struct {
		input  Role
		expect bool
	}{
		{RoleUser, true},
		{RoleAdmin, true},
		{Role("hacker"), false},
	}
	for _, tt := range tests {
		if got := tt.input.Valid(); got != tt.expect {
			t.Fatalf("Role(%q).Valid() = %v, want %v", tt.input, got, tt.expect)
		}
	}
}

func TestPostStatusValues(t *testing.T) {
	values := []PostStatus{PostStatusPending, PostStatusApproved, PostStatusRejected}
	if len(values) != 3 {
		t.Fatalf("expect 3 post statuses, got %d", len(values))
	}
	for _, v := range values {
		if v == "" {
			t.Fatalf("post status must not be empty")
		}
	}
}
