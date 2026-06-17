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
            _viewModel.OnDestinationSaved += RefreshDestinations;
            _viewModel.OnBitrateUpdated += RefreshBitrate;
            _viewModel.OnDestinationToggled += RefreshDestinationToggle;
            
            DataContext = _viewModel;
            
            Closing += (s, e) => _viewModel.StopSSE();
        }

        private void NavigateToDestination(DestinationItem destination)
        {
            var destWindow = new DestinationView(destination.Id, destination.Name ?? "", destination.Configured, destination.StreamKey ?? "", destination.BitLimited == 1);
            destWindow.ShowDialog();

            _viewModel.NotifyDestinationSaved();
            _viewModel.StartSSE();
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

        private void RefreshDestinations()
        {
            _viewModel.RefreshDestinations();
        }

        private void RefreshBitrate(int bitrate)
        {
            _viewModel.ConfiguredBitrate = bitrate;
        }

        private void RefreshDestinationToggle(DestinationItem destination, bool isEnabled)
        {
            _viewModel.RefreshDestinationToggle(destination, isEnabled);
        }
    }
}
