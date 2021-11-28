package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Job struct{
	Name string
	Delay time.Duration
	Number int
}

type Worker struct{
	Id int
	JobQueue chan Job
	WorkerPool chan chan Job
	QuitChan chan bool
}

type Dispatcher struct{
	WorkerPool chan chan Job
	MaxWorkers int
	JobQueue chan Job
}


func NewWorker(id int, workerPool chan chan Job ) *Worker{
	return &Worker{
		Id: id,
		JobQueue: make(chan Job),
		WorkerPool: workerPool,
		QuitChan: make(chan bool),
	}
}

func(w Worker) Start(){
	go func(){
		for{
			w.WorkerPool <- w.JobQueue
			select {
			case job := <-w.JobQueue:
				fmt.Printf("Worker %d initated", w.Id)
				fib := Fibonnacci(job.Number)
				time.Sleep(job.Delay)
				fmt.Printf("Worker %d terminated with a result %d\n", w.Id, fib)
			case <- w.QuitChan:
				fmt.Printf("Worker with id %d stopped \n", w.Id)
			}
		}
	}()
}

func(w Worker) Stop(){
	go func(){
		w.QuitChan <- true
	}()
}

func Fibonnacci(n int) int{
	if n <= 1{
		return n
	}
	return Fibonnacci(n-1) + Fibonnacci(n-2)
}


func NewDispatcher(jobQueue chan Job, maxWorkers int) *Dispatcher{
	worker:= make(chan chan Job, maxWorkers)
	return &Dispatcher{
		JobQueue: jobQueue,
		MaxWorkers: maxWorkers,
		WorkerPool: worker,
	}
}

func (d *Dispatcher) Dispatch(){
	for{
		select{
		case job := <-d.JobQueue:
			go func() {
				workerJobQueue := <-d.WorkerPool
				workerJobQueue <- job
			}()
		}
	}
}

func (d *Dispatcher) Run(){
	for i:=0 ; i< d.MaxWorkers; i++{
		worker := NewWorker(i,d.WorkerPool)
		worker.Start()
	}
	go d.Dispatch()
}

func RequestHandler(w http.ResponseWriter, r *http.Request, jobQueue chan Job){
	if r.Method != "POST"{
		w.Header().Set("Allow","POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}


	delay, err := time.ParseDuration(r.FormValue("delay"))

	if err != nil{
		http.Error(w,"Invalid Delay",http.StatusBadRequest)
		return
	}

	value, err := strconv.Atoi(r.FormValue("value"))
	if err != nil{
		http.Error(w,"Invalid Value",http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == ""{
		http.Error(w, "Invalid Name", http.StatusBadRequest)
	}

	job := Job{
		Name: name,
		Delay: delay,
		Number: value,
	}

	jobQueue <-job
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	var test int
	const(
		maxWorkers = 4
		maxQueueSize = 20
		port = ":8081"
	)
	jobQueue := make(chan Job, maxQueueSize)
	dispatcher := NewDispatcher(jobQueue, maxWorkers)

	dispatcher.Run()

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request){
		RequestHandler(w,r,jobQueue)
	})
	log.Fatal(http.ListenAndServe(port,nil))

}
