using System;
using System.Windows;

namespace StreamLoftApp.Services
{
    public class NavigationService
    {
        private static NavigationService _instance;
        public static NavigationService Instance => _instance ??= new NavigationService();

        public event Action<Type> NavigateRequested;

        public void NavigateTo<T>() where T : Window, new()
        {
            var window = new T();
            window.Show();
            CloseCurrentWindow();
        }

        public void NavigateTo(Window window)
        {
            window.Show();
            CloseCurrentWindow();
        }

        private void CloseCurrentWindow()
        {
            var currentMain = Application.Current.MainWindow;
            currentMain?.Close();
        }

        public void CloseAndOpen<T>() where T : Window, new()
        {
            Window current = Application.Current.MainWindow;
            var newWindow = new T();
            newWindow.Show();
            if (current != null)
            {
                current.Close();
            }
        }

        public void SetMainWindow(Window window)
        {
            Application.Current.MainWindow = window;
        }
    }

    // Screen types for navigation
    public static class Screens
    {
        public const string Login = "LoginView";
        public const string Welcome = "WelcomeView";
        public const string Dashboard = "DashboardView";
        public const string Destination = "DestinationView";
        public const string Events = "EventsView";
    }
}