import { EventsEmit } from "../../wailsjs/runtime/runtime";

export const toast = {
  success: (message: string) => EventsEmit("toast:success", message),
  error: (message: string) => EventsEmit("toast:error", message),
  warning: (message: string) => EventsEmit("toast:warning", message),
};
