package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"taskWorker/model"
	"time"
)

type TaskProcessing struct {
	jobs         chan model.Task
	result       chan model.Task
	DoneTask     []model.Task
	FailedTask   []model.Task
	Wg           *sync.WaitGroup
	Mu           *sync.RWMutex
	ResultString string
}

func NewTaskProcessing() *TaskProcessing {
	return &TaskProcessing{
		jobs:         make(chan model.Task, 1),
		result:       make(chan model.Task, 1),
		DoneTask:     make([]model.Task, 0),
		FailedTask:   make([]model.Task, 0),
		Wg:           &sync.WaitGroup{},
		Mu:           &sync.RWMutex{},
		ResultString: "",
	}
}

func (s *TaskProcessing) Start(t int) {
	s.Wg.Add(1)
	go s.Create(time.NewTicker(time.Duration(t) * time.Second))
	go s.Process()
	go s.Finish()
	go s.MakeList()
	go s.OutputResults()

	s.Wg.Wait()
}

func (s *TaskProcessing) Create(t *time.Ticker) {
	for {
		select {
		case <-t.C:
			close(s.jobs)
			close(s.result)
			s.Wg.Done()
			return
		default:
			var err error
			id := time.Now().UnixNano()
			if time.Now().UnixNano()%2 > 0 { // касательно данного условия, само задание выполнял на ubuntu(go 1.22.3 linux/amd64), сейчас проверяю на маке(go version go1.22.3 darwin/amd64)
				err = errors.New("Create task fail") // и появилась проблема что нано значения округляются, поэтому отмечаю что если ошибки не отображаются то строку необходимо заменить на: if time.Now().UnixMilli()%2 > 0 { 
			}
			cT := time.Unix(0, id).Format(time.RFC3339)
			s.jobs <- model.Task{ID: id, CreateTime: cT, Err: err}
		}
	}
}

func (s *TaskProcessing) Process() {
	for task := range s.jobs {
		time.Sleep(time.Millisecond * 150)

		_, err := time.Parse(time.RFC3339, task.CreateTime)
		if err != nil {
			task.FinishTime = time.Now().Format(time.RFC3339)
			log.Printf("Error Parse Create Time %s", task.CreateTime)
		} else {
			task.FinishTime = time.Now().Format(time.RFC3339)
		}

		s.result <- task
	}
}

func (s *TaskProcessing) Finish() {
	for task := range s.result {
		if task.Err == nil {
			s.Mu.Lock()
			s.DoneTask = append(s.DoneTask, task)
			s.Mu.Unlock()
		} else {
			s.Mu.Lock()
			s.FailedTask = append(s.FailedTask, task)
			s.Mu.Unlock()
		}
	}
}

func (s *TaskProcessing) MakeList() {
	var lastDone, lastFail int
	var done, fail strings.Builder
	var count int
	_, err := done.Write([]byte("Done tasks:\n"))
	if err != nil {
		log.Printf("Error on OutputProcessing task: %s; Error on init buffer: %s", done.String(), err)
	}

	_, err = fail.Write([]byte("Errors:\n"))
	if err != nil {
		log.Printf("Error on OutputProcessing task: %s; Error on init buffer: %s", done.String(), err)
	}

	for {
		if lastDone < len(s.DoneTask)-1 {
			for i := lastDone + 1; i < len(s.DoneTask); i++ {
				s.Mu.RLock()
				done.Write([]byte(fmt.Sprintf("Task with ID: %d start processing at: %s and finish at: %s\n",
					s.DoneTask[i].ID, s.DoneTask[i].CreateTime, s.DoneTask[i].FinishTime)))
				s.Mu.RUnlock()
			}
			lastDone = len(s.DoneTask) - 1
			count++
		}

		if lastFail < len(s.FailedTask)-1 {
			for i := lastFail + 1; i < len(s.FailedTask); i++ {
				s.Mu.RLock()
				fail.Write([]byte(fmt.Sprintf("Task with ID: %d start processing at: %s and finish at: %s with error: %s\n",
					s.FailedTask[i].ID, s.FailedTask[i].CreateTime, s.FailedTask[i].FinishTime, s.FailedTask[i].Err.Error())))
				s.Mu.RUnlock()
			}
			lastFail = len(s.FailedTask) - 1
			count++
		}

		if count > 0 {
			s.Mu.Lock()
			s.ResultString = done.String() + fail.String()
			s.Mu.Unlock()
		}
		count = 0

		time.Sleep(time.Second * 1)
	}
}

func (s *TaskProcessing) OutputResults() {
	for {
		s.Mu.RLock()
		fmt.Println(s.ResultString)
		s.Mu.RUnlock()

		time.Sleep(time.Second * 3)
	}
}
