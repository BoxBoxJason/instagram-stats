package export

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// maxByte is the largest value of a byte.
const maxByte = 0xFF

// Text is a string of the export, repaired from the encoding Instagram writes it in.
//
// Instagram writes each byte of the UTF-8 encoding of its strings as a separate
// character: "é" (bytes C3 A9) is written "Ã©", read "Ã©". Decoding a
// Text reverses that, so it holds the string as it was typed.
type Text string

// UnmarshalJSON decodes a JSON string and repairs its encoding.
func (t *Text) UnmarshalJSON(data []byte) error {
	var decoded string

	err := json.Unmarshal(data, &decoded)
	if err != nil {
		return fmt.Errorf("decoding a text: %w", err)
	}

	*t = Text(repairEncoding(decoded))

	return nil
}

// String returns the text as a string.
func (t Text) String() string {
	return string(t)
}

// repairEncoding returns text with each of its characters taken as a byte, when they
// make valid UTF-8. Otherwise text is not in the broken encoding, and is returned as is.
func repairEncoding(text string) string {
	ascii := true

	for i := range len(text) {
		if text[i] >= utf8.RuneSelf {
			ascii = false

			break
		}
	}

	if ascii {
		return text
	}

	repaired := make([]byte, 0, len(text))

	for _, char := range text {
		if char > maxByte {
			return text
		}

		repaired = append(repaired, byte(char&maxByte))
	}

	if !utf8.Valid(repaired) {
		return text
	}

	return string(repaired)
}
