import { useState, useEffect, useRef } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { Header } from '../components/Header'
import { HistoryModal } from '../components/HistoryModal'
import { PlayingCard } from '../components/PlayingCard'
import { useUser } from '../hooks/useUser'
import { useWebSocket } from '../hooks/useWebSocket'
import { deleteRoom, getRoom } from '../api/room'
import type { StateUpdateSnapshot, CountdownPayload, TurnNotificationPayload, HoleCardsPayload, ShowdownSummary } from '../types/game'
import { getSeatPosition } from '../components/tableUtils'

interface GameLog {
  id: string;
  time: string;
  text: string;
  type: 'system' | 'action' | 'chat';
  senderId?: string;
  senderName?: string;
  senderAvatar?: string;
  timestamp?: number;
}

export default function Table() {
  const { user, loading } = useUser()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [showHistory, setShowHistory] = useState(false)
  const [isOwner, setIsOwner] = useState(false)
  const [validatedRoomNumber, setValidatedRoomNumber] = useState<string | null>(null)
  const [isValidatingRoom, setIsValidatingRoom] = useState(false)
  const [gameState, setGameState] = useState<StateUpdateSnapshot | null>(null)
  const [lastShowdown, setLastShowdown] = useState<ShowdownSummary | null>(null)
  const [showShowdownPanel, setShowShowdownPanel] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [logs, setLogs] = useState<GameLog[]>([
    { id: 'init-1', time: new Date().toLocaleTimeString('zh-CN', { hour12: false }), text: '欢迎来到德州扑克房间', type: 'system' },
    { id: 'init-2', time: new Date().toLocaleTimeString('zh-CN', { hour12: false }), text: '等待玩家加入和准备...', type: 'system' }
  ])
  const logsEndRef = useRef<HTMLDivElement>(null)
  const prevGameStateRef = useRef<StateUpdateSnapshot | null>(null)

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])
  
  const [chatInput, setChatInput] = useState('')
  const [startCountdown, setStartCountdown] = useState<number | null>(null)
  const [actionCountdown, setActionCountdown] = useState<{ playerId: string, seconds: number } | null>(null)
  const [turnNotification, setTurnNotification] = useState<TurnNotificationPayload | null>(null)
  const [betAmount, setBetAmount] = useState<number>(0)
  
  const roomNumber = searchParams.get('room')
  
  useEffect(() => {
    if (startCountdown === null || startCountdown <= 0) return;
    const timer = setInterval(() => {
      setStartCountdown(prev => (prev !== null && prev > 0 ? prev - 1 : null));
    }, 1000);
    return () => clearInterval(timer);
  }, [startCountdown]);

  useEffect(() => {
    if (!actionCountdown || actionCountdown.seconds <= 0) return;
    const timer = setInterval(() => {
      setActionCountdown(prev => {
        if (prev && prev.seconds > 0) {
          return { ...prev, seconds: prev.seconds - 1 };
        }
        return null;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, [actionCountdown?.playerId]);

  // 初始化 WebSocket
  const { messageQueue, isKicked, error, sendMessage } = useWebSocket(validatedRoomNumber, user?.id)
  const processedMessageIndexRef = useRef(0)

  useEffect(() => {
    if (!roomNumber) {
      navigate('/')
      return
    }

    setIsValidatingRoom(true)
    setValidatedRoomNumber(null)
    getRoom(roomNumber)
      .then(() => {
        setValidatedRoomNumber(roomNumber)
      })
      .catch((err: any) => {
        try {
          sessionStorage.setItem('home_notice', err?.message || '房间不存在或已被回收')
        } catch {
          // ignore
        }
        navigate('/')
      })
      .finally(() => {
        setIsValidatingRoom(false)
      })
  }, [roomNumber, navigate])

  // 监听 WebSocket 消息
  useEffect(() => {
    while (processedMessageIndexRef.current < messageQueue.length) {
      const lastMessage = messageQueue[processedMessageIndexRef.current];
      processedMessageIndexRef.current++;

      if (lastMessage.type === 'room.welcome') {
        // 判断当前用户是否是房主
        setIsOwner(lastMessage.payload.owner_id === user?.id);
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: `欢迎来到房间: ${lastMessage.payload.room_number}`,
          type: 'system' as const
        }].slice(-100));
      } else if (lastMessage.type === 'room.destroyed') {
        if (!isOwner) {
          try {
            sessionStorage.setItem('home_notice', '房间已被房主解散');
          } catch {
            // ignore
          }
        }
        navigate('/');
      } else if (lastMessage.type === 'room.chat') {
        const chatPayload = lastMessage.payload;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: chatPayload.message,
          type: 'chat' as const,
          senderId: chatPayload.user_id,
          senderName: chatPayload.user_name,
          senderAvatar: chatPayload.avatar
        }].slice(-100));
      } else if (lastMessage.type === 'room.player_join') {
        const p = lastMessage.payload;
        const isMe = p.user_id === user?.id;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: isMe ? '你进入了房间' : `玩家 [${p.user_name}] 进入了房间`,
          type: 'system' as const
        }].slice(-100));
        // 如果是自己加入房间，此时可能还没有 gameState，尝试从 payload 中获取或等待 state_update
      } else if (lastMessage.type === 'room.player_leave') {
        const p = lastMessage.payload;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: `玩家 [${p.user_name}] 离开了房间`,
          type: 'system' as const
        }].slice(-100));
      } else if (lastMessage.type === 'room.history') {
        const payload = lastMessage.payload;
        const chatHistory = payload.chat_history || [];
        const gameHistory = payload.game_history || [];
        
        const historyLogs: GameLog[] = [];
        
        chatHistory.forEach((entry: any) => {
          const chatPayload = entry.message.payload;
          historyLogs.push({
            id: 'history-chat-' + Math.random(),
            time: new Date(entry.time).toLocaleTimeString('zh-CN', { hour12: false }),
            text: chatPayload.message,
            type: 'chat',
            senderId: chatPayload.user_id,
            senderName: chatPayload.user_name,
            senderAvatar: chatPayload.avatar,
            timestamp: entry.time
          });
        });

        let tempPrevSnap: StateUpdateSnapshot | null = null;
        gameHistory.forEach((entry: any) => {
          const snap = entry.message.payload as StateUpdateSnapshot;
          const reason = entry.message.reason;
          let logText = '';
          
          const getNewPlayerName = () => {
            if (!tempPrevSnap) return '某玩家';
            const newP = snap.players?.find(p => !tempPrevSnap?.players?.find(old => old.id === p.id));
            return newP ? newP.name : '某玩家';
          };
          
          const getRemovedPlayerName = () => {
            if (!tempPrevSnap) return '某玩家';
            const oldP = tempPrevSnap?.players?.find(old => !snap.players?.find(p => p.id === old.id));
            return oldP ? oldP.name : '某玩家';
          };

          const getReadyPlayerName = (isReady: boolean) => {
            if (!tempPrevSnap) return '某玩家';
            const targetState = isReady ? 'ready' : 'waiting';
            const p = snap.players?.find(p => {
              const oldP = tempPrevSnap?.players?.find(old => old.id === p.id);
              return p.state === targetState && oldP?.state !== targetState;
            });
            return p ? p.name : '某玩家';
          };

          const getOfflineChangedPlayerName = (toOffline: boolean) => {
            if (!tempPrevSnap) return '某玩家';
            const p = snap.players?.find(p => {
              const oldP = tempPrevSnap?.players?.find(old => old.id === p.id);
              return p.is_offline === toOffline && oldP?.is_offline !== toOffline;
            });
            return p ? p.name : '某玩家';
          };

          switch (reason) {
            case 'player_reconnected': logText = `${getOfflineChangedPlayerName(false)} 重新连接`; break;
            case 'sit_down': logText = `${getNewPlayerName()} 坐下了`; break;
            case 'stand_up': logText = `${getRemovedPlayerName()} 站起了`; break;
            case 'ready': logText = `${getReadyPlayerName(true)} 已准备`; break;
            case 'cancel_ready': logText = `${getReadyPlayerName(false)} 取消了准备`; break;
            case 'deal_hole_cards': logText = `第 ${snap.hand_count + 1} 局游戏开始，正在发牌...`; break;
            case 'player_action': 
              if (snap.last_action) {
                const p = snap.players?.find(p => p.id === snap.last_action?.player_id);
                const pName = p ? p.name : '某玩家';
                const actionMap: Record<string, string> = {
                  'fold': '弃牌', 'check': '过牌', 'call': '跟注', 'bet': '下注', 'raise': '加注', 'allin': 'All-in'
                };
                const actName = actionMap[snap.last_action.action || ''] || snap.last_action.action;
                logText = `${pName} ${actName} ${snap.last_action.amount ? snap.last_action.amount : ''}`;
              }
              break;
            case 'next_stage': {
              const stageMap: Record<string, string> = {
                'PREFLOP': '翻牌前', 'FLOP': '翻牌圈', 'TURN': '转牌圈', 'RIVER': '河牌圈', 'SHOWDOWN': '摊牌'
              };
              logText = `进入 ${stageMap[snap.stage] || snap.stage} 阶段`; 
              break;
            }
            case 'showdown': logText = '进入摊牌结算'; break;
            case 'hand_finished': logText = `第 ${snap.hand_count} 局游戏结束`; break;
            case 'early_finish': logText = `其他玩家均已弃牌，本局提前结束`; break;
          }
          
          if (logText) {
            historyLogs.push({
              id: 'history-game-' + Math.random(),
              time: new Date(entry.time).toLocaleTimeString('zh-CN', { hour12: false }),
              text: logText,
              type: 'action',
              timestamp: entry.time
            });
          }
          tempPrevSnap = snap;
        });

        // Sort by timestamp
        historyLogs.sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0));
        
        setLogs(prev => {
          return [...prev, ...historyLogs].slice(-100);
        });
      } else if (lastMessage.type === 'texas.state_update') {
        const snap = lastMessage.payload as StateUpdateSnapshot;
        const prevSnap = prevGameStateRef.current;
        
        // 保留旧快照中的手牌数据（如果新快照中没有，且在同一局游戏中）
        if (prevSnap && snap.hand_count === prevSnap.hand_count) {
          snap.players.forEach(p => {
            const oldP = prevSnap.players?.find(old => old.id === p.id);
            if (oldP && oldP.hole_cards && oldP.hole_cards.length > 0 && (!p.hole_cards || p.hole_cards.length === 0)) {
              p.hole_cards = oldP.hole_cards;
            }
          });
        }
        
        // 只有在第一次收到 state_update，或者状态确实有变化时才更新
        setGameState(snap);

        if (snap.showdown_summary) {
          setLastShowdown(snap.showdown_summary);
          // 只有在真正进入 SHOWDOWN 阶段时才自动弹出结算面板
          // 如果是在 WAITING 阶段收到的快照（比如重连），只保存数据不自动弹出
          if (snap.stage === 'SHOWDOWN') {
            setShowShowdownPanel(true);
          }
        } else if (snap.stage === 'PREFLOP' || lastMessage.reason === 'deal_hole_cards') {
          setLastShowdown(null);
          setShowShowdownPanel(false);
        }

        // 如果不是当前玩家的回合，清空通知
        const currentPlayer = snap.players?.find(p => p.seat_number === snap.current_player_index);
        if (currentPlayer?.id !== user?.id) {
          setTurnNotification(null);
        }
        if (!currentPlayer || snap.stage === 'SHOWDOWN' || snap.stage === 'WAITING') {
          setActionCountdown(null);
        }
        if (snap.stage === 'SHOWDOWN' || snap.stage === 'WAITING') {
          setStartCountdown(null);
        }

        const reason = lastMessage.reason;
        if (reason) {
          let logText = '';
          
          const getNewPlayerName = () => {
            const newP = snap.players?.find(p => !prevSnap?.players?.find(old => old.id === p.id));
            return newP ? newP.name : '某玩家';
          };
          
          const getRemovedPlayerName = () => {
            const oldP = prevSnap?.players?.find(old => !snap.players?.find(p => p.id === old.id));
            return oldP ? oldP.name : '某玩家';
          };

          const getReadyPlayerName = (isReady: boolean) => {
            const targetState = isReady ? 'ready' : 'waiting';
            const p = snap.players?.find(p => {
              const oldP = prevSnap?.players?.find(old => old.id === p.id);
              return p.state === targetState && oldP?.state !== targetState;
            });
            return p ? p.name : '某玩家';
          };

          const getOfflineChangedPlayerName = (toOffline: boolean) => {
            const p = snap.players?.find(p => {
              const oldP = prevSnap?.players?.find(old => old.id === p.id);
              return p.is_offline === toOffline && oldP?.is_offline !== toOffline;
            });
            return p ? p.name : '某玩家';
          };

          switch (reason) {
            case 'player_joined': 
              // 忽略 player_joined 的日志，因为 room.player_join 已经处理了
              break;
            case 'player_left': 
              // 忽略 player_left 的日志，因为 room.player_leave 已经处理了
              break;
            case 'player_reconnected': logText = `${getOfflineChangedPlayerName(false)} 重新连接`; break;
            case 'sit_down': logText = `${getNewPlayerName()} 坐下了`; break;
            case 'stand_up': logText = `${getRemovedPlayerName()} 站起了`; break;
            case 'ready': logText = `${getReadyPlayerName(true)} 已准备`; break;
            case 'cancel_ready': logText = `${getReadyPlayerName(false)} 取消了准备`; break;
            case 'deal_hole_cards': logText = `第 ${snap.hand_count + 1} 局游戏开始，正在发牌...`; break;
            case 'player_action': 
              if (snap.last_action) {
                const p = snap.players?.find(p => p.id === snap.last_action?.player_id);
                const pName = p ? p.name : '某玩家';
                const actionMap: Record<string, string> = {
                  'fold': '弃牌', 'check': '过牌', 'call': '跟注', 'bet': '下注', 'raise': '加注', 'allin': 'All-in'
                };
                const actName = actionMap[snap.last_action.action || ''] || snap.last_action.action;
                logText = `${pName} ${actName} ${snap.last_action.amount ? snap.last_action.amount : ''}`;
              }
              break;
            case 'next_stage': {
              const stageMap: Record<string, string> = {
                'PREFLOP': '翻牌前', 'FLOP': '翻牌圈', 'TURN': '转牌圈', 'RIVER': '河牌圈', 'SHOWDOWN': '摊牌'
              };
              logText = `进入 ${stageMap[snap.stage] || snap.stage} 阶段`; 
              break;
            }
            case 'showdown': logText = '进入摊牌结算'; break;
            case 'hand_finished': logText = `第 ${snap.hand_count} 局游戏结束`; break;
            case 'early_finish': logText = `其他玩家均已弃牌，本局提前结束`; break;
          }
          
          if (logText) {
            setLogs(prev => [...prev, {
              id: Date.now() + '-' + Math.random(),
              time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
              text: logText,
              type: 'action' as const
            }].slice(-100));
          }
        }
        
        prevGameStateRef.current = snap;
      } else if (lastMessage.type === 'texas.countdown') {
        const payload = lastMessage.payload as CountdownPayload;
        // 如果 player_id 为空，说明是全局的开局倒计时
        if (!payload.player_id) {
          setStartCountdown(payload.seconds > 0 ? payload.seconds : null);
        } else {
          setActionCountdown({ playerId: payload.player_id, seconds: payload.seconds });
        }
      } else if (lastMessage.type === 'texas.turn_notification') {
        const payload = lastMessage.payload as TurnNotificationPayload;
        setActionCountdown({ playerId: payload.player_id, seconds: payload.timeout_seconds });
        if (payload.player_id === user?.id) {
          setTurnNotification(payload);
          // 默认将滑动条设置在最小加注额
          if (payload.valid_actions.includes('bet')) {
            setBetAmount(payload.action_details.min_bet || 0);
          } else if (payload.valid_actions.includes('raise')) {
            setBetAmount(payload.action_details.min_raise || 0);
          }
        }
      } else if (lastMessage.type === 'texas.hole_cards') {
        const payload = lastMessage.payload as HoleCardsPayload;
        setGameState(prev => {
          if (!prev) return prev;
          const newSnap = {
            ...prev,
            players: prev.players.map(p => 
              p.id === user?.id ? { ...p, hole_cards: payload.cards } : p
            )
          };
          prevGameStateRef.current = newSnap;
          return newSnap;
        });
      }
    }
  }, [messageQueue, user?.id, navigate, isOwner]);

  useEffect(() => {
    if (!isKicked) return;
    try {
      sessionStorage.setItem('kick_notice', error || '该账号已在其他设备登录，你已下线');
    } catch {
      // ignore
    }
    navigate('/');
  }, [isKicked, error, navigate]);

  const handleDeleteRoom = () => {
    setShowDeleteConfirm(true);
  };

  const confirmDeleteRoom = async () => {
    if (!roomNumber || !user?.id) return;
    setIsDeleting(true);
    try {
      await deleteRoom(roomNumber, user.id);
      setShowDeleteConfirm(false);
      // 解散成功后，后端会广播 room.destroyed 消息，前端收到后会自动跳转回大厅
    } catch (err: any) {
      alert(err.message || '解散房间失败');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleSitDown = (seatNumber: number) => {
    sendMessage('texas.sit_down', { seat_number: seatNumber });
    setLastShowdown(null);
    setShowShowdownPanel(false);
  };

  const handleStandUp = () => {
    sendMessage('texas.stand_up', {});
    setLastShowdown(null);
    setShowShowdownPanel(false);
  };

  const handleReady = () => {
    sendMessage('texas.ready', {});
    setLastShowdown(null);
    setShowShowdownPanel(false);
  };

  const handleCancelReady = () => {
    sendMessage('texas.cancel', {});
  };

  const handleSendChat = () => {
    if (!chatInput.trim()) return;
    sendMessage('room.chat', { message: chatInput.trim() });
    setChatInput('');
  };

  const handleAction = (action: string, amount?: number) => {
    sendMessage('texas.action', { action, amount });
    setTurnNotification(null);
  };

  if (loading) {
    return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">加载用户信息中...</div>
  }
  if (!validatedRoomNumber || isValidatingRoom) {
    return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">正在进入房间...</div>
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col lg:flex-row overflow-hidden">
      
      {/* 左侧/上方：游戏主区域 */}
      <div className="flex-1 relative flex flex-col min-h-[60vh] lg:min-h-screen">
        
        <Header 
          userName={user?.nickname || '加载中...'} 
          userAvatar={user?.avatar}
          roomNumber={roomNumber || '未知'}
          isOwner={isOwner}
          onOpenHistory={() => setShowHistory(true)} 
          onOpenSettings={() => {}} 
          onDeleteRoom={handleDeleteRoom}
        />

        {/* 牌桌区域 */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-8 mt-16 lg:mt-0 relative">
          {/* 牌桌 (绿色椭圆) */}
          <div className="relative w-full max-w-2xl lg:max-w-4xl aspect-[2/1] bg-green-800 rounded-[1000px] border-[12px] sm:border-[16px] border-gray-800 shadow-2xl flex items-center justify-center">
            
            {/* 牌桌内圈线 */}
            <div className="absolute inset-3 sm:inset-4 rounded-[1000px] border-2 border-green-700 opacity-50 pointer-events-none"></div>

            {/* 公共牌和底池 */}
            {gameState && (gameState.board_cards?.length > 0 || gameState.pot > 0) && (
              <div className="absolute z-30 flex flex-col items-center justify-center gap-2">
                {gameState.pot > 0 && (
                  <div className="bg-black/60 px-4 py-1 rounded-full text-yellow-400 font-bold text-sm sm:text-base border border-yellow-600/30 shadow-lg backdrop-blur-sm">
                    底池: ${gameState.pot}
                  </div>
                )}
                {gameState.board_cards && gameState.board_cards.length > 0 && (
                  <div className="flex gap-2 items-center justify-center">
                    {gameState.board_cards.map((card, idx) => (
                      <PlayingCard key={idx} card={card} />
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* 开局倒计时 */}
            {startCountdown !== null && startCountdown > 0 && gameState?.stage !== 'SHOWDOWN' && (
              <div className="absolute z-40 flex flex-col items-center justify-center pointer-events-none top-[10%]">
                <div className="text-6xl sm:text-7xl font-bold text-white drop-shadow-[0_0_15px_rgba(255,255,255,0.8)] animate-pulse">
                  {startCountdown}
                </div>
                <div className="text-yellow-400 font-bold mt-2 text-lg sm:text-xl drop-shadow-md">
                  即将开局
                </div>
              </div>
            )}

            {/* 本人回合倒计时 (中心放大显示) */}
            {actionCountdown && actionCountdown.playerId === user?.id && actionCountdown.seconds > 0 && gameState?.stage !== 'SHOWDOWN' && (
              <div className="absolute z-40 flex flex-col items-center justify-center pointer-events-none top-[10%]">
                <div className={`text-6xl sm:text-7xl font-bold drop-shadow-[0_0_15px_rgba(255,255,255,0.8)] ${actionCountdown.seconds <= 5 ? 'text-red-500 animate-ping' : 'text-yellow-400'}`}>
                  {actionCountdown.seconds}
                </div>
                <div className="text-white font-bold mt-2 text-lg sm:text-xl drop-shadow-md">
                  请行动
                </div>
              </div>
            )}

            {!gameState ? (
              <div className="text-gray-400 font-mono text-lg animate-pulse">
                等待游戏数据...
              </div>
            ) : (
              <>
                {/* 座位渲染 */}
                {Array.from({ length: gameState.max_players || 9 }).map((_, index) => {
                  const player = gameState.players?.find(p => p.seat_number === index);
                  const isMySeat = player?.id === user?.id;
                  const position = getSeatPosition(gameState.max_players || 9, index);

                  return (
                    <div 
                      key={`seat-${index}`}
                      className="absolute"
                      style={{ ...position, zIndex: 50 }}
                    >
                      {player ? (
                        <div className="flex flex-col items-center group relative">
                          {/* 玩家头像和信息 */}
                          <div className={`w-14 h-14 sm:w-16 sm:h-16 rounded-full border-4 ${isMySeat ? 'border-yellow-400' : 'border-gray-600'} bg-gray-800 flex items-center justify-center overflow-hidden shadow-lg relative`}>
                            {player.avatar ? (
                              <img src={player.avatar} alt={player.name} className="w-full h-full object-cover" />
                            ) : (
                              <span className="text-xl font-bold text-gray-400">{player.name[0]?.toUpperCase()}</span>
                            )}
                            
                            {/* 离线遮罩 */}
                            {player.is_offline && (
                              <div className="absolute inset-0 bg-black/60 flex items-center justify-center">
                                <span className="text-xs text-red-400 font-bold">离线</span>
                              </div>
                            )}

                            {/* 倒计时遮罩 (他人视角) */}
                            {actionCountdown && actionCountdown.playerId === player.id && actionCountdown.playerId !== user?.id && actionCountdown.seconds > 0 && gameState?.stage !== 'SHOWDOWN' && (
                              <div className="absolute inset-0 bg-black/50 flex items-center justify-center rounded-full">
                                <span className={`text-lg font-bold ${actionCountdown.seconds <= 5 ? 'text-red-500 animate-ping' : 'text-yellow-400'}`}>
                                  {actionCountdown.seconds}
                                </span>
                              </div>
                            )}
                          </div>
                          
                          {/* 玩家名字和筹码 */}
                          <div className="mt-1 bg-black/70 px-2 py-1 rounded-md text-center min-w-[80px]">
                            <div className="text-xs text-gray-300 truncate max-w-[80px]">{player.name}</div>
                            <div className="text-sm font-bold text-yellow-400">${player.chips}</div>
                          </div>

                          {/* 状态标签 */}
                          {player.state !== 'waiting' && player.state !== 'active' && (
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-[10px] px-2 py-0.5 rounded-full whitespace-nowrap shadow-md">
                              {player.state === 'ready' ? '已准备' : player.state === 'folded' ? '已弃牌' : player.state === 'allin' ? 'All-in' : player.state}
                            </div>
                          )}

                          {/* 手牌 */}
                          {(player.state !== 'waiting' || (gameState.stage === 'WAITING' && player.hole_cards && player.hole_cards.length > 0)) && (
                            <div className="absolute top-full mt-1 flex gap-1 z-20">
                              {player.hole_cards && player.hole_cards.length > 0 ? (
                                player.hole_cards.map((card, idx) => (
                                  <PlayingCard key={idx} card={card} className={`scale-75 origin-top ${player.state === 'folded' ? 'opacity-50 grayscale' : ''}`} />
                                ))
                              ) : gameState.stage !== 'WAITING' && player.state !== 'waiting' ? (
                                // 如果没有具体手牌数据（他人视角），且游戏在进行中，显示背面
                                <>
                                  <PlayingCard card={{suit:0, rank:0}} hidden className={`scale-75 origin-top ${player.state === 'folded' ? 'opacity-50 grayscale' : ''}`} />
                                  <PlayingCard card={{suit:0, rank:0}} hidden className={`scale-75 origin-top ${player.state === 'folded' ? 'opacity-50 grayscale' : ''}`} />
                                </>
                              ) : null}
                            </div>
                          )}

                          {/* 起身按钮 (仅对自己显示) */}
                          {isMySeat && gameState.stage === 'WAITING' && (
                            <button 
                              onClick={handleStandUp}
                              className="absolute -right-8 top-0 bg-red-600 hover:bg-red-500 text-white text-xs px-2 py-1 rounded-md shadow-md transition-colors opacity-0 group-hover:opacity-100"
                            >
                              起身
                            </button>
                          )}
                        </div>
                      ) : (
                        <div className="flex flex-col items-center">
                          {/* 空座位 */}
                          <button 
                            onClick={() => handleSitDown(index)}
                            disabled={gameState.stage !== 'WAITING'} // 只要游戏未开始（WAITING阶段），允许任何人（包括已落座玩家换座）坐下
                            className="w-14 h-14 sm:w-16 sm:h-16 rounded-full border-4 border-dashed border-gray-600 bg-black/30 hover:bg-black/50 hover:border-gray-400 flex items-center justify-center transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            <span className="text-gray-500 text-sm">坐下</span>
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </>
            )}

            {/* 结算面板 */}
            {lastShowdown && showShowdownPanel && (
              <div className="absolute inset-0 z-[100] flex items-center justify-center bg-black/60 rounded-[1000px] backdrop-blur-sm">
                <div className="bg-gray-800 border border-gray-700 rounded-xl p-4 sm:p-6 shadow-2xl w-[90%] max-w-md max-h-[80%] overflow-y-auto pointer-events-auto relative">
                  <button 
                    onClick={() => setShowShowdownPanel(false)}
                    className="absolute top-3 right-3 text-gray-400 hover:text-white w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-700 transition-colors"
                  >
                    ✕
                  </button>
                  <h2 className="text-xl sm:text-2xl font-bold text-white mb-4 text-center pr-6">第 {lastShowdown.hand_id} 局结算</h2>
                  
                  {/* 结算面板中的公共牌 */}
                  {lastShowdown.board_cards && lastShowdown.board_cards.length > 0 && (
                    <div className="flex gap-2 mb-4 justify-center">
                      {lastShowdown.board_cards.map((c, i) => (
                        <PlayingCard key={i} card={c} className="scale-[0.8] sm:scale-100 origin-top" />
                      ))}
                    </div>
                  )}

                  <div className="space-y-2">
                    {/* 排序：赢家在前，然后按盈亏降序 */}
                    {[...lastShowdown.player_results].sort((a, b) => {
                      if (a.is_winner && !b.is_winner) return -1;
                      if (!a.is_winner && b.is_winner) return 1;
                      return b.net_profit - a.net_profit;
                    }).map((result, idx) => {
                      const isMe = result.player_id === user?.id;
                      return (
                        <div key={idx} className={`flex justify-between items-center p-2 rounded border ${result.is_winner ? 'border-green-700/50 bg-green-900/20' : 'border-gray-800 bg-gray-800/30'}`}>
                          <div className="flex items-center gap-2">
                            <span className={`text-sm ${isMe ? 'text-blue-400 font-bold' : 'text-white'}`}>
                              {result.player_name} {isMe && '(You)'}
                            </span>
                          </div>
                          <div className="flex items-center gap-3">
                            {result.cards && result.cards.length > 0 ? (
                              <div className="flex gap-1 mr-2">
                                {result.cards.map((c, i) => (
                                  <PlayingCard key={i} card={c} className="scale-[0.5] sm:scale-[0.6] origin-center" />
                                ))}
                              </div>
                            ) : (
                              <span className="text-gray-500 text-xs italic mr-2">未亮牌</span>
                            )}
                            {result.hand_rank > 0 && (
                              <span className="text-gray-400 text-xs">
                                {['高牌', '一对', '两对', '三条', '顺子', '同花', '葫芦', '四条', '同花顺', '皇家同花顺'][result.hand_rank - 1] || '未知牌型'}
                              </span>
                            )}
                            <span className={`${result.net_profit > 0 ? 'text-green-400' : result.net_profit < 0 ? 'text-red-400' : 'text-gray-500'} font-bold text-sm w-16 text-right`}>
                              {result.net_profit > 0 ? '+' : ''}{result.net_profit}
                            </span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            )}

            {/* 重新打开结算面板的按钮 */}
            {lastShowdown && !showShowdownPanel && (
              <button
                onClick={() => setShowShowdownPanel(true)}
                className="absolute top-4 right-4 z-50 bg-gray-800/80 hover:bg-gray-700 text-white text-xs sm:text-sm px-3 py-1.5 rounded-full shadow-lg border border-gray-600 backdrop-blur-sm transition-colors"
              >
                查看上局结算
              </button>
            )}
          </div>
        </main>
      </div>

      {/* 右侧/下方：侧边栏 (聊天与操作) */}
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-[40vh] lg:h-screen">
        {gameState ? (
          <>
            {/* 上方：游戏日志和聊天内容 (约占 2/3) */}
            <div className="flex-1 flex flex-col bg-gray-900 mx-4 mt-4 mb-2 rounded-lg shadow-inner overflow-hidden">
              <div className="bg-gray-800 px-4 py-2 border-b border-gray-700 font-bold text-sm text-gray-300 flex justify-between items-center">
                <span>游戏记录</span>
              </div>
              
              {/* 日志列表 */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {logs.map(log => {
                  if (log.type === 'system' || log.type === 'action') {
                    return (
                      <div key={log.id} className="flex justify-center">
                        <span className="text-[11px] bg-gray-800/80 text-gray-400 px-3 py-1 rounded-full shadow-sm">
                          [{log.time}] {log.text}
                        </span>
                      </div>
                    );
                  }

                  // 聊天气泡
                  const isMe = log.senderId === user?.id;
                  return (
                    <div key={log.id} className={`flex w-full gap-2 ${isMe ? 'flex-row-reverse' : 'flex-row'} items-start`}>
                      {/* 头像 */}
                      <div className="w-8 h-8 rounded-full bg-gray-700 flex-shrink-0 flex items-center justify-center overflow-hidden border border-gray-600 shadow-sm mt-1">
                        {log.senderAvatar ? (
                          <img src={log.senderAvatar} alt={log.senderName} className="w-full h-full object-cover" />
                        ) : (
                          <span className="text-xs text-gray-400 font-bold">{log.senderName?.[0]?.toUpperCase()}</span>
                        )}
                      </div>

                      {/* 消息内容区 */}
                      <div className={`flex flex-col ${isMe ? 'items-end' : 'items-start'} max-w-[75%]`}>
                        <span className="text-[10px] text-gray-500 mb-1 mx-1">
                          {log.senderName} {log.time}
                        </span>
                        <div className={`px-3 py-2 rounded-2xl text-sm shadow-md ${
                          isMe 
                            ? 'bg-blue-600 text-white rounded-tr-sm' 
                            : 'bg-gray-700 text-gray-200 rounded-tl-sm'
                        }`}>
                          {log.text}
                        </div>
                      </div>
                    </div>
                  );
                })}
                <div ref={logsEndRef} />
              </div>

              {/* 聊天输入框 */}
              <div className="p-2 bg-gray-800 border-t border-gray-700 flex">
                <input 
                  type="text" 
                  placeholder="输入聊天内容..." 
                  value={chatInput}
                  onChange={e => setChatInput(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSendChat()}
                  className="flex-1 bg-gray-900 text-white text-sm px-3 py-1.5 rounded-l-md outline-none border border-gray-700 focus:border-gray-500 transition-colors" 
                />
                <button 
                  onClick={handleSendChat}
                  disabled={!chatInput.trim()}
                  className="bg-blue-600 hover:bg-blue-500 px-4 py-1.5 rounded-r-md text-sm text-white font-bold disabled:opacity-50 transition-colors" 
                >
                  发送
                </button>
              </div>
            </div>

            {/* 下方：操作区域 (约占 1/3) */}
            <div className="shrink-0 bg-gray-900 mx-4 mt-2 mb-4 p-4 rounded-lg shadow-inner min-h-[160px] flex flex-col justify-center">
              {(() => {
                const myPlayer = gameState.players?.find(p => p.id === user?.id);
                
                if (!myPlayer) {
                  return <div className="text-sm text-gray-400 text-center">请先在牌桌上找个空位坐下</div>;
                }

                if (gameState.stage === 'WAITING') {
                  if (myPlayer.state === 'ready') {
                    return (
                      <button 
                        onClick={handleCancelReady}
                        className="w-full py-3 bg-red-600 hover:bg-red-500 text-white font-bold rounded-lg transition-colors shadow-lg"
                      >
                        取消准备
                      </button>
                    );
                  } else {
                    return (
                      <button 
                        onClick={handleReady}
                        className="w-full py-3 bg-green-600 hover:bg-green-500 text-white font-bold rounded-lg transition-colors shadow-lg"
                      >
                        准备
                      </button>
                    );
                  }
                }

                // 游戏进行中的操作按钮
                const isMyTurn = turnNotification !== null;
                const validActions = turnNotification?.valid_actions || [];
                const details = turnNotification?.action_details || {};

                return (
                  <div className="flex flex-col gap-3">
                    {!isMyTurn ? (
                      <div className="text-center text-gray-400 py-4">等待其他玩家行动...</div>
                    ) : (
                      <>
                        <div className="grid grid-cols-2 gap-3">
                          <button 
                            onClick={() => handleAction('fold')}
                            disabled={!validActions.includes('fold')}
                            className={`py-3 font-bold rounded-lg shadow-md transition-colors ${validActions.includes('fold') ? 'bg-red-600 hover:bg-red-500 text-white' : 'bg-gray-700 text-gray-500 cursor-not-allowed'}`}
                          >
                            弃牌 (Fold)
                          </button>
                          
                          <button 
                            onClick={() => handleAction('check')}
                            disabled={!validActions.includes('check')}
                            className={`py-3 font-bold rounded-lg shadow-md transition-colors ${validActions.includes('check') ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-gray-700 text-gray-500 cursor-not-allowed'}`}
                          >
                            过牌 (Check)
                          </button>
                          
                          <button 
                            onClick={() => handleAction('call')}
                            disabled={!validActions.includes('call')}
                            className={`py-3 font-bold rounded-lg shadow-md transition-colors ${validActions.includes('call') ? 'bg-blue-600 hover:bg-blue-500 text-white' : 'bg-gray-700 text-gray-500 cursor-not-allowed'}`}
                          >
                            跟注 {details.call_amount ? `($${details.call_amount})` : ''}
                          </button>

                          <button 
                            onClick={() => handleAction('allin')}
                            disabled={!validActions.includes('allin')}
                            className={`py-3 font-bold rounded-lg shadow-md transition-colors ${validActions.includes('allin') ? 'bg-yellow-600 hover:bg-yellow-500 text-white' : 'bg-gray-700 text-gray-500 cursor-not-allowed'}`}
                          >
                            All-in {details.allin_amount ? `($${details.allin_amount})` : ''}
                          </button>
                        </div>
                        
                        {(validActions.includes('bet') || validActions.includes('raise')) && (
                          <div className="flex flex-col gap-2 mt-2 bg-gray-800 p-3 rounded-lg border border-gray-700">
                            <div className="flex justify-between items-center text-sm text-gray-400">
                              <span>选择金额: ${betAmount}</span>
                              <span>Max: ${validActions.includes('bet') ? details.max_bet : details.max_raise}</span>
                            </div>
                            <input 
                              type="range" 
                              min={validActions.includes('bet') ? details.min_bet : details.min_raise} 
                              max={validActions.includes('bet') ? details.max_bet : details.max_raise} 
                              value={betAmount}
                              onChange={(e) => setBetAmount(Number(e.target.value))}
                              className="w-full accent-blue-500 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                            />
                            <button 
                              onClick={() => handleAction(validActions.includes('bet') ? 'bet' : 'raise', betAmount)}
                              className="w-full mt-2 py-3 bg-purple-600 hover:bg-purple-500 text-white font-bold rounded-lg shadow-md transition-colors"
                            >
                              {validActions.includes('bet') ? '下注' : '加注'} ${betAmount}
                            </button>
                          </div>
                        )}
                      </>
                    )}
                  </div>
                );
              })()}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            等待游戏数据...
          </div>
        )}
      </aside>

      {/* 解散房间确认弹窗 */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4">
          <div className="bg-gray-800 rounded-xl shadow-2xl p-6 w-full max-w-sm border border-gray-700">
            <h3 className="text-xl font-bold text-white mb-4">解散房间</h3>
            <p className="text-gray-300 mb-6">确定要解散房间吗？此操作不可恢复，所有玩家将被移出房间。</p>
            <div className="flex justify-end space-x-3">
              <button 
                onClick={() => setShowDeleteConfirm(false)}
                disabled={isDeleting}
                className="px-4 py-2 rounded-lg bg-gray-700 text-white hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                取消
              </button>
              <button 
                onClick={confirmDeleteRoom}
                disabled={isDeleting}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-500 transition-colors disabled:opacity-50 flex items-center"
              >
                {isDeleting ? '解散中...' : '确定解散'}
              </button>
            </div>
          </div>
        </div>
      )}

      <HistoryModal 
        show={showHistory} 
        onClose={() => setShowHistory(false)} 
        userId={user?.id}
        histories={gameState?.histories || []}
      />

    </div>
  )
}
