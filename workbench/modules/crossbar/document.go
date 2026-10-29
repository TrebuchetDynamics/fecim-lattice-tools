package crossbar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const maxDocumentBytes = 1 << 20

func decodeDocument(data []byte, limit int) (Document, error) {
	var doc Document
	if len(data) > maxDocumentBytes {
		return doc, fmt.Errorf("crossbar document exceeds 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return doc, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return doc, fmt.Errorf("trailing crossbar document")
	}
	if err := doc.Config.validate(limit); err != nil {
		return doc, err
	}
	if err := validateData(doc.Config, doc.Matrix, doc.Input); err != nil {
		return doc, err
	}
	return doc, nil
}
