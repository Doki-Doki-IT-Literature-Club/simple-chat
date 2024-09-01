import { formatDate } from '../../lib/format-date';
import styles from './styles.module.css'
type Message = {
  id: string;
  author_id: string;
  channel_id: string;
  content: string;
  created_at: string;
}

type MessageProps = {
  message: Message;
  position: 'left' | 'right'
}

export const Message = ({ message, position = 'right' }: MessageProps) => {
  const messagePosition = () => {
    return { alignSelf:  position === 'left' ? 'flex-start' : 'flex-end' }
  }
  return (
    <div className={styles.Message} style={messagePosition()}>
      <div className={styles.Left}>
        <div>{message.author_id}</div>
        <div>{message.content}</div>
      </div>
      <div className={styles.Date}>{formatDate(message.created_at, { timeStyle: 'medium'})}</div>
    </div>
  )
}