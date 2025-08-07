package headers

import (
	"strings"
	"bytes"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	header := make(map[string]string)
	return header
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	// fmt.Printf("%q\n", data)

	if idx == -1 { // there is no CRLF
		return 0, false, nil
	} else if idx == 0 { // data starts with a CRLF, so i reached end of the header
		return 2, true, nil
	}

	fieldLineString := string(data[:idx])
	fieldName, fieldValue, err := fieldLineFromString(fieldLineString)
	if err != nil {
		return 0, false, err
	}

	fieldName = strings.ToLower(fieldName)
	if _, ok := h[fieldName]; ok {
		h[fieldName] += ", " + fieldValue
	} else {
		h[fieldName] = fieldValue
	}

	return len(data[:idx])+2, false, nil
}

func fieldLineFromString(str string) (fieldName string, fieldValue string, err error) {
	fieldLineParts := strings.Split(str, ":")
	fieldName = strings.TrimLeft(fieldLineParts[0], " ")
	if err := isValidFieldName(fieldName); err != nil {
		return "", "", err
	}

	fieldValue = strings.Join(fieldLineParts[1:], ":")
	fieldValue = strings.Trim(fieldValue, " ")

	if strings.Contains(fieldName, " ") {
		return "", "", fmt.Errorf("error: there must be no spaces betwixt the colon and the field-name")
	}

	return fieldName, fieldValue, nil
}

func isValidFieldName(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("error: the field name length must be of at least 1")
	}
	const allowedSpecialCharacters = "!#$%&'*+-.^_`|~"

	for _, c := range str {
		isLetter := (c > 'A' && c < 'Z' ) || (c > 'a' && c < 'z' )
		isNumber := c >= 0 && c <= 9
		isSpecialCharacter := strings.ContainsRune(allowedSpecialCharacters, c)

		if !isLetter && !isNumber && !isSpecialCharacter {
			return fmt.Errorf("error: invalid character %q in the field name", c)
		}
	}

	return nil
}
