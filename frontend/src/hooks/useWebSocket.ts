import { useState, useEffect, useRef, useCallback } from 'react';

interface Envelope {
  type: string;
  reason?: string;
  payload: any;
}

export const useWebSocket = (roomNumber: string | null, userId: string | undefined) => {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastMessage, setLastMessage] = useState<Envelope | null>(null);
  const [isKicked, setIsKicked] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback(() => {
    if (!roomNumber || !userId) return;
    if (isKicked) return;

    // 根据当前协议 (http/https) 决定 ws 协议 (ws/wss)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws/${roomNumber}?user_id=${userId}`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket Connected');
      setIsConnected(true);
      setError(null);
      setIsKicked(false);
    };

    ws.onmessage = (event) => {
      try {
        const message: Envelope = JSON.parse(event.data);
        console.log('Received WS Message:', message);
        setLastMessage(message);
      } catch (err) {
        console.error('Failed to parse WS message:', event.data, err);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket Error:', event);
      setError('WebSocket connection error');
    };

    ws.onclose = (event) => {
      console.log('WebSocket Disconnected:', event.code, event.reason);
      setIsConnected(false);
      wsRef.current = null;
      if (event.code === 4001) {
        setIsKicked(true);
        setError(event.reason || '该账号已在其他设备登录');
      }
    };

    wsRef.current = ws;
  }, [roomNumber, userId, isKicked]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const sendMessage = useCallback((type: string, payload: any = {}) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const envelope: Envelope = { type, payload };
      wsRef.current.send(JSON.stringify(envelope));
    } else {
      console.warn('WebSocket is not connected. Cannot send message:', type);
    }
  }, []);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    isKicked,
    error,
    lastMessage,
    sendMessage,
    reconnect: connect,
  };
};
