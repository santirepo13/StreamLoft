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
            _viewModel.OnSettingsClick += OpenSettings;
            _viewModel.OnDestinationSaved += RefreshDestinations;
            _viewModel.OnBitrateUpdated += RefreshBitrate;
            _viewModel.OnDestinationToggled += RefreshDestinationToggle;
            
            DataContext = _viewModel;
            
            Closing += (s, e) => _viewModel.StopSSE();
        }

        private void NavigateToDestination(DestinationItem destination)
        {
            var destWindow = new DestinationView(destination.Id, destination.Name ?? "", destination.Configured, destination.StreamKey ?? "");
            destWindow.ShowDialog();

            _viewModel.NotifyDestinationSaved();
            _viewModel.StartSSE();
        }

        private void NavigateToEvents()
        {
            var eventsWindow = new EventsView();
            eventsWindow.ShowDialog();
        }

        private void UserNameButton_Click(object sender, RoutedEventArgs e)
        {
            if (sender is FrameworkElement element && element.ContextMenu != null)
            {
                element.ContextMenu.DataContext = element.DataContext;
                element.ContextMenu.IsOpen = true;
            }
        }

        private void OpenSettings()
        {
            var settingsWindow = new SettingsView(
                _viewModel.RtmpUrl,
                _viewModel.StreamKey,
                _viewModel.ConfiguredBitrate
            );
            settingsWindow.ShowDialog();
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
