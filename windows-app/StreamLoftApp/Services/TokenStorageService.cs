using System;
using System.IO;
using Newtonsoft.Json;

namespace StreamLoftApp.Services
{
    public class TokenStorageService
    {
        private readonly string _tokenFilePath;
        private TokenData _cachedToken;

        public TokenStorageService()
        {
            var appDataPath = Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
                "StreamLoft");
            
            Directory.CreateDirectory(appDataPath);
            _tokenFilePath = Path.Combine(appDataPath, "tokens.json");
        }

        public void SaveTokens(string accessToken, string refreshToken, DateTime expiresAt)
        {
            _cachedToken = new TokenData
            {
                AccessToken = accessToken,
                RefreshToken = refreshToken,
                ExpiresAt = expiresAt
            };

            var json = JsonConvert.SerializeObject(_cachedToken);
            File.WriteAllText(_tokenFilePath, json);
        }

        public TokenData? GetTokens()
        {
            if (_cachedToken != null)
                return _cachedToken;

            if (!File.Exists(_tokenFilePath))
                return null;

            try
            {
                var json = File.ReadAllText(_tokenFilePath);
                _cachedToken = JsonConvert.DeserializeObject<TokenData>(json);
                
                if (_cachedToken != null && _cachedToken.ExpiresAt < DateTime.UtcNow)
                {
                    // Token expired
                    ClearTokens();
                    return null;
                }

                return _cachedToken;
            }
            catch
            {
                return null;
            }
        }

        public void ClearTokens()
        {
            _cachedToken = null;
            if (File.Exists(_tokenFilePath))
            {
                File.Delete(_tokenFilePath);
            }
        }

        public bool HasValidToken()
        {
            var token = GetTokens();
            return token != null && token.ExpiresAt > DateTime.UtcNow;
        }
    }

    public class TokenData
    {
        public string? AccessToken { get; set; }
        public string? RefreshToken { get; set; }
        public DateTime ExpiresAt { get; set; }
    }
}