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
}
