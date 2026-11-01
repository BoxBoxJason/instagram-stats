package export

import (
	"encoding/json"
	"unicode/utf8"
)

// Text is a string of the export, repaired from the encoding Instagram writes it in.
//
// Instagram writes each byte of the UTF-8 encoding of its strings as a separate
// character: "é" (bytes C3 A9) is written "Ã©", read "Ã©". Decoding a
// Text reverses that, so it holds the string as it was typed.
type Text string

// UnmarshalJSON decodes a JSON string and repairs its encoding.
func (t *Text) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = Text(repairEncoding(s))
	return nil
}

// String returns the text as a string.
func (t Text) String() string {
	return string(t)
}

// repairEncoding returns s with each of its characters taken as a byte, when they
// make valid UTF-8. Otherwise s is not in the broken encoding, and is returned as is.
func repairEncoding(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}

	repaired := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
			return s
		}
		repaired = append(repaired, byte(r))
	}
	if !utf8.Valid(repaired) {
		return s
	}
	return string(repaired)
}
