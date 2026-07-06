package proxy

import (
	"encoding/json"
	"io"
)

func Decode(r io.Reader, emit func(Envelope)) error {
	dec := json.NewDecoder(r)

	for {
		var env Envelope
		if err := dec.Decode(&env); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		emit(env)
	}
}
