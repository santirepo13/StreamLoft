using System.Windows;
using StreamLoftApp.ViewModels;

namespace StreamLoftApp.Views
{
    public partial class WelcomeView : Window
    {
        private readonly WelcomeViewModel _viewModel;

        public WelcomeView()
        {
            InitializeComponent();
            
            _viewModel = new WelcomeViewModel();
            _viewModel.OnProceed += NavigateToDashboard;
            _viewModel.OnLogout += NavigateToLogin;
            
            DataContext = _viewModel;
        }

        private void NavigateToDashboard()
        {
            var dashboardWindow = new DashboardView();
            dashboardWindow.Show();
            this.Close();
        }

        private void NavigateToLogin()
        {
            var loginWindow = new LoginView();
            loginWindow.Show();
            this.Close();
        }
    }
}