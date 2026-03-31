package machine

import (
	"fmt"
	"testing"
)

// TestTableFlow_HappyPath 模拟一局正常的德州扑克游戏流程
func TestTableFlow_HappyPath(t *testing.T) {
	// 1. 初始化桌子和玩家
	table := &Table{
		SmallBlind: 10,
		BigBlind:   20,
		Players: []*Player{
			{ID: "P1", Name: "Alice", Chips: 1000},
			{ID: "P2", Name: "Bob", Chips: 1000},
			{ID: "P3", Name: "Charlie", Chips: 1000},
		},
	}

	fmt.Println(">>> 1. 游戏开始")
	table.StartHand()

	// 模拟 PreFlop 阶段的下注
	// 假设 P1 是 Button, P2 是 SB, P3 是 BB
	// 那么第一个说话的是 P1 (UTG)
	fmt.Println("\n>>> 2. PreFlop 阶段下注")
	table.CurrentTurn = 0 // 强制设置为 P1

	// P1 Call 20
	table.HandleAction(PlayerAction{PlayerID: "P1", Action: ActionCall, Amount: 20})

	// P2 (SB) Call 20 (补齐 10)
	table.CurrentTurn = 1
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionCall, Amount: 20})

	// P3 (BB) Check 20
	table.CurrentTurn = 2
	// 假设这里 isRoundOver 返回 true，触发 NextStage 进入 Flop
	table.HandleAction(PlayerAction{PlayerID: "P3", Action: ActionCheck, Amount: 20})

	// 模拟 Flop 阶段的下注
	fmt.Println("\n>>> 3. Flop 阶段下注")
	// Flop 阶段由 SB (P2) 先说话
	table.CurrentTurn = 1

	// P2 Check
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionCheck, Amount: 0})

	// P3 Bet 50
	table.CurrentTurn = 2
	table.HandleAction(PlayerAction{PlayerID: "P3", Action: ActionBet, Amount: 50})

	// P1 Fold
	table.CurrentTurn = 0
	table.HandleAction(PlayerAction{PlayerID: "P1", Action: ActionFold, Amount: 0})

	// P2 Call 50
	table.CurrentTurn = 1
	// 假设这里 isRoundOver 返回 true，触发 NextStage 进入 Turn
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionCall, Amount: 50})

	// 模拟 Turn 阶段
	fmt.Println("\n>>> 4. Turn 阶段")
	table.CurrentTurn = 1
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionCheck, Amount: 0})
	table.CurrentTurn = 2
	table.HandleAction(PlayerAction{PlayerID: "P3", Action: ActionCheck, Amount: 0})

	// 模拟 River 阶段
	fmt.Println("\n>>> 5. River 阶段")
	table.CurrentTurn = 1
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionCheck, Amount: 0})
	table.CurrentTurn = 2
	// 假设这是最后一次操作，触发 Showdown
	table.HandleAction(PlayerAction{PlayerID: "P3", Action: ActionCheck, Amount: 0})

	fmt.Println("\n>>> 游戏模拟结束")
}

// TestTableFlow_EarlyEnd 模拟有人加注，其他人全弃牌导致提前结束的流程
func TestTableFlow_EarlyEnd(t *testing.T) {
	table := &Table{
		SmallBlind: 10,
		BigBlind:   20,
		Players: []*Player{
			{ID: "P1", Name: "Alice", Chips: 1000},
			{ID: "P2", Name: "Bob", Chips: 1000},
			{ID: "P3", Name: "Charlie", Chips: 1000},
		},
	}

	fmt.Println(">>> 1. 游戏开始")
	table.StartHand()

	fmt.Println("\n>>> 2. PreFlop 阶段下注")
	// P1 Raise 到 100
	table.CurrentTurn = 0
	table.HandleAction(PlayerAction{PlayerID: "P1", Action: ActionRaise, Amount: 100})

	// P2 Fold
	table.CurrentTurn = 1
	table.HandleAction(PlayerAction{PlayerID: "P2", Action: ActionFold, Amount: 0})

	// P3 Fold，触发提前结束
	table.CurrentTurn = 2
	// 假设这里触发了 ActivePlayersCount == 1 的条件，调用 EndHandEarly
	table.EndHandEarly()
}
