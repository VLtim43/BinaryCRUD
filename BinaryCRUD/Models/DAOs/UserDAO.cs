using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace BinaryCRUD.Models;

public class UserDAO : FileBinaryDAO<User>
{
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public UserDAO()
        : base("users.bin") { }

    public async Task<User?> AuthenticateAsync(string username, string password)
    {
        var users = await GetAllUsersAsync();
        var user = users.FirstOrDefault(u => !u.IsTombstone && 
                                            u.Username.Equals(username, StringComparison.OrdinalIgnoreCase) && 
                                            u.Password == password);
        return user;
    }

    public async Task<User?> GetUserByUsernameAsync(string username)
    {
        var users = await GetAllUsersAsync();
        return users.FirstOrDefault(u => !u.IsTombstone && 
                                        u.Username.Equals(username, StringComparison.OrdinalIgnoreCase));
    }

    public async Task AddUserAsync(string username, string password, UserRole role = UserRole.User)
    {
        // Check if user already exists
        var existingUser = await GetUserByUsernameAsync(username);
        if (existingUser != null)
        {
            throw new InvalidOperationException($"User '{username}' already exists");
        }

        var user = new User 
        { 
            Username = username, 
            Password = password, 
            Role = role 
        };
        await AddAsync(user);
    }

    public async Task<List<User>> GetAllUsersAsync()
    {
        return await GetAllAsync();
    }

    public async Task<FileHeader?> ReadHeaderAsync()
    {
        return await GetHeaderAsync();
    }

    public async Task DeleteUserAsync(ushort userId)
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(_filePath))
            {
                throw new InvalidOperationException("File does not exist");
            }

            var users = await GetAllUsersAsync();
            var userToDelete = users.FirstOrDefault(u => u.Id == userId);

            if (userToDelete == null)
            {
                throw new InvalidOperationException($"User with ID {userId} not found");
            }

            if (userToDelete.IsTombstone)
            {
                throw new InvalidOperationException($"User with ID {userId} is already deleted");
            }

            userToDelete.IsTombstone = true;

            await RewriteFileAsync(users);
        }
        finally
        {
            _fileLock.Release();
        }
    }

    public async Task DeleteUserByUsernameAsync(string username)
    {
        var user = await GetUserByUsernameAsync(username);
        if (user == null)
        {
            throw new InvalidOperationException($"User '{username}' not found");
        }
        await DeleteUserAsync(user.Id);
    }

    public async Task InitializeDefaultUsersAsync()
    {
        var users = await GetAllUsersAsync();
        
        // Create default admin if no users exist
        if (!users.Any(u => !u.IsTombstone))
        {
            await AddUserAsync("admin", "admin", UserRole.Admin);
            await AddUserAsync("user", "user", UserRole.User);
            
            // Log the created users with their IDs
            var createdUsers = await GetAllUsersAsync();
            foreach (var user in createdUsers.Where(u => !u.IsTombstone))
            {
                System.Console.WriteLine($"[UserDAO] Created user: {user.Username} (ID: {user.Id}, Role: {user.Role})");
            }
        }
    }

    private async Task RewriteFileAsync(List<User> users)
    {
        var tempFilePath = _filePath + ".tmp";

        using (var stream = new FileStream(tempFilePath, FileMode.Create, FileAccess.Write))
        {
            var header = new FileHeader { Count = users.Count };

            var headerBuffer = new byte[4];
            BitConverter.GetBytes(header.Count).CopyTo(headerBuffer, 0);
            await stream.WriteAsync(headerBuffer, 0, headerBuffer.Length);

            foreach (var user in users)
            {
                var entityBytes = user.ToBytes();
                var lengthBytes = BitConverter.GetBytes(entityBytes.Length);
                await stream.WriteAsync(lengthBytes, 0, lengthBytes.Length);
                await stream.WriteAsync(entityBytes, 0, entityBytes.Length);
            }
        }

        File.Move(tempFilePath, _filePath, true);
    }

    public async Task DeleteFileAsync()
    {
        await _fileLock.WaitAsync();
        try
        {
            if (File.Exists(_filePath))
            {
                File.Delete(_filePath);
                System.Console.WriteLine($"[{GetType().Name}] File deleted: {_filePath}");
            }
            else
            {
                System.Console.WriteLine($"[{GetType().Name}] File does not exist: {_filePath}");
            }
        }
        finally
        {
            _fileLock.Release();
        }
    }
}