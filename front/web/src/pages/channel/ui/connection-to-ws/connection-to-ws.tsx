import { useState, useCallback } from "react";

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

type ConnectionToWsProps = {
  onConnect: (userId: string) => void;
}

export const ConnectionToWs = (props: ConnectionToWsProps) => {
  const [name, setName] = useSaveToLocalStorage("name", "Bob");

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      setName(event.target.value || "");
    },
    [setName]
  );

  const connect = () => {
    props.onConnect(name);
  }


  return (
    <div>
      <input
        onChange={handleChange}
        value={name}
        type="text"
        name="username"
        placeholder="username"
      />
      <button
        onClick={connect}
      >
        connect
      </button>
    </div>
  )
}