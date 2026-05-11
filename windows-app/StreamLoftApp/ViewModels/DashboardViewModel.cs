using System;
using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Input;
using System.Windows.Threading;
using StreamLoftApp.Models;
using StreamLoftApp.Services;

namespace StreamLoftApp.ViewModels
{
    public class DashboardViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private DispatcherTimer _statusTimer;

        private string _rtmpUrl;
        private string _streamKey;
        private bool _isLive;
        private bool _bitrateWarning;
        private bool _isLoading;
        private string _statusText;
        private int? _configuredBitrate;

        public string RtmpUrl
        {
            get => _rtmpUrl;
            set => SetProperty(ref _rtmpUrl, value);
        }

        public string StreamKey
        {
            get => _streamKey;
            set => SetProperty(ref _streamKey, value);
        }

        public bool IsLive
        {
            get => _isLive;
            set
            {
                if (SetProperty(ref _isLive, value))
                {
                    StatusText = value ? "Live" : "Offline";
                }
            }
        }

        public bool BitrateWarning
        {
            get => _bitrateWarning;
            set => SetProperty(ref _bitrateWarning, value);
        }

        public int? ConfiguredBitrate
        {
            get => _configuredBitrate;
            set => SetProperty(ref _configuredBitrate, value);
        }

        public bool IsLoading
        {
            get => _isLoading;
            set => SetProperty(ref _isLoading, value);
        }

        public string StatusText
        {
            get => _statusText;
            set => SetProperty(ref _statusText, value);
        }

        public ObservableCollection<DestinationItem> Destinations { get; } = new ObservableCollection<DestinationItem>();

        public ICommand CopyRtmpUrlCommand { get; }
        public ICommand CopyStreamKeyCommand { get; }
        public ICommand DestinationClickCommand { get; }
        public ICommand LogoutCommand { get; }
        public ICommand EventsCommand { get; }
        public ICommand UpdateBitrateCommand { get; }

        public event Action<DestinationItem> OnDestinationClick;
        public event Action OnEventsClick;
        public event Action OnLogout;

        public DashboardViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            CopyRtmpUrlCommand = new RelayCommand(CopyRtmpUrl);
            CopyStreamKeyCommand = new RelayCommand(CopyStreamKey);
            DestinationClickCommand = new RelayCommand<DestinationItem>(OnDestination);
            LogoutCommand = new RelayCommand(async _ => await ExecuteLogoutAsync(), _ => !IsLoading);
            EventsCommand = new RelayCommand(_ => OnEventsClick?.Invoke(), _ => !IsLoading);
            UpdateBitrateCommand = new RelayCommand(async _ => await UpdateBitrateAsync(), _ => !IsLoading);

            _statusTimer = new DispatcherTimer
            {
                Interval = TimeSpan.FromSeconds(3)
            };
            _statusTimer.Tick += async (s, e) => await RefreshStatusAsync();

            LoadDataAsync();
        }

        public void StartPolling()
        {
            _statusTimer.Start();
        }

        public void StopPolling()
        {
            _statusTimer.Stop();
        }

        private async void LoadDataAsync()
        {
            IsLoading = true;

            try
            {
                var user = await _apiService.GetUserAsync();
                RtmpUrl = user.RtmpUrl ?? "";
                StreamKey = user.StreamKey ?? "";
                ConfiguredBitrate = user.Bitrate;

                var destinations = await _apiService.GetDestinationsAsync();
                Destinations.Clear();
                foreach (var dest in destinations)
                {
                    Destinations.Add(new DestinationItem
                    {
                        Id = dest.Id,
                        Name = dest.Name,
                        RtmpUrl = dest.RtmpUrl,
                        Configured = dest.Configured,
                        StreamKey = dest.StreamKey ?? ""
                    });
                }

                await RefreshStatusAsync();
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to load data: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
            finally
            {
                IsLoading = false;
                StartPolling();
            }
        }

        private async System.Threading.Tasks.Task RefreshStatusAsync()
        {
            try
            {
                var status = await _apiService.GetStreamStatusAsync();
                IsLive = status.Status?.ToLower() == "live";
                BitrateWarning = status.BitrateWarning;
            }
            catch
            {
                // Ignore status refresh errors
            }
        }

        private void CopyRtmpUrl(object parameter)
        {
            if (!string.IsNullOrEmpty(RtmpUrl))
            {
                Clipboard.SetText(RtmpUrl);
            }
        }

        private void CopyStreamKey(object parameter)
        {
            if (!string.IsNullOrEmpty(StreamKey))
            {
                Clipboard.SetText(StreamKey);
            }
        }

        private void OnDestination(DestinationItem destination)
        {
            OnDestinationClick?.Invoke(destination);
        }

        private async System.Threading.Tasks.Task ExecuteLogoutAsync()
        {
            StopPolling();
            IsLoading = true;

            try
            {
                await _apiService.LogoutAsync();
            }
            catch
            {
                // Continue with logout
            }
            finally
            {
                IsLoading = false;
                OnLogout?.Invoke();
            }
        }

        private async System.Threading.Tasks.Task UpdateBitrateAsync()
        {
            if (!ConfiguredBitrate.HasValue || ConfiguredBitrate.Value <= 0)
            {
                MessageBox.Show("Please enter a valid bitrate (positive number)", "Invalid Bitrate", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            IsLoading = true;

            try
            {
                await _apiService.UpdateBitrateAsync(ConfiguredBitrate.Value);
                MessageBox.Show($"Bitrate updated to {ConfiguredBitrate.Value} kbps", "Success", MessageBoxButton.OK, MessageBoxImage.Information);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to update bitrate: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
            finally
            {
                IsLoading = false;
            }
        }
    }

    public class DestinationItem
    {
        public int Id { get; set; }
        public string Name { get; set; }
        public string RtmpUrl { get; set; }
        public bool Configured { get; set; }
        public string StreamKey { get; set; }
    }
}