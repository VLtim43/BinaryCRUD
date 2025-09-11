using System;
using System.Collections.Generic;
using System.Linq;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][ItemCount:2bytes][ItemIds:2bytes*ItemCount][TotalPrice:4bytes]
// Variable Size: 7 + (ItemCount * 2) bytes
// ItemId Format: [ItemId:2bytes] = 2 bytes each (multiple instances represent multiple items)
public class Order : InterfaceSerializable
{
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public List<ushort> ItemIds { get; set; } = new List<ushort>();
    public float TotalPrice { get; set; } = 0.0f;

    // Property for XAML binding - shows individual item IDs
    public string ItemsDisplayText
    {
        get
        {
            return string.Join(", ", ItemIds.Select(id => $"ID:{id}"));
        }
    }

    public byte[] ToBytes()
    {
        // Calculate total size: IsTombstone(1) + Id(2) + ItemCount(2) + ItemIds(2*count) + TotalPrice(4)
        var totalSize = 1 + sizeof(ushort) + sizeof(ushort) + (ItemIds.Count * sizeof(ushort)) + sizeof(float);
        var result = new byte[totalSize];
        int offset = 0;

        // Write tombstone bit (1 byte)
        result[offset] = IsTombstone ? (byte)1 : (byte)0;
        offset += 1;

        // Write ID (2 bytes)
        BitConverter.GetBytes(Id).CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write ItemCount (2 bytes)
        BitConverter.GetBytes((ushort)ItemIds.Count).CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write ItemIds (2 bytes each)
        foreach (var itemId in ItemIds)
        {
            BitConverter.GetBytes(itemId).CopyTo(result, offset);
            offset += sizeof(ushort);
        }

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

        // Read ItemCount (2 bytes)
        var itemCount = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

        // Read ItemIds (2 bytes each)
        ItemIds = new List<ushort>();
        for (int i = 0; i < itemCount; i++)
        {
            var itemId = BitConverter.ToUInt16(data, offset);
            ItemIds.Add(itemId);
            offset += sizeof(ushort);
        }

        // Read TotalPrice (4 bytes)
        TotalPrice = BitConverter.ToSingle(data, offset);
    }
}