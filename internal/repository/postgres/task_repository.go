package postgres

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
    const query = `
        INSERT INTO tasks (title, description, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, title, description, status, created_at, updated_at
    `

    row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
    created, err := scanTask(row)
    if err != nil {
        return nil, err
    }

    if task.Recurrence != nil {
        const recQuery = `
            INSERT INTO task_recurrences (task_id, type, value, interval)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `
        err = r.pool.QueryRow(ctx, recQuery, created.ID, task.Recurrence.Type, task.Recurrence.Value, task.Recurrence.Interval).Scan(&task.Recurrence.ID)
        if err != nil {
            return nil, err
        }
        created.Recurrence = task.Recurrence
        created.Recurrence.TaskID = int(created.ID)
    }

    return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
    const query = `
        SELECT 
            t.id, t.title, t.description, t.status, t.created_at, t.updated_at,
            tr.id, tr.type, tr.value, tr.interval
        FROM tasks t
        LEFT JOIN task_recurrences tr ON t.id = tr.task_id
        WHERE t.id = $1
    `

    row := r.pool.QueryRow(ctx, query, id)
    found, err := scanTaskWithRecurrence(row)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("not found")
        }
        return nil, err
    }

    return found, nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
    const query = `
        SELECT 
            t.id, t.title, t.description, t.status, t.created_at, t.updated_at,
            tr.id, tr.type, tr.value, tr.interval
        FROM tasks t
        LEFT JOIN task_recurrences tr ON t.id = tr.task_id
        ORDER BY t.id DESC
    `

    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    tasks := make([]taskdomain.Task, 0)
    for rows.Next() {
        task, err := scanTaskWithRecurrence(rows)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, *task)
    }

    return tasks, nil
}

type taskScanner interface {
    Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
    var (
        task   taskdomain.Task
        status string
    )
    if err := scanner.Scan(&task.ID, &task.Title, &task.Description, &status, &task.CreatedAt, &task.UpdatedAt); err != nil {
        return nil, err
    }
    task.Status = taskdomain.Status(status)
    return &task, nil
}

func scanTaskWithRecurrence(scanner taskScanner) (*taskdomain.Task, error) {
    var (
        task        taskdomain.Task
        status      string
        recID       *int
        recType     *string
        recValue    *string
        recInterval *int
    )

    err := scanner.Scan(
        &task.ID, &task.Title, &task.Description, &status, &task.CreatedAt, &task.UpdatedAt,
        &recID, &recType, &recValue, &recInterval,
    )
    if err != nil {
        return nil, err
    }

    task.Status = taskdomain.Status(status)

    if recID != nil {
        task.Recurrence = &taskdomain.TaskRecurrence{
            ID:       *recID,
            TaskID:   int(task.ID),
            Type:     *recType,
            Value:    *recValue,
            Interval: *recInterval,
        }
    }

    return &task, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
    const query = `UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=$4 WHERE id=$5 RETURNING id, title, description, status, created_at, updated_at`
    row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
    return scanTask(row)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
    const query = `DELETE FROM tasks WHERE id = $1`
    _, err := r.pool.Exec(ctx, query, id)
    return err
}
