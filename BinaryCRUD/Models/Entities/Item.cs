using System;
using System.Text;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][Price:4bytes][ContentLength:2bytes][Content:variable]
// Total Size: 9 bytes + Content length
public class Item : InterfaceSerializable
{
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public string Content { get; set; } = string.Empty;
    public float Price { get; set; } = 0.0f;

    public byte[] ToBytes()
    {
        var contentBytes = Encoding.UTF8.GetBytes(Content);
        var idBytes = BitConverter.GetBytes(Id);
        var priceBytes = BitConverter.GetBytes(Price);

        var result = new byte[1 + sizeof(ushort) + sizeof(float) + sizeof(ushort) + contentBytes.Length];

        int offset = 0;

        // Write tombstone bit (1 byte)
        result[offset] = IsTombstone ? (byte)1 : (byte)0;
        offset += 1;

        // Write ID (2 bytes)
        idBytes.CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write price (4 bytes)
        priceBytes.CopyTo(result, offset);
        offset += sizeof(float);

        // Write content length (2 bytes)
        BitConverter.GetBytes((ushort)contentBytes.Length).CopyTo(result, offset);
        offset += sizeof(ushort);

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

        // Read ID (2 bytes)
        Id = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

        // Read price (4 bytes)
        Price = BitConverter.ToSingle(data, offset);
        offset += sizeof(float);

        // Read content length (2 bytes)
        var contentLength = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

        // Read content
        Content = Encoding.UTF8.GetString(data, offset, contentLength);
    }
}
