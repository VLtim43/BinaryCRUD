import "./App.scss";
import "./components/Toast.scss";
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
import { ToastContainer } from "./components/Toast";
import { logService, LogEntry } from "./services/logService";
import { toast } from "./utils/toast";

type TabType = "item" | "order" | "promotion" | "debug";

export const App = () => {
  const [activeTab, setActiveTab] = useState<TabType>("item");
  const [message, setMessage] = useState("Enter item text below 👇");
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [logsPanelOpen, setLogsPanelOpen] = useState(false);
  const [debugSubTab, setDebugSubTab] = useState<"tools" | "print" | "compress">("tools");

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
      toast.error("Failed to load logs");
    }
  };

  const handleClearLogs = async () => {
    try {
      await logService.clear();
      setLogs([]);
      toast.success("Logs cleared");
    } catch (err) {
      toast.error("Failed to clear logs");
    }
  };

  const handleCopyLogs = () => {
    const logText = logs
      .map((log) => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join("\n");
    navigator.clipboard.writeText(logText).then(() => {
      toast.success("Logs copied to clipboard!");
    });
  };

  const getDefaultMessage = (tab: TabType): string => {
    switch (tab) {
      case "item":
        return "Enter item text below 👇";
      case "order":
        return "Manage orders";
      case "promotion":
        return "Manage promotions";
      case "debug":
        return "Debug tools and utilities";
    }
  };

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab);
    setMessage(getDefaultMessage(tab));
  };

  const hideLogo = activeTab === "order" || activeTab === "promotion" || (activeTab === "debug" && (debugSubTab === "print" || debugSubTab === "compress"));

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
        {!hideLogo && <img src={logo} id="logo" alt="logo" />}

        <div id="result" className="result">
          {message}
        </div>

        {activeTab === "item" && <ItemTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "order" && <OrderTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "promotion" && <PromotionTab onMessage={setMessage} onRefreshLogs={refreshLogs} />}
        {activeTab === "debug" && <DebugTab onMessage={setMessage} onRefreshLogs={refreshLogs} subTab={debugSubTab} onSubTabChange={setDebugSubTab} />}
      </div>

      <LogsPanel
        logs={logs}
        isOpen={logsPanelOpen}
        onToggle={() => setLogsPanelOpen(!logsPanelOpen)}
        onClear={handleClearLogs}
        onCopy={handleCopyLogs}
      />

      <ToastContainer position="top-right" />
    </div>
  );
};
