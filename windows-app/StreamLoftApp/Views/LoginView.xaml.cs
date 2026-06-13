using System.Windows;
using StreamLoftApp.ViewModels;

namespace StreamLoftApp.Views
{
    public partial class LoginView : Window
    {
        private readonly LoginViewModel _viewModel;

        public LoginView()
        {
            InitializeComponent();
            
            _viewModel = new LoginViewModel();
            _viewModel.OnLoginSuccess += NavigateToWelcome;
            _viewModel.OnLoginFailed += ShowError;
            
            DataContext = _viewModel;
            
            Loaded += async (s, e) => await TryAutoLogin();
        }

        private async System.Threading.Tasks.Task TryAutoLogin()
        {
            if (_viewModel.LoginCommand.CanExecute(null))
            {
                _viewModel.AutoLoginCommand.Execute(null);
            }
        }

        private void NavigateToWelcome()
        {
            var welcomeWindow = new WelcomeView();
            welcomeWindow.Show();
            this.Close();
        }

        private void ShowError(string message)
        {
            // Error is displayed via binding
        }
    }
}