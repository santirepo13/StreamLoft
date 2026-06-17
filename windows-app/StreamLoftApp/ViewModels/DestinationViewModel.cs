using System;
using System.Windows;
using System.Windows.Input;
using StreamLoftApp.Services;

namespace StreamLoftApp.ViewModels
{
    public class DestinationViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private readonly int _destinationId;
        private readonly string _destinationName;

        private string _streamKeyInput = null!;
        private bool _isConfigured;
        private bool _isSaving;
        private string _errorMessage = null!;
        private bool _showSavedIndicator;
        private bool _bitLimited;

        public string DestinationName
        {
            get => _destinationName;
        }

        public string StreamKeyInput
        {
            get => _streamKeyInput;
            set
            {
                if (SetProperty(ref _streamKeyInput, value))
                {
                    ErrorMessage = string.Empty;
                    ShowSavedIndicator = false;
                }
            }
        }

        public bool IsConfigured
        {
            get => _isConfigured;
            set => SetProperty(ref _isConfigured, value);
        }

        public bool IsSaving
        {
            get => _isSaving;
            set => SetProperty(ref _isSaving, value);
        }

        public string ErrorMessage
        {
            get => _errorMessage;
            set => SetProperty(ref _errorMessage, value);
        }

        public bool ShowSavedIndicator
        {
            get => _showSavedIndicator;
            set => SetProperty(ref _showSavedIndicator, value);
        }

        public bool BitLimited
        {
            get => _bitLimited;
            set => SetProperty(ref _bitLimited, value);
        }

        public ICommand SaveCommand { get; }
        public ICommand CancelCommand { get; }

        public event Action? OnSaved;
        public event Action? OnCancel;
        public event Action? OnDestinationSaved;

        public DestinationViewModel(int destinationId, string destinationName, bool isConfigured, string streamKey, bool bitLimited = false)
        {
            _apiService = new ApiService(new TokenStorageService());
            _destinationId = destinationId;
            _destinationName = destinationName;
            _isConfigured = isConfigured;

            // Show actual stream key in the field
            StreamKeyInput = streamKey ?? "";
            _isConfigured = !string.IsNullOrEmpty(StreamKeyInput);
            _bitLimited = bitLimited;

            SaveCommand = new RelayCommand(async _ => await ExecuteSaveAsync(), _ => !IsSaving);
            CancelCommand = new RelayCommand(_ => OnCancel?.Invoke());
        }

        private async System.Threading.Tasks.Task ExecuteSaveAsync()
        {
            // Validate stream key
            if (!string.IsNullOrEmpty(StreamKeyInput))
            {
                var validationError = ValidationHelper.GetStreamKeyError(StreamKeyInput);
                if (!string.IsNullOrEmpty(validationError))
                {
                    ErrorMessage = validationError;
                    return;
                }
            }

            IsSaving = true;
            ErrorMessage = string.Empty;

            try
            {
                await _apiService.UpdateDestinationAsync(_destinationId, StreamKeyInput ?? "", BitLimited);

                // Update configured status based on saved value
                IsConfigured = !string.IsNullOrEmpty(StreamKeyInput);
                ShowSavedIndicator = true;

                OnSaved?.Invoke();
                OnDestinationSaved?.Invoke();
            }
            catch (Exception ex)
            {
                ErrorMessage = "Failed to save. Please try again.";
                MessageBox.Show($"Save failed: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
            finally
            {
                IsSaving = false;
            }
        }
    }
}