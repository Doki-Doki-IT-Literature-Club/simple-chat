import { useQuery } from '@tanstack/react-query';
import styles from './styles.module.css'
import { Message } from '../message/message';
import { ReactNode, useRef } from 'react';
import React from 'react';
import { formatDate } from '../../lib/format-date';

const useMessageHistory = (channelId: string) => {
  return useQuery({
    queryKey: ["messages", channelId],
    queryFn: async () => {
      const response = await fetch(
        `http://localhost:8080/rest/messages?channel_id=${channelId}`
      );
      return response.json() as Promise<
        | {
          id: string;
          content: string;
          author_id: string;
          channel_id: string;
          created_at: string;
        }[]
        | null
      >;
    },
    enabled: Boolean(channelId),
  });
};

type MessagesHistoryProps = {
  channelId: string;
  // TODO
  currentUserId: string;
}

const options: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: "2-digit",
  day: "numeric"
};

export const MessagesHistory = (props: MessagesHistoryProps) => {
  const { data } = useMessageHistory(props.channelId);
  const lastStoredDate = useRef<string>();
  const messagePosition = (messageAuthor: string) => {
    return messageAuthor === props.currentUserId ? 'right' : 'left';
  }

  const showDateIfNeeded = (date: string) => {
    if (!areDatesEqual(date, lastStoredDate.current || '')) {
      lastStoredDate.current = date;
      return (
        <span className={styles.Date} key={date}>{formatDate(date, options)}</span>
      )
    } else return null;
  }

  const areDatesEqual = (date1: string, date2: string) => {
    const d1 = new Date(date1);
    const d2 = new Date(date2);

    return d1.getFullYear() === d2.getFullYear() &&
      d1.getMonth() === d2.getMonth() &&
      d1.getDate() === d2.getDate();
  }

  const Messages = (): ReactNode => {
    const res: ReactNode[] = [];
    if (!data) return res;
    for (let i = data.length - 1; i >= 0; i--) {
      const msg = data[i];
      res.push(<React.Fragment key={msg.id}>
        {showDateIfNeeded(msg.created_at)}
        <Message message={msg} position={messagePosition(msg.author_id)} />
      </React.Fragment>)
    }
    return res;
  }


  return (
    <div className={styles.Container}>
      <Messages />
    </div>
  )
}