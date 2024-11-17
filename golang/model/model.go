package model

type Task struct {
	ID         int64
	CreateTime string // время создания
	FinishTime string // время выполнения
	Err        error
}
