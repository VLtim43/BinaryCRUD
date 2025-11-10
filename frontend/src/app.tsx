import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState, useEffect } from "preact/hooks";
import { h } from "preact";
import { Button } from "./components/Button";
import { ItemTab } from "./components/tabs/ItemTab";
import { OrderTab } from "./components/tabs/OrderTab";
import { PromotionTab } from "./components/tabs/PromotionTab";
import { DebugTab } from "./components/tabs/DebugTab";
import { LogsPanel } from "./components/LogsPanel";
import { logService, LogEntry } from "./services/logService";

type TabType = "item" | "order" | "promotion" | "debug";

export const App = () => {
  const [activeTab, setActiveTab] = useState<TabType>("item");
  const [message, setMessage] = useState("Enter item text below 👇");
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [logsPanelOpen, setLogsPanelOpen] = useState(false);

  useEffect(() => {
    if (logsPanelOpen) {
      refreshLogs();
    }
  }, [logsPanelOpen]);

  const refreshLogs = async () => {
    try {
      const newLogs = await logService.getAll();
      setLogs(newLogs);
    } catch (err) {
      console.error("Error loading logs:", err);
    }
  };

  const handleClearLogs = async () => {
    try {
      await logService.clear();
      setLogs([]);
      setMessage("Logs cleared");
    } catch (err: any) {
      setMessage(`Error clearing logs: ${err}`);
    }
  };

  const handleCopyLogs = () => {
    const logText = logs
      .map((log) => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join("\n");
    navigator.clipboard.writeText(logText).then(() => {
      setMessage("Logs copied to clipboard!");
    });
  };

  const getDefaultMessage = (tab: TabType): string => {
    switch (tab) {
      case "item":
        return "Enter item text below 👇";
      case "order":
        return "Select items to add to your order";
      case "promotion":
        return "Create a new promotion by selecting items";
      case "debug":
        return "Debug tools and utilities";
    }
  };

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab);
    setMessage(getDefaultMessage(tab));
  };

  const showCreateTab = (activeTab === "order" || activeTab === "promotion");

  return (
    <div className={`app-container ${logsPanelOpen ? "logs-open" : ""}`}>
      <button className="close-btn" onClick={() => Quit()}>
        ×
      </button>

      <div className="tabs">
        <Button className={`tab ${activeTab === "item" ? "active" : ""}`} onClick={() => handleTabChange("item")}>
          Item
        </Button>
        <Button className={`tab ${activeTab === "order" ? "active" : ""}`} onClick={() => handleTabChange("order")}>
          Order
        </Button>
        <Button className={`tab ${activeTab === "promotion" ? "active" : ""}`} onClick={() => handleTabChange("promotion")}>
          Promotion
        </Button>
        <Button className={`tab ${activeTab === "debug" ? "active" : ""}`} onClick={() => handleTabChange("debug")}>
          Debug
        </Button>
      </div>

      <div id="App">
        {activeTab !== "order" && activeTab !== "promotion" && <img src={logo} id="logo" alt="logo" />}

        {!showCreateTab && (
          <div id="result" className="result">
            {message}
          </div>
        )}

        {activeTab === "item" && <ItemTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "order" && <OrderTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "promotion" && <PromotionTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "debug" && <DebugTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
      </div>

      <LogsPanel
        logs={logs}
        isOpen={logsPanelOpen}
        onToggle={() => setLogsPanelOpen(!logsPanelOpen)}
        onClear={handleClearLogs}
        onCopy={handleCopyLogs}
      />
    </div>
  );
};
