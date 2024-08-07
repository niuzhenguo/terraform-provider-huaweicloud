package cae

import (
	"encoding/json"
	"log"
)

func unmarshalJsonFormatParamster(paramName, paramVal string) map[string]interface{} {
	parseResult := make(map[string]interface{})
	err := json.Unmarshal([]byte(paramVal), &parseResult)
	if err != nil {
		log.Printf("[ERROR] Invalid type of the %s, it's not JSON format", paramName)
	}
	return parseResult
}

func marshalJsonFormatParamster(paramName, paramVal interface{}) interface{} {
	jsonDetail, err := json.Marshal(paramVal)
	if err != nil {
		log.Printf("[ERROR] unable to convert the %s, it's not JSON format", paramName)
		return nil
	}
	return string(jsonDetail)
}
