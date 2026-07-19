package demo

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	v2models "github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
)

// DemoUserPrefix is the prefix for all demo user names.
const DemoUserPrefix = "demo_"

// DemoConfig holds configuration for demo mode.
type DemoConfig struct {
	// Domain is the OAST domain used for demo payloads.
	Domain string
	// ResetInterval is how often demo data is reset.
	ResetInterval time.Duration
	// NumUsers is the number of demo users to create.
	NumUsers int
}

// DefaultConfig returns a default demo configuration.
func DefaultConfig(domain string) DemoConfig {
	return DemoConfig{
		Domain:        domain,
		ResetInterval: 6 * time.Hour,
		NumUsers:      3,
	}
}

// Manager manages demo mode: creating seed data, resetting it periodically,
// and enforcing permission restrictions for demo users.
type Manager struct {
	cfg    DemoConfig
	orm    *xorm.Engine
	mu     sync.Mutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewManager creates a demo manager.
func NewManager(orm *xorm.Engine, cfg DemoConfig) *Manager {
	return &Manager{
		cfg:    cfg,
		orm:    orm,
		stopCh: make(chan struct{}),
	}
}

// IsDemoUser checks if a username is a demo user.
func IsDemoUser(username string) bool {
	return strings.HasPrefix(username, DemoUserPrefix)
}

// IsDemoUserID checks if a user ID string belongs to a demo user.
// Demo users are identified by their name prefix in the database.
func (m *Manager) IsDemoUserID(userID string) bool {
	if m == nil || m.orm == nil {
		return false
	}
	// Check by querying the user table
	var user oldmodels.TblUser
	has, err := m.orm.Where("id = ?", userID).Get(&user)
	if err != nil || !has {
		return false
	}
	return IsDemoUser(user.Name)
}

// InitSeedData creates demo users and their associated sample data.
// It is safe to call multiple times; existing demo data will be reset.
func (m *Manager) InitSeedData() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First, clean up any existing demo data
	if err := m.resetDemoData(); err != nil {
		logrus.Warnf("[demo] failed to reset existing demo data: %v", err)
	}

	for i := 0; i < m.cfg.NumUsers; i++ {
		if err := m.createDemoUser(i); err != nil {
			logrus.Errorf("[demo] failed to create demo user %d: %v", i, err)
			continue
		}
	}

	logrus.Infof("[demo] initialized %d demo users with sample data", m.cfg.NumUsers)
	return nil
}

// createDemoUser creates a single demo user with sample Case, Payload, and Interaction.
func (m *Manager) createDemoUser(index int) error {
	userName := fmt.Sprintf("%suser%d", DemoUserPrefix, index+1)
	email := fmt.Sprintf("demo%d@godnslog.demo", index+1)
	password := fmt.Sprintf("demo%d123", index+1)

	// Create user
	user := &oldmodels.TblUser{
		Name:          userName,
		Email:         email,
		ShortId:       fmt.Sprintf("dm%d%s", index+1, randString(4)),
		Pass:          password, // Already hashed by caller's makePassword equivalent
		Token:         randString(32),
		Role:          oldmodels.RoleNormal,
		Lang:          "en-US",
		CleanInterval: 3600,
	}
	_, err := m.orm.InsertOne(user)
	if err != nil {
		return fmt.Errorf("insert demo user: %w", err)
	}
	logrus.Infof("[demo] created demo user: %s (password: %s)", userName, password)

	userIDStr := fmt.Sprintf("%d", user.Id)

	// Create sample Case
	caseID := uuid.New().String()
	demoCase := &v2models.Case{
		ID:          caseID,
		Title:       fmt.Sprintf("Demo SSRF Test #%d", index+1),
		Description: "Sample case demonstrating SSRF vulnerability detection via OAST callback",
		Status:      v2models.CaseStatusActive,
		Tags:        v2models.Tags{"ssrf", "demo"},
		Type:        "single",
		CreatedBy:   userIDStr,
	}
	if _, err := m.orm.InsertOne(demoCase); err != nil {
		logrus.Warnf("[demo] failed to create demo case: %v", err)
	}

	// Create sample Payload
	token := randString(8)
	payload := &v2models.Payload{
		ID:               uuid.New().String(),
		CaseID:           caseID,
		Token:            token,
		TemplateID:       "ssrf-basic",
		TemplateRendered: fmt.Sprintf("http://%s.%s/", token, m.cfg.Domain),
		Variables:        v2models.Variables{"token": token, "domain": m.cfg.Domain},
		Status:           v2models.PayloadStatusActive,
		ExpectedProtocol: "dns",
		CreatedBy:        userIDStr,
	}
	if _, err := m.orm.InsertOne(payload); err != nil {
		logrus.Warnf("[demo] failed to create demo payload: %v", err)
	}

	// Create sample Interaction (DNS hit)
	now := time.Now()
	domain := fmt.Sprintf("%s.%s", token, m.cfg.Domain)
	sourceIP := fmt.Sprintf("10.0.%d.%d", index+1, randInt(1, 254))
	interaction := &v2models.Interaction{
		ID:        uuid.New().String(),
		Type:      "dns",
		CaseID:    &caseID,
		PayloadID: &payload.ID,
		Token:     &token,
		Timestamp: now.Add(-2 * time.Hour),
		SourceIP:  sourceIP,
		Domain:    &domain,
		DNSType:   strPtr("A"),
		RawData:   fmt.Sprintf("DNS query from %s for %s", sourceIP, domain),
	}
	if _, err := m.orm.InsertOne(interaction); err != nil {
		logrus.Warnf("[demo] failed to create demo interaction: %v", err)
	}

	// Create a second interaction (HTTP hit)
	httpPath := fmt.Sprintf("/log/%s/", token)
	interaction2 := &v2models.Interaction{
		ID:        uuid.New().String(),
		Type:      "http",
		CaseID:    &caseID,
		PayloadID: &payload.ID,
		Token:     &token,
		Timestamp: now.Add(-1 * time.Hour),
		SourceIP:  sourceIP,
		Method:    strPtr("GET"),
		Path:      &httpPath,
		Headers:   v2models.Headers{"User-Agent": "Mozilla/5.0 (Demo)"},
		RawData:   fmt.Sprintf("GET %s from %s", httpPath, sourceIP),
	}
	if _, err := m.orm.InsertOne(interaction2); err != nil {
		logrus.Warnf("[demo] failed to create demo HTTP interaction: %v", err)
	}

	return nil
}

// StartResetLoop starts a background goroutine that periodically resets demo data.
func (m *Manager) StartResetLoop() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.cfg.ResetInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				logrus.Info("[demo] starting periodic data reset")
				if err := m.InitSeedData(); err != nil {
					logrus.Errorf("[demo] periodic reset failed: %v", err)
				}
			}
		}
	}()
	logrus.Infof("[demo] reset loop started (interval=%v)", m.cfg.ResetInterval)
}

// Stop shuts down the demo manager.
func (m *Manager) Stop() {
	close(m.stopCh)
	m.wg.Wait()
	logrus.Info("[demo] manager stopped")
}

// resetDemoData removes all demo users and their associated data.
func (m *Manager) resetDemoData() error {
	// Find all demo user IDs
	var demoUsers []oldmodels.TblUser
	if err := m.orm.Where("name LIKE ?", DemoUserPrefix+"%").Find(&demoUsers); err != nil {
		return fmt.Errorf("find demo users: %w", err)
	}

	if len(demoUsers) == 0 {
		return nil
	}

	var userIDs []string
	for _, u := range demoUsers {
		userIDs = append(userIDs, fmt.Sprintf("%d", u.Id))
	}

	// Delete interactions associated with demo users' cases
	if _, err := m.orm.In("case_id",
		// Find case IDs created by demo users
		func() []string {
			var cases []v2models.Case
			m.orm.In("created_by", userIDs).Find(&cases)
			var caseIDs []string
			for _, c := range cases {
				caseIDs = append(caseIDs, c.ID)
			}
			return caseIDs
		}(),
	).Delete(&v2models.Interaction{}); err != nil {
		logrus.Warnf("[demo] failed to delete demo interactions: %v", err)
	}

	// Delete payloads for demo users' cases
	if _, err := m.orm.In("created_by", userIDs).Delete(&v2models.Payload{}); err != nil {
		logrus.Warnf("[demo] failed to delete demo payloads: %v", err)
	}

	// Delete cases created by demo users
	if _, err := m.orm.In("created_by", userIDs).Delete(&v2models.Case{}); err != nil {
		logrus.Warnf("[demo] failed to delete demo cases: %v", err)
	}

	// Delete demo users
	if _, err := m.orm.Where("name LIKE ?", DemoUserPrefix+"%").Delete(&oldmodels.TblUser{}); err != nil {
		return fmt.Errorf("delete demo users: %w", err)
	}

	logrus.Infof("[demo] reset %d demo users and associated data", len(demoUsers))
	return nil
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}

// randString generates a random alphanumeric string of the given length.
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// randInt returns a random int in [min, max).
func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}
