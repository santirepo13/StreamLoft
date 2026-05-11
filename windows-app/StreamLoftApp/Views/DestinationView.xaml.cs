using System.Windows;
using StreamLoftApp.ViewModels;

namespace StreamLoftApp.Views
{
    public partial class DestinationView : Window
    {
        private readonly DestinationViewModel _viewModel;

        public DestinationView(int destinationId, string destinationName, bool isConfigured)
        {
            InitializeComponent();
            
            _viewModel = new DestinationViewModel(destinationId, destinationName, isConfigured);
            _viewModel.OnSaved += OnSaved;
            _viewModel.OnCancel += OnCancel;
            
            DataContext = _viewModel;
        }

        private void OnSaved()
        {
            // Close after save
            this.Close();
        }

        private void OnCancel()
        {
            this.Close();
        }
    }
}