export type HistoryDetail = {
  role: string;
  roleClass: string;
  borderClass: string;
  bgClass: string;
  name: string;
  cards: string[];
  handType: string;
  amountStr: string;
  amountClass: string;
  italic?: boolean;
  isMe?: boolean;
};

export type History = {
  id: number;
  pot: number;
  board: string[];
  details: HistoryDetail[];
};

export const mockHistories: History[] = [
  {
    id: 4,
    pot: 3200,
    board: ['K♠', 'K♥', '9♦', '2♣', '5♠'],
    details: [
      { role: '主池赢家', roleClass: 'text-yellow-400', borderClass: 'border-green-900/30', bgClass: 'bg-gray-800', name: 'Player 1', cards: ['9♠', '9♣'], handType: '葫芦', amountStr: '+$1,200', amountClass: 'text-green-400' },
      { role: '边池赢家', roleClass: 'text-yellow-400', borderClass: 'border-green-900/30', bgClass: 'bg-gray-800', name: 'MyNickname', isMe: true, cards: ['A♠', 'K♦'], handType: '三条', amountStr: '+$2,000', amountClass: 'text-green-400' },
      { role: '输家', roleClass: 'text-gray-500', borderClass: 'border-transparent', bgClass: 'bg-gray-800/50', name: 'Player 3', cards: ['Q♠', 'Q♥'], handType: '两对', amountStr: '-$1,600', amountClass: 'text-red-400' }
    ]
  },
  {
    id: 3,
    pot: 1500,
    board: ['A♥', 'K♠', 'Q♦', 'J♣', '10♥'],
    details: [
      { role: '赢家', roleClass: 'text-yellow-400', borderClass: 'border-green-900/30', bgClass: 'bg-gray-800', name: 'Player 1', cards: ['A♦', 'A♠'], handType: '顺子', amountStr: '+$1,500', amountClass: 'text-green-400' },
      { role: '输家', roleClass: 'text-gray-500', borderClass: 'border-transparent', bgClass: 'bg-gray-800/50', name: 'MyNickname', isMe: true, cards: ['K♠', 'Q♠'], handType: '两对', amountStr: '-$500', amountClass: 'text-red-400' }
    ]
  },
  {
    id: 2,
    pot: 450,
    board: ['7♠', '2♣', '9♥', '', ''],
    details: [
      { role: '赢家', roleClass: 'text-yellow-400', borderClass: 'border-green-900/30', bgClass: 'bg-gray-800', name: 'MyNickname', isMe: true, cards: [], handType: '其他玩家弃牌', amountStr: '+$450', amountClass: 'text-green-400', italic: true }
    ]
  },
  {
    id: 1,
    pot: 2000,
    board: ['A♠', 'K♠', 'Q♠', 'J♠', '10♠'],
    details: [
      { role: '平分', roleClass: 'text-blue-400', borderClass: 'border-blue-900/30', bgClass: 'bg-gray-800', name: 'Player 2', cards: ['2♥', '3♣'], handType: '皇家同花顺', amountStr: '+$1,000', amountClass: 'text-blue-400' },
      { role: '平分', roleClass: 'text-blue-400', borderClass: 'border-blue-900/30', bgClass: 'bg-gray-800', name: 'Player 5', cards: ['4♦', '5♠'], handType: '皇家同花顺', amountStr: '+$1,000', amountClass: 'text-blue-400' }
    ]
  }
];

export const mockSettlementData = [
  { title: '主池 (Main Pot)', winner: 'Player 1', hand: '葫芦', amount: '+$1,200', isMe: false },
  { title: '边池 (Side Pot)', winner: 'MyNickname', hand: '三条', amount: '+$2,000', isMe: true }
];
