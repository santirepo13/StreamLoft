using System.IO;
using Newtonsoft.Json;

namespace StreamLoftApp.Configuration
{
    public class AppSettings
    {
        private static AppSettings _instance;
        public string ApiBaseUrl { get; set; }

        public static AppSettings Instance
        {
            get
            {
                if (_instance == null)
                {
                    _instance = Load();
                }
                return _instance;
            }
        }

        private static AppSettings Load()
        {
            try
            {
                var appPath = AppDomain.CurrentDomain.BaseDirectory;
                var settingsPath = Path.Combine(appPath, "appsettings.json");

                if (File.Exists(settingsPath))
                {
                    var json = File.ReadAllText(settingsPath);
                    return JsonConvert.DeserializeObject<AppSettings>(json) ?? new AppSettings();
                }
            }
            catch
            {
                // Fall through to default
            }

            return new AppSettings { ApiBaseUrl = "http://172.86.73.79:8080" };
        }
    }
}