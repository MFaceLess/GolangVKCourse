package main

import (
	"errors"
)

type Command struct {
	CommandName  string
	ArgsNum      int
	ErrorMessage string
}

func (command *Command) Error() string {
	return command.ErrorMessage
}

var (
	CommandLookAround = Command{CommandName: "осмотреться", ArgsNum: 0}
	CommandWalk       = Command{CommandName: "идти", ArgsNum: 1}
	CommandWear       = Command{CommandName: "надеть", ArgsNum: 1}
	CommandTake       = Command{CommandName: "взять", ArgsNum: 1}
	CommandApply      = Command{CommandName: "применить", ArgsNum: 2}

	UnknownCommand = Command{CommandName: "неизвестная команда", ArgsNum: 0}
)

var (
	errInvalidCommand = errors.New("ошибка формата команды")
	errItemNotInRoom  = errors.New("нет такого")
)
