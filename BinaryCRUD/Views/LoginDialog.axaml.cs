using System;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Interactivity;
using BinaryCRUD.Services;

namespace BinaryCRUD.Views;

public partial class LoginDialog : Window
{
    private readonly AuthenticationService _authService;
    
    public bool LoginSuccessful { get; private set; } = false;

    public LoginDialog(AuthenticationService authService)
    {
        _authService = authService;
        InitializeComponent();
        
        var loginButton = this.FindControl<Button>("LoginButton");
        var cancelButton = this.FindControl<Button>("CancelButton");
        
        if (loginButton != null)
            loginButton.Click += LoginButton_Click;
        if (cancelButton != null)
            cancelButton.Click += CancelButton_Click;
            
        // Focus on username textbox
        var usernameTextBox = this.FindControl<TextBox>("UsernameTextBox");
        if (usernameTextBox != null)
        {
            usernameTextBox.Focus();
        }
    }

    private async void LoginButton_Click(object? sender, RoutedEventArgs e)
    {
        var usernameTextBox = this.FindControl<TextBox>("UsernameTextBox");
        var passwordTextBox = this.FindControl<TextBox>("PasswordTextBox");
        var loginButton = this.FindControl<Button>("LoginButton");
        
        if (usernameTextBox == null || passwordTextBox == null || loginButton == null)
            return;

        var username = usernameTextBox.Text ?? "";
        var password = passwordTextBox.Text ?? "";

        if (string.IsNullOrWhiteSpace(username) || string.IsNullOrWhiteSpace(password))
        {
            await ShowErrorMessage("Please enter both username and password.");
            return;
        }

        // Disable button during login attempt
        loginButton.IsEnabled = false;
        loginButton.Content = "Logging in...";

        try
        {
            var success = await _authService.LoginAsync(username, password);
            
            if (success)
            {
                LoginSuccessful = true;
                Close();
            }
            else
            {
                await ShowErrorMessage("Invalid username or password.");
                passwordTextBox.Text = "";
                passwordTextBox.Focus();
            }
        }
        finally
        {
            loginButton.IsEnabled = true;
            loginButton.Content = "Login";
        }
    }

    private void CancelButton_Click(object? sender, RoutedEventArgs e)
    {
        LoginSuccessful = false;
        Close();
    }

    private async System.Threading.Tasks.Task ShowErrorMessage(string message)
    {
        var errorDialog = new Window
        {
            Title = "Login Error",
            Width = 300,
            Height = 150,
            WindowStartupLocation = WindowStartupLocation.CenterOwner,
            CanResize = false,
            Content = new StackPanel
            {
                Margin = new Thickness(20),
                Spacing = 15,
                Children =
                {
                    new TextBlock
                    {
                        Text = message,
                        TextWrapping = Avalonia.Media.TextWrapping.Wrap,
                        FontSize = 14,
                        HorizontalAlignment = Avalonia.Layout.HorizontalAlignment.Center
                    },
                    new Button
                    {
                        Content = "OK",
                        Width = 80,
                        Height = 30,
                        HorizontalAlignment = Avalonia.Layout.HorizontalAlignment.Center
                    }
                }
            }
        };

        var okButton = ((StackPanel)errorDialog.Content).Children[1] as Button;
        if (okButton != null)
        {
            okButton.Click += (s, e) => errorDialog.Close();
        }

        await errorDialog.ShowDialog(this);
    }
}