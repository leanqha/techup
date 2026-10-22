-- +goose Up
-- +goose StatementBegin
-- Curriculum subjects without an assigned teacher used to be dropped when a
-- group's schedule was generated. Allow a lesson to exist without a teacher so
-- those pairs are kept instead of ignored. The FK still applies to non-NULL
-- values, so a teacher that is set must still reference an existing account.
ALTER TABLE lessons ALTER COLUMN teacher_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Re-add NOT NULL. Any teacher-less lessons must be removed first, otherwise the
-- constraint cannot be re-applied.
DELETE FROM lessons WHERE teacher_id IS NULL;
ALTER TABLE lessons ALTER COLUMN teacher_id SET NOT NULL;
-- +goose StatementEnd
