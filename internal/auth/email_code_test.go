package auth

import "testing"

func TestEmailCodeManagerGenerateAndVerify(t *testing.T) {
	manager, err := NewEmailCodeManager(
		"01234567890123456789012345678901",
	)
	if err != nil {
		t.Fatalf("create email code manager: %v", err)
	}

	plainCode, codeHash, err := manager.Generate()
	if err != nil {
		t.Fatalf("generate email code: %v", err)
	}

	if len(plainCode) != 6 {
		t.Fatalf("expected code length 6, got %d", len(plainCode))
	}

	for _, symbol := range plainCode {
		if symbol < '0' || symbol > '9' {
			t.Fatalf("code contains non-digit: %q", plainCode)
		}
	}

	if !manager.Verify(plainCode, codeHash) {
		t.Fatal("expected generated code to pass verification")
	}

	wrongCode := plainCode
	if plainCode[0] == '0' {
		wrongCode = "1" + plainCode[1:]
	} else {
		wrongCode = "0" + plainCode[1:]
	}

	if manager.Verify(wrongCode, codeHash) {
		t.Fatal("expected changed code to fail verification")
	}

	if manager.Verify(plainCode, "not-hex") {
		t.Fatal("expected malformed hash to fail verification")
	}
}
