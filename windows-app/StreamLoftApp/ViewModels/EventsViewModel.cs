using System;
using System.Collections.ObjectModel;
using System.Linq;
using System.Windows;
using System.Windows.Input;
using StreamLoftApp.Services;
using StreamLoftApp.Views;

namespace StreamLoftApp.ViewModels
{
    public class EventsViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private bool _isLoading;
        private BroadcastGroupItem _selectedBroadcast = null!;

        public bool IsLoading
        {
            get => _isLoading;
            set => SetProperty(ref _isLoading, value);
        }

        public BroadcastGroupItem SelectedBroadcast
        {
            get => _selectedBroadcast;
            set => SetProperty(ref _selectedBroadcast, value);
        }

        public ObservableCollection<BroadcastGroupItem> Broadcasts { get; } = new ObservableCollection<BroadcastGroupItem>();

        public ICommand RefreshCommand { get; }
        public ICommand BackCommand { get; }
        public ICommand DeleteCommand { get; }

        public event Action? OnBack;

        public EventsViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            RefreshCommand = new RelayCommand(async _ => await LoadBroadcastsAsync(), _ => !IsLoading);
            BackCommand = new RelayCommand(_ => OnBack?.Invoke());
            DeleteCommand = new RelayCommand<BroadcastGroupItem>(async item => await DeleteBroadcastAsync(item), item => item != null);

            _ = LoadBroadcastsAsync();
        }

        private async System.Threading.Tasks.Task LoadBroadcastsAsync()
        {
            IsLoading = true;

            try
            {
                var broadcasts = await _apiService.GetBroadcastsAsync();

                Broadcasts.Clear();
                foreach (var b in broadcasts)
                {
                    Broadcasts.Add(new BroadcastGroupItem
                    {
                        DestinationName = b.DestinationName,
                        Date = b.Date,
                        DurationText = $"{b.TotalMinutes} min"
                    });
                }
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to load broadcasts: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
            finally
            {
                IsLoading = false;
            }
        }

        private async System.Threading.Tasks.Task DeleteBroadcastAsync(BroadcastGroupItem item)
        {
            var owner = Application.Current.Windows
                .OfType<System.Windows.Window>()
                .FirstOrDefault(w => w.IsActive) ?? Application.Current.MainWindow;

            var dialog = new DeleteConfirmDialog(item.DestinationName ?? "", item.Date ?? "", item.DurationText ?? "")
            {
                Owner = owner
            };

            if (dialog.ShowDialog() != true)
                return;

            IsLoading = true;

            try
            {
                await _apiService.DeleteBroadcastAsync(item.DestinationName ?? "", item.Date ?? "");
                Broadcasts.Remove(item);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to delete broadcast: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
            finally
            {
                IsLoading = false;
            }
        }
    }

    public class BroadcastGroupItem
    {
        public string? DestinationName { get; set; }
        public string? Date { get; set; }
        public string? DurationText { get; set; }
    }
}
