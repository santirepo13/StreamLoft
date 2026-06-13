using System.Windows;
using StreamLoftApp.ViewModels;

namespace StreamLoftApp.Views
{
    public partial class EventsView : Window
    {
        private readonly EventsViewModel _viewModel;

        public EventsView()
        {
            InitializeComponent();
            
            _viewModel = new EventsViewModel();
            _viewModel.OnBack += OnBack;
            
            DataContext = _viewModel;
        }

        private void OnBack()
        {
            this.Close();
        }
    }
}