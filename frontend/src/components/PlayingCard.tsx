import React from 'react';
import type { Card } from '../types/game';

interface PlayingCardProps {
  card: Card;
  className?: string;
  hidden?: boolean;
}

const getSuitSymbol = (suit: number) => {
  switch (suit) {
    case 0: return '♠';
    case 1: return '♥';
    case 2: return '♦';
    case 3: return '♣';
    default: return '?';
  }
};

const getSuitColor = (suit: number) => {
  return suit === 1 || suit === 2 ? 'text-red-600' : 'text-gray-900';
};

const getRankString = (rank: number) => {
  if (rank >= 2 && rank <= 10) return rank.toString();
  switch (rank) {
    case 11: return 'J';
    case 12: return 'Q';
    case 13: return 'K';
    case 14: return 'A';
    default: return '?';
  }
};

export const PlayingCard: React.FC<PlayingCardProps> = ({ card, className = '', hidden = false }) => {
  if (hidden) {
    return (
      <div className={`w-10 h-14 sm:w-12 sm:h-16 bg-blue-800 rounded border-2 border-white shadow-md flex items-center justify-center ${className}`}>
        <div className="w-full h-full m-1 border border-blue-400 rounded-sm opacity-50 bg-[repeating-linear-gradient(45deg,transparent,transparent_2px,#ffffff33_2px,#ffffff33_4px)]"></div>
      </div>
    );
  }

  const symbol = getSuitSymbol(card.suit);
  const colorClass = getSuitColor(card.suit);
  const rankStr = getRankString(card.rank);

  return (
    <div className={`w-10 h-14 sm:w-12 sm:h-16 bg-white rounded border border-gray-300 shadow-md flex flex-col justify-between p-1 ${colorClass} ${className}`}>
      <div className="text-xs sm:text-sm font-bold leading-none">{rankStr}</div>
      <div className="text-lg sm:text-xl text-center leading-none flex-1 flex items-center justify-center">{symbol}</div>
    </div>
  );
};
