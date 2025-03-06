package dbgrpc

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/db_service/internal/models"
	servicemocks "github.com/db_service/internal/servicemoks"
	database "github.com/stipochka/protos/gen/go/db"
	"github.com/stretchr/testify/assert"
)

func TestGetAllRecords(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(m *servicemocks.RecordGetter)
		expectedValue interface{}
		errExpected   bool
	}{
		{
			name: "success",
			mockSetup: func(m *servicemocks.RecordGetter) {
				expectedRecords := []models.Record{
					{ID: 1, Data: "temperature: test, humidity: test"},
					{ID: 2, Data: "temperature: test, humidity: test"},
				}
				m.On("GetAllRecords", context.Background()).Return(expectedRecords, nil)
			},
			expectedValue: &database.RecordsResponse{
				Record: []*database.RecordResponse{
					{Id: 1, Data: "temperature: test, humidity: test"},
					{Id: 2, Data: "temperature: test, humidity: test"},
				},
			},
		},
		{
			name: "internal error",
			mockSetup: func(m *servicemocks.RecordGetter) {
				m.On("GetAllRecords", context.Background()).Return(nil, errors.New("internal error"))
			},
			expectedValue: "rpc error: code = Internal desc = failed to get all records",
			errExpected:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := new(servicemocks.RecordGetter)
			tc.mockSetup(mock)

			server := ServerAPI{storageService: mock}

			resp, err := server.GetAllRecords(context.Background(), nil)

			if err != nil && tc.errExpected {
				assert.Equal(t, tc.expectedValue, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, resp)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestGetRecordByID(t *testing.T) {
	tests := []struct {
		name          string
		recordID      int
		mockSetup     func(m *servicemocks.RecordGetter)
		expectedValue interface{}
		errorMessage  string
	}{
		{
			name:     "success with recordID = 1",
			recordID: 1,
			mockSetup: func(m *servicemocks.RecordGetter) {
				expectedRecord := models.Record{ID: 1, Data: "temperature: test, humidity: test"}
				m.On("GetRecordByID", context.Background(), 1).Return(expectedRecord, nil)
			},
			expectedValue: &database.RecordResponse{Id: 1, Data: "temperature: test, humidity: test"},
		},
		{
			name:     "invalid recordID",
			recordID: -2,
			mockSetup: func(m *servicemocks.RecordGetter) {
				//m.On("GetRecordByID", context.Background(), -2).Return(models.Record{}, nil)
			},
			expectedValue: nil,
			errorMessage:  "rpc error: code = InvalidArgument desc = ID is incorrect",
		},
		{
			name:     "internal error",
			recordID: 1,
			mockSetup: func(m *servicemocks.RecordGetter) {
				m.On("GetRecordByID", context.Background(), 1).Return(models.Record{}, errors.New("internal error"))
			},
			expectedValue: nil,
			errorMessage:  "rpc error: code = NotFound desc = failed to find record",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := new(servicemocks.RecordGetter)
			tc.mockSetup(mock)

			server := ServerAPI{storageService: mock}

			resp, err := server.GetRecordByID(context.Background(), &database.GetByIdRequest{RecordID: int64(tc.recordID)})

			if tc.expectedValue != nil && err != nil {
				log.Fatalf("unexpected error: %v", err)
			}
			if tc.expectedValue == nil && err != nil {
				assert.Equal(t, tc.errorMessage, err.Error())

			} else if err == nil {
				assert.Equal(t, tc.expectedValue, resp)
			}

			mock.AssertExpectations(t)
		})
	}
}
