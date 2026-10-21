-- +goose Up
-- +goose StatementBegin
-- Учебный план группы: какие предметы группа изучает в конкретном семестре,
-- с опциональным закреплённым преподавателем. Служит источником блоков для
-- конструктора расписания (левая колонка) и наполняет вкладку «Учебный план».
CREATE TABLE IF NOT EXISTS group_subjects (
    id SERIAL PRIMARY KEY,
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    subject_id INT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    semester INT NOT NULL,
    teacher_id INT REFERENCES accounts(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_subjects_unique
    ON group_subjects (group_id, subject_id, semester);

CREATE INDEX IF NOT EXISTS idx_group_subjects_group_sem
    ON group_subjects (group_id, semester);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_subjects;
-- +goose StatementEnd
