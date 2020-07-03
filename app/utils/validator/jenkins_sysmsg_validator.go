package validator

import "errors"

// ValidateJenkinsSystemMessage validates the Jenkins System Message
func ValidateJenkinsSystemMessage(input string) error {
	// a Jenkins System Message cannot be longer than 255 characters
	if len(input) > 255 {
		return errors.New("Should not be longer than 255 characters. ")
	}
	return nil
}
