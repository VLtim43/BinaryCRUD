import { GetLogs, ClearLogs } from "../../wailsjs/go/main/App";

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
}

export const logService = {
  getAll: async (): Promise<LogEntry[]> => {
    return GetLogs();
  },

  clear: async (): Promise<void> => {
    return ClearLogs();
  },
};
