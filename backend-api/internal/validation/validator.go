package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"streamloft-api/internal/errors"
)

var (
	userIDRegex   = regexp.MustCompile(`^[1-9][0-9]{0,5}$`)
	streamKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	nameRegex     = regexp.MustCompile(`^.+$`)
)

func ValidateUserID(input string) error {
	if !userIDRegex.MatchString(input) {
		return errors.BadRequest("ID must be a number between 1 and 999999")
	}

	id, err := strconv.Atoi(input)
	if err != nil {
		return errors.BadRequest("ID must be a number")
	}

	if id < 1 || id > 999999 {
		return errors.BadRequest("ID must be between 1 and 999999")
	}

	return nil
}

func ValidateUserName(input string) error {
	if strings.TrimSpace(input) == "" {
		return errors.BadRequest("Name is required")
	}

	if len(input) > 100 {
		return errors.BadRequest("Name must be 100 characters or less")
	}

	return nil
}

func ValidateStreamKey(input string) error {
	if strings.TrimSpace(input) == "" {
		return errors.BadRequest("Stream key is required")
	}

	if len(input) > 256 {
		return errors.BadRequest("Stream key must be 256 characters or less")
	}

	if !streamKeyRegex.MatchString(input) {
		return errors.BadRequest("Stream key contains invalid characters")
	}

	return nil
}

func ValidateRTMPURL(input string) error {
	if strings.TrimSpace(input) == "" {
		return errors.BadRequest("RTMP URL is required")
	}

	if len(input) > 500 {
		return errors.BadRequest("RTMP URL must be 500 characters or less")
	}

	if !strings.HasPrefix(input, "rtmp://") {
		return errors.BadRequest("Invalid RTMP URL")
	}

	return nil
}

func ValidateDestinationName(input string) error {
	if strings.TrimSpace(input) == "" {
		return errors.BadRequest("Name is required")
	}

	if len(input) > 100 {
		return errors.BadRequest("Name must be 100 characters or less")
	}

	return nil
}

func ValidateBitrate(input string) error {
	if input == "" {
		return nil
	}

	bitrate, err := strconv.Atoi(input)
	if err != nil {
		return errors.BadRequest("Bitrate must be a number")
	}

	if bitrate <= 0 {
		return errors.BadRequest("Bitrate must be positive")
	}

	return nil
}

func ValidateMachineID(input string) error {
	if strings.TrimSpace(input) == "" {
		return errors.BadRequest("Machine ID is required")
	}

	if len(input) > 255 {
		return errors.BadRequest("Machine ID is too long")
	}

	return nil
}

type ValidationResult struct {
	Field   string
	Message string
}

func ValidateAll(rules map[string]string) []ValidationResult {
	var results []ValidationResult

	for field, value := range rules {
		var err error

		switch field {
		case "user_id":
			err = ValidateUserID(value)
		case "name":
			err = ValidateUserName(value)
		case "stream_key":
			err = ValidateStreamKey(value)
		case "rtmp_url":
			err = ValidateRTMPURL(value)
		case "destination_name":
			err = ValidateDestinationName(value)
		case "bitrate":
			err = ValidateBitrate(value)
		case "machine_id":
			err = ValidateMachineID(value)
		default:
			continue
		}

		if err != nil {
			results = append(results, ValidationResult{Field: field, Message: err.Error()})
		}
	}

	return results
}

func FormatValidationErrors(results []ValidationResult) string {
	var msgs []string
	for _, r := range results {
		msgs = append(msgs, fmt.Sprintf("%s: %s", r.Field, r.Message))
	}
	return strings.Join(msgs, ", ")
}