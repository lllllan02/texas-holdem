package texas

import "github.com/spf13/cast"

type Table struct {
	maxPlayers int

	seats []*Player

	players map[string]*Player
}

func NewTable(param map[string]any) *Table {
	maxPlayers := cast.ToInt(param["maxPlayers"])

	return &Table{
		maxPlayers: maxPlayers,
		seats:      make([]*Player, maxPlayers),
		players:    make(map[string]*Player),
	}
}
