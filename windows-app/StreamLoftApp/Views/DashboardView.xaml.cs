using System.Windows;
using StreamLoftApp.ViewModels;

namespace StreamLoftApp.Views
{
    public partial class DashboardView : Window
    {
        private readonly DashboardViewModel _viewModel;

        public DashboardView()
        {
            InitializeComponent();
            
            _viewModel = new DashboardViewModel();
            _viewModel.OnDestinationClick += NavigateToDestination;
            _viewModel.OnEventsClick += NavigateToEvents;
            _viewModel.OnLogout += NavigateToLogin;
            
            DataContext = _viewModel;
            
            Closing += (s, e) => _viewModel.StopPolling();
        }

        private void NavigateToDestination(DestinationItem destination)
        {
            var destWindow = new DestinationView(destination.Id, destination.Name, destination.Configured);
            destWindow.ShowDialog();
            
            // Refresh data after returning
            _viewModel.StartPolling();
        }

        private void NavigateToEvents()
        {
            var eventsWindow = new EventsView();
            eventsWindow.ShowDialog();
        }

        private void NavigateToLogin()
        {
            var loginWindow = new LoginView();
            loginWindow.Show();
            this.Close();
        }
    }
}