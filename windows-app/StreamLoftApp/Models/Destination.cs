using System.Collections.Generic;

namespace StreamLoftApp.Models
{
    public class Destination
    {
        public int Id { get; set; }
        public string Name { get; set; }
        public string RtmpUrl { get; set; }
        public string StreamKey { get; set; }
        public bool Enabled { get; set; }
    }
}
