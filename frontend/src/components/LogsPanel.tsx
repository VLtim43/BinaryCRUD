import { h } from "preact";
import { useEffect, useRef } from "preact/hooks";
import { LogEntry } from "../services/logService";

interface LogsPanelProps {
  logs: LogEntry[];
  isOpen: boolean;
  onToggle: () => void;
  onClear: () => void;
  onCopy: () => void;
}

const getLogLevelClass = (level: string) => {
  switch (level.toUpperCase()) {
    case "DEBUG":
      return "log-level-debug";
    case "INFO":
      return "log-level-info";
    case "WARN":
      return "log-level-warn";
    case "ERROR":
      return "log-level-error";
    default:
      return "log-level-info";
  }
};

export const LogsPanel = ({ logs, isOpen, onToggle, onClear, onCopy }: LogsPanelProps) => {
  const logsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className={`logs-panel ${isOpen ? "open" : "closed"}`}>
      <button className="logs-toggle" onClick={onToggle}>
        {isOpen ? "»" : "«"}
      </button>

      {isOpen && (
        <div className="logs-content">
          <div className="logs-header">
            <h3>Logs</h3>
            <div className="logs-controls">
              <button className="btn-icon" onClick={onCopy} title="Copy Logs">
                📋
              </button>
              <button className="btn-icon btn-danger" onClick={onClear} title="Clear Logs">
                🗑
              </button>
            </div>
          </div>
          <div className="logs-container" ref={logsContainerRef}>
            {logs.length === 0 ? (
              <div className="logs-empty">No logs yet</div>
            ) : (
              logs.map((log, index) => (
                <div key={index} className={`log-entry ${getLogLevelClass(log.level)}`}>
                  <span className="log-timestamp">[{log.timestamp}]</span>
                  <span className="log-level">[{log.level}]</span>
                  <span className="log-message">{log.message}</span>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};
