using System.Windows;

namespace StreamLoftApp.Views
{
    public partial class DeleteConfirmDialog : Window
    {
        public DeleteConfirmDialog(string destinationName, string date, string durationText)
        {
            InitializeComponent();

            Style = (Style)Resources["DialogWindowStyle"];

            MessageText.Text =
                $"This will permanently delete this broadcast event from the database.\n\n" +
                $"Destination: {destinationName}\n" +
                $"Date: {date}\n" +
                $"Duration: {durationText}\n\n" +
                $"This action cannot be undone. Continue?";
        }

        private void Yes_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = true;
            Close();
        }

        private void Cancel_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = false;
            Close();
        }
    }
}
