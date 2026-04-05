package rootkeys

// RootKey represents a real upstream API credential stored server-side.
//
// APIKey holds the plaintext value in memory only — it is never persisted
// directly. EncryptedAPIKey stores the AES-256-GCM ciphertext in the database.
// KeyHint exposes the last four characters so operators can identify which
// credential is stored without revealing the full value.
type RootKey struct {
	ID              string `json:"id"               gorm:"primaryKey;size:255"`
	APIKey          string `json:"api_key,omitempty" gorm:"-"`
	EncryptedAPIKey []byte `json:"-"                 gorm:"column:encrypted_api_key"`
	KeyHint         string `json:"key_hint,omitempty" gorm:"column:key_hint;size:8"`
}

func (RootKey) TableName() string { return "root_keys" }

// hint returns the last four characters of s, or the full string if shorter.
func hint(s string) string {
	if len(s) <= 4 {
		return s
	}
	return s[len(s)-4:]
}
