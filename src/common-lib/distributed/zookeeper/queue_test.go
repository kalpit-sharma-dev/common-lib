package zookeeper

import (
	"errors"
	"reflect"
	"testing"

	"github.com/maraino/go-mock"
	"github.com/samuel/go-zookeeper/zk"
)

func TestCreate(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	type args struct {
		scenario      string
		queueName     string
		errorExpected bool
		expectedError error
	}

	tests := []args{
		{
			scenario:      "success",
			queueName:     "test-1",
			errorExpected: false,
		},
		{
			scenario:      "failure",
			queueName:     "test-2",
			errorExpected: true,
			expectedError: errors.New("injected-error"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			if !test.errorExpected {
				zkMockObj.When("CreateRecursive", getQueueZkPath(test.queueName), nil, int32(0), mock.Any).Return(test.queueName, nil)
				name, err := Queue.Create(test.queueName)

				if err != nil {
					t.Fatalf("expected no error but got %v", err)
				}
				if name != test.queueName {
					t.Fatalf("expected result: %s, got: %s", test.queueName, name)
				}
			} else {
				zkMockObj.When("CreateRecursive", getQueueZkPath(test.queueName), nil, int32(0), mock.Any).Return("", test.expectedError)
				_, err := Queue.Create(test.queueName)

				if err == nil {
					t.Fatalf("expected %v error but did not get any", test.expectedError)
				}
				if err.Error() != test.expectedError.Error() {
					t.Fatalf("expected error: %s, got: %s", test.expectedError.Error(), err.Error())
				}
			}
		})
		zkMockObj.Reset()
	}
}

func TestExists(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	type args struct {
		queueName     string
		scenario      string
		errorExpected bool
		queueExists   bool
		expectedError error
	}

	tests := []args{
		{
			scenario:      "success",
			errorExpected: false,
		},
		{
			scenario:      "failure",
			errorExpected: true,
			expectedError: errors.New("injected-error"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			zkMockObj.When("Exists", getQueueZkPath(test.queueName)).Return(test.queueExists, &zk.Stat{}, test.expectedError)
			exists, err := Queue.Exists(test.queueName)

			if exists != test.queueExists {
				t.Fatalf("expected result: %t, got: %t", test.queueExists, exists)
			}

			if !test.errorExpected && err != nil {
				t.Fatalf("expected no error but got %v", err)

			}

			if test.errorExpected && err == nil {
				t.Fatalf("expected %v error but did not get any", test.expectedError)
				if err.Error() != test.expectedError.Error() {
					t.Fatalf("expected error: %s, got: %s", test.expectedError.Error(), err.Error())
				}
			}
		})
		zkMockObj.Reset()
	}
}

func TestGetList(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	expectedResult := []string{"test"}

	zkMockObj.When("Children", mock.Any).Return(expectedResult, &zk.Stat{}, nil)

	arr, err := Queue.GetList("test")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(arr, expectedResult) {
		t.Errorf("expected result: %s, got: %s", expectedResult, arr)
	}
}

func TestCreateItem(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	expectedResult := "expected result"

	zkMockObj.When("CreateRecursive", mock.Any, mock.Any, mock.Any, mock.Any).Return(expectedResult, nil)

	result, err := Queue.CreateItem([]byte("test"), "test_name")
	if err != nil {
		t.Fatal(err)
	}
	if result != expectedResult {
		t.Errorf("expected result: %s, got: %s", expectedResult, result)
	}
}

func TestGetItemData(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	expectedResult := []byte("test")

	zkMockObj.When("Get", mock.Any).Return(expectedResult, &zk.Stat{}, nil)

	arr, err := Queue.GetItemData("test", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(arr, expectedResult) {
		t.Errorf("expected result: %s, got: %s", expectedResult, arr)
	}
}

func TestRemoveItem(t *testing.T) {
	zkMockObj, originalClient := InitMock()
	defer Restore(originalClient)

	expectedErr := errors.New("some error")

	zkMockObj.When("Delete", mock.Any, mock.Any).Return(expectedErr)

	err := Queue.RemoveItem("test", "test")
	if err != expectedErr {
		t.Errorf("expected err: %s, got: %s", expectedErr, err)
	}
}
