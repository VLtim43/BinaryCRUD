import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import { AddItem, PrintBinaryFile, DeleteAllFiles } from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = () => {
  const [activeTab, setActiveTab] = useState<"create" | "read" | "debug">("create");
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const [recordId, setRecordId] = useState("");
  const updateItemText = (e: any) => setItemText(e.target.value);
  const updateRecordId = (e: any) => setRecordId(e.target.value);
  const updateResultText = (result: string) => setResultText(result);

  const addItem = () => {
    // Validate input before sending to backend
    if (!itemText || itemText.trim().length === 0) {
      updateResultText("Error: Cannot add empty item");
      return;
    }

    AddItem(itemText)
      .then(() => {
        updateResultText(`Item saved: ${itemText}`);
        setItemText("");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const printFile = () => {
    PrintBinaryFile()
      .then(() => {
        updateResultText("Binary file printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const getRecordById = () => {
    // TODO: Implement logic to fetch record by ID
    updateResultText(`Getting record with ID: ${recordId}`);
  };

  const deleteAllFiles = () => {
    DeleteAllFiles()
      .then(() => {
        updateResultText("All generated files deleted successfully!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  return (
    <>
      <button className="close-btn" onClick={() => Quit()}>×</button>
      <div className="tabs">
        <button
          className={`tab ${activeTab === "create" ? "active" : ""}`}
          onClick={() => setActiveTab("create")}
        >
          Create
        </button>
        <button
          className={`tab ${activeTab === "read" ? "active" : ""}`}
          onClick={() => setActiveTab("read")}
        >
          Read
        </button>
        <button
          className={`tab ${activeTab === "debug" ? "active" : ""}`}
          onClick={() => setActiveTab("debug")}
        >
          Debug
        </button>
      </div>
      <div id="App">
        <img src={logo} id="logo" alt="logo" />
        <div id="result" className="result">
          {resultText}
        </div>

        {activeTab === "create" && (
          <div id="input" className="input-box">
            <input
              id="name"
              className="input"
              onChange={updateItemText}
              autoComplete="off"
              name="input"
              type="text"
              value={itemText}
            />
            <button className="btn" onClick={addItem}>
              Add Item
            </button>
            <button className="btn" onClick={printFile}>
              Print
            </button>
          </div>
        )}

        {activeTab === "read" && (
          <div id="read-input" className="input-box">
            <input
              id="record-id"
              className="input"
              onChange={updateRecordId}
              autoComplete="off"
              name="record-id"
              type="text"
              placeholder="Enter Record ID"
              value={recordId}
            />
            <button className="btn" onClick={getRecordById}>
              Get Record
            </button>
          </div>
        )}

        {activeTab === "debug" && (
          <div id="debug-controls" className="input-box">
            <button className="btn btn-danger" onClick={deleteAllFiles}>
              Delete All Files
            </button>
          </div>
        )}
      </div>
    </>
  );
}
