using System;
using System.Text;

namespace BinaryCRUD.Models;

public class Order : InterfaceSerializable
{
    public string Content { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public byte[] ToBytes()
    {
        var contentBytes = Encoding.UTF8.GetBytes(Content);
        var createdAtBytes = BitConverter.GetBytes(CreatedAt.Ticks);

        var result = new byte[contentBytes.Length + sizeof(long) + sizeof(int)];

        // Write content length, then content, then timestamp
        BitConverter.GetBytes(contentBytes.Length).CopyTo(result, 0);
        contentBytes.CopyTo(result, sizeof(int));
        createdAtBytes.CopyTo(result, sizeof(int) + contentBytes.Length);

        return result;
    }

    public void FromBytes(byte[] data)
    {
        var contentLength = BitConverter.ToInt32(data, 0);
        Content = Encoding.UTF8.GetString(data, sizeof(int), contentLength);
        var ticks = BitConverter.ToInt64(data, sizeof(int) + contentLength);
        CreatedAt = new DateTime(ticks, DateTimeKind.Utc);
    }
}
