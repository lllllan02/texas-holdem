package texas

// SidePot 边池
// 当有玩家 All-in 时，如果他的筹码不足以跟注当前最高下注额，奖池会发生拆分
type SidePot struct {
	PotNumber int      // 奖池编号（1通常为主池，2,3为边池）
	Amount    int      // 该池内的筹码总额
	Winners   []string // 赢家ID列表 (单赢家时数组长度为1，平分时长度>1)
	Players   []string // 有资格竞争该奖池的玩家 ID 列表
	HandRank  HandRank // 赢的牌型等级（如 1 代表 One Pair）
	IsMainPot bool     // true 表示主池，false 表示边池
}
