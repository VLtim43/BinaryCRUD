using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class FileOrderDAO : IOrderDAO, IDisposable
{
    private const string DataDirectory = "Data";
    private const string OrdersFileName = "orders.bin";
    private readonly string _filePath;
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    // create the folder and path
    public FileOrderDAO()
    {
        if (!Directory.Exists(DataDirectory))
            Directory.CreateDirectory(DataDirectory);

        _filePath = Path.Combine(DataDirectory, OrdersFileName);
    }

    public async Task AddOrderAsync(string content)
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                await CreateNewFileWithHeaderAsync();
            }

            // Read current header
            var header = await ReadHeaderAsync();
            var previousCount = header.Count;
            header.Count++;
            header.LastUpdated = DateTime.UtcNow;

            Console.WriteLine($"[DAO] Updating header: {previousCount} > {header.Count} orders");

            await UpdateHeaderAsync(header);

            await AppendOrderToFileAsync(content);

            Console.WriteLine($"[DAO] Order appended to: {_filePath}");
        }
        finally
        {
            _fileLock.Release();
        }
    }

    private async Task CreateNewFileWithHeaderAsync()
    {
        var header = new OrderHeader { Count = 0 };
        using var stream = new FileStream(_filePath, FileMode.Create, FileAccess.Write);
        await WriteHeaderAsync(stream, header);
    }

    private async Task<OrderHeader> ReadHeaderAsync()
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.Read);
        return await ReadHeaderFromStreamAsync(stream);
    }

    private async Task UpdateHeaderAsync(OrderHeader header)
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.ReadWrite);
        stream.Seek(0, SeekOrigin.Begin); // Go to beginning
        await WriteHeaderAsync(stream, header);
    }

    private async Task AppendOrderToFileAsync(string content)
    {
        using var stream = new FileStream(_filePath, FileMode.Open, FileAccess.Write);
        stream.Seek(0, SeekOrigin.End); // Go to end of file
        await WriteOrderAsync(stream, content);
    }

    private async Task<OrderHeader> ReadHeaderFromStreamAsync(Stream stream)
    {
        var buffer = new byte[GetHeaderSize()];
        await stream.ReadExactlyAsync(buffer);

        var count = BitConverter.ToInt32(buffer, 0);
        var ticks = BitConverter.ToInt64(buffer, 4);

        return new OrderHeader
        {
            Count = count,
            LastUpdated = new DateTime(ticks, DateTimeKind.Utc),
        };
    }

    private async Task WriteHeaderAsync(Stream stream, OrderHeader header)
    {
        var buffer = new byte[GetHeaderSize()];
        BitConverter.GetBytes(header.Count).CopyTo(buffer, 0);
        BitConverter.GetBytes(header.LastUpdated.Ticks).CopyTo(buffer, 4);

        await stream.WriteAsync(buffer, 0, buffer.Length);
    }

    private async Task WriteOrderAsync(Stream stream, string content)
    {
        var contentBytes = Encoding.UTF8.GetBytes(content);
        var lengthBytes = BitConverter.GetBytes(contentBytes.Length);

        await stream.WriteAsync(lengthBytes, 0, lengthBytes.Length);
        await stream.WriteAsync(contentBytes, 0, contentBytes.Length);
    }

    private static int GetHeaderSize() => sizeof(int) + sizeof(long); // Count + LastUpdated ticks

    protected virtual void Dispose(bool disposing)
    {
        if (disposing)
        {
            _fileLock?.Dispose();
            if (File.Exists(_filePath + ".tmp"))
                File.Delete(_filePath + ".tmp");
        }
    }

    public void Dispose()
    {
        Dispose(true);
        GC.SuppressFinalize(this);
    }
}
