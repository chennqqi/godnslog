package case

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/models"
	"xorm.io/xorm"
)

func TestCreateCase(t *testing.T) {
	// This is a basic unit test structure
	// In a real scenario, you would need to set up a test database
	
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblCase))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)
	
	testCase := &models.TblCase{
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = service.CreateCase(testCase.Title, testCase.Description, testCase.Target, testCase.Status, []string{}, testCase.CreatedBy)
	if err != nil {
		t.Errorf("Failed to create case: %v", err)
	}
}

func TestGetCaseByID(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblCase))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)
	
	// Create a test case first
	testCaseModel := &models.TblCase{
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	// Test getting the case
	retrieved, err := service.GetCaseByID(testCaseModel.Id)
	if err != nil {
		t.Errorf("Failed to get case: %v", err)
	}
	if retrieved.Title != testCaseModel.Title {
		t.Errorf("Expected title %s, got %s", testCaseModel.Title, retrieved.Title)
	}
}

func TestUpdateCase(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblCase))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)
	
	testCaseModel := &models.TblCase{
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	// Test updating the case
	err = service.UpdateCase(testCaseModel.Id, "Updated Title", "Updated Description", "updated.target.com", "completed", []string{})
	if err != nil {
		t.Errorf("Failed to update case: %v", err)
	}

	// Verify the update
	retrieved, err := service.GetCaseByID(testCaseModel.Id)
	if err != nil {
		t.Errorf("Failed to get updated case: %v", err)
	}
	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected updated title, got %s", retrieved.Title)
	}
}

func TestDeleteCase(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblCase))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)
	
	testCaseModel := &models.TblCase{
		Title:       "Test Case",
		Description: "Test Description",
		Target:      "test.example.com",
		Status:      "active",
		CreatedBy:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	_, err = engine.Insert(testCaseModel)
	if err != nil {
		t.Fatalf("Failed to insert test case: %v", err)
	}

	// Test deleting the case
	err = service.DeleteCase(testCaseModel.Id)
	if err != nil {
		t.Errorf("Failed to delete case: %v", err)
	}

	// Verify the deletion
	_, err = service.GetCaseByID(testCaseModel.Id)
	if err == nil {
		t.Error("Expected error when getting deleted case, got nil")
	}
}

func TestListCases(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblCase))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)
	
	// Insert test cases
	for i := 0; i < 5; i++ {
		testCaseModel := &models.TblCase{
			Title:       "Test Case",
			Description: "Test Description",
			Target:      "test.example.com",
			Status:      "active",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_, err = engine.Insert(testCaseModel)
		if err != nil {
			t.Fatalf("Failed to insert test case: %v", err)
		}
	}

	// Test listing cases
	cases, total, err := service.ListCases(1, 10, "", "")
	if err != nil {
		t.Errorf("Failed to list cases: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected 5 cases, got %d", total)
	}
	if len(cases) != 5 {
		t.Errorf("Expected 5 cases in list, got %d", len(cases))
	}
}
