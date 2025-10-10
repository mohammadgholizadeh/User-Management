package domain

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type NationalID string

func (n NationalID) Validate() error {
	actualLen := len(n)

	if actualLen > 10 || actualLen < 8 {
		zap.L().Warn("National ID length out of bounds",
			zap.String("entity", "national_id"),
			zap.String("value", string(n)),
			zap.Int("length", actualLen),
		)
		return fmt.Errorf("national ID length must be between 8 and 10 characters")
	}

	paddedN := string(n)
	if actualLen < 10 {
		paddedN = strings.Repeat("0", 10-actualLen) + paddedN
	}

	digits := strings.Split(paddedN, "")
	var sum int

	controlDigit, err := strconv.Atoi(digits[9])
	if err != nil {
		zap.L().Warn("National ID control digit is not a number",
			zap.String("entity", "national_id"),
			zap.String("value", string(n)),
			zap.Error(err),
		)
		return fmt.Errorf("national ID must only contain numeric characters")
	}

	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(digits[i])
		if err != nil {
			zap.L().Warn("National ID digit is not a number",
				zap.String("entity", "national_id"),
				zap.String("value", string(n)),
				zap.Int("index", i),
				zap.Error(err),
			)
			return fmt.Errorf("national ID must only contain numeric characters")
		}
		sum += digit * (10 - i)
	}

	remainder := sum % 11
	isValid := (remainder < 2 && remainder == controlDigit) || (remainder >= 2 && (11-remainder) == controlDigit)

	if !isValid {
		zap.L().Warn("National ID checksum mismatch",
			zap.String("entity", "national_id"),
			zap.String("value", string(n)),
			zap.Int("remainder", remainder),
			zap.Int("control_digit", controlDigit),
		)
		return fmt.Errorf("national ID is invalid")
	}

	return nil
}
