using System.ComponentModel;
using System.Runtime.CompilerServices;

namespace StreamLoftApp.Models
{
    public class User : INotifyPropertyChanged
    {
        private string _userId;
        private string _name;
        private string _streamKey;

        public string UserId
        {
            get => _userId;
            set { _userId = value; OnPropertyChanged(); }
        }

        public string Name
        {
            get => _name;
            set { _name = value; OnPropertyChanged(); }
        }

        public string StreamKey
        {
            get => _streamKey;
            set { _streamKey = value; OnPropertyChanged(); }
        }

        public event PropertyChangedEventHandler PropertyChanged;

        protected virtual void OnPropertyChanged([CallerMemberName] string propertyName = null)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }
    }
}
