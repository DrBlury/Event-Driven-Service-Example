package domain

import (
	"testing"
)

func TestDateGetters(t *testing.T) {
	t.Parallel()

	date := &Date{
		Year:  2024,
		Month: 6,
		Day:   15,
	}

	if date.GetYear() != 2024 {
		t.Errorf("GetYear() = %d, want 2024", date.GetYear())
	}
	if date.GetMonth() != 6 {
		t.Errorf("GetMonth() = %d, want 6", date.GetMonth())
	}
	if date.GetDay() != 15 {
		t.Errorf("GetDay() = %d, want 15", date.GetDay())
	}
}

func TestDateGettersNil(t *testing.T) {
	t.Parallel()

	var date *Date

	if date.GetYear() != 0 {
		t.Errorf("GetYear() on nil = %d, want 0", date.GetYear())
	}
	if date.GetMonth() != 0 {
		t.Errorf("GetMonth() on nil = %d, want 0", date.GetMonth())
	}
	if date.GetDay() != 0 {
		t.Errorf("GetDay() on nil = %d, want 0", date.GetDay())
	}
}

func TestDateProtoMethods(t *testing.T) {
	t.Parallel()

	date := &Date{
		Year:  2024,
		Month: 12,
		Day:   25,
	}

	// Test String
	str := date.String()
	if str == "" {
		t.Error("String() should not be empty")
	}

	// Test Reset
	date.Reset()
	if date.GetYear() != 0 || date.GetMonth() != 0 || date.GetDay() != 0 {
		t.Error("Reset() should zero all fields")
	}

	// Test ProtoMessage (just ensure it doesn't panic)
	date.ProtoMessage()

	// Test ProtoReflect
	pr := date.ProtoReflect()
	if pr == nil {
		t.Error("ProtoReflect() should not return nil")
	}

	// Test Descriptor
	desc, nums := date.Descriptor()
	if len(desc) == 0 {
		t.Error("Descriptor() should return non-empty bytes")
	}
	if len(nums) != 1 || nums[0] != 0 {
		t.Errorf("Descriptor() index = %v, want [0]", nums)
	}
}

func TestExampleRecordGetters(t *testing.T) {
	t.Parallel()

	record := &ExampleRecord{
		RecordId:    "test-123",
		Title:       "Test Title",
		Description: "Test Description",
		Tags:        []string{"tag1", "tag2"},
		Meta: &ExampleMeta{
			RequestedBy:      "user",
			RequiresFollowUp: true,
			Priority:         5,
		},
	}

	if record.GetRecordId() != "test-123" {
		t.Errorf("GetRecordId() = %q, want 'test-123'", record.GetRecordId())
	}
	if record.GetTitle() != "Test Title" {
		t.Errorf("GetTitle() = %q, want 'Test Title'", record.GetTitle())
	}
	if record.GetDescription() != "Test Description" {
		t.Errorf("GetDescription() = %q, want 'Test Description'", record.GetDescription())
	}
	if len(record.GetTags()) != 2 {
		t.Errorf("GetTags() length = %d, want 2", len(record.GetTags()))
	}
	if record.GetMeta() == nil {
		t.Error("GetMeta() should not be nil")
	}
}

func TestExampleRecordGettersNil(t *testing.T) {
	t.Parallel()

	var record *ExampleRecord

	if record.GetRecordId() != "" {
		t.Errorf("GetRecordId() on nil = %q, want ''", record.GetRecordId())
	}
	if record.GetTitle() != "" {
		t.Errorf("GetTitle() on nil = %q, want ''", record.GetTitle())
	}
	if record.GetDescription() != "" {
		t.Errorf("GetDescription() on nil = %q, want ''", record.GetDescription())
	}
	if record.GetTags() != nil {
		t.Errorf("GetTags() on nil = %v, want nil", record.GetTags())
	}
	if record.GetMeta() != nil {
		t.Error("GetMeta() on nil should be nil")
	}
}

func TestExampleRecordProtoMethods(t *testing.T) {
	t.Parallel()

	record := &ExampleRecord{
		RecordId: "test-456",
		Title:    "Proto Test",
	}

	// Test String
	str := record.String()
	if str == "" {
		t.Error("String() should not be empty")
	}

	// Test Reset
	record.Reset()
	if record.GetRecordId() != "" || record.GetTitle() != "" {
		t.Error("Reset() should zero all fields")
	}

	// Test ProtoMessage
	record.ProtoMessage()

	// Test ProtoReflect
	pr := record.ProtoReflect()
	if pr == nil {
		t.Error("ProtoReflect() should not return nil")
	}

	// Test Descriptor
	desc, nums := record.Descriptor()
	if len(desc) == 0 {
		t.Error("Descriptor() should return non-empty bytes")
	}
	if len(nums) != 1 {
		t.Errorf("Descriptor() index length = %d, want 1", len(nums))
	}
}

func TestExampleMetaGetters(t *testing.T) {
	t.Parallel()

	meta := &ExampleMeta{
		RequestedBy:      "test-user",
		RequiresFollowUp: true,
		Priority:         3,
		DesiredStartDate: &Date{
			Year:  2024,
			Month: 7,
			Day:   1,
		},
	}

	if meta.GetRequestedBy() != "test-user" {
		t.Errorf("GetRequestedBy() = %q, want 'test-user'", meta.GetRequestedBy())
	}
	if !meta.GetRequiresFollowUp() {
		t.Error("GetRequiresFollowUp() should be true")
	}
	if meta.GetPriority() != 3 {
		t.Errorf("GetPriority() = %d, want 3", meta.GetPriority())
	}
	if meta.GetDesiredStartDate() == nil {
		t.Error("GetDesiredStartDate() should not be nil")
	}
}

func TestExampleMetaGettersNil(t *testing.T) {
	t.Parallel()

	var meta *ExampleMeta

	if meta.GetRequestedBy() != "" {
		t.Errorf("GetRequestedBy() on nil = %q, want ''", meta.GetRequestedBy())
	}
	if meta.GetRequiresFollowUp() {
		t.Error("GetRequiresFollowUp() on nil should be false")
	}
	if meta.GetPriority() != 0 {
		t.Errorf("GetPriority() on nil = %d, want 0", meta.GetPriority())
	}
	if meta.GetDesiredStartDate() != nil {
		t.Error("GetDesiredStartDate() on nil should be nil")
	}
}

func TestExampleMetaProtoMethods(t *testing.T) {
	t.Parallel()

	meta := &ExampleMeta{
		RequestedBy: "proto-test",
		Priority:    1,
	}

	// Test String
	str := meta.String()
	if str == "" {
		t.Error("String() should not be empty")
	}

	// Test Reset
	meta.Reset()
	if meta.GetRequestedBy() != "" || meta.GetPriority() != 0 {
		t.Error("Reset() should zero all fields")
	}

	// Test ProtoMessage
	meta.ProtoMessage()

	// Test ProtoReflect
	pr := meta.ProtoReflect()
	if pr == nil {
		t.Error("ProtoReflect() should not return nil")
	}

	// Test Descriptor
	desc, nums := meta.Descriptor()
	if len(desc) == 0 {
		t.Error("Descriptor() should return non-empty bytes")
	}
	if len(nums) != 1 {
		t.Errorf("Descriptor() index length = %d, want 1", len(nums))
	}
}

func TestExampleResultGetters(t *testing.T) {
	t.Parallel()

	result := &ExampleResult{
		RecordId: "result-123",
		Status:   "completed",
		Note:     "Test note",
		ProcessedOn: &Date{
			Year:  2024,
			Month: 8,
			Day:   15,
		},
	}

	if result.GetRecordId() != "result-123" {
		t.Errorf("GetRecordId() = %q, want 'result-123'", result.GetRecordId())
	}
	if result.GetStatus() != "completed" {
		t.Errorf("GetStatus() = %q, want 'completed'", result.GetStatus())
	}
	if result.GetNote() != "Test note" {
		t.Errorf("GetNote() = %q, want 'Test note'", result.GetNote())
	}
	if result.GetProcessedOn() == nil {
		t.Error("GetProcessedOn() should not be nil")
	}
}

func TestExampleResultGettersNil(t *testing.T) {
	t.Parallel()

	var result *ExampleResult

	if result.GetRecordId() != "" {
		t.Errorf("GetRecordId() on nil = %q, want ''", result.GetRecordId())
	}
	if result.GetStatus() != "" {
		t.Errorf("GetStatus() on nil = %q, want ''", result.GetStatus())
	}
	if result.GetNote() != "" {
		t.Errorf("GetNote() on nil = %q, want ''", result.GetNote())
	}
	if result.GetProcessedOn() != nil {
		t.Error("GetProcessedOn() on nil should be nil")
	}
}

func TestExampleResultProtoMethods(t *testing.T) {
	t.Parallel()

	result := &ExampleResult{
		RecordId: "proto-result",
		Status:   "pending",
	}

	// Test String
	str := result.String()
	if str == "" {
		t.Error("String() should not be empty")
	}

	// Test Reset
	result.Reset()
	if result.GetRecordId() != "" || result.GetStatus() != "" {
		t.Error("Reset() should zero all fields")
	}

	// Test ProtoMessage
	result.ProtoMessage()

	// Test ProtoReflect
	pr := result.ProtoReflect()
	if pr == nil {
		t.Error("ProtoReflect() should not return nil")
	}

	// Test Descriptor
	desc, nums := result.Descriptor()
	if len(desc) == 0 {
		t.Error("Descriptor() should return non-empty bytes")
	}
	if len(nums) != 1 {
		t.Errorf("Descriptor() index length = %d, want 1", len(nums))
	}
}

func TestInfoGetters(t *testing.T) {
	t.Parallel()

	info := &Info{
		Version:    "1.0.0",
		BuildDate:  "2024-01-01",
		Details:    "Test details",
		CommitHash: "abc123",
		CommitDate: "2024-01-01T00:00:00Z",
	}

	if info.GetVersion() != "1.0.0" {
		t.Errorf("GetVersion() = %q, want '1.0.0'", info.GetVersion())
	}
	if info.GetBuildDate() != "2024-01-01" {
		t.Errorf("GetBuildDate() = %q, want '2024-01-01'", info.GetBuildDate())
	}
	if info.GetDetails() != "Test details" {
		t.Errorf("GetDetails() = %q, want 'Test details'", info.GetDetails())
	}
	if info.GetCommitHash() != "abc123" {
		t.Errorf("GetCommitHash() = %q, want 'abc123'", info.GetCommitHash())
	}
	if info.GetCommitDate() != "2024-01-01T00:00:00Z" {
		t.Errorf("GetCommitDate() = %q, want '2024-01-01T00:00:00Z'", info.GetCommitDate())
	}
}

func TestInfoGettersNil(t *testing.T) {
	t.Parallel()

	var info *Info

	if info.GetVersion() != "" {
		t.Errorf("GetVersion() on nil = %q, want ''", info.GetVersion())
	}
	if info.GetBuildDate() != "" {
		t.Errorf("GetBuildDate() on nil = %q, want ''", info.GetBuildDate())
	}
	if info.GetDetails() != "" {
		t.Errorf("GetDetails() on nil = %q, want ''", info.GetDetails())
	}
	if info.GetCommitHash() != "" {
		t.Errorf("GetCommitHash() on nil = %q, want ''", info.GetCommitHash())
	}
	if info.GetCommitDate() != "" {
		t.Errorf("GetCommitDate() on nil = %q, want ''", info.GetCommitDate())
	}
}

func TestInfoProtoMethods(t *testing.T) {
	t.Parallel()

	info := &Info{
		Version:   "2.0.0",
		BuildDate: "2024-06-01",
	}

	// Test String
	str := info.String()
	if str == "" {
		t.Error("String() should not be empty")
	}

	// Test Reset
	info.Reset()
	if info.GetVersion() != "" || info.GetBuildDate() != "" {
		t.Error("Reset() should zero all fields")
	}

	// Test ProtoMessage
	info.ProtoMessage()

	// Test ProtoReflect
	pr := info.ProtoReflect()
	if pr == nil {
		t.Error("ProtoReflect() should not return nil")
	}

	// Test Descriptor
	desc, nums := info.Descriptor()
	if len(desc) == 0 {
		t.Error("Descriptor() should return non-empty bytes")
	}
	if len(nums) != 1 {
		t.Errorf("Descriptor() index length = %d, want 1", len(nums))
	}
}
