using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace StreamLoftApp.Services
{
    public class ApiService
    {
        private readonly HttpClient _httpClient;
        private readonly string _baseUrl = "http://172.86.73.79:8080";

        public ApiService()
        {
            _httpClient = new HttpClient();
            _httpClient.BaseAddress = new Uri(_baseUrl);
        }

        public async Task<User> LoginAsync(string userId)
        {
            var response = await _httpClient.PostAsync($"/auth/login", 
                new StringContent(JsonSerializer.Serialize(new { UserId = userId }), Encoding.UTF8, "application/json"));
            
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                return JsonSerializer.Deserialize<User>(content);
            }
            
            throw new Exception($"Login failed: {response.StatusCode}");
        }

        public async Task<List<Destination>> GetDestinationsAsync()
        {
            var response = await _httpClient.GetAsync("/destinations");
            if (response.IsSuccessStatusCode)
            {
                var content = await response.Content.ReadAsStringAsync();
                return JsonSerializer.Deserialize<List<Destination>>(content);
            }
            throw new Exception($"Failed to get destinations: {response.StatusCode}");
        }
    }
}
