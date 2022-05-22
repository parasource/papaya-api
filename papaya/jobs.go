/*
 * Copyright 2022 LightSwitch.Digital
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package papaya

import (
	"github.com/go-co-op/gocron"
	"sync"
	"time"
)

type interval int

const (
	IntervalDaily = iota
	IntervalMonthly
	IntervalHourly
	IntervalEveryMinute
)

type Job struct {
	Name     string
	F        func()
	Interval interval
	Error    error
}

type JobsManager struct {
	sched *gocron.Scheduler

	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewJobsManager(jobs []*Job) (*JobsManager, error) {
	s := gocron.NewScheduler(time.UTC)

	j := &JobsManager{
		sched: s,
		jobs:  make(map[string]*Job),
	}

	for _, job := range jobs {
		err := j.AddJob(job)
		if err != nil {
			return nil, err
		}
	}

	j.sched.StartAsync()
	j.PreflightRunAllJobs()

	return j, nil
}

func (j *JobsManager) PreflightRunAllJobs() {
	for _, job := range j.jobs {
		job.F()
	}
}

func (j *JobsManager) AddJob(job *Job) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.jobs[job.Name]; ok {
		return nil
	}

	jb := j.sched.Every(1)
	switch job.Interval {
	case IntervalDaily:
		jb = jb.Day().At("00:00")
	case IntervalMonthly:
		jb = jb.Month().At("00:00")
	case IntervalHourly:
		jb = jb.Month().At("00:00")
	case IntervalEveryMinute:
		jb = jb.Minute()
	}

	_, err := jb.Do(job.F)
	if err != nil {
		return err
	}

	j.jobs[job.Name] = job

	return nil
}
