using System;
using System.Windows;
using System.Windows.Input;
using StreamLoftApp.Services;

namespace StreamLoftApp.ViewModels
{
    public class LoginViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private readonly TokenStorageService _tokenStorage;
        private readonly MachineIdService _machineIdService;

        private string _userId;
        private string _errorMessage;
        private bool _isLoading;

        public string UserId
        {
            get => _userId;
            set
            {
                if (SetProperty(ref _userId, value))
                {
                    ErrorMessage = string.Empty;
                }
            }
        }

        public string ErrorMessage
        {
            get => _errorMessage;
            set => SetProperty(ref _errorMessage, value);
        }

        public bool IsLoading
        {
            get => _isLoading;
            set => SetProperty(ref _isLoading, value);
        }

        public ICommand LoginCommand { get; }
        public ICommand AutoLoginCommand { get; }

        public event Action OnLoginSuccess;
        public event Action<string> OnLoginFailed;

        public LoginViewModel()
        {
            _tokenStorage = new TokenStorageService();
            _machineIdService = new MachineIdService();
            _apiService = new ApiService(_tokenStorage);

            LoginCommand = new RelayCommand(async _ => await ExecuteLoginAsync(), _ => !IsLoading);
            AutoLoginCommand = new RelayCommand(async _ => await TryAutoLoginAsync(), _ => !IsLoading);
        }

        private async System.Threading.Tasks.Task ExecuteLoginAsync()
        {
            // Validate numeric ID
            var validationError = ValidationHelper.GetUserIdError(UserId);
            if (!string.IsNullOrEmpty(validationError))
            {
                ErrorMessage = validationError;
                return;
            }

            IsLoading = true;
            ErrorMessage = string.Empty;

            try
            {
                var machineId = _machineIdService.GetMachineId();
                var response = await _apiService.LoginAsync(UserId, machineId);

                // Store user info for Welcome screen
                App.Current.Properties["UserName"] = response.Name;
                App.Current.Properties["NumericId"] = response.NumericId;
                App.Current.Properties["RtmpUrl"] = response.RtmpUrl;
                App.Current.Properties["StreamKey"] = response.StreamKey;
                App.Current.Properties["Bitrate"] = response.Bitrate;

                OnLoginSuccess?.Invoke();
            }
            catch (Exception ex)
            {
                ErrorMessage = "Login failed. Please check your ID.";
                OnLoginFailed?.Invoke(ex.Message);
            }
            finally
            {
                IsLoading = false;
            }
        }

        private async System.Threading.Tasks.Task TryAutoLoginAsync()
        {
            if (!_tokenStorage.HasValidToken())
            {
                return;
            }

            IsLoading = true;

            try
            {
                var user = await _apiService.GetUserAsync();
                
                // Store user info
                App.Current.Properties["UserName"] = user.Name;
                App.Current.Properties["NumericId"] = user.NumericId;
                App.Current.Properties["RtmpUrl"] = user.RtmpUrl;
                App.Current.Properties["StreamKey"] = user.StreamKey;
                App.Current.Properties["Bitrate"] = user.Bitrate;

                OnLoginSuccess?.Invoke();
            }
            catch
            {
                // Token invalid - clear and stay on login
                _tokenStorage.ClearTokens();
            }
            finally
            {
                IsLoading = false;
            }
        }
    }
}