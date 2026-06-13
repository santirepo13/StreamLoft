using System.Windows;
using System.Windows.Controls;
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

        private void DeleteMenuItem_Click(object sender, RoutedEventArgs e)
        {
            if (sender is MenuItem menuItem &&
                menuItem.Parent is ContextMenu contextMenu &&
                contextMenu.PlacementTarget is Border border &&
                border.DataContext is BroadcastGroupItem item)
            {
                _viewModel.DeleteCommand.Execute(item);
            }
        }
    }
}
