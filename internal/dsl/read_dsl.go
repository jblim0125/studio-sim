package dsl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// SampleDSL dsl.json 구조
type SampleDSL struct {
	DSLs map[string]interface{}
}

// ReadSampleDSL 주어진 경로 있는 json 파일을 읽어들인다.
func ReadSampleDSL(path string) (*map[string]interface{}, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fail[ %s ]", err.Error())
	}
	sample := SampleDSL{}
	err = json.Unmarshal(b, &sample)
	if err != nil {
		return nil, fmt.Errorf("fail unmarshal from file[ %s ]", err.Error())
	}
	return &sample.DSLs, nil
}
