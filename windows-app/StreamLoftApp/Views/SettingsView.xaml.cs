using System;
using System.Windows;
using StreamLoftApp.Services;

namespace StreamLoftApp.Views
{
    public partial class SettingsView : Window
    {
        private readonly ApiService _apiService;

        public SettingsView(string rtmpUrl, string streamKey, int? configuredBitrate)
        {
            InitializeComponent();

            _apiService = new ApiService(new TokenStorageService());

            RtmpUrlTextBox.Text = rtmpUrl;
            StreamKeyTextBox.Text = streamKey;
            BitrateTextBox.Text = configuredBitrate?.ToString() ?? "";
        }

        private void CopyRtmpUrl_Click(object sender, RoutedEventArgs e)
        {
            if (!string.IsNullOrEmpty(RtmpUrlTextBox.Text))
                Clipboard.SetText(RtmpUrlTextBox.Text);
        }

        private void CopyStreamKey_Click(object sender, RoutedEventArgs e)
        {
            if (!string.IsNullOrEmpty(StreamKeyTextBox.Text))
                Clipboard.SetText(StreamKeyTextBox.Text);
        }

        private async void SetBitrate_Click(object sender, RoutedEventArgs e)
        {
            if (!int.TryParse(BitrateTextBox.Text, out int bitrate) || bitrate <= 0)
            {
                MessageBox.Show("Please enter a valid bitrate (positive number)", "Invalid Bitrate", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            try
            {
                await _apiService.UpdateBitrateAsync(bitrate);
                MessageBox.Show($"Bitrate updated to {bitrate} kbps", "Success", MessageBoxButton.OK, MessageBoxImage.Information);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Failed to update bitrate: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private void Close_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
