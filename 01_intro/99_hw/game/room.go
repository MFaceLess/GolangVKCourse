package main

type RoomTarget struct {
	Item           *Item
	IsNeededToFind bool
}
type Room struct {
	Name            string
	Description     string
	WalkDescription string

	Connections    []*Room
	ConnectionsSet map[*Room]bool // Учитывается возможность прохода из одной комнаты в другую

	Items         []*RoomTarget
	IsOutsideHome bool
}
