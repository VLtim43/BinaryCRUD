using System;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][ItemId:2bytes][TotalPrice:4bytes]
// Total Size: 9 bytes
public class Order : InterfaceSerializable
{
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public ushort ItemId { get; set; }
    public float TotalPrice { get; set; } = 0.0f;

    public byte[] ToBytes()
    {
        var result = new byte[1 + sizeof(ushort) + sizeof(ushort) + sizeof(float)];
        int offset = 0;

        // Write tombstone bit (1 byte)
        result[offset] = IsTombstone ? (byte)1 : (byte)0;
        offset += 1;

        // Write ID (2 bytes)
        BitConverter.GetBytes(Id).CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write ItemId (2 bytes)
        BitConverter.GetBytes(ItemId).CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write TotalPrice (4 bytes)
        BitConverter.GetBytes(TotalPrice).CopyTo(result, offset);

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

        // Read ItemId (2 bytes)
        ItemId = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

        // Read TotalPrice (4 bytes)
        TotalPrice = BitConverter.ToSingle(data, offset);
    }
}