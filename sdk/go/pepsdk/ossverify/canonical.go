package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// maxSafeInteger is the largest integer exactly representable by IEEE-754
// double, which RFC 8785 requires for interoperable numbers.
const maxSafeInteger int64 = 9007199254740991

// Marshal serializes the constrained draft subset of RFC 8785 JCS:
// maps with string keys, arrays, strings, booleans, null and safe integers.
// Floating point and non-integral numbers are rejected until full JCS number
// formatting is reviewed. This is contract tooling, not an enabled issuer.
func Marshal(value any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeCanonical(&buffer, value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeCanonical(buffer *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		buffer.WriteString("null")
	case bool:
		if typed {
			buffer.WriteString("true")
		} else {
			buffer.WriteString("false")
		}
	case string:
		return writeCanonicalString(buffer, typed)
	case json.Number:
		return writeCanonicalNumber(buffer, typed.String())
	case float64:
		return writeCanonicalFloat(buffer, typed)
	case int:
		return writeCanonicalInteger(buffer, int64(typed))
	case int64:
		return writeCanonicalInteger(buffer, typed)
	case []any:
		buffer.WriteByte('[')
		for index, element := range typed {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeCanonical(buffer, element); err != nil {
				return err
			}
		}
		buffer.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(left, right int) bool { return utf16Less(keys[left], keys[right]) })
		buffer.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeCanonicalString(buffer, key); err != nil {
				return err
			}
			buffer.WriteByte(':')
			if err := writeCanonical(buffer, typed[key]); err != nil {
				return err
			}
		}
		buffer.WriteByte('}')
	default:
		return fmt.Errorf("canonical JSON does not support %T", value)
	}
	return nil
}

func writeCanonicalString(buffer *bytes.Buffer, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("canonical JSON requires valid UTF-8")
	}
	buffer.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			buffer.WriteString(`\"`)
		case '\\':
			buffer.WriteString(`\\`)
		case '\b':
			buffer.WriteString(`\b`)
		case '\f':
			buffer.WriteString(`\f`)
		case '\n':
			buffer.WriteString(`\n`)
		case '\r':
			buffer.WriteString(`\r`)
		case '\t':
			buffer.WriteString(`\t`)
		default:
			if character < 0x20 {
				fmt.Fprintf(buffer, `\u%04x`, character)
				continue
			}
			buffer.WriteRune(character)
		}
	}
	buffer.WriteByte('"')
	return nil
}

func writeCanonicalNumber(buffer *bytes.Buffer, value string) error {
	if !validJSONNumber(value) {
		return fmt.Errorf("invalid JSON number %q", value)
	}
	integer := true
	for _, character := range value {
		if character == '.' || character == 'e' || character == 'E' {
			integer = false
			break
		}
	}
	if integer {
		if value == "-0" {
			buffer.WriteByte('0')
			return nil
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("canonical integer is out of range: %q", value)
		}
		return writeCanonicalInteger(buffer, parsed)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid JSON number %q", value)
	}
	return writeCanonicalFloat(buffer, parsed)
}

func writeCanonicalFloat(buffer *bytes.Buffer, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("canonical JSON rejects NaN and infinities")
	}
	buffer.WriteString(formatES6Number(value))
	return nil
}

// formatES6Number implements the ECMAScript Number::toString ordering required
// by RFC 8785 for the shortest round-trippable double.
func formatES6Number(value float64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	absolute := math.Abs(value)
	representation := strconv.FormatFloat(absolute, 'e', -1, 64)
	mantissa, exponentText, _ := strings.Cut(representation, "e")
	exponent, _ := strconv.Atoi(exponentText)
	digits := strings.Replace(mantissa, ".", "", 1)
	k := len(digits)
	n := exponent + 1

	var formatted string
	switch {
	case k <= n && n <= 21:
		formatted = digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		formatted = digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		formatted = "0." + strings.Repeat("0", -n) + digits
	default:
		exponential := n - 1
		sign := "+"
		if exponential < 0 {
			sign = "-"
			exponential = -exponential
		}
		if k == 1 {
			formatted = digits
		} else {
			formatted = digits[:1] + "." + digits[1:]
		}
		formatted += "e" + sign + strconv.Itoa(exponential)
	}
	if negative {
		return "-" + formatted
	}
	return formatted
}

func validJSONNumber(value string) bool {
	if value == "" {
		return false
	}
	index := 0
	if value[index] == '-' {
		index++
		if index == len(value) {
			return false
		}
	}
	if value[index] == '0' {
		index++
		if index < len(value) && value[index] >= '0' && value[index] <= '9' {
			return false
		}
	} else if value[index] >= '1' && value[index] <= '9' {
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			index++
		}
	} else {
		return false
	}
	if index < len(value) && value[index] == '.' {
		index++
		start := index
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			index++
		}
		if index == start {
			return false
		}
	}
	if index < len(value) && (value[index] == 'e' || value[index] == 'E') {
		index++
		if index < len(value) && (value[index] == '+' || value[index] == '-') {
			index++
		}
		start := index
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			index++
		}
		if index == start {
			return false
		}
	}
	return index == len(value)
}

func writeCanonicalInteger(buffer *bytes.Buffer, value int64) error {
	if value > maxSafeInteger || value < -maxSafeInteger {
		return fmt.Errorf("integer %d exceeds the interoperable safe range", value)
	}
	buffer.WriteString(strconv.FormatInt(value, 10))
	return nil
}

func utf16Less(left, right string) bool {
	leftUnits := utf16.Encode([]rune(left))
	rightUnits := utf16.Encode([]rune(right))
	for index := 0; index < len(leftUnits) && index < len(rightUnits); index++ {
		if leftUnits[index] != rightUnits[index] {
			return leftUnits[index] < rightUnits[index]
		}
	}
	return len(leftUnits) < len(rightUnits)
}
