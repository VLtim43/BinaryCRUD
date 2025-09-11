using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][ItemCount:2bytes][ItemIds:2bytes*ItemCount][TotalPrice:4bytes][UserId:2bytes][AdditionalInfoLength:2bytes][AdditionalInfo:variable]
// Fixed Size: 13 + (ItemCount * 2) + AdditionalInfoLength bytes
// ItemId Format: [ItemId:2bytes] = 2 bytes each (multiple instances represent multiple items)
// String Format: [Length:2bytes][UTF-8 Data:variable] - Length=0 for null/empty strings
public class Order : InterfaceSerializable
{
    private static UserDAO? _userDAO;
    private static UserDAO UserDAO => _userDAO ??= new UserDAO();
    
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public List<ushort> ItemIds { get; set; } = new List<ushort>();
    public float TotalPrice { get; set; } = 0.0f;
    public ushort UserId { get; set; }
    public string? AdditionalInfo { get; set; } = null;

    // Method to get username from UserId
    public async Task<string> GetUsernameAsync()
    {
        try
        {
            var users = await UserDAO.GetAllUsersAsync();
            var user = users.FirstOrDefault(u => u.Id == UserId && !u.IsTombstone);
            return user?.Username ?? "Unknown User";
        }
        catch
        {
            return "Unknown User";
        }
    }

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
            // Note: This is synchronous for UI binding, username will be resolved asynchronously in UI
            var display = $"ID:{Id} - User:{UserId}";
            if (!string.IsNullOrEmpty(AdditionalInfo))
                display += $" ({AdditionalInfo})";
            return display;
        }
    }

    // Property for XAML binding - shows user information
    public string UserDisplayText
    {
        get { return $"User ID: {UserId}"; }
    }

    public byte[] ToBytes()
    {
        // Prepare string data
        byte[] additionalInfoBytes = string.IsNullOrEmpty(AdditionalInfo) ? Array.Empty<byte>() : Encoding.UTF8.GetBytes(AdditionalInfo);

        // Calculate total size: IsTombstone(1) + Id(2) + ItemCount(2) + ItemIds(2*count) + TotalPrice(4) + UserId(2) + AdditionalInfoLength(2) + AdditionalInfo
        var totalSize = 1 + sizeof(ushort) + sizeof(ushort) + (ItemIds.Count * sizeof(ushort)) + sizeof(float) + sizeof(ushort) + sizeof(ushort) + additionalInfoBytes.Length;
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

        // Write UserId (2 bytes)
        BitConverter.GetBytes(UserId).CopyTo(result, offset);
        offset += sizeof(ushort);

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

        // Read UserId (2 bytes)
        UserId = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

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
