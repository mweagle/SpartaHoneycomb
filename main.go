package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"

	"sync"

	gocf "github.com/crewjam/go-cloudformation"
	"github.com/honeycombio/libhoney-go"
	sparta "github.com/mweagle/Sparta"
	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaVault "github.com/mweagle/SpartaVault/encrypt"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////

// HoneycombWriteKey is the SpartaVault encrypted WriteKey
var HoneycombWriteKey = &spartaVault.KMSEncryptedValue{
	KMSKeyARNOrGUID: "4f2f62e1-41e0-49e2-8da4-3a7ec511f498",
	PropertyName:    "HoneycombWriteKey",
	Key:             "AQEDAHi8zBTBrgXJ4OyfnaJ8C9B2H/WAF54D9vPaarH9Dob2wwAAAH4wfAYJKoZIhvcNAQcGoG8wbQIBADBoBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDBJWM9Dux5p3p4Y0fwIBEIA7sQkB36bHSDbe/wNlQBByUtyn2UH71fI6sVwTMS9N/w8OsAsX5glKFzBoJmjG6j95Qu+RSQSOivZIoCM=",
	Nonce:           "CyThLSrKbxMSklte",
	Value:           "ydRruv9+fUtJAIndvAquKwqTfMi3DEs7YzwLv/36U9tjc9eZ1MM+XUvOMUjAf9ZG",
	Created:         "2017-05-27T19:16:26-07:00",
}

// one time logger registration
var oneTime sync.Once

////////////////////////////////////////////////////////////////////////////////
// Honeycomb.io Logrus hook
////////////////////////////////////////////////////////////////////////////////
type honeycombHook struct {
}

func (hook *honeycombHook) Fire(entry *logrus.Entry) error {
	eventBuilder := libhoney.NewBuilder()
	honeycombEvent := eventBuilder.NewEvent()
	for eachKey, eachValue := range entry.Data {
		honeycombEvent.AddField(eachKey, eachValue)
	}
	honeycombEvent.AddField("ts", entry.Time)
	honeycombEvent.Send()
	return nil
}

func (hook *honeycombHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

////////////////////////////////////////////////////////////////////////////////
// Return a new Honeycomb.io logrus hook
////////////////////////////////////////////////////////////////////////////////
func newHoneycombHook(writeKey string, datasetName string) logrus.Hook {
	libhoney.Init(libhoney.Config{
		WriteKey: writeKey,
		Dataset:  datasetName,
	})

	// We want every event to include the number of currently running goroutines
	libhoney.AddDynamicField("num_goroutines",
		func() interface{} { return runtime.NumGoroutine() })
	return &honeycombHook{}
}

// Honeycomb.io function
func helloHoneycomb(ctx context.Context) (string, error) {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Lazily register the logrus logging hook
	oneTime.Do(func() {
		key, keyErr := HoneycombWriteKey.Decrypt(spartaAWS.NewSession(logger))
		if nil != keyErr {
			logger.Error("Failed to decrypt WriteKey: " + keyErr.Error())
		} else {
			honeycombLoggingHook := newHoneycombHook(string(key),
				"LambdaDataset")
			logger.Hooks.Add(honeycombLoggingHook)
			logger.Out = ioutil.Discard
		}
	})

	logger.WithFields(logrus.Fields{
		"bee_stings": rand.Int31n(10),
	}).Info("'tis only a flesh wound")
	return "Hello 🐝", nil
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Ensure that the IAM role definition includes access to the
	// KMS key used to d/encrypt the Honeycomb write key
	var iamLambdaRole = sparta.IAMRoleDefinition{}
	kmsARN := gocf.Join("",
		gocf.String("arn:aws:kms:"),
		gocf.Ref("AWS::Region"),
		gocf.String(":"),
		gocf.Ref("AWS::AccountId"),
		gocf.String(":key/"),
		gocf.String(HoneycombWriteKey.KMSKeyARNOrGUID))
	iamLambdaRole.Privileges = append(iamLambdaRole.Privileges, sparta.IAMRolePrivilege{
		Actions:  []string{"kms:Decrypt"},
		Resource: kmsARN})
	lambdaFn := sparta.HandleAWSLambda("helloHoneycomb",
		helloHoneycomb,
		sparta.IAMRoleDefinition{})
	lambdaFn.Options = &sparta.LambdaFunctionOptions{
		Description: "Sample Honeycomb.io function",
		Timeout:     10,
		MemorySize:  128,
	}

	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	// Use the CGO version of this function
	stackName := spartaCF.UserScopedStackName("SpartaHoneycomb")
	err := sparta.Main(stackName,
		fmt.Sprintf("Test sending events to Honeycomb.io"),
		lambdaFunctions,
		nil,
		nil)
	if err != nil {
		os.Exit(1)
	}
}
