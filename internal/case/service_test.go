package casemgmt

import (
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestCreateCase(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Case))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &CaseCreateRequest{
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Tags:        []string{"tag-1"},
	}

	testCase, err := service.CreateCase(req, "user-1")
	if err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}
	if testCase == nil || testCase.ID == "" {
		t.Fatal("expected created case with ID")
	}
	if testCase.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, testCase.Title)
	}
}

func TestGetCaseByID(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Case))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	testCaseModel := &Case{
		ID:          "case-1",
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   "user-1",
	}

	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	retrieved, err := service.GetCaseByID(testCaseModel.ID)
	if err != nil {
		t.Fatalf("Failed to get case: %v", err)
	}
	if retrieved.Title != testCaseModel.Title {
		t.Errorf("Expected title %s, got %s", testCaseModel.Title, retrieved.Title)
	}
}

func TestUpdateCase(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Case))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	testCaseModel := &Case{
		ID:          "case-2",
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   "user-1",
	}

	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	req := &CaseUpdateRequest{
		Title:       "Updated Title",
		Description: "Updated Description",
		Target:      "updated.target.com",
		Status:      "completed",
		Tags:        []string{"tag-2"},
	}

	err = service.UpdateCase(testCaseModel.ID, req)
	if err != nil {
		t.Fatalf("Failed to update case: %v", err)
	}

	retrieved, err := service.GetCaseByID(testCaseModel.ID)
	if err != nil {
		t.Fatalf("Failed to get updated case: %v", err)
	}
	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected updated title, got %s", retrieved.Title)
	}
}

func TestDeleteCase(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Case))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	testCaseModel := &Case{
		ID:          "case-3",
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   "user-1",
	}

	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	err = service.DeleteCase(testCaseModel.ID)
	if err != nil {
		t.Fatalf("Failed to delete case: %v", err)
	}

	_, err = service.GetCaseByID(testCaseModel.ID)
	if err == nil {
		t.Error("Expected error when getting deleted case, got nil")
	}
}

func TestListCases(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Case))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	for i := 0; i < 5; i++ {
		testCaseModel := &Case{
			ID:          fmt.Sprintf("case-list-%d", i),
			Title:       "Test Case",
			Description: "Test Description",
			Target:      "test.example.com",
			Status:      "active",
			CreatedBy:   "user-1",
		}
		_, err = engine.Insert(testCaseModel)
		if err != nil {
			t.Fatalf("Failed to insert test case: %v", err)
		}
	}

	resp, err := service.ListCases("active", "", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list cases: %v", err)
	}
	if resp.Total != 5 {
		t.Errorf("Expected 5 cases, got %d", resp.Total)
	}
	if len(resp.Items) != 5 {
		t.Errorf("Expected 5 cases in list, got %d", len(resp.Items))
	}
}
