using System;
using System.Windows.Input;
using StreamLoftApp.Services;

namespace StreamLoftApp.ViewModels
{
    public class WelcomeViewModel : ViewModelBase
    {
        private readonly ApiService _apiService;
        private string _userName;
        private string _numericId;
        private bool _isLoading;

        public string UserName
        {
            get => _userName;
            set => SetProperty(ref _userName, value);
        }

        public string NumericId
        {
            get => _numericId;
            set => SetProperty(ref _numericId, value);
        }

        public bool IsLoading
        {
            get => _isLoading;
            set => SetProperty(ref _isLoading, value);
        }

        public ICommand ProceedCommand { get; }
        public ICommand LogoutCommand { get; }

        public event Action OnProceed;
        public event Action OnLogout;

        public WelcomeViewModel()
        {
            _apiService = new ApiService(new TokenStorageService());

            ProceedCommand = new RelayCommand(_ => OnProceed?.Invoke(), _ => !IsLoading);
            LogoutCommand = new RelayCommand(async _ => await ExecuteLogoutAsync(), _ => !IsLoading);

            LoadUserInfo();
        }

        private void LoadUserInfo()
        {
            UserName = App.Current.Properties["UserName"] as string ?? "User";
            NumericId = App.Current.Properties["NumericId"] as string ?? "";
        }

        private async System.Threading.Tasks.Task ExecuteLogoutAsync()
        {
            IsLoading = true;

            try
            {
                await _apiService.LogoutAsync();
            }
            catch
            {
                // Continue with logout even if API call fails
            }
            finally
            {
                IsLoading = false;
                OnLogout?.Invoke();
            }
        }
    }
}