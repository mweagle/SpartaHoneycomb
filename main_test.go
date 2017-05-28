package main

import (
	"encoding/json"
	"testing"

	"net/http/httptest"

	libhoney "github.com/honeycombio/libhoney-go"
	sparta "github.com/mweagle/Sparta"
)

var mockEventJSON = `{
	"key3": "value3",
	"key2": "value2",
	"key1": "value1"
}`

func mockContext() *sparta.LambdaContext {
	return &sparta.LambdaContext{
		AWSRequestID:       "75b4d36b-426e-11e7-a8e8-29a80f7c49a8",
		InvokeID:           "",
		LogGroupName:       "/aws/lambda/SpartaHoneycombTest-mweagleSpartaHoneycombTest-1IYQNUEKTGEGV",
		LogStreamName:      "2017/05/26/[$LATEST]cfdcafcb1a4244bdbf3ffd16cb199654",
		FunctionName:       "SpartaHoneycombTest-mweagleSpartaHoneycombTest-1IYQNUEKTGEGV",
		MemoryLimitInMB:    "128",
		FunctionVersion:    "$LATEST",
		InvokedFunctionARN: "arn:aws:lambda:us-west-2:027159405834:function:SpartaHoneycombTest-mweagleSpartaHoneycombTest-1IYQNUEKTGEGV",
	}
}

func TestHoneycomb(t *testing.T) {
	// Setup a mock event and a request recorder
	requestRecorder := httptest.NewRecorder()
	logger, loggerErr := sparta.NewLogger("info")
	if nil != loggerErr {
		t.Error(loggerErr)
	}
	mockEvent := json.RawMessage(mockEventJSON)
	helloHoneycomb(&mockEvent,
		mockContext(),
		requestRecorder,
		logger)
	libhoney.Close()
}
