package texas

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lllllan02/texas-holdem/pkg/user"
)

// mockMessenger 模拟消息发送和定时器
type mockMessenger struct{}

func (m *mockMessenger) Broadcast(msgType string, action string, payload interface{}) {
	// fmt.Printf("[Broadcast] %s: %+v\n", action, payload)
}

func (m *mockMessenger) SendTo(userID string, msgType string, action string, payload interface{}) {
	// fmt.Printf("[SendTo %s] %s: %+v\n", userID, action, payload)
}

func (m *mockMessenger) Execute(task func()) {
	task()
}

func TestGameFlow_HappyPath(t *testing.T) {
	// 1. 初始化桌子
	table := NewTable()
	options := `{"player_count": 3, "small_blind": 10, "big_blind": 20, "initial_chips": 1000}`
	err := table.OnInit(&mockMessenger{}, []byte(options))
	if err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	// 2. 玩家加入并坐下
	u1 := &user.User{ID: "u1", Nickname: "Alice"}
	u2 := &user.User{ID: "u2", Nickname: "Bob"}
	u3 := &user.User{ID: "u3", Nickname: "Charlie"}

	table.OnPlayerJoin(u1)
	table.OnPlayerJoin(u2)
	table.OnPlayerJoin(u3)

	sitPayload1, _ := json.Marshal(map[string]interface{}{"seat_number": 0})
	table.HandleMessage(u1.ID, MsgTypeSitDown, sitPayload1)

	sitPayload2, _ := json.Marshal(map[string]interface{}{"seat_number": 1})
	table.HandleMessage(u2.ID, MsgTypeSitDown, sitPayload2)

	sitPayload3, _ := json.Marshal(map[string]interface{}{"seat_number": 2})
	table.HandleMessage(u3.ID, MsgTypeSitDown, sitPayload3)

	// 3. 玩家准备
	table.HandleMessage(u1.ID, MsgTypeReady, nil)
	table.HandleMessage(u2.ID, MsgTypeReady, nil)
	table.HandleMessage(u3.ID, MsgTypeReady, nil)

	// 此时游戏应该已经自动开始了 (因为 checkAndAutoStart)
	// 或者倒计时中。我们等待一下倒计时（如果有的话）
	// 倒计时是 3 秒
	time.Sleep(3500 * time.Millisecond)

	if table.CurrentHand == nil {
		t.Fatalf("Game did not start automatically")
	}

	// 此时应该是 PreFlop 阶段
	if table.CurrentHand.Stage != HandStagePreFlop {
		t.Fatalf("Expected PreFlop stage, got %v", table.CurrentHand.Stage)
	}

	// 打印当前行动玩家
	currIdx := table.CurrentHand.CurrentPlayerIndex
	currPlayer := table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)

	// 模拟玩家操作
	// 打印一下当前的行动顺序
	fmt.Printf("ActionOrder: %v\n", table.CurrentHand.ActionOrder)

	// 让当前玩家 Call
	actionPayloadCall, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCall, "amount": 20})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCall)

	// 下一个玩家
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCall)

	// 下一个玩家
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadCheck, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCheck, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)

	// 此时应该进入 Flop 阶段
	if table.CurrentHand.Stage != HandStageFlop {
		t.Fatalf("Expected Flop stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 2. 进入 Flop 阶段")
	fmt.Printf("公共牌: %v\n", table.CurrentHand.BoardCards)

	// Flop 阶段，大家都 Check
	for i := 0; i < 3; i++ {
		currIdx = table.CurrentHand.CurrentPlayerIndex
		currPlayer = table.Seats[currIdx]
		table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)
	}

	// 此时应该进入 Turn 阶段
	if table.CurrentHand.Stage != HandStageTurn {
		t.Fatalf("Expected Turn stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 3. 进入 Turn 阶段")
	fmt.Printf("公共牌: %v\n", table.CurrentHand.BoardCards)

	// Turn 阶段，大家都 Check
	for i := 0; i < 3; i++ {
		currIdx = table.CurrentHand.CurrentPlayerIndex
		currPlayer = table.Seats[currIdx]
		table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)
	}

	// 此时应该进入 River 阶段
	if table.CurrentHand.Stage != HandStageRiver {
		t.Fatalf("Expected River stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 4. 进入 River 阶段")
	fmt.Printf("公共牌: %v\n", table.CurrentHand.BoardCards)

	// River 阶段，大家都 Check
	for i := 0; i < 3; i++ {
		currIdx = table.CurrentHand.CurrentPlayerIndex
		currPlayer = table.Seats[currIdx]
		table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)
	}

	// 此时应该进入 Showdown 阶段，并且游戏结束回到 Waiting 状态
	// 因为 Showdown 后会重置状态
	// 但由于使用了 timer，可能需要等一下或者手动触发
	time.Sleep(3 * time.Second) // showdown timer is 2s

	fmt.Println(">>> 5. 游戏结束")
	for _, p := range table.Seats {
		if p != nil {
			fmt.Printf("玩家 %s 筹码: %d\n", p.User.Nickname, p.Chips)
		}
	}
}

func TestGameFlow_EarlyEnd(t *testing.T) {
	// 1. 初始化桌子
	table := NewTable()
	options := `{"player_count": 3, "small_blind": 10, "big_blind": 20, "initial_chips": 1000}`
	err := table.OnInit(&mockMessenger{}, []byte(options))
	if err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	// 2. 玩家加入并坐下
	u1 := &user.User{ID: "u1", Nickname: "Alice"}
	u2 := &user.User{ID: "u2", Nickname: "Bob"}
	u3 := &user.User{ID: "u3", Nickname: "Charlie"}

	table.OnPlayerJoin(u1)
	table.OnPlayerJoin(u2)
	table.OnPlayerJoin(u3)

	sitPayload1, _ := json.Marshal(map[string]interface{}{"seat_number": 0})
	table.HandleMessage(u1.ID, MsgTypeSitDown, sitPayload1)

	sitPayload2, _ := json.Marshal(map[string]interface{}{"seat_number": 1})
	table.HandleMessage(u2.ID, MsgTypeSitDown, sitPayload2)

	sitPayload3, _ := json.Marshal(map[string]interface{}{"seat_number": 2})
	table.HandleMessage(u3.ID, MsgTypeSitDown, sitPayload3)

	// 3. 玩家准备
	table.HandleMessage(u1.ID, MsgTypeReady, nil)
	table.HandleMessage(u2.ID, MsgTypeReady, nil)
	table.HandleMessage(u3.ID, MsgTypeReady, nil)

	time.Sleep(3500 * time.Millisecond)

	if table.CurrentHand == nil {
		t.Fatalf("Game did not start automatically")
	}

	// 此时应该是 PreFlop 阶段
	if table.CurrentHand.Stage != HandStagePreFlop {
		t.Fatalf("Expected PreFlop stage, got %v", table.CurrentHand.Stage)
	}

	fmt.Println(">>> 1. 游戏开始 (PreFlop)")

	// 让当前玩家 Raise
	currIdx := table.CurrentHand.CurrentPlayerIndex
	currPlayer := table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadRaise, _ := json.Marshal(map[string]interface{}{"action": ActionTypeRaise, "amount": 100})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadRaise)

	// 下一个玩家 Fold
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadFold, _ := json.Marshal(map[string]interface{}{"action": ActionTypeFold, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadFold)

	// 最后一个玩家 Fold
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadFold)

	// 此时游戏应该已经提前结束，因为只剩下一个未弃牌玩家
	time.Sleep(3 * time.Second) // showdown timer is 2s

	fmt.Println(">>> 2. 游戏提前结束")
	for _, p := range table.Seats {
		if p != nil {
			fmt.Printf("玩家 %s 筹码: %d\n", p.User.Nickname, p.Chips)
		}
	}
}

func TestGameFlow_EarlyEnd_OnFlop(t *testing.T) {
	// 1. 初始化桌子
	table := NewTable()
	options := `{"player_count": 3, "small_blind": 10, "big_blind": 20, "initial_chips": 1000}`
	err := table.OnInit(&mockMessenger{}, []byte(options))
	if err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	// 2. 玩家加入并坐下
	u1 := &user.User{ID: "u1", Nickname: "Alice"}
	u2 := &user.User{ID: "u2", Nickname: "Bob"}
	u3 := &user.User{ID: "u3", Nickname: "Charlie"}

	table.OnPlayerJoin(u1)
	table.OnPlayerJoin(u2)
	table.OnPlayerJoin(u3)

	sitPayload1, _ := json.Marshal(map[string]interface{}{"seat_number": 0})
	table.HandleMessage(u1.ID, MsgTypeSitDown, sitPayload1)

	sitPayload2, _ := json.Marshal(map[string]interface{}{"seat_number": 1})
	table.HandleMessage(u2.ID, MsgTypeSitDown, sitPayload2)

	sitPayload3, _ := json.Marshal(map[string]interface{}{"seat_number": 2})
	table.HandleMessage(u3.ID, MsgTypeSitDown, sitPayload3)

	// 3. 玩家准备
	table.HandleMessage(u1.ID, MsgTypeReady, nil)
	table.HandleMessage(u2.ID, MsgTypeReady, nil)
	table.HandleMessage(u3.ID, MsgTypeReady, nil)

	time.Sleep(3500 * time.Millisecond)

	if table.CurrentHand == nil {
		t.Fatalf("Game did not start automatically")
	}

	fmt.Println(">>> 1. 游戏开始 (PreFlop)")

	// 此时应该是 PreFlop 阶段
	// 假设行动顺序：Alice (UTG/BTN) -> Bob (SB) -> Charlie (BB)

	// Alice Call 20
	currIdx := table.CurrentHand.CurrentPlayerIndex
	currPlayer := table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadCall, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCall, "amount": 20})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCall)

	// Bob Call 20 (补齐 10)
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCall)

	// Charlie Check
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadCheck, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCheck, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)

	// 进入 Flop
	if table.CurrentHand.Stage != HandStageFlop {
		t.Fatalf("Expected Flop stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 2. 进入 Flop 阶段")
	fmt.Printf("公共牌: %v\n", table.CurrentHand.BoardCards)

	// Flop 阶段，从 SB (Bob) 开始说话
	// Bob Bet 50
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadBet, _ := json.Marshal(map[string]interface{}{"action": ActionTypeBet, "amount": 50})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadBet)

	// Charlie Fold
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	actionPayloadFold, _ := json.Marshal(map[string]interface{}{"action": ActionTypeFold, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadFold)

	// Alice Fold
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s\n", currPlayer.User.Nickname)
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadFold)

	// 此时游戏应该已经提前结束，Bob 赢下底池
	time.Sleep(3 * time.Second) // showdown timer is 2s

	fmt.Println(">>> 3. 游戏在 Flop 阶段提前结束")
	for _, p := range table.Seats {
		if p != nil {
			fmt.Printf("玩家 %s 筹码: %d\n", p.User.Nickname, p.Chips)
		}
	}
}

func TestGameFlow_ReRaise_And_Call(t *testing.T) {
	// 1. 初始化桌子
	table := NewTable()
	options := `{"player_count": 3, "small_blind": 10, "big_blind": 20, "initial_chips": 1000}`
	err := table.OnInit(&mockMessenger{}, []byte(options))
	if err != nil {
		t.Fatalf("OnInit failed: %v", err)
	}

	// 2. 玩家加入并坐下
	u1 := &user.User{ID: "u1", Nickname: "Alice"}
	u2 := &user.User{ID: "u2", Nickname: "Bob"}
	u3 := &user.User{ID: "u3", Nickname: "Charlie"}

	table.OnPlayerJoin(u1)
	table.OnPlayerJoin(u2)
	table.OnPlayerJoin(u3)

	sitPayload1, _ := json.Marshal(map[string]interface{}{"seat_number": 0})
	table.HandleMessage(u1.ID, MsgTypeSitDown, sitPayload1)

	sitPayload2, _ := json.Marshal(map[string]interface{}{"seat_number": 1})
	table.HandleMessage(u2.ID, MsgTypeSitDown, sitPayload2)

	sitPayload3, _ := json.Marshal(map[string]interface{}{"seat_number": 2})
	table.HandleMessage(u3.ID, MsgTypeSitDown, sitPayload3)

	// 3. 玩家准备
	table.HandleMessage(u1.ID, MsgTypeReady, nil)
	table.HandleMessage(u2.ID, MsgTypeReady, nil)
	table.HandleMessage(u3.ID, MsgTypeReady, nil)

	time.Sleep(3500 * time.Millisecond)

	if table.CurrentHand == nil {
		t.Fatalf("Game did not start automatically")
	}

	fmt.Println(">>> 1. 游戏开始 (PreFlop)")

	// 此时应该是 PreFlop 阶段
	// 假设行动顺序：Alice (UTG/BTN) -> Bob (SB) -> Charlie (BB)

	// Alice (BTN) Raise 到 60 (当前盲注是 20，加注到 60 是合法的)
	currIdx := table.CurrentHand.CurrentPlayerIndex
	currPlayer := table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Raise to 60\n", currPlayer.User.Nickname)
	actionPayloadRaise1, _ := json.Marshal(map[string]interface{}{"action": ActionTypeRaise, "amount": 60})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadRaise1)

	// Bob (SB) Re-raise 到 150 (3-bet)
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Raise to 150\n", currPlayer.User.Nickname)
	actionPayloadRaise2, _ := json.Marshal(map[string]interface{}{"action": ActionTypeRaise, "amount": 150})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadRaise2)

	// Charlie (BB) 觉得太贵了，Fold
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Fold\n", currPlayer.User.Nickname)
	actionPayloadFold, _ := json.Marshal(map[string]interface{}{"action": ActionTypeFold, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadFold)

	// 此时行动权应该回到 Alice，因为 Bob 加注了，Alice 需要补齐差额
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Call (补齐到 150)\n", currPlayer.User.Nickname)
	if currPlayer.User.ID != u1.ID {
		t.Fatalf("Expected Alice to act, got %s", currPlayer.User.Nickname)
	}
	actionPayloadCall, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCall, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCall)

	// 进入 Flop
	if table.CurrentHand.Stage != HandStageFlop {
		t.Fatalf("Expected Flop stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 2. 进入 Flop 阶段")
	fmt.Printf("公共牌: %v\n", table.CurrentHand.BoardCards)
	fmt.Printf("当前底池: %d\n", table.CurrentHand.Pot) // 应该是 150(Alice) + 150(Bob) + 20(Charlie) = 320

	// Flop 阶段，从 SB (Bob) 开始说话
	// Bob Check
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Check\n", currPlayer.User.Nickname)
	actionPayloadCheck, _ := json.Marshal(map[string]interface{}{"action": ActionTypeCheck, "amount": 0})
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)

	// Alice Check
	currIdx = table.CurrentHand.CurrentPlayerIndex
	currPlayer = table.Seats[currIdx]
	fmt.Printf("当前行动玩家: %s, 动作: Check\n", currPlayer.User.Nickname)
	table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)

	// 进入 Turn
	if table.CurrentHand.Stage != HandStageTurn {
		t.Fatalf("Expected Turn stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 3. 进入 Turn 阶段")

	// Turn 阶段，Bob Check, Alice Check
	for i := 0; i < 2; i++ {
		currIdx = table.CurrentHand.CurrentPlayerIndex
		currPlayer = table.Seats[currIdx]
		table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)
	}

	// 进入 River
	if table.CurrentHand.Stage != HandStageRiver {
		t.Fatalf("Expected River stage, got %v", table.CurrentHand.Stage)
	}
	fmt.Println(">>> 4. 进入 River 阶段")

	// River 阶段，Bob Check, Alice Check
	for i := 0; i < 2; i++ {
		currIdx = table.CurrentHand.CurrentPlayerIndex
		currPlayer = table.Seats[currIdx]
		table.HandleMessage(currPlayer.User.ID, MsgTypeAction, actionPayloadCheck)
	}

	// 此时应该进入 Showdown 阶段
	time.Sleep(3 * time.Second) // showdown timer is 2s

	fmt.Println(">>> 5. 游戏结束 (Showdown)")
	for _, p := range table.Seats {
		if p != nil {
			fmt.Printf("玩家 %s 筹码: %d\n", p.User.Nickname, p.Chips)
		}
	}
}
