using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][ItemCount:2bytes][ItemIds:2bytes*ItemCount][TotalPrice:4bytes][NameLength:2bytes][Name:variable][AdditionalInfoLength:2bytes][AdditionalInfo:variable]
// Variable Size: 11 + (ItemCount * 2) + NameLength + AdditionalInfoLength bytes
// ItemId Format: [ItemId:2bytes] = 2 bytes each (multiple instances represent multiple items)
// String Format: [Length:2bytes][UTF-8 Data:variable] - Length=0 for null/empty strings
public class Order : InterfaceSerializable
{
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public List<ushort> ItemIds { get; set; } = new List<ushort>();
    public float TotalPrice { get; set; } = 0.0f;
    public string Name { get; set; } = string.Empty;
    public string? AdditionalInfo { get; set; } = null;

    // Property for XAML binding - shows individual item IDs
    public string ItemsDisplayText
    {
        get { return string.Join(", ", ItemIds.Select(id => $"ID:{id}")); }
    }

    // Property for XAML binding - checks if additional info should be visible
    public bool HasAdditionalInfo
    {
        get { return !string.IsNullOrEmpty(AdditionalInfo); }
    }

    // Property for XAML binding - shows display information
    public string DisplayText
    {
        get 
        { 
            var display = $"ID:{Id} - {Name}";
            if (!string.IsNullOrEmpty(AdditionalInfo))
                display += $" ({AdditionalInfo})";
            return display;
        }
    }

    public byte[] ToBytes()
    {
        // Prepare string data
        byte[] nameBytes = string.IsNullOrEmpty(Name) ? Array.Empty<byte>() : Encoding.UTF8.GetBytes(Name);
        byte[] additionalInfoBytes = string.IsNullOrEmpty(AdditionalInfo) ? Array.Empty<byte>() : Encoding.UTF8.GetBytes(AdditionalInfo);

        // Calculate total size: IsTombstone(1) + Id(2) + ItemCount(2) + ItemIds(2*count) + TotalPrice(4) + NameLength(2) + Name + AdditionalInfoLength(2) + AdditionalInfo
        var totalSize = 1 + sizeof(ushort) + sizeof(ushort) + (ItemIds.Count * sizeof(ushort)) + sizeof(float) + sizeof(ushort) + nameBytes.Length + sizeof(ushort) + additionalInfoBytes.Length;
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
        offset += sizeof(float);

        // Write Name length (2 bytes) and data
        BitConverter.GetBytes((ushort)nameBytes.Length).CopyTo(result, offset);
        offset += sizeof(ushort);
        if (nameBytes.Length > 0)
        {
            nameBytes.CopyTo(result, offset);
            offset += nameBytes.Length;
        }

        // Write AdditionalInfo length (2 bytes) and data
        BitConverter.GetBytes((ushort)additionalInfoBytes.Length).CopyTo(result, offset);
        offset += sizeof(ushort);
        if (additionalInfoBytes.Length > 0)
        {
            additionalInfoBytes.CopyTo(result, offset);
        }

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
        offset += sizeof(float);

        // Read Name length (2 bytes) and data
        var nameLength = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);
        if (nameLength > 0)
        {
            Name = Encoding.UTF8.GetString(data, offset, nameLength);
            offset += nameLength;
        }
        else
        {
            Name = string.Empty;
        }

        // Read AdditionalInfo length (2 bytes) and data
        var additionalInfoLength = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);
        if (additionalInfoLength > 0)
        {
            AdditionalInfo = Encoding.UTF8.GetString(data, offset, additionalInfoLength);
        }
        else
        {
            AdditionalInfo = null;
        }
    }
}
