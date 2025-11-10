import { useState } from "preact/hooks";

export const useMessage = (initialMessage: string = "") => {
  const [message, setMessage] = useState(initialMessage);

  const showSuccess = (msg: string) => setMessage(msg);
  const showError = (error: any) => setMessage(`Error: ${error}`);
  const clear = () => setMessage("");

  return {
    message,
    setMessage,
    showSuccess,
    showError,
    clear,
  };
};
