// MIT License
//
// (C) Copyright [2018, 2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package base

import (
	"log"
)

///////////////////////////////////////////////////////////////////////////////
// Job interface
///////////////////////////////////////////////////////////////////////////////

type JobType int
type JobStatus int

const (
	JSTAT_DEFAULT    JobStatus = 0
	JSTAT_QUEUED     JobStatus = 1
	JSTAT_PROCESSING JobStatus = 2
	JSTAT_COMPLETE   JobStatus = 3
	JSTAT_CANCELLED  JobStatus = 4
	JSTAT_ERROR      JobStatus = 5
	JSTAT_MAX        JobStatus = 6
)

var JStatString = map[JobStatus]string{
	JSTAT_DEFAULT:    "JSTAT_DEFAULT",
	JSTAT_QUEUED:     "JSTAT_QUEUED",
	JSTAT_PROCESSING: "JSTAT_PROCESSING",
	JSTAT_COMPLETE:   "JSTAT_COMPLETE",
	JSTAT_CANCELLED:  "JSTAT_CANCELLED",
	JSTAT_ERROR:      "JSTAT_ERROR",
	JSTAT_MAX:        "JSTAT_MAX",
}

type Job interface {
	Log(format string, a ...interface{})
	//New(t JobType)
	Type() JobType
	//JobSCN() JobSCN
	Run()
	GetStatus() (JobStatus, error)
	SetStatus(JobStatus, error) (JobStatus, error)
	Cancel() JobStatus
}

///////////////////////////////////////////////////////////////////////////////
// Workers
///////////////////////////////////////////////////////////////////////////////
type Worker struct {
	WorkerPool  chan chan Job
	JobChannel  chan Job
	StopChannel chan bool
}

// Create a new worker
func NewWorker(workerPool chan chan Job) Worker {
	jChan := make(chan Job)
	stopChan := make(chan bool)
	return Worker{workerPool, jChan, stopChan}
}

// Start a worker to start consuming Jobs
func (w Worker) Start() {
	go func() {
		for {
			// Tell the dispatcher that this worker is available
			w.WorkerPool <- w.JobChannel

			select {
			case <-w.StopChannel:
				// Received a stop signal
				log.Print("Worker Stopping")
				return
			case job := <-w.JobChannel:
				// Received a job!
				job.Run()
				if status, _ := job.GetStatus(); status != JSTAT_ERROR {
					job.SetStatus(JSTAT_COMPLETE, nil)
				}
			}
		}
	}()
}

// Send as stop signal to the worker
func (w Worker) Stop() {
	go func() {
		w.StopChannel <- true
	}()
}

///////////////////////////////////////////////////////////////////////////////
// WorkerPool
///////////////////////////////////////////////////////////////////////////////
type WorkerPool struct {
	Workers     []Worker
	Pool        chan chan Job
	JobQueue    chan Job
	StopChannel chan bool
}

// Create a new pool of workers
func NewWorkerPool(maxWorkers, maxJobQueue int) *WorkerPool {
	workers := make([]Worker, maxWorkers)
	pool := make(chan chan Job, maxWorkers)
	jobQueue := make(chan Job, maxJobQueue)
	stopChan := make(chan bool)
	return &WorkerPool{workers, pool, jobQueue, stopChan}
}

// Starts all of the workers and the job dispatcher
func (p *WorkerPool) Run() {
	// Start the workers
	for i, _ := range p.Workers {
		worker := NewWorker(p.Pool)
		worker.Start()
		p.Workers[i] = worker
	}

	go p.dispatch()
}

// Hands out jobs to available workers
func (p *WorkerPool) dispatch() {
	for {
		// Wait for a job or a stop signal
		select {
		case <-p.StopChannel:
			break
		case job := <-p.JobQueue:
			// Wait for a free worker or a stop signal
			select {
			case <-p.StopChannel:
				break
			case jobChannel := <-p.Pool:
				if status, _ := job.GetStatus(); status != JSTAT_CANCELLED {
					job.SetStatus(JSTAT_PROCESSING, nil)
					// Send the job to the worker
					jobChannel <- job
				}
			}
		}
	}
	log.Print("Stopping Workers")
	// Send a stop signal to all of the workers
	for _, worker := range p.Workers {
		worker.Stop()
	}
	log.Print("Stopping Dispatcher")
}

// Queue a job. Returns 1 if the operation would
// block because the work queue is full.
func (p *WorkerPool) Queue(job Job) int {
	if job == nil {
		//Error
		return -1
	}
	select {
	case p.JobQueue <- job:
		job.SetStatus(JSTAT_QUEUED, nil)
		//Job queued
	default:
		//WOULDBLOCK
		return 1
	}
	return 0
}

// Command the dispatcher to stop itself and all workers
func (p *WorkerPool) Stop() {
	go func() {
		p.StopChannel <- true
	}()
}
