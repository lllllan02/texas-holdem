export const getSeatPosition = (total: number, index: number) => {
  const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
  const rx = 54; 
  const ry = 62; 
  const x = 50 + rx * Math.cos(angle);
  const y = 50 + ry * Math.sin(angle);
  return { 
    left: `${x}%`, 
    top: `${y}%`,
    transform: 'translate(-50%, -50%)'
  };
};

export const getCardsPosition = (total: number, index: number) => {
  if (index === 0) return {};

  const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
  const offset = 95; 
  
  const dx = -offset * Math.cos(angle);
  const dy = -offset * Math.sin(angle);

  return {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px))`,
    zIndex: 30,
  };
};

export const getBetPosition = (total: number, index: number) => {
  const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
  const offset = index === 0 ? 110 : 155; 
  
  const dx = -offset * Math.cos(angle);
  const dy = -offset * Math.sin(angle);

  return {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px))`,
    zIndex: 40,
  };
};

export const getMockPosition = (total: number, index: number) => {
  const dealerIndex = 1;
  const offset = (index - dealerIndex + total) % total;
  
  if (total === 2) return offset === 0 ? 'BTN' : 'BB';
  if (total === 3) return ['BTN', 'SB', 'BB'][offset];
  if (total === 4) return ['BTN', 'SB', 'BB', 'UTG'][offset];
  if (total === 5) return ['BTN', 'SB', 'BB', 'UTG', 'CO'][offset];
  if (total === 6) return ['BTN', 'SB', 'BB', 'UTG', 'HJ', 'CO'][offset];
  if (total === 7) return ['BTN', 'SB', 'BB', 'UTG', 'MP', 'HJ', 'CO'][offset];
  if (total === 8) return ['BTN', 'SB', 'BB', 'UTG', 'UTG+1', 'MP', 'HJ', 'CO'][offset];
  if (total === 9) return ['BTN', 'SB', 'BB', 'UTG', 'UTG+1', 'UTG+2', 'MP', 'HJ', 'CO'][offset];
  return '';
};
