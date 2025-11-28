import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";

export type ToastType = "success" | "error" | "info" | "warning";

interface ToastMessage {
  id: number;
  message: string;
  type: ToastType;
  fading: boolean;
}

interface ToastContainerProps {
  position?: "top-right" | "top-left" | "bottom-right" | "bottom-left";
}

let toastId = 0;

export const ToastContainer = ({
  position = "bottom-right",
}: ToastContainerProps) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  useEffect(() => {
    const handleToast = (message: string, type: ToastType = "info") => {
      const id = ++toastId;
      setToasts((prev) => [...prev, { id, message, type, fading: false }]);

      // Start fade out after 3 seconds
      setTimeout(() => {
        setToasts((prev) =>
          prev.map((t) => (t.id === id ? { ...t, fading: true } : t))
        );
      }, 1600);

      // Remove after fade animation (3s + 0.3s)
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, 3300);
    };

    // Listen for toast events from Go
    EventsOn("toast", handleToast);
    EventsOn("toast:success", (msg: string) => handleToast(msg, "success"));
    EventsOn("toast:error", (msg: string) => handleToast(msg, "error"));
    EventsOn("toast:warning", (msg: string) => handleToast(msg, "warning"));
    EventsOn("toast:info", (msg: string) => handleToast(msg, "info"));

    return () => {
      EventsOff(
        "toast",
        "toast:success",
        "toast:error",
        "toast:warning",
        "toast:info"
      );
    };
  }, []);

  const removeToast = (id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  const positionClass = {
    "top-right": "toast-container-top-right",
    "top-left": "toast-container-top-left",
    "bottom-right": "toast-container-bottom-right",
    "bottom-left": "toast-container-bottom-left",
  }[position];

  return (
    <div className={`toast-container ${positionClass}`}>
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`toast toast-${toast.type}${
            toast.fading ? " toast-fading" : ""
          }`}
          onClick={() => removeToast(toast.id)}
        >
          <span className="toast-icon">
            {toast.type === "success" && "✓"}
            {toast.type === "error" && "✕"}
            {toast.type === "warning" && "⚠"}
            {toast.type === "info" && "ℹ"}
          </span>
          <span className="toast-message">{toast.message}</span>
        </div>
      ))}
    </div>
  );
};
