using System;
using System.Text.RegularExpressions;

namespace StreamLoftApp.Services
{
    public static class ValidationHelper
    {
        private static readonly Regex NumericOnlyRegex = new Regex(@"^\d+$", RegexOptions.Compiled);
        
        /// <summary>
        /// Validates user ID is numeric only (per VAL-001)
        /// </summary>
        public static bool IsValidUserId(string userId)
        {
            if (string.IsNullOrWhiteSpace(userId))
                return false;

            return NumericOnlyRegex.IsMatch(userId);
        }

        /// <summary>
        /// Validates user ID is numeric and positive (per VAL-001)
        /// </summary>
        public static bool IsValidUserIdRange(string userId)
        {
            if (!long.TryParse(userId, out long id))
                return false;

            return id >= 1;
        }

        /// <summary>
        /// Validates stream key is non-empty and max 256 chars (per VAL-003)
        /// </summary>
        public static bool IsValidStreamKey(string streamKey)
        {
            if (streamKey == null)
                return false;

            // Empty string is valid (used to disable destination)
            if (streamKey == string.Empty)
                return true;

            return streamKey.Length > 0 && streamKey.Length <= 256;
        }

        /// <summary>
        /// Validates name is non-empty and max 100 chars (per VAL-002)
        /// </summary>
        public static bool IsValidName(string name)
        {
            if (string.IsNullOrWhiteSpace(name))
                return false;

            return name.Length > 0 && name.Length <= 100;
        }

        /// <summary>
        /// Validates bitrate is positive integer (per VAL-007)
        /// </summary>
        public static bool IsValidBitrate(int bitrate)
        {
            return bitrate > 0;
        }

        /// <summary>
        /// Gets validation error message for user ID
        /// </summary>
        public static string GetUserIdError(string userId)
        {
            if (string.IsNullOrWhiteSpace(userId))
                return "ID is required";
            
            if (!IsValidUserId(userId))
                return "ID must be a number";
            
            if (!IsValidUserIdRange(userId))
                return "ID must be a positive number";
            
            return string.Empty;
        }

        /// <summary>
        /// Gets validation error message for stream key
        /// </summary>
        public static string GetStreamKeyError(string streamKey)
        {
            if (streamKey != null && streamKey.Length > 256)
                return "Stream key must be 256 characters or less";
            
            return string.Empty;
        }
    }
}