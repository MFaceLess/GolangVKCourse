package main

import (
	"sync"
)

type TaskCounter struct {
	TasksCounter int
	*sync.RWMutex
}

type UserChats struct {
	UserChats map[string]int64
	*sync.RWMutex
}

type TasksInfo struct {
	TasksNames    map[int]string
	TasksAssignee map[int]string
	TasksCreators map[int]string
	*sync.RWMutex
}

var usersChats = UserChats{UserChats: make(map[string]int64), RWMutex: &sync.RWMutex{}}
var tasksInfo = TasksInfo{TasksNames: make(map[int]string), TasksAssignee: make(map[int]string), TasksCreators: make(map[int]string), RWMutex: &sync.RWMutex{}}
var taskCounter = TaskCounter{RWMutex: &sync.RWMutex{}}
