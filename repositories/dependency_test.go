package repositories

import "testing"

func TestDependency_ValidateSuccess(t *testing.T) {
	dep := Dependency{
		Id:              "test",
		InteractionType: "config",
	}
	err := dep.Validate()
	if err != nil {
		t.Error(err)
	}
}

func TestDependency_ValidateFailNoId(t *testing.T) {
	dep := Dependency{}
	err := dep.Validate()
	if err == nil {
		t.Error("Expected error")
	} else {
		expectedMsg := "dependency id is required" // Adjust to match the actual expected message
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	}

}

func TestDependency_ValidateFailInvalidInteractionType(t *testing.T) {
	dep := Dependency{
		Id:              "test",
		InteractionType: "invalid",
	}
	err := dep.Validate()
	if err == nil {
		t.Error("Expected error")
	}
}
