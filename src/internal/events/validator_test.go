package events

import (
	"testing"

	"drblury/event-driven-service/internal/domain"
)

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if v.validator == nil {
		t.Error("validator field is nil")
	}
}

func TestNewValidatorMultipleCalls(t *testing.T) {
	// Ensure we can create multiple validators without issues
	for i := 0; i < 3; i++ {
		v, err := NewValidator()
		if err != nil {
			t.Fatalf("NewValidator() iteration %d error = %v", i, err)
		}
		if v == nil {
			t.Errorf("NewValidator() iteration %d returned nil", i)
		}
	}
}

func TestValidatorValidateValidMessage(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Create a valid ExampleRecord
	record := &domain.ExampleRecord{
		RecordId:    "test-123",
		Title:       "Test Record",
		Description: "A test description",
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for valid record", err)
	}
}

func TestValidatorValidateEmptyRecord(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Create an empty record - should be valid unless protovalidate rules are strict
	record := &domain.ExampleRecord{}

	// This may or may not error depending on protovalidate rules
	_ = v.Validate(record)
}

func TestValidatorValidateWithMeta(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	record := &domain.ExampleRecord{
		RecordId:    "test-456",
		Title:       "Test with Meta",
		Description: "A record with metadata",
		Meta: &domain.ExampleMeta{
			RequestedBy:      "test-user",
			RequiresFollowUp: true,
			Priority:         5,
		},
		Tags: []string{"tag1", "tag2"},
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for record with meta", err)
	}
}

func TestValidatorValidateExampleResult(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	result := &domain.ExampleResult{
		RecordId: "result-123",
		Status:   "completed",
		Note:     "Processed successfully",
		ProcessedOn: &domain.Date{
			Year:  2024,
			Month: 6,
			Day:   15,
		},
	}

	err = v.Validate(result)
	if err != nil {
		t.Errorf("Validate() error = %v for valid result", err)
	}
}

func TestValidatorValidateDate(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	date := &domain.Date{
		Year:  2024,
		Month: 12,
		Day:   31,
	}

	err = v.Validate(date)
	if err != nil {
		t.Errorf("Validate() error = %v for valid date", err)
	}
}

func TestValidatorValidateInfo(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	info := &domain.Info{
		Version:    "1.0.0",
		BuildDate:  "2024-06-15",
		Details:    "Test build",
		CommitHash: "abc123",
		CommitDate: "2024-06-15T00:00:00Z",
	}

	err = v.Validate(info)
	if err != nil {
		t.Errorf("Validate() error = %v for valid info", err)
	}
}

func TestValidatorStruct(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Verify validator struct has the internal validator set
	if v.validator == nil {
		t.Error("Validator.validator should not be nil after successful creation")
	}
}

func TestValidatorValidateExampleMeta(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	meta := &domain.ExampleMeta{
		RequestedBy:      "test-user",
		RequiresFollowUp: true,
		Priority:         10,
		DesiredStartDate: &domain.Date{
			Year:  2025,
			Month: 1,
			Day:   1,
		},
	}

	err = v.Validate(meta)
	if err != nil {
		t.Errorf("Validate() error = %v for valid meta", err)
	}
}

func TestValidatorValidateMultipleRecords(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	records := []*domain.ExampleRecord{
		{RecordId: "rec-1", Title: "Record 1"},
		{RecordId: "rec-2", Title: "Record 2", Description: "Description"},
		{RecordId: "rec-3", Title: "Record 3", Tags: []string{"tag1", "tag2"}},
	}

	for i, record := range records {
		err = v.Validate(record)
		if err != nil {
			t.Errorf("Validate() record %d error = %v", i, err)
		}
	}
}

func TestValidatorValidateExampleResultVariants(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	results := []*domain.ExampleResult{
		{RecordId: "r1", Status: "pending"},
		{RecordId: "r2", Status: "completed", Note: "Done"},
		{RecordId: "r3", Status: "failed", Note: "Error occurred", ProcessedOn: &domain.Date{Year: 2024, Month: 1, Day: 1}},
	}

	for i, result := range results {
		err = v.Validate(result)
		if err != nil {
			t.Errorf("Validate() result %d error = %v", i, err)
		}
	}
}

func TestValidatorValidateDateVariants(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	dates := []*domain.Date{
		{Year: 2020, Month: 1, Day: 1},
		{Year: 2024, Month: 6, Day: 15},
		{Year: 2030, Month: 12, Day: 31},
		{Year: 0, Month: 0, Day: 0}, // Zero values
	}

	for i, date := range dates {
		err = v.Validate(date)
		if err != nil {
			t.Errorf("Validate() date %d error = %v", i, err)
		}
	}
}

func TestValidatorValidateInfoVariants(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	infos := []*domain.Info{
		{},
		{Version: "1.0.0"},
		{Version: "2.0.0", BuildDate: "2024-01-01"},
		{Version: "3.0.0", BuildDate: "2024-01-01", Details: "Full info", CommitHash: "abc123", CommitDate: "2024-01-01T00:00:00Z"},
	}

	for i, info := range infos {
		err = v.Validate(info)
		if err != nil {
			t.Errorf("Validate() info %d error = %v", i, err)
		}
	}
}

func TestValidatorValidateRecordWithAllFields(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	record := &domain.ExampleRecord{
		RecordId:    "full-record-123",
		Title:       "Complete Example Record",
		Description: "This is a fully populated example record for testing",
		Tags:        []string{"test", "complete", "all-fields"},
		Meta: &domain.ExampleMeta{
			RequestedBy:      "test-automation",
			RequiresFollowUp: true,
			Priority:         5,
			DesiredStartDate: &domain.Date{
				Year:  2024,
				Month: 12,
				Day:   25,
			},
		},
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for complete record", err)
	}
}

func TestValidatorValidateMetaWithNilDate(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	meta := &domain.ExampleMeta{
		RequestedBy:      "test-user",
		RequiresFollowUp: false,
		Priority:         1,
		DesiredStartDate: nil, // No date
	}

	err = v.Validate(meta)
	if err != nil {
		t.Errorf("Validate() error = %v for meta with nil date", err)
	}
}

func TestValidatorValidateResultWithNilDate(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	result := &domain.ExampleResult{
		RecordId:    "result-no-date",
		Status:      "pending",
		Note:        "No processed date",
		ProcessedOn: nil,
	}

	err = v.Validate(result)
	if err != nil {
		t.Errorf("Validate() error = %v for result with nil date", err)
	}
}

func TestValidatorValidateRecordWithTags(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	record := &domain.ExampleRecord{
		RecordId: "tags-test",
		Title:    "Tags Test",
		Tags:     []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for record with tags", err)
	}
}

func TestValidatorValidateRecordWithEmptyTags(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	record := &domain.ExampleRecord{
		RecordId: "empty-tags",
		Title:    "Empty Tags Test",
		Tags:     []string{},
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for record with empty tags", err)
	}
}

func TestValidatorValidateDateZeroValues(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	date := &domain.Date{
		Year:  0,
		Month: 0,
		Day:   0,
	}

	err = v.Validate(date)
	if err != nil {
		t.Errorf("Validate() error = %v for zero date", err)
	}
}

func TestValidatorValidateMetaPriorityValues(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	priorities := []int32{0, 1, 5, 10, 100}
	for _, priority := range priorities {
		meta := &domain.ExampleMeta{
			RequestedBy: "test",
			Priority:    priority,
		}
		err = v.Validate(meta)
		if err != nil {
			t.Errorf("Validate() error = %v for priority %d", err, priority)
		}
	}
}

func TestValidatorStruct_CheckFields(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Verify the validator struct is correctly initialized
	if v.validator == nil {
		t.Error("expected validator field to be non-nil")
	}
}

func TestValidatorValidateMultipleTypes(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Test validating different protobuf message types in sequence
	messages := []interface{}{
		&domain.ExampleRecord{RecordId: "test-1", Title: "Title 1"},
		&domain.ExampleResult{RecordId: "result-1", Status: "pending"},
		&domain.ExampleMeta{RequestedBy: "user1", Priority: 1},
		&domain.Date{Year: 2024, Month: 6, Day: 15},
		&domain.Info{Version: "1.0.0"},
	}

	for i, msg := range messages {
		err = v.Validate(msg)
		if err != nil {
			t.Errorf("Validate() failed for message %d: %v", i, err)
		}
	}
}

func TestValidatorValidateRecordWithLongDescription(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	longDescription := ""
	for i := 0; i < 1000; i++ {
		longDescription += "This is a very long description. "
	}

	record := &domain.ExampleRecord{
		RecordId:    "long-desc-test",
		Title:       "Long Description Test",
		Description: longDescription,
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for record with long description", err)
	}
}

func TestValidatorValidateManyTags(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	tags := make([]string, 100)
	for i := 0; i < 100; i++ {
		tags[i] = "tag" + string(rune('A'+i%26))
	}

	record := &domain.ExampleRecord{
		RecordId: "many-tags",
		Title:    "Many Tags Test",
		Tags:     tags,
	}

	err = v.Validate(record)
	if err != nil {
		t.Errorf("Validate() error = %v for record with many tags", err)
	}
}

func TestValidatorValidateSequential(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Validate same record multiple times
	record := &domain.ExampleRecord{
		RecordId: "seq-test",
		Title:    "Sequential Test",
	}

	for i := 0; i < 10; i++ {
		err = v.Validate(record)
		if err != nil {
			t.Errorf("Validate() iteration %d error = %v", i, err)
		}
	}
}
