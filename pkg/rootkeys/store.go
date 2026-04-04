package rootkeys

import (
	"errors"
	"sync"

	"gorm.io/gorm"

	"github.com/farovictor/bifrost/pkg/crypto"
	"github.com/farovictor/bifrost/pkg/database"
)

// Store defines persistence behavior for RootKey objects.
type Store interface {
	Create(RootKey) error
	Get(id string) (RootKey, error)
	Delete(id string) error
	Update(RootKey) error
	List() []RootKey
}

// MemoryStore keeps RootKeys in memory with concurrency safety.
type MemoryStore struct {
	mu     sync.RWMutex
	keys   map[string]RootKey
	encKey []byte // nil means encryption disabled
}

// NewMemoryStore creates an initialized MemoryStore without encryption.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]RootKey)}
}

// NewMemoryStoreWithKey creates an initialized MemoryStore with AES-256-GCM encryption.
func NewMemoryStoreWithKey(encKey []byte) *MemoryStore {
	return &MemoryStore{keys: make(map[string]RootKey), encKey: encKey}
}

// SQLStore persists RootKeys in a SQL database.
type SQLStore struct {
	db     *gorm.DB
	encKey []byte // nil means encryption disabled
}

// NewSQLStore creates a SQL-backed store without encryption.
func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&RootKey{})
	return &SQLStore{db: db}
}

// NewSQLStoreWithKey creates a SQL-backed store with AES-256-GCM encryption.
func NewSQLStoreWithKey(db *gorm.DB, encKey []byte) *SQLStore {
	db.AutoMigrate(&RootKey{})
	return &SQLStore{db: db, encKey: encKey}
}

func encryptRootKey(k *RootKey, encKey []byte) error {
	if k.APIKey == "" {
		return nil
	}
	k.KeyHint = hint(k.APIKey)
	if encKey == nil {
		// No encryption key — store plaintext bytes for uniformity.
		k.EncryptedAPIKey = []byte(k.APIKey)
		k.APIKey = ""
		return nil
	}
	ct, err := crypto.Encrypt(k.APIKey, encKey)
	if err != nil {
		return err
	}
	k.EncryptedAPIKey = ct
	k.APIKey = ""
	return nil
}

func decryptRootKey(k *RootKey, encKey []byte) error {
	if len(k.EncryptedAPIKey) == 0 {
		return nil
	}
	if encKey == nil {
		k.APIKey = string(k.EncryptedAPIKey)
		return nil
	}
	pt, err := crypto.Decrypt(k.EncryptedAPIKey, encKey)
	if err != nil {
		return err
	}
	k.APIKey = pt
	return nil
}

// ── MemoryStore ──────────────────────────────────────────────────────────────

func (s *MemoryStore) Create(k RootKey) error {
	if err := encryptRootKey(&k, s.encKey); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

func (s *MemoryStore) Get(id string) (RootKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.keys[id]
	if !ok {
		return RootKey{}, ErrKeyNotFound
	}
	if err := decryptRootKey(&k, s.encKey); err != nil {
		return RootKey{}, err
	}
	return k, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

func (s *MemoryStore) Update(k RootKey) error {
	if err := encryptRootKey(&k, s.encKey); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; !ok {
		return ErrKeyNotFound
	}
	s.keys[k.ID] = k
	return nil
}

// List returns all RootKeys. APIKey is not populated — use Get for plaintext.
func (s *MemoryStore) List() []RootKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RootKey, 0, len(s.keys))
	for _, k := range s.keys {
		k.APIKey = ""
		out = append(out, k)
	}
	return out
}

// ── SQLStore ─────────────────────────────────────────────────────────────────

func (s *SQLStore) Create(k RootKey) error {
	if err := encryptRootKey(&k, s.encKey); err != nil {
		return err
	}
	if err := s.db.Create(&k).Error; err != nil {
		if database.IsDuplicateError(err) {
			return ErrKeyExists
		}
		return err
	}
	return nil
}

func (s *SQLStore) Get(id string) (RootKey, error) {
	var k RootKey
	if err := s.db.First(&k, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RootKey{}, ErrKeyNotFound
		}
		return RootKey{}, err
	}
	if err := decryptRootKey(&k, s.encKey); err != nil {
		return RootKey{}, err
	}
	return k, nil
}

func (s *SQLStore) Delete(id string) error {
	res := s.db.Delete(&RootKey{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

func (s *SQLStore) Update(k RootKey) error {
	if err := encryptRootKey(&k, s.encKey); err != nil {
		return err
	}
	res := s.db.Model(&RootKey{}).Where("id = ?", k.ID).Updates(map[string]any{
		"encrypted_api_key": k.EncryptedAPIKey,
		"key_hint":          k.KeyHint,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// List returns all root keys. APIKey is not populated — use Get for plaintext.
func (s *SQLStore) List() []RootKey {
	var out []RootKey
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

var (
	ErrKeyNotFound = errors.New("root key not found")
	ErrKeyExists   = errors.New("root key already exists")
)
