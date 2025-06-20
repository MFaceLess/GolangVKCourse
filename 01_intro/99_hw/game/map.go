package main

var gameMap Map

type Map struct {
	Rooms map[string]*Room
}

func (gameMap *Map) ClearMap() {
	gameMap.Rooms = make(map[string]*Room)
}

func (gameMap *Map) AddRoom(r *Room) {
	gameMap.Rooms[r.Name] = r
}
