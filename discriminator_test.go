package main

import (
	"github.com/gagliardetto/utilz"
	"testing"
)

func TestAccountDiscriminator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"binArrayBitmapExtension", "BinArrayBitmapExtension"},
		{"user_account_balance", "UserAccountBalance"},
		{"UserID", "UserID"},
		{"my-event-name", "MyEventName"},
	}

	for _, tt := range tests {
		result := utilz.ToCamel(tt.input)
		if result != tt.expected {
			t.Errorf("For %s expected %s, got %s", tt.input, tt.expected, result)
		}
		d := AccountDiscriminator(tt.input)
		t.Log(d)
	}

}
