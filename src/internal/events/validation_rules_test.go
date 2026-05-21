package events

// validation_rules_test.go demonstrates the power of protovalidate CEL rules
// compared to plain OpenAPI / JSON-Schema field validation.
//
// Each test group is organised around a validation category:
//   1. Field-level built-in rules (string length, int range)
//   2. CEL regex rules (UUID format, character-set constraints)
//   3. Collection rules (max items, item constraints, uniqueness)
//   4. Cross-field / conditional CEL rules (the main protovalidate advantage)
//
// All tests that EXPECT validation failures call t.Error if the validator
// returns nil, showing that the domain refuses the payload at the earliest
// possible point in the request lifecycle.

import (
	"strings"
	"testing"

	datepb "google.golang.org/genproto/googleapis/type/date"

	"drblury/event-driven-service/internal/domain"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func newValidator(t *testing.T) *Validator {
	t.Helper()
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	return v
}

func assertValid(t *testing.T, v *Validator, msg any, description string) {
	t.Helper()
	if err := v.Validate(msg); err != nil {
		t.Errorf("expected valid (%s) but got error: %v", description, err)
	}
}

func assertInvalid(t *testing.T, v *Validator, msg any, wantFragment, description string) {
	t.Helper()
	err := v.Validate(msg)
	if err == nil {
		t.Errorf("expected validation error (%s) but got nil", description)
		return
	}
	if wantFragment != "" && !strings.Contains(err.Error(), wantFragment) {
		t.Errorf("expected error to contain %q (%s) but got: %v", wantFragment, description, err)
	}
}

// ─── ExampleRecord: field-level constraints ──────────────────────────────────

func TestValidation_ExampleRecord_ValidMinimal(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440000",
		Title:    "Min",
	}, "minimal valid record")
}

func TestValidation_ExampleRecord_TitleTooShort(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// Title shorter than 3 chars should fail string.min_len
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440001",
		Title:    "Hi",
	}, "value length must be at least 3", "title with 2 characters")
}

func TestValidation_ExampleRecord_TitleTooLong(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440002",
		Title:    strings.Repeat("x", 101),
	}, "value length must be at most 100", "title with 101 characters")
}

func TestValidation_ExampleRecord_TitleAtBoundaries(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// 3 chars is valid
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440003",
		Title:    "ABC",
	}, "title at 3 chars (lower boundary)")
	// 100 chars is valid
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440004",
		Title:    strings.Repeat("x", 100),
	}, "title at 100 chars (upper boundary)")
}

func TestValidation_ExampleRecord_DescriptionTooLong(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId:    "550e8400-e29b-41d4-a716-446655440005",
		Title:       "Test",
		Description: strings.Repeat("d", 501),
	}, "value length must be at most 500", "description exceeding 500 chars")
}

func TestValidation_ExampleRecord_DescriptionEmpty(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// Empty description is ignored (IGNORE_IF_ZERO_VALUE)
	assertValid(t, v, &domain.ExampleRecord{
		RecordId:    "550e8400-e29b-41d4-a716-446655440006",
		Title:       "No description",
		Description: "",
	}, "empty description (optional field)")
}

// ─── ExampleRecord: UUID CEL rule ────────────────────────────────────────────

// This CEL regex rule cannot be expressed as a JSON Schema / OpenAPI constraint
// on a plain string field — OpenAPI's `format: uuid` is advisory only and many
// validators do not enforce it. Protovalidate enforces it at the domain level,
// so every consumer (API, events, CLI) gets the same guarantee for free.

func TestValidation_ExampleRecord_InvalidUUID(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	invalidIDs := []struct {
		id   string
		desc string
	}{
		{"not-a-uuid", "plain string"},
		{"test-123", "hyphenated non-uuid"},
		{"550e8400e29b41d4a716446655440000", "uuid without hyphens"},
		{"550e8400-e29b-41d4-a716-44665544000Z", "uuid with invalid char"},
		{"", "empty string"},
	}

	for _, tc := range invalidIDs {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			assertInvalid(t, v, &domain.ExampleRecord{
				RecordId: tc.id,
				Title:    "Title",
			}, "record_id must be a valid UUID", tc.desc)
		})
	}
}

func TestValidation_ExampleRecord_ValidUUIDFormats(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	validIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"00000000-0000-0000-0000-000000000000",
		"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
	}

	for _, id := range validIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			assertValid(t, v, &domain.ExampleRecord{
				RecordId: id,
				Title:    "Title",
			}, "valid UUID format")
		})
	}
}

// ─── ExampleRecord: collection constraints ───────────────────────────────────

func TestValidation_ExampleRecord_TooManyTags(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	tags := make([]string, 11)
	for i := range tags {
		tags[i] = strings.Repeat("a", i+1) // unique, non-empty tags
	}
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440010",
		Title:    "Too many tags",
		Tags:     tags,
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "value must contain no more than 10 item", "11 tags (limit is 10)")
}

func TestValidation_ExampleRecord_TagAtMaxLength(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// A single tag exactly 30 chars is valid
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440011",
		Title:    "Tag length boundary",
		Tags:     []string{strings.Repeat("x", 30)},
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "tag at 30 chars (upper boundary)")
}

func TestValidation_ExampleRecord_TagTooLong(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440012",
		Title:    "Tag too long",
		Tags:     []string{strings.Repeat("x", 31)},
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "value length must be at most 30", "tag with 31 characters")
}

// DuplicateTags: the uniqueness check uses CEL `this.all(x, this.filter(y, y==x).size()==1)`
// — this is IMPOSSIBLE to express in OpenAPI JSON Schema without custom validators.
func TestValidation_ExampleRecord_DuplicateTags(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440013",
		Title:    "Duplicate tags",
		Tags:     []string{"alpha", "beta", "alpha"}, // "alpha" appears twice
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "tags must be unique", "duplicate tag 'alpha'")
}

func TestValidation_ExampleRecord_MaxUniqueTags(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// Exactly 10 unique tags is valid
	tags := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440014",
		Title:    "Max unique tags",
		Tags:     tags,
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "10 unique tags (exact limit)")
}

// ─── ExampleRecord: cross-field CEL ─────────────────────────────────────────

// This cross-field rule is the hallmark advantage of protovalidate over field-level
// JSON Schema validation: "if tags are present, meta.requested_by must be set".
// OpenAPI cannot express this relationship without custom x-extensions.

func TestValidation_ExampleRecord_TagsWithoutMeta(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440020",
		Title:    "Tags but no meta",
		Tags:     []string{"important"},
		// Meta is nil / empty — cross-field rule fires
	}, "meta.requested_by must be set", "tags without meta.requested_by")
}

func TestValidation_ExampleRecord_TagsWithMeta(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440021",
		Title:    "Tags with meta",
		Tags:     []string{"important"},
		Meta:     &domain.ExampleMeta{RequestedBy: "owner", Priority: 1},
	}, "tags with meta.requested_by set")
}

func TestValidation_ExampleRecord_EmptyTagsNoMeta(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// No tags → cross-field rule does not apply
	assertValid(t, v, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440022",
		Title:    "No tags, no meta",
	}, "empty tags list skips cross-field rule")
}

// ─── ExampleMeta: field-level constraints ────────────────────────────────────

func TestValidation_ExampleMeta_PriorityRange(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	tests := []struct {
		priority int32
		valid    bool
		desc     string
	}{
		{0, false, "priority below minimum (0)"},
		{1, true, "priority at minimum (1)"},
		{3, true, "priority mid-range (3)"},
		{5, true, "priority at maximum (5)"},
		{6, false, "priority above maximum (6)"},
		{100, false, "priority far above maximum (100)"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			meta := &domain.ExampleMeta{
				RequestedBy: "user",
				Priority:    tc.priority,
			}
			if tc.valid {
				assertValid(t, v, meta, tc.desc)
			} else {
				assertInvalid(t, v, meta, "value must be greater than or equal to 1 and less than or equal to 5", tc.desc)
			}
		})
	}
}

func TestValidation_ExampleMeta_RequestedByTooShort(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleMeta{
		RequestedBy: "x", // 1 char, minimum is 2
		Priority:    1,
	}, "value length must be at least 2", "requested_by with 1 character")
}

func TestValidation_ExampleMeta_RequestedByTooLong(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleMeta{
		RequestedBy: strings.Repeat("a", 65),
		Priority:    1,
	}, "value length must be at most 64", "requested_by with 65 characters")
}

// RequestedBy CEL regex: only alphanumeric, dot, underscore, hyphen.
// Again, this level of format enforcement is not possible in standard JSON Schema.
func TestValidation_ExampleMeta_RequestedByInvalidChars(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	invalidNames := []struct {
		name string
		desc string
	}{
		{"user name", "space in username"},
		{"user@domain", "@ character"},
		{"user/name", "slash character"},
		{"user name!", "exclamation mark"},
	}

	for _, tc := range invalidNames {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			assertInvalid(t, v, &domain.ExampleMeta{
				RequestedBy: tc.name,
				Priority:    1,
			}, "requested_by must contain only", tc.desc)
		})
	}
}

func TestValidation_ExampleMeta_RequestedByValidFormats(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	validNames := []string{
		"john",
		"john.doe",
		"john_doe",
		"john-doe",
		"service-account.v2",
		"CI_BOT",
	}

	for _, name := range validNames {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertValid(t, v, &domain.ExampleMeta{
				RequestedBy: name,
				Priority:    1,
			}, "valid requested_by format")
		})
	}
}

// ─── ExampleMeta: date CEL rule ──────────────────────────────────────────────

// The desired_start_date year constraint shows how protovalidate can validate
// nested message fields using CEL — the `has()` function checks presence and
// `.year` accesses the nested Date message field. This requires zero extra code
// in the application layer.

func TestValidation_ExampleMeta_DateYearTooEarly(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertInvalid(t, v, &domain.ExampleMeta{
		RequestedBy: "user",
		Priority:    1,
		DesiredStartDate: &datepb.Date{
			Year: 2019, Month: 12, Day: 31,
		},
	}, "desired_start_date year must be 2020", "start date year 2019")
}

func TestValidation_ExampleMeta_DateYearValid(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	assertValid(t, v, &domain.ExampleMeta{
		RequestedBy: "user",
		Priority:    1,
		DesiredStartDate: &datepb.Date{
			Year: 2020, Month: 1, Day: 1,
		},
	}, "start date year 2020 (boundary)")
}

func TestValidation_ExampleMeta_DateAbsent(t *testing.T) {
	t.Parallel()
	v := newValidator(t)
	// Absent date is not checked (has() returns false → short-circuit)
	assertValid(t, v, &domain.ExampleMeta{
		RequestedBy:      "user",
		Priority:         1,
		DesiredStartDate: nil,
	}, "absent desired_start_date skips year check")
}

// ─── ExampleMeta: cross-field CEL ────────────────────────────────────────────

// This is the canonical example of a cross-field rule:
//   "if requires_follow_up is true, priority must be ≥ 3"
//
// This is a conditional requirement that involves two separate fields.
// OpenAPI 3.x cannot express this with standard schema keywords — you need
// to add custom middleware or validate in business logic code. With
// protovalidate the rule lives in the proto definition and is automatically
// enforced by every runtime that touches the message.

func TestValidation_ExampleMeta_FollowUpRequiresHighPriority(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	cases := []struct {
		followUp bool
		priority int32
		valid    bool
		desc     string
	}{
		{false, 1, true, "no follow-up, low priority — allowed"},
		{false, 5, true, "no follow-up, high priority — allowed"},
		{true, 3, true, "follow-up with priority=3 — minimum allowed"},
		{true, 4, true, "follow-up with priority=4 — allowed"},
		{true, 5, true, "follow-up with priority=5 — allowed"},
		{true, 1, false, "follow-up with priority=1 — rejected (< 3)"},
		{true, 2, false, "follow-up with priority=2 — rejected (< 3)"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			meta := &domain.ExampleMeta{
				RequestedBy:      "user",
				Priority:         tc.priority,
				RequiresFollowUp: tc.followUp,
			}
			if tc.valid {
				assertValid(t, v, meta, tc.desc)
			} else {
				assertInvalid(t, v, meta, "priority must be at least 3", tc.desc)
			}
		})
	}
}

// ─── ExampleResult: status CEL enum-like rule ────────────────────────────────

// The `in` operator provides an enum-like constraint on a plain string field.
// This is more flexible than a proto enum (easy to add values without breaking
// wire compatibility) while still being strictly validated at the domain layer.

func TestValidation_ExampleResult_ValidStatuses(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	validCases := []struct {
		status string
		note   string
	}{
		{"pending", ""},
		{"processing", ""},
		{"completed", ""},
		// "failed" requires a non-empty note (cross-field rule)
		{"failed", "timeout connecting to upstream service"},
	}
	for _, tc := range validCases {
		tc := tc
		t.Run(tc.status, func(t *testing.T) {
			t.Parallel()
			assertValid(t, v, &domain.ExampleResult{
				RecordId: "550e8400-e29b-41d4-a716-446655440050",
				Status:   tc.status,
				Note:     tc.note,
			}, "valid status: "+tc.status)
		})
	}
}

func TestValidation_ExampleResult_InvalidStatus(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	invalidStatuses := []string{"unknown", "PENDING", "done", "error", ""}
	for _, status := range invalidStatuses {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			assertInvalid(t, v, &domain.ExampleResult{
				RecordId: "550e8400-e29b-41d4-a716-446655440051",
				Status:   status,
			}, "status must be one of", "invalid status: "+status)
		})
	}
}

// ─── ExampleResult: cross-field CEL ─────────────────────────────────────────

// Conditional requirement: when status is "failed", note must explain the reason.
// This is a classic cross-field validation that is cumbersome to implement in
// OpenAPI and must otherwise live in handler / service code. With protovalidate
// it is a single CEL expression in the proto definition.

func TestValidation_ExampleResult_FailedRequiresNote(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	// failed + empty note → violation
	assertInvalid(t, v, &domain.ExampleResult{
		RecordId: "550e8400-e29b-41d4-a716-446655440060",
		Status:   "failed",
		Note:     "",
	}, "note is required when status is 'failed'", "failed without note")
}

func TestValidation_ExampleResult_FailedWithNote(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	assertValid(t, v, &domain.ExampleResult{
		RecordId: "550e8400-e29b-41d4-a716-446655440061",
		Status:   "failed",
		Note:     "database connection timeout",
	}, "failed with explanation note")
}

func TestValidation_ExampleResult_NonFailedEmptyNote(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	for _, status := range []string{"pending", "processing", "completed"} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			// non-failed statuses do not require a note
			assertValid(t, v, &domain.ExampleResult{
				RecordId: "550e8400-e29b-41d4-a716-446655440062",
				Status:   status,
				Note:     "",
			}, status+" without note is allowed")
		})
	}
}

func TestValidation_ExampleResult_NoteTooLong(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	assertInvalid(t, v, &domain.ExampleResult{
		RecordId: "550e8400-e29b-41d4-a716-446655440063",
		Status:   "completed",
		Note:     strings.Repeat("x", 1001),
	}, "value length must be at most 1000", "note exceeding 1000 chars")
}

// ─── Multiple violations in one message ──────────────────────────────────────

// Protovalidate collects ALL violations before returning so callers get a
// complete picture, not just the first failure.

func TestValidation_ExampleRecord_MultipleViolations(t *testing.T) {
	t.Parallel()
	v := newValidator(t)

	err := v.Validate(&domain.ExampleRecord{
		RecordId:    "not-a-uuid",        // violates UUID CEL
		Title:       "Hi",                // violates min_len=3
		Description: strings.Repeat("d", 501), // violates max_len=500
	})
	if err == nil {
		t.Fatal("expected multiple validation errors but got nil")
	}

	// Confirm the error message mentions multiple numbered violations
	errStr := err.Error()
	if !strings.Contains(errStr, "1:") || !strings.Contains(errStr, "2:") {
		t.Errorf("expected at least two numbered violations, got: %v", errStr)
	}
}
