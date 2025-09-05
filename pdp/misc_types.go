package pdp

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TimePrecision sets the precision of times and dates within this library. This
// has an influence on the precision of times when comparing expiry or other
// related time fields. Furthermore, it is also the precision of times when
// serializing.
//
// For backwards compatibility the default precision is set to seconds, so that
// no fractional timestamps are generated.
var TimePrecision = time.Second

// MarshalSingleStringAsArray modifies the behavior of the ClaimStrings type,
// especially its MarshalJSON function.
//
// If it is set to true (the default), it will always serialize the type as an
// array of strings, even if it just contains one element, defaulting to the
// behavior of the underlying []string. If it is set to false, it will serialize
// to a single string, if it contains one element. Otherwise, it will serialize
// to an array of strings.
var MarshalSingleStringAsArray = true

// NewNumericDate constructs a new *NumericDate from a standard library time.Time struct.
// It will truncate the timestamp according to the precision specified in TimePrecision.
func NewNumericDate(t time.Time) *jwt.NumericDate {
	return &jwt.NumericDate{Time: t.Truncate(TimePrecision)}
}

// newNumericDateFromSeconds creates a new *NumericDate out of a float64 representing a
// UNIX epoch with the float fraction representing non-integer seconds.
func newNumericDateFromSeconds(f float64) *jwt.NumericDate {
	round, frac := math.Modf(f)
	return NewNumericDate(time.Unix(int64(round), int64(frac*1e9)))
}

// ClaimStrings is basically just a slice of strings, but it can be either
// serialized from a string array or just a string. This type is necessary,
// since the "aud" claim can either be a single string or an array.
type ClaimStrings []string

func (s *ClaimStrings) UnmarshalJSON(data []byte) (err error) {
	var value any

	if err = json.Unmarshal(data, &value); err != nil {
		return err
	}

	var aud []string

	switch v := value.(type) {
	case string:
		aud = append(aud, v)
	case []string:
		aud = ClaimStrings(v)
	case []any:
		for _, vv := range v {
			vs, ok := vv.(string)
			if !ok {
				return jwt.ErrInvalidType
			}
			aud = append(aud, vs)
		}
	case nil:
		return nil
	default:
		return jwt.ErrInvalidType
	}

	*s = aud

	return
}

func (s ClaimStrings) MarshalJSON() (b []byte, err error) {
	// This handles a special case in the JWT RFC. If the string array, e.g.
	// used by the "aud" field, only contains one element, it MAY be serialized
	// as a single string. This may or may not be desired based on the ecosystem
	// of other JWT library used, so we make it configurable by the variable
	// MarshalSingleStringAsArray.
	if len(s) == 1 && !MarshalSingleStringAsArray {
		return json.Marshal(s[0])
	}

	return json.Marshal([]string(s))
}

// newError creates a new error message with a detailed error message. The
// message will be prefixed with the contents of the supplied error type.
// Additionally, more errors, that provide more context can be supplied which
// will be appended to the message. This makes use of Go 1.20's possibility to
// include more than one %w formatting directive in [fmt.Errorf].
//
// For example,
//
//	newError("no keyfunc was provided", ErrTokenUnverifiable)
//
// will produce the error string
//
//	"token is unverifiable: no keyfunc was provided"
func newError(message string, err error, more ...error) error {
	var format string
	var args []any
	if message != "" {
		format = "%w: %s"
		args = []any{err, message}
	} else {
		format = "%w"
		args = []any{err}
	}

	for _, e := range more {
		format += ": %w"
		args = append(args, e)
	}

	err = fmt.Errorf(format, args...)
	return err
}
