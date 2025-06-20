package main

type Item struct {
	Name         string
	CanApply     map[string]string
	ItemAction   string
	ActionResult string
	ItemPosition string

	UseItem func(*User)
}
