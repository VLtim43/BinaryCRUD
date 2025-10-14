import "./App.scss";
import logo from "./assets/images/logo-universal.png";
import { AddItem, PrintBinaryFile } from "../wailsjs/go/main/App";
import { useState } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = (props: any) => {
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const updateItemText = (e: any) => setItemText(e.target.value);
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

  const handleKeyDown = (e: any) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addItem();
    }
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

  return (
    <>
      <div id="App">
        <img src={logo} id="logo" alt="logo" />
        <div id="result" className="result">
          {resultText}
        </div>
        <div id="input" className="input-box">
          <input
            id="name"
            className="input"
            onChange={updateItemText}
            onKeyDown={handleKeyDown}
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
      </div>
    </>
  );
}
