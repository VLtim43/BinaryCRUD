import { h } from "preact";
import { useState } from "preact/hooks";
import { Button } from "../Button";
import { systemService } from "../../services/systemService";

interface DebugTabProps {
  onMessage: (msg: string) => void;
  onRefreshLogs: () => void;
}

export const DebugTab = ({ onMessage, onRefreshLogs }: DebugTabProps) => {
  const [indexData, setIndexData] = useState<any>(null);

  const handlePopulate = async () => {
    onMessage("Populating data from items.json and promotions.json...");
    try {
      await systemService.populateInventory();
      onMessage("Data populated successfully! Check logs for details.");
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error populating data: ${err}`);
      onRefreshLogs();
    }
  };

  const handlePrintIndex = async () => {
    onMessage("Loading index contents...");
    try {
      const data = await systemService.getIndexContents();
      setIndexData(data);
      onMessage(`Index loaded: ${data.count} entries. Scroll down to see details.`);
      onRefreshLogs();
    } catch (err: any) {
      setIndexData(null);
      onMessage(`Error loading index: ${err}`);
    }
  };

  const handleDeleteAll = async () => {
    try {
      await systemService.deleteAllFiles();
      setIndexData(null);
      onMessage("All generated files deleted successfully!");
      onRefreshLogs();
    } catch (err: any) {
      onMessage(`Error: ${err}`);
    }
  };

  return (
    <>
      <div className="input-box">
        <Button onClick={handlePopulate}>Populate Data</Button>
        <Button onClick={handlePrintIndex}>Print Index</Button>
        <Button variant="danger" onClick={handleDeleteAll}>
          Delete All Files
        </Button>
      </div>

      {indexData && (
        <div
          style={{
            marginTop: "20px",
            padding: "20px",
            backgroundColor: "rgba(255, 255, 255, 0.05)",
            borderRadius: "8px",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            maxHeight: "400px",
            overflowY: "auto",
          }}
        >
          <h3 style={{ margin: "0 0 15px 0", color: "#fff" }}>B+ Tree Index Contents</h3>
          <div style={{ marginBottom: "10px", color: "#aaa" }}>Total entries: {indexData.count}</div>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              color: "#fff",
            }}
          >
            <thead>
              <tr style={{ borderBottom: "2px solid rgba(255, 255, 255, 0.2)" }}>
                <th style={{ padding: "8px", textAlign: "left" }}>Item ID</th>
                <th style={{ padding: "8px", textAlign: "left" }}>File Offset</th>
              </tr>
            </thead>
            <tbody>
              {indexData.entries.map((entry: any, idx: number) => (
                <tr
                  key={idx}
                  style={{
                    borderBottom: "1px solid rgba(255, 255, 255, 0.1)",
                    backgroundColor: idx % 2 === 0 ? "rgba(0, 0, 0, 0.2)" : "transparent",
                  }}
                >
                  <td style={{ padding: "8px" }}>{entry.id}</td>
                  <td style={{ padding: "8px", fontFamily: "monospace", color: "#888" }}>
                    {entry.offset} bytes
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
};
