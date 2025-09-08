using System;
using System.Text;

namespace BinaryCRUD.Models;

public class Item : InterfaceSerializable
{
    public string Content { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public decimal Price { get; set; } = 0.0m;

    public byte[] ToBytes()
    {
        var contentBytes = Encoding.UTF8.GetBytes(Content);
        var createdAtBytes = BitConverter.GetBytes(CreatedAt.Ticks);
        var priceBytes = decimal.GetBits(Price);
        var priceByteArray = new byte[16]; // decimal is 16 bytes (4 int32s)
        
        Buffer.BlockCopy(priceBytes, 0, priceByteArray, 0, 16);

        var result = new byte[contentBytes.Length + sizeof(long) + sizeof(int) + 16];

        // Write content length, then content, then timestamp, then price
        BitConverter.GetBytes(contentBytes.Length).CopyTo(result, 0);
        contentBytes.CopyTo(result, sizeof(int));
        createdAtBytes.CopyTo(result, sizeof(int) + contentBytes.Length);
        priceByteArray.CopyTo(result, sizeof(int) + contentBytes.Length + sizeof(long));

        return result;
    }

    public void FromBytes(byte[] data)
    {
        var contentLength = BitConverter.ToInt32(data, 0);
        Content = Encoding.UTF8.GetString(data, sizeof(int), contentLength);
        var ticks = BitConverter.ToInt64(data, sizeof(int) + contentLength);
        CreatedAt = new DateTime(ticks, DateTimeKind.Utc);
        
        // Read price (16 bytes starting after content + timestamp)
        var priceOffset = sizeof(int) + contentLength + sizeof(long);
        var priceBytes = new int[4];
        Buffer.BlockCopy(data, priceOffset, priceBytes, 0, 16);
        Price = new decimal(priceBytes);
    }
}