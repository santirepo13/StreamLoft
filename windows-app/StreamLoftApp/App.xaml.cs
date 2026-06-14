using System.Windows;
using System.Windows.Markup;

namespace StreamLoftApp
{
    public partial class App : Application
    {
        protected override void OnStartup(StartupEventArgs e)
        {
            AppDomain.CurrentDomain.UnhandledException += (s, args) =>
                MessageBox.Show(args.ExceptionObject.ToString(), "Fatal Error");

            try
            {
                base.OnStartup(e);
            }
            catch (XamlParseException ex)
            {
                string msg = $"XAML Error in {ex.BaseUri}\nLine {ex.LineNumber}, Pos {ex.LinePosition}\n\n{ex.Message}\n\nInner:\n{ex.InnerException}";
                MessageBox.Show(msg, "XAML Parse Error");
                throw;
            }
        }
    }
}
