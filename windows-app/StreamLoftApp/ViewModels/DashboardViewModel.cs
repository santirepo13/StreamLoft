using System;
using System.Collections.ObjectModel;
using System.Linq;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Input;
using StreamLoftApp.Models;
using StreamLoftApp.Services;
using System.Windows.Controls;
using System.Windows.Media;

namespace StreamLoftApp.ViewModels
{
    public class DashboardViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private CancellationTokenSource _sseCts;
        private Task _sseTask;

        private string _rtmpUrl;
        private string _streamKey;
        private bool _isLive;
        private bool _bitrateWarning;
        private bool _isLoading;
        private string _statusText = "Offline";
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
                    UpdateDestinationsActiveStatus();
                }
            }
        }

        private void UpdateDestinationsActiveStatus()
        {
            foreach (var dest in Destinations)
            {
                dest.IsActive = dest.Enabled == 1 && dest.Configured && IsLive;
            }
            // Force ObservableCollection refresh
            var items = Destinations.ToList();
            Destinations.Clear();
            foreach (var item in items)
                Destinations.Add(item);
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
        public ICommand DestinationBadgeCommand { get; }
        public ICommand LogoutCommand { get; }
        public ICommand EventsCommand { get; }
        public ICommand UpdateBitrateCommand { get; }

        public event Action<DestinationItem> OnDestinationClick;
        public event Action OnEventsClick;
        public event Action OnLogout;
        public event Action OnDestinationSaved;
        public event Action<int> OnBitrateUpdated;
        public event Action<DestinationItem, bool> OnDestinationToggled;

        public void RefreshDestinations()
        {
            _ = RefreshDestinationsAsync();
        }

        public void RefreshBitrate(int bitrate)
        {
            ConfiguredBitrate = bitrate;
        }

        public void RefreshDestinationToggle(DestinationItem destination, bool isEnabled)
        {
            var dest = Destinations.FirstOrDefault(d => d.Id == destination.Id);
            if (dest != null)
            {
                dest.Enabled = isEnabled ? 1 : 0;
                dest.IsActive = isEnabled && dest.Configured && IsLive;
                UpdateDestinationsActiveStatus();
            }
        }

        private async System.Threading.Tasks.Task RefreshDestinationsAsync()
        {
            try
            {
                var destinations = await _apiService.GetDestinationsAsync();
                Application.Current?.Dispatcher.Invoke(() =>
                {
                    Destinations.Clear();
                    foreach (var dest in destinations)
                    {
                        Destinations.Add(new DestinationItem
                        {
                            Id = dest.Id,
                            Name = dest.Name,
                            RtmpUrl = dest.RtmpUrl,
                            Configured = dest.Configured,
                            Enabled = dest.Enabled,
                            StreamKey = dest.StreamKey ?? "",
                            IsActive = dest.Enabled == 1 && dest.Configured && IsLive,
                            BitLimited = dest.BitLimited
                        });
                    }
                });
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to refresh destinations: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        public DashboardViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            CopyRtmpUrlCommand = new RelayCommand(CopyRtmpUrl);
            CopyStreamKeyCommand = new RelayCommand(CopyStreamKey);
            DestinationClickCommand = new RelayCommand<DestinationItem>(OnDestination);
            DestinationBadgeCommand = new RelayCommand<DestinationItem>(OnDestinationBadge);
            LogoutCommand = new RelayCommand(async _ => await ExecuteLogoutAsync(), _ => !IsLoading);
            EventsCommand = new RelayCommand(_ => OnEventsClick?.Invoke(), _ => !IsLoading);
            UpdateBitrateCommand = new RelayCommand(async _ => await UpdateBitrateAsync(), _ => !IsLoading);

            LoadDataAsync();
        }

        public void StartSSE()
        {
            StopSSE();
            _sseCts = new CancellationTokenSource();
            _sseTask = Task.Run(async () => await SubscribeToStreamEvents(_sseCts.Token));
        }

        public void StopSSE()
        {
            _sseCts?.Cancel();
        }

        private async System.Threading.Tasks.Task SubscribeToStreamEvents(CancellationToken ct)
        {
            while (!ct.IsCancellationRequested)
            {
                try
                {
                    await foreach (var evt in _apiService.SubscribeStreamEventsAsync(ct))
                    {
                        Application.Current?.Dispatcher.Invoke(() =>
                        {
                            IsLive = evt.Status?.ToLower() == "live";
                            BitrateWarning = evt.BitrateWarning;
                        });
                    }
                }
                catch (OperationCanceledException)
                {
                    break;
                }
                catch
                {
                    // SSE disconnected — fall back to polling
                }

                if (!ct.IsCancellationRequested)
                {
                    // Poll status while waiting to retry SSE
                    try
                    {
                        var status = await _apiService.GetStreamStatusAsync();
                        Application.Current?.Dispatcher.Invoke(() =>
                        {
                            IsLive = status.Status?.ToLower() == "live";
                            BitrateWarning = status.BitrateWarning;
                        });
                    }
                    catch
                    {
                        // Ignore polling errors
                    }

                    try
                    {
                        await System.Threading.Tasks.Task.Delay(5000, ct);
                    }
                    catch (OperationCanceledException)
                    {
                        break;
                    }
                }
            }
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
                        Enabled = dest.Enabled,
                        StreamKey = dest.StreamKey ?? "",
                        IsActive = dest.Enabled == 1 && dest.Configured && IsLive,
                        BitLimited = dest.BitLimited
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
                StartSSE();
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
                SafeSetClipboardText(RtmpUrl);
            }
        }

        private void CopyStreamKey(object parameter)
        {
            if (!string.IsNullOrEmpty(StreamKey))
            {
                SafeSetClipboardText(StreamKey);
            }
        }

        private static void SafeSetClipboardText(string text)
        {
            for (int i = 0; i < 5; i++)
            {
                try
                {
                    Clipboard.SetText(text);
                    return;
                }
                catch (COMException ex) when ((uint)ex.ErrorCode == 0x800401D0)
                {
                    if (i == 4) throw;
                    Thread.Sleep(50);
                }
            }
        }

        private void OnDestination(DestinationItem destination)
        {
            OnDestinationClick?.Invoke(destination);
        }

        private async void OnDestinationBadge(DestinationItem destination)
        {
            // If not configured, open config window
            if (!destination.Configured)
            {
                OnDestinationClick?.Invoke(destination);
                return;
            }

            // Toggle enabled state
            try
            {
                int newEnabled = destination.Enabled == 1 ? 0 : 1;
                await _apiService.ToggleDestinationAsync(destination.Id, newEnabled);
                destination.Enabled = newEnabled;
                UpdateDestinationsActiveStatus();
                OnDestinationToggled?.Invoke(destination, newEnabled == 1);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to toggle destination: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private async System.Threading.Tasks.Task ExecuteLogoutAsync()
        {
            StopSSE();
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
                OnBitrateUpdated?.Invoke(ConfiguredBitrate.Value);
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
        public int Enabled { get; set; } // 1=true, 0=false
        public bool IsActive { get; set; }
        public string StreamKey { get; set; }
        public int BitLimited { get; set; }
    }
}