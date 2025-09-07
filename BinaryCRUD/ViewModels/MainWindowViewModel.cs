using System.IO;
using System.Text;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace BinaryCRUD.ViewModels;

public partial class MainWindowViewModel : ViewModelBase
{
    [ObservableProperty]
    private string text = string.Empty;

    [RelayCommand]
    private void Save()
    {
        SaveToBinaryFile();
    }

    private void SaveToBinaryFile()
    {
        if (string.IsNullOrEmpty(Text))
            return;

        byte[] data = Encoding.UTF8.GetBytes(Text);
        File.WriteAllBytes("data.bin", data);
    }
}
