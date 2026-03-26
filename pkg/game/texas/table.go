package texas

type Table struct {
	// 静态配置 (OnInit 时确定)
	MaxPlayers int
	SmallBlind int
	BigBlind   int
}
