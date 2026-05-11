using System.Windows;
using StreamLoftApp.Views;

namespace StreamLoftApp
{
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
        }

        private void Logout_Click(object sender, RoutedEventArgs e)
        {
            // TODO: Implement logout
            LoginView loginView = new LoginView();
            loginView.Show();
            this.Close();
        }
    }
}
