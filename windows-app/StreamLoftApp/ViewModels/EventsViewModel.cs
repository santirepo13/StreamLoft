using System;
using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Input;
using StreamLoftApp.Services;

namespace StreamLoftApp.ViewModels
{
    public class EventsViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private bool _isLoading;
        private BroadcastGroupItem _selectedBroadcast;

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

        public event Action OnBack;

        public EventsViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            RefreshCommand = new RelayCommand(async _ => await LoadBroadcastsAsync(), _ => !IsLoading);
            BackCommand = new RelayCommand(_ => OnBack?.Invoke());
            DeleteCommand = new RelayCommand<BroadcastGroupItem>(async item => await DeleteBroadcastAsync(item), item => item != null);

            LoadBroadcastsAsync();
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
            var result = MessageBox.Show(
                $"This will permanently delete this broadcast event from the database.\n\nDestination: {item.DestinationName}\nDate: {item.Date}\nDuration: {item.DurationText}\n\nThis action cannot be undone. Continue?",
                "Delete Broadcast Event",
                MessageBoxButton.YesNo,
                MessageBoxImage.Warning);

            if (result != MessageBoxResult.Yes)
                return;

            IsLoading = true;

            try
            {
                await _apiService.DeleteBroadcastAsync(item.DestinationName, item.Date);
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
        public string DestinationName { get; set; }
        public string Date { get; set; }
        public string DurationText { get; set; }
    }
}
