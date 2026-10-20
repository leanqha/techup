-- +goose Up
-- +goose StatementBegin
-- Ячейки конструктора расписания: повторяющееся занятие на чётной/нечётной
-- неделе в конкретный день недели и время. Это шаблон на две недели, из
-- которого на весь семестр генерируются конкретные пары (таблица lessons).
CREATE TABLE IF NOT EXISTS schedule_slots (
    id SERIAL PRIMARY KEY,
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    semester INT NOT NULL,
    week_parity TEXT NOT NULL CHECK (week_parity IN ('odd', 'even')), -- нечётная / чётная
    weekday INT NOT NULL CHECK (weekday BETWEEN 1 AND 7),             -- 1=Пн .. 7=Вс
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'Другое',
    teacher_id INT REFERENCES accounts(id) ON DELETE SET NULL,
    classroom TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_schedule_slots_group_sem
    ON schedule_slots (group_id, semester);

CREATE INDEX IF NOT EXISTS idx_schedule_slots_teacher
    ON schedule_slots (teacher_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedule_slots;
-- +goose StatementEnd
