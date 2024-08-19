import { useCallback, useState } from "react";
import "./App.css";
import { useQuery, useQueryClient } from "@tanstack/react-query";

const useSaveToLocalStorage = (key: string, value: string) => {
  const [state, setState] = useState(() => {
    const localStorageValue = localStorage.getItem(key);
    return localStorageValue !== null ? localStorageValue : value;
  });

  const setLocalStorage = useCallback(
    (newValue: string) => {
      setState(newValue);
      localStorage.setItem(key, newValue);
    },
    [key]
  );

  return [state, setLocalStorage] as const;
};

const useMessageHistory = (channelId: string) => {
  return useQuery({
    queryKey: ["messages", channelId],
    queryFn: async () => {
      const response = await fetch(
        `http://localhost:8080/rest/messages?channel_id=${channelId}`
      );
      return response.json() as Promise<
        {
          id: string;
          content: string;
          author_id: string;
          channel_id: string;
          created_at: string;
        }[]
      >;
    },
    enabled: Boolean(channelId),
  });
};

function App() {
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [message, setMessage] = useState("");
  const [name, setName] = useSaveToLocalStorage("name", "Bob");
  const channelId = "chubra";
  const { data = [] } = useMessageHistory(channelId);
  const queryClient = useQueryClient();

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      setName(event.target.value || "");
    },
    [setName]
  );

  return (
    <div>
      <div>
        <input
          onChange={handleChange}
          value={name}
          type="text"
          name="username"
          placeholder="username"
        />
        <button
          onClick={() => {
            const ws = new WebSocket(
              `ws://localhost:8080/ws/channels/chubra?name=${name}`
            );

            setWs(ws);

            ws.onopen = () => {
              console.log("connected");
            };

            ws.onmessage = () => {
              queryClient.invalidateQueries({
                queryKey: ["messages", channelId],
              });
            };

            ws.onclose = () => {
              console.log("disconnected");
            };

            ws.onerror = (error) => {
              console.log("error", error);
            };
          }}
        >
          connect
        </button>
      </div>
      <div>
        <textarea
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          placeholder="message"
        />
        <button
          onClick={() => {
            ws?.send(message);
            setMessage("");
          }}
        >
          send
        </button>
      </div>
      <div style={{ color: "white" }}>
        {data.map((message) => (
          <div key={message.id}>
            <div>{message.author_id}</div>
            <div>{message.content}</div>
            <div>{message.created_at}</div>
            <hr />
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
