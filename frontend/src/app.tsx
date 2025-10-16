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
  GetItems,
  CreateOrder,
  PrintOrderBinaryFile,
  GetOrderByID,
} from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";
import { useState, useEffect } from "preact/hooks";
import { h, Fragment } from "preact";

export const App = () => {
  const [activeTab, setActiveTab] = useState<"item" | "order" | "debug">("item");
  const [itemSubTab, setItemSubTab] = useState<"create" | "read" | "delete">("create");
  const [orderSubTab, setOrderSubTab] = useState<"create" | "read" | "delete">("create");
  const [resultText, setResultText] = useState("Enter item text below 👇");
  const [itemText, setItemText] = useState("");
  const [recordId, setRecordId] = useState("");
  const [deleteRecordId, setDeleteRecordId] = useState("");
  const [orderRecordId, setOrderRecordId] = useState("");
  const [jsonFilePath, setJsonFilePath] = useState("inventory.json");
  const [allItems, setAllItems] = useState<Array<{ id: number; name: string }>>(
    []
  );
  const [selectedItemId, setSelectedItemId] = useState<number | "">("");
  const [cart, setCart] = useState<Array<{ id: number; name: string; quantity: number }>>(
    []
  );
  const [cartSortBy, setCartSortBy] = useState<"id" | "name">("id");
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
  const updateOrderRecordId = (e: any) => {
    const value = e.target.value;
    // Only allow empty string or non-negative integers
    if (value === "" || /^\d+$/.test(value)) {
      setOrderRecordId(value);
    }
  };
  const updateJsonFilePath = (e: any) => setJsonFilePath(e.target.value);
  const updateResultText = (result: string) => setResultText(result);

  // Get default text for each tab/subtab combination
  const getDefaultText = (
    tab: "item" | "order" | "debug",
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

    if (tab === "order") {
      switch (subTab) {
        case "create":
          return "Select items for your order";
        case "read":
          return "Enter an order ID to fetch 👇";
        case "delete":
          return "Enter an order ID to delete 👇";
        default:
          return "";
      }
    }

    return "";
  };

  // Handle main tab changes
  const handleTabChange = (tab: "item" | "order" | "debug") => {
    setActiveTab(tab);
    if (tab === "item") {
      setResultText(getDefaultText(tab, itemSubTab));
    } else if (tab === "order") {
      setResultText(getDefaultText(tab, orderSubTab));
    } else {
      setResultText(getDefaultText(tab));
    }
  };

  // Handle item subtab changes
  const handleItemSubTabChange = (subTab: "create" | "read" | "delete") => {
    setItemSubTab(subTab);
    setResultText(getDefaultText("item", subTab));
  };

  // Handle order subtab changes
  const handleOrderSubTabChange = (subTab: "create" | "read" | "delete") => {
    setOrderSubTab(subTab);
    setResultText(getDefaultText("order", subTab));
  };

  // Load all items when order create subtab becomes active
  useEffect(() => {
    if (activeTab === "order" && orderSubTab === "create") {
      GetItems()
        .then((items: Array<{ id: number; name: string }>) => {
          setAllItems(items);
        })
        .catch((err: any) => {
          updateResultText(`Error loading items: ${err}`);
        });
    }
  }, [activeTab, orderSubTab]);

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
    // Validate file path
    if (!jsonFilePath || jsonFilePath.trim().length === 0) {
      updateResultText("Error: Please enter a file path");
      return;
    }

    PopulateInventory(jsonFilePath)
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

  const addToCart = () => {
    if (selectedItemId === "") {
      updateResultText("Error: Please select an item");
      return;
    }

    const selectedItem = allItems.find((item) => item.id === selectedItemId);
    if (!selectedItem) {
      updateResultText("Error: Item not found");
      return;
    }

    const existingCartItem = cart.find((item) => item.id === selectedItemId);
    if (existingCartItem) {
      setCart(
        cart.map((item) =>
          item.id === selectedItemId
            ? { ...item, quantity: item.quantity + 1 }
            : item
        )
      );
      updateResultText(`Added another ${selectedItem.name} to cart`);
    } else {
      setCart([...cart, { ...selectedItem, quantity: 1 }]);
      updateResultText(`Added ${selectedItem.name} to cart`);
    }
    setSelectedItemId("");
  };

  const removeFromCart = (id: number) => {
    const item = cart.find((item) => item.id === id);
    if (item && item.quantity > 1) {
      setCart(
        cart.map((item) =>
          item.id === id ? { ...item, quantity: item.quantity - 1 } : item
        )
      );
    } else {
      setCart(cart.filter((item) => item.id !== id));
    }
  };

  const clearCart = () => {
    setCart([]);
    updateResultText("Cart cleared");
  };

  const toggleCartSort = () => {
    setCartSortBy(cartSortBy === "id" ? "name" : "id");
  };

  const getSortedCart = () => {
    const sortedCart = [...cart];
    if (cartSortBy === "id") {
      sortedCart.sort((a, b) => a.id - b.id);
    } else {
      sortedCart.sort((a, b) => a.name.localeCompare(b.name));
    }
    return sortedCart;
  };

  const createOrder = () => {
    if (cart.length === 0) {
      updateResultText("Error: Cart is empty");
      return;
    }

    // Convert cart items to the format expected by the backend
    const orderItems = cart.map((item) => ({
      itemId: item.id,
      quantity: item.quantity,
    }));

    CreateOrder(orderItems)
      .then(() => {
        updateResultText(`Order created successfully with ${cart.length} items`);
        setCart([]);
      })
      .catch((err: any) => {
        updateResultText(`Error creating order: ${err}`);
      });
  };

  const printOrderFile = () => {
    PrintOrderBinaryFile()
      .then(() => {
        updateResultText("Order binary file printed to application console!");
      })
      .catch((err: any) => {
        updateResultText(`Error: ${err}`);
      });
  };

  const getOrderById = () => {
    // Validate input before sending to backend
    if (!orderRecordId || orderRecordId.trim().length === 0) {
      updateResultText("Error: Please enter an order ID");
      return;
    }

    // Parse the record ID as a number
    const id = parseInt(orderRecordId, 10);
    if (isNaN(id) || id < 0) {
      updateResultText("Error: Order ID must be a valid non-negative number");
      return;
    }

    GetOrderByID(id)
      .then((orderDetails: string) => {
        updateResultText(orderDetails);
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
          className={`tab ${activeTab === "order" ? "active" : ""}`}
          onClick={() => handleTabChange("order")}
        >
          Order
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

      {activeTab === "order" && (
        <div className="sub_tabs">
          <button
            className={`tab ${orderSubTab === "create" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("create")}
          >
            Create
          </button>
          <button
            className={`tab ${orderSubTab === "read" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("read")}
          >
            Read
          </button>
          <button
            className={`tab ${orderSubTab === "delete" ? "active" : ""}`}
            onClick={() => handleOrderSubTabChange("delete")}
          >
            Delete
          </button>
        </div>
      )}

      <div id="App">

        {!(activeTab === "order" && orderSubTab === "create") && (
          <img src={logo} id="logo" alt="logo" />
        )}
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

        {activeTab === "order" && orderSubTab === "create" && (
          <div id="order-section">
            <div className="cart-container">
              <div className="cart-header">
                <div className="cart-title">
                  <h3>Shopping Cart</h3>
                  <button className="btn btn-sort" onClick={toggleCartSort}>
                    Sort: {cartSortBy === "id" ? "ID" : "A-Z"}
                  </button>
                </div>
                {cart.length > 0 && (
                  <button className="btn btn-secondary" onClick={clearCart}>
                    Clear Cart
                  </button>
                )}
              </div>
              <div className="cart-items">
                {cart.length > 0 ? (
                  getSortedCart().map((item) => (
                    <div key={item.id} className="cart-item">
                      <div className="cart-item-info">
                        <span className="cart-item-name">{item.name}</span>
                        <span className="cart-item-id">ID: {item.id}</span>
                      </div>
                      <div className="cart-item-controls">
                        <span className="cart-item-quantity">x{item.quantity}</span>
                        <button
                          className="btn btn-small btn-danger"
                          onClick={() => removeFromCart(item.id)}
                        >
                          -
                        </button>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="cart-empty">Cart is empty</div>
                )}
              </div>
              <div className="cart-footer">
                <select
                  className="input cart-select"
                  value={selectedItemId}
                  onChange={(e: any) =>
                    setSelectedItemId(
                      e.target.value === "" ? "" : Number(e.target.value)
                    )
                  }
                >
                  <option value="">Select an item</option>
                  {allItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} (ID: {item.id})
                    </option>
                  ))}
                </select>
                <button className="btn" onClick={addToCart}>
                  Add to Cart
                </button>
                <button className="btn btn-primary" onClick={createOrder}>
                  Create Order
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === "order" && orderSubTab === "read" && (
          <div id="order-read" className="input-box">
            <input
              id="order-record-id"
              className="input"
              onChange={updateOrderRecordId}
              autoComplete="off"
              name="order-record-id"
              placeholder="Enter Order ID"
              value={orderRecordId}
            />
            <button className="btn" onClick={getOrderById}>
              Get Order
            </button>
            <button className="btn" onClick={printOrderFile}>
              Print All
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
