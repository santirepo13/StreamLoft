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

        public bool IsLoading
        {
            get => _isLoading;
            set => SetProperty(ref _isLoading, value);
        }

        public ObservableCollection<BroadcastItem> Broadcasts { get; } = new ObservableCollection<BroadcastItem>();

        public ICommand RefreshCommand { get; }
        public ICommand BackCommand { get; }

        public event Action OnBack;

        public EventsViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            RefreshCommand = new RelayCommand(async _ => await LoadBroadcastsAsync(), _ => !IsLoading);
            BackCommand = new RelayCommand(_ => OnBack?.Invoke());

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
                    Broadcasts.Add(new BroadcastItem
                    {
                        Id = b.Id,
                        DestinationName = b.DestinationName,
                        Date = b.Date,
                        StartedAt = b.StartedAt,
                        EndedAt = b.EndedAt,
                        DurationMinutes = b.DurationMinutes
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
    }

    public class BroadcastItem
    {
        public int Id { get; set; }
        public string DestinationName { get; set; }
        public string Date { get; set; }
        public string StartedAt { get; set; }
        public string EndedAt { get; set; }
        public int DurationMinutes { get; set; }

        public string DurationText => DurationMinutes > 0 
            ? $"{DurationMinutes} min" 
            : "In progress";
    }
}