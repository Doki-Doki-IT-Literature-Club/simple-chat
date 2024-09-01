import { MessagesHistory } from "./ui/messages-history/messages-history"
import styles from './styles.module.css'
import { NewMessage } from "./ui/new-message/new-message";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { WebSocketCLient } from "../../shared/api/websocket/client";
import { ConnectionToWs } from "./ui/connection-to-ws/connection-to-ws";


export const ChannelPage = () => {
  const queryClient = useQueryClient();
  const [ws, setWs] = useState<WebSocketCLient | null>(null);
  const channelId = "chubra";
  const [userId, setUserId] = useState('');
  


  const createWebSocketConnection = () => {
    const url = `ws://localhost:8080/ws/channels/${channelId}?name=${userId}`;
    const ws = new WebSocketCLient(url, onMessageHandler);
    setWs(ws);
  }

  const connectToWs = (userId: string) => {
    setUserId(userId);
    createWebSocketConnection()
  }

  const onMessageHandler = () => {
    queryClient.invalidateQueries({
      queryKey: ["messages", channelId],
    });
  }

  const sendNewMessage = (msg: string) => {
    ws?.sendMessage(msg);
  }

  return (
    <div className={styles.ChatPage}>
      <ConnectionToWs onConnect={connectToWs} />
      <MessagesHistory channelId={channelId} currentUserId={userId} />
      <NewMessage onSend={sendNewMessage} />
    </div>
  )
}