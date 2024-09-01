import { useState } from "react";

type NewMessageProps = {
  onSend: (msg: string) => void;
}

export const NewMessage = (props: NewMessageProps) => {
  const [message, setMessage] = useState("");

  return (
    <div>
    <textarea
      value={message}
      onChange={(event) => setMessage(event.target.value)}
      placeholder="message"
    />
    <button
      onClick={() => {
        props.onSend(message)
        setMessage("");
      }}
    >
      send
    </button>
  </div>
  )
}