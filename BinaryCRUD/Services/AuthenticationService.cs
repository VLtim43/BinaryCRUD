using System;
using System.Threading.Tasks;
using BinaryCRUD.Models;

namespace BinaryCRUD.Services;

public class AuthenticationService
{
    private readonly UserDAO _userDAO;

    public User? CurrentUser { get; private set; }
    public bool IsLoggedIn => CurrentUser != null;
    public bool IsAdmin => CurrentUser?.Role == UserRole.Admin;
    public bool IsUser => CurrentUser?.Role == UserRole.User;

    public event EventHandler<User?>? UserChanged;

    public AuthenticationService()
    {
        _userDAO = new UserDAO();
    }

    public async Task<bool> LoginAsync(string username, string password)
    {
        try
        {
            var user = await _userDAO.AuthenticateAsync(username, password);
            if (user != null)
            {
                CurrentUser = user;
                UserChanged?.Invoke(this, user);
                System.Console.WriteLine($"[AUTH] User '{username}' logged in as {user.Role}");
                return true;
            }
            else
            {
                System.Console.WriteLine($"[AUTH] Login failed for user '{username}'");
                return false;
            }
        }
        catch (Exception ex)
        {
            System.Console.WriteLine($"[AUTH] Login error: {ex.Message}");
            return false;
        }
    }

    public void Logout()
    {
        var previousUser = CurrentUser;
        CurrentUser = null;
        UserChanged?.Invoke(this, null);
        System.Console.WriteLine($"[AUTH] User '{previousUser?.Username}' logged out");
    }

    public async Task InitializeAsync()
    {
        await _userDAO.InitializeDefaultUsersAsync();
    }

    public bool HasPermission(string action)
    {
        if (!IsLoggedIn)
            return false;

        return action switch
        {
            "create_items" => IsAdmin,
            "delete_items" => IsAdmin,
            "populate_inventory" => IsAdmin,
            "delete_files" => IsAdmin,
            "create_orders" => IsUser, // Only regular users can create orders
            "delete_orders" => IsAdmin,
            "view_items" => IsLoggedIn,
            "view_orders" => IsAdmin, // Only admins can view order list
            _ => false,
        };
    }
}
