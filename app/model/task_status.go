package model

import "database/sql/driver"

type TaskStatus string

const (
	QUEUED     TaskStatus = "QUEUED"
	SUCCESS    TaskStatus = "SUCCESS"
	INPROGRESS TaskStatus = "INPROGRESS"
	FAILED     TaskStatus = "FAILED"
	INITIAL    TaskStatus = "INITIAL"
)

func (ts *TaskStatus) Scan(value interface{}) error {
	*ts = TaskStatus(value.(string))
	return nil
}

func (ts TaskStatus) Value() (driver.Value, error) {
	return string(ts), nil
}
