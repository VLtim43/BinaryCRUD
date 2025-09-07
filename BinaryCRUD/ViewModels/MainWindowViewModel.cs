using Avalonia;
using Avalonia.Controls.ApplicationLifetimes;
using BinaryCRUD.Models;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace BinaryCRUD.ViewModels;

public partial class MainWindowViewModel : ViewModelBase
{
    private readonly IOrderDAO _orderDAO;

    public MainWindowViewModel(IOrderDAO orderDAO)
    {
        _orderDAO = orderDAO;
    }

    [ObservableProperty]
    private string text = string.Empty;

    [RelayCommand]
    private async System.Threading.Tasks.Task SaveAsync()
    {
        if (string.IsNullOrEmpty(Text))
        {
            System.Console.WriteLine("[WARNING] Cannot save empty order");
            return;
        }

        try
        {
            System.Console.WriteLine($"[INFO] Creating order: '{Text}'");
            await _orderDAO.AddOrderAsync(Text);
            System.Console.WriteLine("[SUCCESS] Order created successfully");
            Text = string.Empty; // Clear the text field after saving
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to create order: {ex.Message}");
        }
    }

    [RelayCommand]
    private async System.Threading.Tasks.Task ReadOrdersAsync()
    {
        try
        {
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
        }
        catch (System.Exception ex)
        {
            System.Console.WriteLine($"[ERROR] Failed to read orders: {ex.Message}");
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
