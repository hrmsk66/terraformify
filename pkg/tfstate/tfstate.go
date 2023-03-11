package tfstate

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/itchyny/gojq"
)

type TFState struct {
	Value interface{}
}

type TFStateWithTemplate struct {
	*template.Template
	*TFState
}

func (s *TFState) AddTemplate(tmpl string) (*TFStateWithTemplate, error) {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid template: %w", err)
	}

	return &TFStateWithTemplate{t, s}, nil
}

func Load(workingDir string) (state *TFState, err error) {
	file := filepath.Join(workingDir, "terraform.tfstate")
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = f.Close()
	}()

	var s TFState
	if err := json.NewDecoder(f).Decode(&s.Value); err != nil {
		return nil, fmt.Errorf("tfstate: invalid json: %w", err)
	}

	return &s, nil
}

func (s TFState) Bytes() []byte {
	switch v := (s.Value).(type) {
	case string:
		return []byte(v)
	default:
		b, _ := json.Marshal(v)
		return b
	}
}

func (s TFState) String() string {
	return string(s.Bytes())
}

func (s *TFState) Query(query string) (*TFState, error) {
	jq, err := gojq.Parse(query)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	iter := jq.Run(s.Value)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		results = append(results, v)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("tfstate: %s is not found in the state", query)
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("tfstate: %s returned multiple results", query)
	}
	return &TFState{Value: results[0]}, nil
}
