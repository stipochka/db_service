package dbgrpc

import (
	"context"
	"fmt"

	"github.com/db_service/internal/models"
	"github.com/db_service/internal/service"
	database "github.com/stipochka/protos/gen/go/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAPI struct {
	database.UnimplementedDatabaseServer
	storageService service.RecordGetter
}

//GetRecordByID(context.Context, *GetByIdRequest) (*RecordResponse, error)
//	GetAllRecords(context.Context, *GetAllRecordsRequest) (*RecordResponse, error)

func Register(gRPCServer *grpc.Server, storageService service.RecordGetter) {
	database.RegisterDatabaseServer(gRPCServer, &ServerAPI{storageService: storageService})
}

func (s *ServerAPI) GetRecordByID(ctx context.Context, req *database.GetByIdRequest) (*database.RecordResponse, error) {
	if req.RecordID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "ID is incorrect")
	}

	record, err := s.storageService.GetRecordByID(ctx, int(req.GetRecordID()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "failed to find record")
	}

	return RecordToRecordResponse(record), nil

}

func (s *ServerAPI) GetAllRecords(ctx context.Context, req *database.GetAllRecordsRequest) (*database.RecordsResponse, error) {
	recordsResponse := make([]*database.RecordResponse, 0)

	records, err := s.storageService.GetAllRecords(ctx)

	if err != nil {
		fmt.Println("error", err.Error())
		return nil, status.Error(codes.Internal, "failed to get all records")
	}

	for _, record := range records {
		recordsResponse = append(recordsResponse, RecordToRecordResponse(record))
	}

	return &database.RecordsResponse{Records: recordsResponse}, nil
}

func RecordToRecordResponse(record models.Record) *database.RecordResponse {
	return &database.RecordResponse{
		Id:            int32(record.ID),
		Status:        record.Status,
		PollingPeriod: int32(record.PollingPeriod),
		Temperature:   int32(record.Temperature),
		LampStatus:    record.LampStatus,
		Voltage:       int32(record.Voltage),
		Thresholds: &database.Thresholds{
			Temperature: SensorDataToSensorDataResponse(record.Thresholds.Temperature),
			Humidity:    SensorDataToSensorDataResponse(record.Thresholds.Humidity),
			Voltage:     SensorDataToSensorDataResponse(record.Thresholds.Voltage),
		},
	}
}

func SensorDataToSensorDataResponse(records []models.SensorData) []*database.SensorData {
	sensorDataResponse := make([]*database.SensorData, 0)

	for _, record := range records {
		sensorDataResponse = append(sensorDataResponse, &database.SensorData{
			Value:         int32(record.Value),
			PollingPeriod: int32(record.PollingPeriod),
		})
	}

	return sensorDataResponse
}
