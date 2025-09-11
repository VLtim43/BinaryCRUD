using System;
using System.Collections.ObjectModel;
using System.IO;
using System.Text;
using CommunityToolkit.Mvvm.ComponentModel;

namespace BinaryCRUD.Services;

public partial class ConsoleService : ObservableObject
{
    private readonly MultiTextWriter _multiTextWriter;
    private readonly TextWriter _originalOutput;
    
    [ObservableProperty]
    private ObservableCollection<ConsoleEntry> entries = new();
    
    public ConsoleService()
    {
        _originalOutput = Console.Out;
        _multiTextWriter = new MultiTextWriter(_originalOutput);
        Console.SetOut(_multiTextWriter);
        
        _multiTextWriter.TextWritten += OnTextWritten;
    }
    
    private void OnTextWritten(object? sender, string text)
    {
        if (string.IsNullOrWhiteSpace(text)) return;
        
        var entry = new ConsoleEntry
        {
            Timestamp = DateTime.Now,
            Message = text.Trim(),
            Level = DetermineLogLevel(text)
        };
        
        Avalonia.Threading.Dispatcher.UIThread.InvokeAsync(() =>
        {
            Entries.Add(entry);
            
            // Keep only the last 1000 entries to prevent memory issues
            while (Entries.Count > 1000)
            {
                Entries.RemoveAt(0);
            }
        });
    }
    
    private ConsoleEntryLevel DetermineLogLevel(string text)
    {
        var upperText = text.ToUpperInvariant();
        
        if (upperText.Contains("[ERROR]") || upperText.Contains("ERROR"))
            return ConsoleEntryLevel.Error;
        if (upperText.Contains("[WARNING]") || upperText.Contains("WARNING"))
            return ConsoleEntryLevel.Warning;
        if (upperText.Contains("[INFO]") || upperText.Contains("INFO"))
            return ConsoleEntryLevel.Info;
        if (upperText.Contains("[HEADER]"))
            return ConsoleEntryLevel.Header;
        
        return ConsoleEntryLevel.Default;
    }
    
    public void Clear()
    {
        Entries.Clear();
    }
    
    public void Dispose()
    {
        Console.SetOut(_originalOutput);
        _multiTextWriter.Dispose();
    }
}

public class ConsoleEntry
{
    public DateTime Timestamp { get; set; }
    public string Message { get; set; } = string.Empty;
    public ConsoleEntryLevel Level { get; set; }
    
    public string TimestampString => Timestamp.ToString("HH:mm:ss.fff");
    
    public string LevelColor => Level switch
    {
        ConsoleEntryLevel.Error => "#FF6B6B",
        ConsoleEntryLevel.Warning => "#FFB347",
        ConsoleEntryLevel.Info => "#4ECDC4",
        ConsoleEntryLevel.Header => "#9B59B6",
        _ => "#FFFFFF"
    };
}

public enum ConsoleEntryLevel
{
    Default,
    Info,
    Warning,
    Error,
    Header
}

public class MultiTextWriter : TextWriter
{
    private readonly TextWriter[] _writers;
    
    public event EventHandler<string>? TextWritten;
    
    public MultiTextWriter(params TextWriter[] writers)
    {
        _writers = writers;
    }
    
    public override Encoding Encoding => _writers[0].Encoding;
    
    public override void Write(char value)
    {
        foreach (var writer in _writers)
        {
            writer.Write(value);
        }
    }
    
    public override void WriteLine(string? value)
    {
        foreach (var writer in _writers)
        {
            writer.WriteLine(value);
        }
        
        if (!string.IsNullOrEmpty(value))
        {
            TextWritten?.Invoke(this, value);
        }
    }
    
    public override void Write(string? value)
    {
        foreach (var writer in _writers)
        {
            writer.Write(value);
        }
        
        if (!string.IsNullOrEmpty(value))
        {
            TextWritten?.Invoke(this, value);
        }
    }
    
    protected override void Dispose(bool disposing)
    {
        if (disposing)
        {
            foreach (var writer in _writers)
            {
                writer.Dispose();
            }
        }
        base.Dispose(disposing);
    }
}