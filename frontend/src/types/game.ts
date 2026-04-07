export interface Card {
  suit: number;
  rank: number;
}

export interface PlayerSnapshot {
  id: string;
  name: string;
  avatar: string;
  position?: string;
  seat_number: number;
  chips: number;
  current_bet: number;
  state: 'waiting' | 'ready' | 'active' | 'folded' | 'allin';
  has_acted_this_round: boolean;
  is_offline: boolean;
  buy_in_count: number;
  hole_cards?: Card[];
}

export interface ActionInfo {
  player_id?: string;
  action?: string;
  amount?: number;
}

export interface SidePot {
  amount: number;
  eligible_player_ids: string[];
}

export interface ShowdownSummary {
  // ... (will add details later if needed)
}

export interface CountdownPayload {
  player_id?: string;
  seconds: number;
}

export interface HoleCardsPayload {
  cards: Card[];
}

export interface ActionDetails {
  call_amount?: number;
  min_bet?: number;
  max_bet?: number;
  min_raise?: number;
  max_raise?: number;
  allin_amount?: number;
}

export interface TurnNotificationPayload {
  player_id: string;
  valid_actions: string[];
  action_details: ActionDetails;
  timeout_seconds: number;
}

export interface StateUpdateSnapshot {
  hand_count: number;
  button_seat: number;
  max_players: number;
  stage: 'WAITING' | 'PREFLOP' | 'FLOP' | 'TURN' | 'RIVER' | 'SHOWDOWN';
  pot: number;
  current_bet: number;
  min_raise: number;
  board_cards: Card[];
  current_player_index: number;
  is_paused: boolean;
  action_order: number[];
  side_pots: SidePot[];
  showdown_summary?: ShowdownSummary;
  last_action?: ActionInfo;
  histories: ShowdownSummary[];
  players: PlayerSnapshot[];
}
