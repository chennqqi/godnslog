package demo

import (
	"fmt"
	"testing"

	v2models "github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupTestDB(t *testing.T) *xorm.Engine {
	t.Helper()
	orm, err := xorm.NewEngine("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	if err := orm.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	orm.Sync2(&oldmodels.TblUser{}, &v2models.Case{}, &v2models.Payload{}, &v2models.Interaction{})
	return orm
}

func TestIsDemoUser(t *testing.T) {
	tests := []struct {
		username string
		want     bool
	}{
		{"demo_user1", true},
		{"demo_admin", true},
		{"admin", false},
		{"user", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsDemoUser(tt.username); got != tt.want {
			t.Errorf("IsDemoUser(%q) = %v, want %v", tt.username, got, tt.want)
		}
	}
}

func TestInitSeedData(t *testing.T) {
	orm := setupTestDB(t)
	m := NewManager(orm, DemoConfig{
		Domain:        "demo.example.com",
		ResetInterval: 0, // disable loop
		NumUsers:      2,
	})

	if err := m.InitSeedData(); err != nil {
		t.Fatalf("InitSeedData: %v", err)
	}

	// Verify demo users were created
	count, err := orm.Where("name LIKE ?", DemoUserPrefix+"%").Count(&oldmodels.TblUser{})
	if err != nil {
		t.Fatalf("count demo users: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 demo users, got %d", count)
	}

	// Verify cases were created
	caseCount, err := orm.Count(&v2models.Case{})
	if err != nil {
		t.Fatalf("count cases: %v", err)
	}
	if caseCount != 2 {
		t.Fatalf("expected 2 demo cases, got %d", caseCount)
	}

	// Verify payloads were created
	payloadCount, err := orm.Count(&v2models.Payload{})
	if err != nil {
		t.Fatalf("count payloads: %v", err)
	}
	if payloadCount != 2 {
		t.Fatalf("expected 2 demo payloads, got %d", payloadCount)
	}

	// Verify interactions were created (2 per user: 1 DNS + 1 HTTP)
	interactionCount, err := orm.Count(&v2models.Interaction{})
	if err != nil {
		t.Fatalf("count interactions: %v", err)
	}
	if interactionCount != 4 {
		t.Fatalf("expected 4 demo interactions, got %d", interactionCount)
	}
}

func TestResetDemoData(t *testing.T) {
	orm := setupTestDB(t)
	m := NewManager(orm, DemoConfig{
		Domain:   "demo.example.com",
		NumUsers: 3,
	})

	// Initialize
	if err := m.InitSeedData(); err != nil {
		t.Fatalf("InitSeedData: %v", err)
	}

	// Verify data exists
	count, _ := orm.Where("name LIKE ?", DemoUserPrefix+"%").Count(&oldmodels.TblUser{})
	if count != 3 {
		t.Fatalf("expected 3 demo users after init, got %d", count)
	}

	// Reset
	if err := m.InitSeedData(); err != nil {
		t.Fatalf("re-init (reset): %v", err)
	}

	// Verify data was reset and recreated
	count, _ = orm.Where("name LIKE ?", DemoUserPrefix+"%").Count(&oldmodels.TblUser{})
	if count != 3 {
		t.Fatalf("expected 3 demo users after reset, got %d", count)
	}
}

func TestIsDemoUserID(t *testing.T) {
	orm := setupTestDB(t)
	m := NewManager(orm, DemoConfig{
		Domain:   "demo.example.com",
		NumUsers: 1,
	})

	if err := m.InitSeedData(); err != nil {
		t.Fatalf("InitSeedData: %v", err)
	}

	// Find the demo user
	var user oldmodels.TblUser
	has, err := orm.Where("name LIKE ?", DemoUserPrefix+"%").Get(&user)
	if err != nil || !has {
		t.Fatalf("could not find demo user: %v %v", err, has)
	}

	if !m.IsDemoUserID(fmt.Sprintf("%d", user.Id)) {
		t.Fatal("IsDemoUserID should return true for demo user")
	}

	if m.IsDemoUserID("999999") {
		t.Fatal("IsDemoUserID should return false for non-existent user")
	}
}

func TestNilManagerIsDemoUserID(t *testing.T) {
	var m *Manager
	if m.IsDemoUserID("1") {
		t.Fatal("nil manager should return false")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test.example.com")
	if cfg.Domain != "test.example.com" {
		t.Fatalf("expected domain test.example.com, got %s", cfg.Domain)
	}
	if cfg.NumUsers != 3 {
		t.Fatalf("expected 3 users, got %d", cfg.NumUsers)
	}
	if cfg.ResetInterval <= 0 {
		t.Fatal("reset interval should be positive")
	}
}
