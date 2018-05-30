package bystander

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"

	"bystander/structuredstream"

	"github.com/boltdb/bolt"
)

type silencer struct {
	Filters map[string]string `json:"filters"`
	Until   int64             `json:"until"`
	Reason  string            `json:"reason"`
}

func (s *silencer) json() string {
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (s *silencer) key() string {
	return string(serializeFilters(s.Filters))
}

func silencerFromJSON(b []byte) (*silencer, error) {
	s := &silencer{}
	err := json.Unmarshal(b, s)
	return s, err
}

func serializeFilters(f map[string]string) []byte {
	keys := []string{}
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	w := structuredstream.NewWriter(&buf)
	w.Write(uint16(0))

	for _, k := range keys {
		w.WriteUint16PrefixedString(k)
		w.WriteUint16PrefixedString(f[k])
	}
	if err := w.Error(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func deserializeFilters(b []byte) (map[string]string, error) {
	r := structuredstream.NewReader(bytes.NewReader(b))
	version := r.ReadUint16()
	if version != 0 {
		panic("bad version")
	}

	filters := map[string]string{}
	for {
		key := r.ReadUint16PrefixedString()
		if err := r.Error(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		val := r.ReadUint16PrefixedString()
		if err := r.Error(); err != nil {
			//if an EOF happens here, then we read a key without a value, this is bad.
			panic(err)
		}

		filters[key] = val
	}
	return filters, nil
}

func serializeSilencer(s *silencer) ([]byte, []byte) {
	key := serializeFilters(s.Filters)

	var buf bytes.Buffer
	w := structuredstream.NewWriter(&buf)
	w.Write(s.Until)
	w.WriteUint16PrefixedString(s.Reason)
	if err := w.Error(); err != nil {
		panic(err)
	}
	val := buf.Bytes()

	return key, val
}

func deserializeSilencer(key, val []byte) (*silencer, error) {
	filters, err := deserializeFilters(key)
	if err != nil {
		return nil, err
	}

	r := structuredstream.NewReader(bytes.NewReader(val))
	s := &silencer{
		Filters: filters,
		Until:   r.ReadInt64(),
		Reason:  r.ReadUint16PrefixedString(),
	}
	if err = r.Error(); err != nil {
		return nil, err
	}
	return s, nil
}

func loadSilencers(db *bolt.DB) (map[string]*silencer, error) {
	silencers := map[string]*silencer{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("silencers"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			s, err := deserializeSilencer(k, v)
			if err != nil {
				return err
			}
			silencers[s.key()] = s
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return silencers, nil
}
