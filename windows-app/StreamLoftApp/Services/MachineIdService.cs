using System;
using System.IO;
using System.Management;

namespace StreamLoftApp.Services
{
    public class MachineIdService
    {
        private readonly string _machineIdFilePath;
        private string _cachedMachineId;

        public MachineIdService()
        {
            var appDataPath = Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
                "StreamLoft");
            
            Directory.CreateDirectory(appDataPath);
            _machineIdFilePath = Path.Combine(appDataPath, "machine_id.txt");
        }

        public string GetMachineId()
        {
            if (!string.IsNullOrEmpty(_cachedMachineId))
                return _cachedMachineId;

            // Try to load from file
            if (File.Exists(_machineIdFilePath))
            {
                try
                {
                    _cachedMachineId = File.ReadAllText(_machineIdFilePath).Trim();
                    if (!string.IsNullOrEmpty(_cachedMachineId))
                        return _cachedMachineId;
                }
                catch
                {
                    // Fall through to generate new
                }
            }

            // Generate new machine ID
            _cachedMachineId = GenerateMachineId();
            SaveMachineId(_cachedMachineId);
            return _cachedMachineId;
        }

        private void SaveMachineId(string machineId)
        {
            try
            {
                File.WriteAllText(_machineIdFilePath, machineId);
            }
            catch
            {
                // Ignore save errors
            }
        }

        private string GenerateMachineId()
        {
            try
            {
                // Use hardware identifiers for uniqueness
                var machineName = Environment.MachineName;
                var userName = Environment.UserName;
                
                // Combine with a unique identifier
                var combined = $"{machineName}-{userName}-{Guid.NewGuid()}";
                
                // Create a simple hash
                var hash = 0;
                foreach (var c in combined)
                {
                    hash = ((hash << 5) - hash) + c;
                    hash = hash & hash; // Keep within int bounds
                }
                
                return Math.Abs(hash).ToString("X8");
            }
            catch
            {
                // Fallback to random
                return Guid.NewGuid().ToString("N").Substring(0, 16);
            }
        }
    }
}