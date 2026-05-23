using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;
using StreamLoftApp.Configuration;
using StreamLoftApp.Models;

namespace StreamLoftApp.Services
{
    public class ApiService
    {
        private readonly HttpClient _httpClient;
        private readonly TokenStorageService _tokenStorage;
        private string _accessToken;

        public ApiService(TokenStorageService tokenStorage)
        {
            _tokenStorage = tokenStorage;
            _httpClient = new HttpClient();
            _httpClient.BaseAddress = new Uri(AppSettings.Instance.ApiBaseUrl);
            
            // Load cached token
            var token = _tokenStorage.GetTokens();
            if (token != null)
            {
                _accessToken = token.AccessToken;
            }
        }

        private void SetAuthHeader()
        {
            if (!string.IsNullOrEmpty(_accessToken))
            {
                _httpClient.DefaultRequestHeaders.Authorization = 
                    new AuthenticationHeaderValue("Bearer", _accessToken);
            }
        }

        public async Task<LoginResponse> LoginAsync(string numericId, string machineId)
        {
            var response = await _httpClient.PostAsync("/auth/login", 
                new StringContent(
                    JsonConvert.SerializeObject(new { numeric_id = numericId, machine_id = machineId }),
                    Encoding.UTF8, "application/json"));
            
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                var loginResponse = JsonConvert.DeserializeObject<LoginResponse>(content);
                
                if (loginResponse != null)
                {
                    _accessToken = loginResponse.AccessToken;
                    _tokenStorage.SaveTokens(
                        loginResponse.AccessToken,
                        loginResponse.RefreshToken,
                        DateTime.UtcNow.AddDays(3));
                }
                
                return loginResponse;
            }
            
            throw new Exception($"Login failed: {response.StatusCode}");
        }

        public async Task<string> RefreshTokenAsync()
        {
            var token = _tokenStorage.GetTokens();
            if (token?.RefreshToken == null)
                throw new Exception("No refresh token available");

            var response = await _httpClient.PostAsync("/auth/refresh",
                new StringContent(
                    JsonConvert.SerializeObject(new { refresh_token = token.RefreshToken }),
                    Encoding.UTF8, "application/json"));

            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                var refreshResponse = JsonConvert.DeserializeObject<RefreshResponse>(content);
                
                if (refreshResponse != null)
                {
                    _accessToken = refreshResponse.AccessToken;
                    _tokenStorage.SaveTokens(
                        refreshResponse.AccessToken,
                        token.RefreshToken,
                        DateTime.UtcNow.AddDays(3));
                    
                    return refreshResponse.AccessToken;
                }
            }

            throw new Exception("Token refresh failed");
        }

        public async Task LogoutAsync()
        {
            SetAuthHeader();
            
            try
            {
                await _httpClient.PostAsync("/auth/logout", null);
            }
            finally
            {
                _tokenStorage.ClearTokens();
                _accessToken = null;
            }
        }

        public async Task<UserResponse> GetUserAsync()
        {
            SetAuthHeader();
            
            var response = await _httpClient.GetAsync("/user");
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                return JsonConvert.DeserializeObject<UserResponse>(content);
            }
            
            throw new Exception($"Failed to get user: {response.StatusCode}");
        }

        public async Task<List<DestinationResponse>> GetDestinationsAsync()
        {
            SetAuthHeader();
            
            var response = await _httpClient.GetAsync("/destinations");
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                var destinationsResponse = JsonConvert.DeserializeObject<DestinationsResponse>(content);
                return destinationsResponse?.Destinations ?? new List<DestinationResponse>();
            }
            
            throw new Exception($"Failed to get destinations: {response.StatusCode}");
        }

        public async Task UpdateDestinationAsync(int destinationId, string streamKey)
        {
            SetAuthHeader();
            
            var response = await _httpClient.PutAsync($"/destinations/{destinationId}",
                new StringContent(
                    JsonConvert.SerializeObject(new { stream_key = streamKey }),
                    Encoding.UTF8, "application/json"));

            if (!response.IsSuccessStatusCode)
            {
                throw new Exception($"Failed to update destination: {response.StatusCode}");
            }
        }

        public async Task<StreamStatusResponse> GetStreamStatusAsync()
        {
            SetAuthHeader();
            
            var response = await _httpClient.GetAsync("/user/stream/status");
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                return JsonConvert.DeserializeObject<StreamStatusResponse>(content);
            }
            
            throw new Exception($"Failed to get stream status: {response.StatusCode}");
        }

        public async Task<List<BroadcastResponse>> GetBroadcastsAsync()
        {
            SetAuthHeader();
            
            var response = await _httpClient.GetAsync("/broadcasts");
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                return JsonConvert.DeserializeObject<List<BroadcastResponse>>(content);
            }
            
            throw new Exception($"Failed to get broadcasts: {response.StatusCode}");
        }

        public async Task UpdateBitrateAsync(int bitrate)
        {
            SetAuthHeader();
            
            var response = await _httpClient.PutAsync("/user/bitrate",
                new StringContent(
                    JsonConvert.SerializeObject(new { bitrate = bitrate }),
                    Encoding.UTF8, "application/json"));

            if (!response.IsSuccessStatusCode)
            {
                throw new Exception($"Failed to update bitrate: {response.StatusCode}");
            }
        }

        public async IAsyncEnumerable<StreamEvent> SubscribeStreamEventsAsync(
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken ct = default)
        {
            SetAuthHeader();

            using var request = new HttpRequestMessage(HttpMethod.Get, "/user/stream/events");
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", _accessToken);

            using var response = await _httpClient.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, ct);
            response.EnsureSuccessStatusCode();

            using var stream = await response.Content.ReadAsStreamAsync(ct);
            using var reader = new System.IO.StreamReader(stream);

            while (!reader.EndOfStream && !ct.IsCancellationRequested)
            {
                var line = await reader.ReadLineAsync(ct);
                if (line == null) break;

                if (line.StartsWith("data: "))
                {
                    var json = line.Substring(6);
                    var evt = JsonConvert.DeserializeObject<StreamEvent>(json);
                    if (evt != null)
                        yield return evt;
                }
            }
        }

        public void ClearToken()
        {
            _accessToken = null;
            _tokenStorage.ClearTokens();
        }

        public bool HasValidToken()
        {
            return _tokenStorage.HasValidToken();
        }
    }

    // Response models
    public class LoginResponse
    {
        [JsonProperty("numeric_id")]
        public string NumericId { get; set; }

        [JsonProperty("name")]
        public string Name { get; set; }

        [JsonProperty("access_token")]
        public string AccessToken { get; set; }

        [JsonProperty("refresh_token")]
        public string RefreshToken { get; set; }

        [JsonProperty("stream_key")]
        public string StreamKey { get; set; }

        [JsonProperty("rtmp_url")]
        public string RtmpUrl { get; set; }

        public List<DestinationResponse> Destinations { get; set; }
    }

    public class RefreshResponse
    {
        public string AccessToken { get; set; }
    }

    public class UserResponse
    {
        [JsonProperty("numeric_id")]
        public string NumericId { get; set; }

        [JsonProperty("name")]
        public string Name { get; set; }

        [JsonProperty("stream_key")]
        public string StreamKey { get; set; }

        [JsonProperty("rtmp_url")]
        public string RtmpUrl { get; set; }

        public int? Bitrate { get; set; }
        public List<DestinationResponse> Destinations { get; set; }
    }

    public class DestinationResponse
    {
        public int Id { get; set; }

        [JsonProperty("name")]
        public string Name { get; set; }

        [JsonProperty("rtmp_url")]
        public string RtmpUrl { get; set; }

        public bool Configured { get; set; }
        
        [JsonProperty("stream_key")]
        public string StreamKey { get; set; }
    }

    public class DestinationsResponse
    {
        [JsonProperty("destinations")]
        public List<DestinationResponse> Destinations { get; set; }
    }

    public class StreamStatusResponse
    {
        [JsonProperty("status")]
        public string Status { get; set; }

        [JsonProperty("bitrate_warning")]
        public bool BitrateWarning { get; set; }
    }

    public class StreamEvent
    {
        [JsonProperty("type")]
        public string Type { get; set; }

        [JsonProperty("status")]
        public string Status { get; set; }

        [JsonProperty("bitrate_warning")]
        public bool BitrateWarning { get; set; }
    }

    public class BroadcastResponse
    {
        public int Id { get; set; }

        [JsonProperty("destination_name")]
        public string DestinationName { get; set; }

        public string Date { get; set; }
        public string StartedAt { get; set; }
        public string EndedAt { get; set; }
        public int DurationMinutes { get; set; }
    }
}