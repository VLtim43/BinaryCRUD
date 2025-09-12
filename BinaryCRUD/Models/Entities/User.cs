using System;
using System.Text;

namespace BinaryCRUD.Models;

// Binary Layout: [IsTombstone:1byte][Id:2bytes][UsernameLength:2bytes][Username:variable][PasswordLength:2bytes][Password:variable][Role:1byte]
// Variable Size: 6 + UsernameLength + PasswordLength bytes
// String Format: [Length:2bytes][UTF-8 Data:variable] - Length=0 for null/empty strings
public class User : InterfaceSerializable
{
    public ushort Id { get; set; }
    public bool IsTombstone { get; set; } = false;
    public string Username { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
    public UserRole Role { get; set; } = UserRole.User;

    public string DisplayText
    {
        get { return $"{Username} ({Role})"; }
    }

    public byte[] ToBytes()
    {
        // Convert strings to UTF-8 bytes
        var usernameBytes = string.IsNullOrEmpty(Username)
            ? new byte[0]
            : Encoding.UTF8.GetBytes(Username);
        var passwordBytes = string.IsNullOrEmpty(Password)
            ? new byte[0]
            : Encoding.UTF8.GetBytes(Password);

        // Calculate total size: IsTombstone(1) + Id(2) + UsernameLength(2) + Username + PasswordLength(2) + Password + Role(1)
        var totalSize =
            1
            + sizeof(ushort)
            + sizeof(ushort)
            + usernameBytes.Length
            + sizeof(ushort)
            + passwordBytes.Length
            + 1;
        var result = new byte[totalSize];
        int offset = 0;

        // Write tombstone bit (1 byte)
        result[offset] = IsTombstone ? (byte)1 : (byte)0;
        offset += 1;

        // Write ID (2 bytes)
        BitConverter.GetBytes(Id).CopyTo(result, offset);
        offset += sizeof(ushort);

        // Write Username length and data
        BitConverter.GetBytes((ushort)usernameBytes.Length).CopyTo(result, offset);
        offset += sizeof(ushort);
        if (usernameBytes.Length > 0)
        {
            usernameBytes.CopyTo(result, offset);
            offset += usernameBytes.Length;
        }

        // Write Password length and data
        BitConverter.GetBytes((ushort)passwordBytes.Length).CopyTo(result, offset);
        offset += sizeof(ushort);
        if (passwordBytes.Length > 0)
        {
            passwordBytes.CopyTo(result, offset);
            offset += passwordBytes.Length;
        }

        // Write Role (1 byte)
        result[offset] = (byte)Role;

        return result;
    }

    public void FromBytes(byte[] data)
    {
        int offset = 0;

        // Read tombstone (1 byte)
        IsTombstone = data[offset] == 1;
        offset += 1;

        // Read ID (2 bytes)
        Id = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);

        // Read Username length and data
        var usernameLength = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);
        if (usernameLength > 0)
        {
            Username = Encoding.UTF8.GetString(data, offset, usernameLength);
            offset += usernameLength;
        }
        else
        {
            Username = string.Empty;
        }

        // Read Password length and data
        var passwordLength = BitConverter.ToUInt16(data, offset);
        offset += sizeof(ushort);
        if (passwordLength > 0)
        {
            Password = Encoding.UTF8.GetString(data, offset, passwordLength);
            offset += passwordLength;
        }
        else
        {
            Password = string.Empty;
        }

        // Read Role (1 byte)
        Role = (UserRole)data[offset];
    }
}

public enum UserRole : byte
{
    User = 0,
    Admin = 1,
}
