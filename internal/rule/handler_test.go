package rule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handlerMockStore implements Store for handler tests
type handlerMockStore struct {
	Store
	rules      map[string]*Rule
	executions map[string][]*RuleExecution
	err        error
}

func newHandlerMockStore() *handlerMockStore {
	return &handlerMockStore{
		rules:      make(map[string]*Rule),
		executions: make(map[string][]*RuleExecution),
	}
}

func (m *handlerMockStore) CreateRule(_ context.Context, rule *Rule) error {
	if m.err != nil {
		return m.err
	}
	id := generateID()
	rule.ID = id
	m.rules[id] = rule
	return nil
}

func (m *handlerMockStore) GetRule(_ context.Context, id string) (*Rule, error) {
	if m.err != nil {
		return nil, m.err
	}
	r, ok := m.rules[id]
	if !ok {
		return nil, errors.New("rule not found")
	}
	return r, nil
}

func (m *handlerMockStore) ListRules(_ context.Context, page, pageSize int) ([]*Rule, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var all []*Rule
	for _, r := range m.rules {
		all = append(all, r)
	}
	total := int64(len(all))
	start := (page - 1) * pageSize
	if start >= len(all) {
		return []*Rule{}, total, nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (m *handlerMockStore) UpdateRule(_ context.Context, rule *Rule) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.rules[rule.ID]; !ok {
		return errors.New("rule not found")
	}
	m.rules[rule.ID] = rule
	return nil
}

func (m *handlerMockStore) DeleteRule(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.rules[id]; !ok {
		return errors.New("rule not found")
	}
	delete(m.rules, id)
	return nil
}

func (m *handlerMockStore) GetExecutions(_ context.Context, ruleID string, page, pageSize int) ([]*RuleExecution, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	execs := m.executions[ruleID]
	total := int64(len(execs))
	start := (page - 1) * pageSize
	if start >= len(execs) {
		return []*RuleExecution{}, total, nil
	}
	end := start + pageSize
	if end > len(execs) {
		end = len(execs)
	}
	return execs[start:end], total, nil
}

// updateErrorStore wraps a Store and returns an error only on UpdateRule
type updateErrorStore struct {
	Store
}

func (s *updateErrorStore) UpdateRule(_ context.Context, _ *Rule) error {
	return errors.New("update failed")
}

func setupRuleHandler() (*Handler, *handlerMockStore) {
	gin.SetMode(gin.TestMode)
	store := newHandlerMockStore()
	h := NewHandler(store)
	return h, store
}

func TestHandler_NewHandler(t *testing.T) {
	h, store := setupRuleHandler()
	assert.NotNil(t, h)
	assert.Equal(t, store, h.store)
}

func TestHandler_CreateRule(t *testing.T) {
	h, _ := setupRuleHandler()

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := CreateRuleRequest{
			Name:    "test-rule",
			Enabled: true,
			Priority: 10,
			Conditions: Conditions{
				Protocol: []string{"http"},
			},
			Actions: Actions{
				DiscardNoise: true,
			},
		}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPost, "/rules", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateRule(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		assert.Equal(t, "success", resp["message"])
		assert.NotNil(t, resp["data"])
	})

	t.Run("bad request - invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/rules", bytes.NewReader([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateRule(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("store error", func(t *testing.T) {
		h2, s2 := setupRuleHandler()
		s2.err = errors.New("db error")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := CreateRuleRequest{Name: "fail"}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPost, "/rules", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h2.CreateRule(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetRule(t *testing.T) {
	h, store := setupRuleHandler()

	// Pre-create a rule
	rule := &Rule{Name: "get-test"}
	require.NoError(t, store.CreateRule(context.Background(), rule))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: rule.ID}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/"+rule.ID, nil)

		h.GetRule(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/nonexistent", nil)

		h.GetRule(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_ListRules(t *testing.T) {
	h, store := setupRuleHandler()

	// Pre-create rules
	for i := 0; i < 5; i++ {
		store.rules[generateID()] = &Rule{Name: "rule"}
	}

	t.Run("success with defaults", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/rules", nil)

		h.ListRules(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(5), data["total"])
	})

	t.Run("with pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/rules?page=1&page_size=3", nil)

		h.ListRules(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(3), data["page_size"])
	})

	t.Run("invalid page defaults to 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/rules?page=abc", nil)

		h.ListRules(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("store error", func(t *testing.T) {
		h2, s2 := setupRuleHandler()
		s2.err = errors.New("db error")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/rules", nil)

		h2.ListRules(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_UpdateRule(t *testing.T) {
	h, store := setupRuleHandler()

	// Pre-create a rule
	rule := &Rule{Name: "original", Enabled: false}
	require.NoError(t, store.CreateRule(context.Background(), rule))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: rule.ID}}

		enabled := true
		req := UpdateRuleRequest{Enabled: &enabled}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPut, "/rules/"+rule.ID, bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateRule(c)
		assert.Equal(t, http.StatusOK, w.Code)

		updated, _ := store.GetRule(context.Background(), rule.ID)
		assert.True(t, updated.Enabled)
	})

	t.Run("bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: rule.ID}}
		c.Request = httptest.NewRequest(http.MethodPut, "/rules/"+rule.ID, bytes.NewReader([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateRule(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}

		enabled := true
		req := UpdateRuleRequest{Enabled: &enabled}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPut, "/rules/nonexistent", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateRule(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("store error on update", func(t *testing.T) {
		_, s2 := setupRuleHandler()
		r2 := &Rule{Name: "err-test"}
		require.NoError(t, s2.CreateRule(context.Background(), r2))

		// Wrap with err-only-on-update store
		updateErrStore := &updateErrorStore{Store: s2}
		h3 := NewHandler(updateErrStore)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: r2.ID}}

		name := "new name"
		req := UpdateRuleRequest{Name: &name}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPut, "/rules/"+r2.ID, bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h3.UpdateRule(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_DeleteRule(t *testing.T) {
	h, store := setupRuleHandler()

	rule := &Rule{Name: "to-delete"}
	require.NoError(t, store.CreateRule(context.Background(), rule))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: rule.ID}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/rules/"+rule.ID, nil)

		h.DeleteRule(c)
		assert.Equal(t, http.StatusOK, w.Code)

		_, err := store.GetRule(context.Background(), rule.ID)
		assert.Error(t, err)
	})

	t.Run("store error", func(t *testing.T) {
		h2, s2 := setupRuleHandler()
		s2.err = errors.New("delete failed")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "any"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/rules/any", nil)

		h2.DeleteRule(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetExecutions(t *testing.T) {
	h, store := setupRuleHandler()

	ruleID := "rule-exec-1"
	store.executions[ruleID] = []*RuleExecution{
		{ID: "exec-1", RuleID: ruleID, Matched: true},
		{ID: "exec-2", RuleID: ruleID, Matched: false},
	}

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: ruleID}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/"+ruleID+"/executions", nil)

		h.GetExecutions(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(2), data["total"])
	})

	t.Run("with pagination params", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: ruleID}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/"+ruleID+"/executions?page=1&page_size=1", nil)

		h.GetExecutions(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid page defaults", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: ruleID}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/"+ruleID+"/executions?page=abc&page_size=xyz", nil)

		h.GetExecutions(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("store error", func(t *testing.T) {
		h2, s2 := setupRuleHandler()
		s2.err = errors.New("exec error")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: ruleID}}
		c.Request = httptest.NewRequest(http.MethodGet, "/rules/"+ruleID+"/executions", nil)

		h2.GetExecutions(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
