using Avalonia;
using Avalonia.Controls.ApplicationLifetimes;
using BinaryCRUD.Models;
using BinaryCRUD.Services;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System;
using System.Collections.Generic;
using System.Linq;

namespace BinaryCRUD.ViewModels;

public enum InputMode
{
    Orders,
    Users,
    Items
}

public partial class MainWindowViewModel : ViewModelBase
{
    private readonly OrderDAO _orderDAO;
    private readonly UserDAO _userDAO;
    private readonly ItemDAO _itemDAO;
    public ToastService ToastService { get; }

    public MainWindowViewModel(OrderDAO orderDAO, UserDAO userDAO, ItemDAO itemDAO)
    {
        _orderDAO = orderDAO;
        _userDAO = userDAO;
        _itemDAO = itemDAO;
        ToastService = new ToastService();
        
        AvailableModes = Enum.GetValues<InputMode>().ToList();
        SelectedMode = InputMode.Orders;
    }

    [ObservableProperty]
    private string text = string.Empty;
    
    [ObservableProperty]
    private InputMode selectedMode = InputMode.Orders;
    
    [ObservableProperty]
    private string saveButtonText = "Save Order";
    
    [ObservableProperty]
    private string readButtonText = "Read Orders";
    
    [ObservableProperty]
    private string inputPlaceholder = "Enter order details...";
    
    [ObservableProperty]
    private decimal priceInput = 0.0m;
    
    [ObservableProperty]
    private bool isPriceVisible = false;
    
    public List<InputMode> AvailableModes { get; }
    
    partial void OnSelectedModeChanged(InputMode value)
    {
        UpdateUILabels();
    }
    
    private void UpdateUILabels()
    {
        switch (SelectedMode)
        {
            case InputMode.Orders:
                SaveButtonText = "Save Order";
                ReadButtonText = "Read Orders";
                InputPlaceholder = "Enter order details...";
                IsPriceVisible = false;
                break;
            case InputMode.Users:
                SaveButtonText = "Save User";
                ReadButtonText = "Read Users";
                InputPlaceholder = "Enter user details...";
                IsPriceVisible = false;
                break;
            case InputMode.Items:
                SaveButtonText = "Save Item";
                ReadButtonText = "Read Items";
                InputPlaceholder = "Enter item name...";
                IsPriceVisible = true;
                break;
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task SaveAsync()
    {
        if (string.IsNullOrEmpty(Text))
        {
            System.Console.WriteLine($"[WARNING] Cannot save empty {SelectedMode.ToString().ToLower().TrimEnd('s')}");
            return;
        }

        try
        {
            switch (SelectedMode)
            {
                case InputMode.Orders:
                    System.Console.WriteLine($"[INFO] Creating order: '{Text}'");
                    await _orderDAO.AddOrderAsync(Text);
                    ToastService.ShowSuccess($"Order '{Text}' created successfully");
                    break;
                case InputMode.Users:
                    System.Console.WriteLine($"[INFO] Creating user: '{Text}'");
                    await _userDAO.AddUserAsync(Text);
                    ToastService.ShowSuccess($"User '{Text}' created successfully");
                    break;
                case InputMode.Items:
                    System.Console.WriteLine($"[INFO] Creating item: '{Text}' with price: ${PriceInput}");
                    await _itemDAO.AddItemAsync(Text, PriceInput);
                    ToastService.ShowSuccess($"Item '{Text}' (${PriceInput:F2}) created successfully");
                    break;
            }
            Text = string.Empty;
            PriceInput = 0.0m; // Clear input fields after saving
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create {SelectedMode.ToString().ToLower().TrimEnd('s')}: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadOrdersAsync()
    {
        try
        {
            switch (SelectedMode)
            {
                case InputMode.Orders:
                    System.Console.WriteLine("[INFO] Reading orders from file...");
                    var orders = await _orderDAO.GetAllOrdersAsync();
                    
                    if (orders.Count == 0)
                    {
                        System.Console.WriteLine("[INFO] No orders found");
                        return;
                    }
                    
                    for (int i = 0; i < orders.Count; i++)
                    {
                        var order = orders[i];
                        System.Console.WriteLine(
                            $"[ORDER {i + 1}] Content: '{order.Content}', Created: {order.CreatedAt:yyyy-MM-dd HH:mm:ss}"
                        );
                    }
                    
                    System.Console.WriteLine($"[INFO] Found {orders.Count} orders total");
                    break;
                    
                case InputMode.Users:
                    System.Console.WriteLine("[INFO] Reading users from file...");
                    var users = await _userDAO.GetAllUsersAsync();
                    
                    if (users.Count == 0)
                    {
                        System.Console.WriteLine("[INFO] No users found");
                        return;
                    }
                    
                    for (int i = 0; i < users.Count; i++)
                    {
                        var user = users[i];
                        System.Console.WriteLine(
                            $"[USER {i + 1}] Content: '{user.Content}', Created: {user.CreatedAt:yyyy-MM-dd HH:mm:ss}"
                        );
                    }
                    
                    System.Console.WriteLine($"[INFO] Found {users.Count} users total");
                    break;
                    
                case InputMode.Items:
                    System.Console.WriteLine("[INFO] Reading items from file...");
                    var items = await _itemDAO.GetAllItemsAsync();
                    
                    if (items.Count == 0)
                    {
                        System.Console.WriteLine("[INFO] No items found");
                        return;
                    }
                    
                    for (int i = 0; i < items.Count; i++)
                    {
                        var item = items[i];
                        System.Console.WriteLine(
                            $"[ITEM {i + 1}] Name: '{item.Content}', Price: ${item.Price:F2}, Created: {item.CreatedAt:yyyy-MM-dd HH:mm:ss}"
                        );
                    }
                    
                    System.Console.WriteLine($"[INFO] Found {items.Count} items total");
                    break;
            }
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read {SelectedMode.ToString().ToLower()}: {ex.Message}");
        }
    }

    [RelayCommand]
    private static void Quit()
    {
        System.Console.WriteLine("[INFO] Quitting application...");
        if (
            Application.Current?.ApplicationLifetime
            is IClassicDesktopStyleApplicationLifetime desktop
        )
        {
            desktop.Shutdown();
        }
    }
}
