import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import {
  AddItem,
  PrintBinaryFile,
  DeleteAllFiles,
  GetItemByID,
  PopulateInventory,
  DeleteItem,
  PrintIndex,
  RebuildIndex,
} from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = () => {
  const [activeTab, setActiveTab] = useState<"item" | "debug">("item");
  const [itemSubTab, setItemSubTab] = useState<"create" | "read" | "delete">("create");
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteRecordId, setDeleteRecordId] = useState("");
  const updateItemText = (e: any) => setItemText(e.target.value);
  const updateRecordId = (e: any) => {
    const value = e.target.value;
    // Only allow empty string or non-negative integers
    if (value === "" || /^\d+$/.test(value)) {
      setRecordId(value);
    }
  };
  const updateDeleteRecordId = (e: any) => {
    const value = e.target.value;
    // Only allow empty string or non-negative integers
    if (value === "" || /^\d+$/.test(value)) {
      setDeleteRecordId(value);
    }
  };
  const updateResultText = (result: string) => setResultText(result);

  // Get default text for each tab/subtab combination
  const getDefaultText = (
    tab: "item" | "debug",
    subTab?: "create" | "read" | "delete"
  ) => {
    if (tab === "debug") {
      return "Debug tools and utilities";
    }

    if (tab === "item") {
      switch (subTab) {
        case "create":
          return "Enter item text below 👇";
        case "read":
          return "Enter a record ID to fetch 👇";
        case "delete":
          return "Enter a record ID to delete 👇";
        default:
          return "";
      }
    }

    return "";
  };

  // Handle main tab changes
  const handleTabChange = (tab: "item" | "debug") => {
    setActiveTab(tab);
    if (tab === "item") {
      setResultText(getDefaultText(tab, itemSubTab));
    } else {
      setResultText(getDefaultText(tab));
    }
  };

  // Handle item subtab changes
  const handleItemSubTabChange = (subTab: "create" | "read" | "delete") => {
    setItemSubTab(subTab);
    setResultText(getDefaultText("item", subTab));
  };

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
    // Validate input before sending to backend
    if (!recordId || recordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      return;
    }

    // Parse the record ID as a number
    const id = parseInt(recordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Record ID must be a valid non-negative number");
      return;
    }

    GetItemByID(id)
      .then((itemName: string) => {
        updateResultText(`Record ${id}: ${itemName}`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
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

  const populateInventory = () => {
    PopulateInventory("inventory.json")
      .then((result: string) => {
        updateResultText(`Inventory populated! ${result}`);
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const printIndex = () => {
    PrintIndex()
      .then(() => {
        updateResultText("Index printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const rebuildIndex = () => {
    RebuildIndex()
      .then(() => {
        updateResultText("Index rebuilt successfully!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const deleteItem = () => {
    // Validate input before sending to backend
    if (!deleteRecordId || deleteRecordId.trim().length === 0) {
      updateResultText("Error: Please enter a record ID");
      return;
    }

    // Parse the record ID as a number
    const id = parseInt(deleteRecordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Record ID must be a valid non-negative number");
      return;
    }

    DeleteItem(id)
      .then((itemName: string) => {
        updateResultText(`Record [${id}] [${itemName}] was deleted`);
        setDeleteRecordId("");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  return (
    <>
      <button className="close-btn" onClick={() => Quit()}>
        ×
      </button>
      <div className="tabs">
        <button
          className={`tab ${activeTab === "item" ? "active" : ""}`}
          onClick={() => handleTabChange("item")}
        >
          Item
        </button>
        <button
          className={`tab ${activeTab === "debug" ? "active" : ""}`}
          onClick={() => handleTabChange("debug")}
        >
          Debug
        </button>
      </div>

      {activeTab === "item" && (
        <div className="sub_tabs">
          <button
            className={`tab ${itemSubTab === "create" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("create")}
          >
            Create
          </button>
          <button
            className={`tab ${itemSubTab === "read" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("read")}
          >
            Read
          </button>
          <button
            className={`tab ${itemSubTab === "delete" ? "active" : ""}`}
            onClick={() => handleItemSubTabChange("delete")}
          >
            Delete
          </button>
        </div>
      )}

      <div id="App">
        <img src={logo} id="logo" alt="logo" />
        <div id="result" className="result">
          {resultText}
        </div>

        {activeTab === "item" && itemSubTab === "create" && (
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
          </div>
        )}

        {activeTab === "item" && itemSubTab === "read" && (
          <div id="read-input" className="input-box">
            <input
              id="record-id"
              className="input"
              onChange={updateRecordId}
              autoComplete="off"
              name="record-id"
              placeholder="Enter Record ID"
              value={recordId}
            />
            <button className="btn" onClick={getRecordById}>
              Get Record
            </button>
            <button className="btn" onClick={printFile}>
              Print
            </button>
          </div>
        )}

        {activeTab === "item" && itemSubTab === "delete" && (
          <div id="delete-input" className="input-box">
            <input
              id="delete-record-id"
              className="input"
              onChange={updateDeleteRecordId}
              autoComplete="off"
              name="delete-record-id"
              placeholder="Enter Record ID"
              value={deleteRecordId}
            />
            <button className="btn btn-danger" onClick={deleteItem}>
              Delete Record
            </button>
          </div>
        )}

        {activeTab === "debug" && (
          <div id="debug-controls" className="debug-section">
            <div className="input-box">
              <button className="btn btn-warning" onClick={populateInventory}>
                Populate Inventory
              </button>
              <button className="btn" onClick={printIndex}>
                Print Index
              </button>
              <button className="btn" onClick={rebuildIndex}>
                Rebuild Index
              </button>
              <button className="btn btn-danger" onClick={deleteAllFiles}>
                Delete All Files
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
};
