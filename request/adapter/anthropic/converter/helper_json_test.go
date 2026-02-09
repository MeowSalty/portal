package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeToolInput_ExistingString(t *testing.T) {
	existingArgs := `{"key":"value"}`
	result, err := SerializeToolInput(nil, &existingArgs)
	assert.NoError(t, err)
	assert.Equal(t, existingArgs, result)
}

func TestSerializeToolInput_PayloadSerialization(t *testing.T) {
	payload := map[string]interface{}{"key": "value"}
	result, err := SerializeToolInput(payload, nil)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"key":"value"}`, result)
}

func TestSerializePayload(t *testing.T) {
	payload := map[string]interface{}{"key": "value", "number": 123}
	result, err := SerializePayload(payload)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"key":"value","number":123}`, result)
}

func TestDeserializeToolInput(t *testing.T) {
	args := `{"key":"value","number":123}`
	result, err := DeserializeToolInput(args)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
	assert.Equal(t, float64(123), result["number"])
}

func TestDeserializePayload(t *testing.T) {
	jsonStr := `{"key":"value","nested":{"inner":"value"}}`
	result, err := DeserializePayload(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
	assert.NotNil(t, result["nested"])
}

func TestSerializeToolInput_ErrorHandling(t *testing.T) {
	result, err := SerializeToolInput(make(chan int), nil)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestDeserializeToolInput_ErrorHandling(t *testing.T) {
	result, err := DeserializeToolInput("invalid json")
	assert.Error(t, err)
	assert.Nil(t, result)
}
