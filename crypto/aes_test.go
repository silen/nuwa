package crypto

import "testing"

func TestDecryptAESRejectsShortCiphertext(t *testing.T) {
	t.Parallel()

	if _, err := DecryptAES(DataSourceKey, "0011"); err == nil {
		t.Fatalf("expected short ciphertext error")
	}
}

func TestPKCS7UnpaddingRejectsEmptyInput(t *testing.T) {
	t.Parallel()

	if _, err := pkcs7Unpadding(nil); err == nil {
		t.Fatalf("expected empty padding error")
	}
}
