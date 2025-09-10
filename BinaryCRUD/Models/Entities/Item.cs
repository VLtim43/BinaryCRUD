using System;
using System.Text;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:8bytes][Price:16bytes][ContentLength:4bytes][Content:variable]
// Total Size: 29 bytes + Content length
public class Item : InterfaceSerializable
{
    public long Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public string Content { get; set; } = string.Empty;
    public decimal Price { get; set; } = 0.0m;

    public byte[] ToBytes()
    {
        var contentBytes = Encoding.UTF8.GetBytes(Content);
        var idBytes = BitConverter.GetBytes(Id);
        var priceBytes = decimal.GetBits(Price);
        var priceByteArray = new byte[16]; // decimal is 16 bytes

        Buffer.BlockCopy(priceBytes, 0, priceByteArray, 0, 16);

        var result = new byte[1 + sizeof(long) + sizeof(int) + contentBytes.Length + 16];

        int offset = 0;

        // Write tombstone bit (1 byte)
        result[offset] = IsTombstone ? (byte)1 : (byte)0;
        offset += 1;

        // Write ID (8 bytes)
        idBytes.CopyTo(result, offset);
        offset += sizeof(long);

        // Write price (16 bytes)
        priceByteArray.CopyTo(result, offset);
        offset += 16;

        // Write content length (4 bytes)
        BitConverter.GetBytes(contentBytes.Length).CopyTo(result, offset);
        offset += sizeof(int);

        // Write content
        contentBytes.CopyTo(result, offset);

        return result;
    }

    public void FromBytes(byte[] data)
    {
        int offset = 0;

        // Read tombstone bit (1 byte)
        IsTombstone = data[offset] == 1;
        offset += 1;

        // Read ID (8 bytes)
        Id = BitConverter.ToInt64(data, offset);
        offset += sizeof(long);

        // Read price (16 bytes)
        var priceBytes = new int[4];
        Buffer.BlockCopy(data, offset, priceBytes, 0, 16);
        Price = new decimal(priceBytes);
        offset += 16;

        // Read content length (4 bytes)
        var contentLength = BitConverter.ToInt32(data, offset);
        offset += sizeof(int);

        // Read content
        Content = Encoding.UTF8.GetString(data, offset, contentLength);
    }
}
