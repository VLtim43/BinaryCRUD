using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;

namespace BinaryCRUD.Services;

public partial class ToastMessage : ObservableObject
{
    [ObservableProperty]
    private string message = string.Empty;

    [ObservableProperty]
    private bool isVisible = true;

    public DateTime CreatedAt { get; init; } = DateTime.Now;
}

public partial class ToastService : ObservableObject
{
    [ObservableProperty]
    private ObservableCollection<ToastMessage> toasts = new();

    public void ShowSuccess(string message)
    {
        var toast = new ToastMessage { Message = $"✅ {message}" };
        Toasts.Add(toast);

        _ = Task.Run(async () =>
        {
            await Task.Delay(3000);
            await RemoveToast(toast);
        });
    }

    private async Task RemoveToast(ToastMessage toast)
    {
        await Task.Delay(100);
        if (Toasts.Contains(toast))
        {
            Toasts.Remove(toast);
        }
    }
}
