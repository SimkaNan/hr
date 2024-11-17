package service

type Service struct {
	TaskWorker
}

type TaskWorker interface {
	Start(t int)
}

func New() *Service {
	return &Service{
		TaskWorker: NewTaskProcessing(),
	}
}
