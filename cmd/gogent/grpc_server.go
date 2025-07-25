package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"gogent/internal/auth"
	"gogent/internal/types"
	pb "gogent/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCServer implements the GogentServiceServer interface
type GRPCServer struct {
	pb.UnimplementedGogentServiceServer
	businessLogic *BusinessLogic
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer() (*GRPCServer, error) {
	businessLogic, err := NewBusinessLogic()
	if err != nil {
		return nil, fmt.Errorf("failed to create business logic: %w", err)
	}

	return &GRPCServer{
		businessLogic: businessLogic,
	}, nil
}

// Close closes the server resources
func (s *GRPCServer) Close() error {
	if s.businessLogic != nil {
		return s.businessLogic.Close()
	}
	return nil
}

// =============================================================================
// AUTHENTICATION & USER MANAGEMENT
// =============================================================================

func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, token, expiresAt, err := s.businessLogic.LoginUser(req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Login failed: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.LoginResponse{
		Token:     token,
		User:      protoUser,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

func (s *GRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, token, err := s.businessLogic.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Registration failed: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.RegisterResponse{
		User:  protoUser,
		Token: token,
	}, nil
}

func (s *GRPCServer) CreateTemporaryUser(ctx context.Context, req *pb.CreateTemporaryUserRequest) (*pb.CreateTemporaryUserResponse, error) {
	user, tempPassword, token, err := s.businessLogic.CreateTemporaryUser(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create temporary user: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.CreateTemporaryUserResponse{
		User:              protoUser,
		TemporaryPassword: tempPassword,
		Token:             token,
	}, nil
}

func (s *GRPCServer) SaveTemporaryAccount(ctx context.Context, req *pb.SaveTemporaryAccountRequest) (*pb.SaveTemporaryAccountResponse, error) {
	user, emailSent, err := s.businessLogic.SaveTemporaryAccount(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save temporary account: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.SaveTemporaryAccountResponse{
		User:      protoUser,
		EmailSent: emailSent,
	}, nil
}

func (s *GRPCServer) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	user, verified, err := s.businessLogic.VerifyEmail(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Email verification failed: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.VerifyEmailResponse{
		User:     protoUser,
		Verified: verified,
	}, nil
}

func (s *GRPCServer) GetCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.GetCurrentUserResponse, error) {
	user, err := s.businessLogic.GetCurrentUser()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Failed to get current user: %v", err)
	}

	protoUser := s.convertUserToProto(user)
	return &pb.GetCurrentUserResponse{
		User: protoUser,
	}, nil
}

// =============================================================================
// EXECUTION MANAGEMENT
// =============================================================================

func (s *GRPCServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	// Convert proto request to internal types
	multiReq := &types.MultiExecutionRequest{
		ExecutionRunName:      req.ExecutionRunName,
		Description:           req.Description,
		BasePrompt:            req.BasePrompt,
		Context:               req.Context,
		EnableFunctionCalling: req.EnableFunctionCalling,
		Configurations:        s.convertProtoConfigurations(req.Configurations),
	}

	// Create additional config for external APIs
	additionalConfig := &types.GeminiClientConfig{
		OpenWeatherAPIKey: req.OpenweatherApiKey,
		Neo4jURL:          req.Neo4JUrl,
		Neo4jUsername:     req.Neo4JUsername,
		Neo4jPassword:     req.Neo4JPassword,
		Neo4jDatabase:     req.Neo4JDatabase,
	}

	executionID, executionRun, err := s.businessLogic.StartExecution(multiReq, req.UseMock, additionalConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to start execution: %v", err)
	}

	protoExecutionRun := s.convertExecutionRunToProto(executionRun)
	return &pb.ExecuteResponse{
		ExecutionId:  executionID,
		Message:      "Execution started. Use GetExecutionStatus to check progress.",
		ExecutionRun: protoExecutionRun,
	}, nil
}

func (s *GRPCServer) GetExecutionStatus(ctx context.Context, req *pb.GetExecutionStatusRequest) (*pb.GetExecutionStatusResponse, error) {
	execStatus, startTime, endTime, errorMessage, result, err := s.businessLogic.GetExecutionStatus(ctx, req.ExecutionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	response := &pb.GetExecutionStatusResponse{
		Status:    execStatus,
		StartTime: timestamppb.New(startTime),
	}

	if endTime != nil {
		response.EndTime = timestamppb.New(*endTime)
	}

	if errorMessage != "" {
		response.ErrorMessage = errorMessage
	}

	if result != nil {
		protoResult, err := s.convertExecutionResultToProto(result)
		if err == nil {
			response.Result = protoResult
		}
	}

	return response, nil
}

func (s *GRPCServer) GetExecutionResult(ctx context.Context, req *pb.GetExecutionResultRequest) (*pb.GetExecutionResultResponse, error) {
	result, err := s.businessLogic.GetExecutionResult(ctx, req.ExecutionRunId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Execution result not found: %v", err)
	}

	protoResult, err := s.convertExecutionResultToProto(result)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert result: %v", err)
	}

	return &pb.GetExecutionResultResponse{
		Result: protoResult,
	}, nil
}

func (s *GRPCServer) ListExecutionRuns(ctx context.Context, req *pb.ListExecutionRunsRequest) (*pb.ListExecutionRunsResponse, error) {
	runs, err := s.businessLogic.ListExecutionRuns(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list execution runs: %v", err)
	}

	var protoRuns []*pb.ExecutionRun
	for _, run := range runs {
		protoRun := s.convertExecutionRunToProto(run)
		protoRuns = append(protoRuns, protoRun)
	}

	return &pb.ListExecutionRunsResponse{
		ExecutionRuns: protoRuns,
		TotalCount:    int32(len(protoRuns)),
	}, nil
}

func (s *GRPCServer) DeleteExecutionRun(ctx context.Context, req *pb.DeleteExecutionRunRequest) (*pb.DeleteExecutionRunResponse, error) {
	err := s.businessLogic.DeleteExecutionRun(ctx, req.ExecutionRunId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete execution run: %v", err)
	}

	return &pb.DeleteExecutionRunResponse{
		Message: fmt.Sprintf("Execution run %s deleted successfully", req.ExecutionRunId),
	}, nil
}

// =============================================================================
// CONFIGURATION MANAGEMENT
// =============================================================================

func (s *GRPCServer) ListConfigurations(ctx context.Context, req *pb.ListConfigurationsRequest) (*pb.ListConfigurationsResponse, error) {
	configs := s.businessLogic.GetDefaultConfigurations()
	var protoConfigs []*pb.APIConfiguration

	for _, config := range configs {
		protoConfig := s.convertConfigurationToProto(&config)
		protoConfigs = append(protoConfigs, protoConfig)
	}

	return &pb.ListConfigurationsResponse{
		Configurations: protoConfigs,
	}, nil
}

func (s *GRPCServer) CreateConfiguration(ctx context.Context, req *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error) {
	config := s.convertProtoConfigurationToInternal(req.Configuration)

	createdConfig, err := s.businessLogic.CreateConfiguration(config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create configuration: %v", err)
	}

	protoConfig := s.convertConfigurationToProto(createdConfig)
	return &pb.CreateConfigurationResponse{
		Configuration: protoConfig,
	}, nil
}

func (s *GRPCServer) UpdateConfiguration(ctx context.Context, req *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error) {
	config := s.convertProtoConfigurationToInternal(req.Configuration)

	updatedConfig, err := s.businessLogic.UpdateConfiguration(req.Id, config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update configuration: %v", err)
	}

	protoConfig := s.convertConfigurationToProto(updatedConfig)
	return &pb.UpdateConfigurationResponse{
		Configuration: protoConfig,
	}, nil
}

func (s *GRPCServer) DeleteConfiguration(ctx context.Context, req *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error) {
	err := s.businessLogic.DeleteConfiguration(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete configuration: %v", err)
	}

	return &pb.DeleteConfigurationResponse{
		Message: fmt.Sprintf("Configuration %s deleted successfully", req.Id),
	}, nil
}

// =============================================================================
// FUNCTION MANAGEMENT
// =============================================================================

func (s *GRPCServer) ListFunctions(ctx context.Context, req *pb.ListFunctionsRequest) (*pb.ListFunctionsResponse, error) {
	functions, err := s.businessLogic.ListFunctions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list functions: %v", err)
	}

	var protoFunctions []*pb.FunctionDefinition
	for _, function := range functions {
		protoFunction := s.convertFunctionToProto(function)
		protoFunctions = append(protoFunctions, protoFunction)
	}

	return &pb.ListFunctionsResponse{
		Functions: protoFunctions,
	}, nil
}

func (s *GRPCServer) GetFunction(ctx context.Context, req *pb.GetFunctionRequest) (*pb.GetFunctionResponse, error) {
	function, err := s.businessLogic.GetFunction(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Function not found: %v", err)
	}

	protoFunction := s.convertFunctionToProto(function)
	return &pb.GetFunctionResponse{
		Function: protoFunction,
	}, nil
}

func (s *GRPCServer) CreateFunction(ctx context.Context, req *pb.CreateFunctionRequest) (*pb.CreateFunctionResponse, error) {
	function := s.convertProtoFunctionToInternal(req.Function)

	createdFunction, err := s.businessLogic.CreateFunction(function)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create function: %v", err)
	}

	protoFunction := s.convertFunctionToProto(createdFunction)
	return &pb.CreateFunctionResponse{
		Function: protoFunction,
	}, nil
}

func (s *GRPCServer) UpdateFunction(ctx context.Context, req *pb.UpdateFunctionRequest) (*pb.UpdateFunctionResponse, error) {
	function := s.convertProtoFunctionToInternal(req.Function)

	updatedFunction, err := s.businessLogic.UpdateFunction(req.Id, function)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update function: %v", err)
	}

	protoFunction := s.convertFunctionToProto(updatedFunction)
	return &pb.UpdateFunctionResponse{
		Function: protoFunction,
	}, nil
}

func (s *GRPCServer) DeleteFunction(ctx context.Context, req *pb.DeleteFunctionRequest) (*pb.DeleteFunctionResponse, error) {
	err := s.businessLogic.DeleteFunction(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete function: %v", err)
	}

	return &pb.DeleteFunctionResponse{
		Message: fmt.Sprintf("Function %s deleted successfully", req.Id),
	}, nil
}

func (s *GRPCServer) TestFunction(ctx context.Context, req *pb.TestFunctionRequest) (*pb.TestFunctionResponse, error) {
	success, usedMockData, executionTimeMs, responseData, errorMessage, err := s.businessLogic.TestFunction(req.FunctionId, req.UseMockData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to test function: %v", err)
	}

	response, _ := structpb.NewStruct(responseData)
	return &pb.TestFunctionResponse{
		Success:         success,
		UsedMockData:    usedMockData,
		ExecutionTimeMs: executionTimeMs,
		Response:        response,
		ErrorMessage:    errorMessage,
	}, nil
}

// =============================================================================
// DATABASE MANAGEMENT
// =============================================================================

func (s *GRPCServer) GetDatabaseStats(ctx context.Context, req *pb.GetDatabaseStatsRequest) (*pb.GetDatabaseStatsResponse, error) {
	totalExecutionRuns, totalApiRequests, totalApiResponses, totalFunctionCalls, avgResponseTime, successRate := s.businessLogic.GetDatabaseStats()

	return &pb.GetDatabaseStatsResponse{
		TotalExecutionRuns: totalExecutionRuns,
		TotalApiRequests:   totalApiRequests,
		TotalApiResponses:  totalApiResponses,
		TotalFunctionCalls: totalFunctionCalls,
		AvgResponseTime:    avgResponseTime,
		SuccessRate:        successRate,
	}, nil
}

func (s *GRPCServer) ListDatabaseTables(ctx context.Context, req *pb.ListDatabaseTablesRequest) (*pb.ListDatabaseTablesResponse, error) {
	tables := s.businessLogic.ListDatabaseTables()

	return &pb.ListDatabaseTablesResponse{
		Tables: tables,
	}, nil
}

func (s *GRPCServer) GetTableData(ctx context.Context, req *pb.GetTableDataRequest) (*pb.GetTableDataResponse, error) {
	columns, rows, totalRows, err := s.businessLogic.GetTableData(req.TableName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get table data: %v", err)
	}

	var protoRows []*structpb.ListValue
	for _, row := range rows {
		rowProto, err := structpb.NewList(row)
		if err != nil {
			continue
		}
		protoRows = append(protoRows, rowProto)
	}

	return &pb.GetTableDataResponse{
		TableName: req.TableName,
		Columns:   columns,
		Rows:      protoRows,
		TotalRows: totalRows,
	}, nil
}

// =============================================================================
// HEALTH & SYSTEM
// =============================================================================

func (s *GRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	status, version, database, geminiAPI := s.businessLogic.GetHealthStatus()

	return &pb.HealthResponse{
		Status:    status,
		Version:   version,
		Timestamp: timestamppb.Now(),
		Database:  database,
		GeminiApi: geminiAPI,
	}, nil
}

// =============================================================================
// PROTO CONVERSION HELPERS
// =============================================================================

func (s *GRPCServer) convertUserToProto(user *auth.User) *pb.User {
	protoUser := &pb.User{
		Id:            user.ID,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		IsTemporary:   user.IsTemporary,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
	}

	if user.Email != nil {
		protoUser.Email = *user.Email
	}

	if user.LastLoginAt != nil {
		protoUser.LastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return protoUser
}

func (s *GRPCServer) convertExecutionRunToProto(run *types.ExecutionRun) *pb.ExecutionRun {
	return &pb.ExecutionRun{
		Id:                    run.ID,
		UserId:                "current-user-1", // TODO: Get from actual data
		Name:                  run.Name,
		Description:           run.Description,
		EnableFunctionCalling: run.EnableFunctionCalling,
		Status:                run.Status,
		CreatedAt:             timestamppb.New(run.CreatedAt),
		UpdatedAt:             timestamppb.New(run.UpdatedAt),
	}
}

func (s *GRPCServer) convertConfigurationToProto(config *types.APIConfiguration) *pb.APIConfiguration {
	protoConfig := &pb.APIConfiguration{
		Id:            config.ID,
		VariationName: config.VariationName,
		ModelName:     config.ModelName,
		SystemPrompt:  config.SystemPrompt,
		CreatedAt:     timestamppb.New(config.CreatedAt),
	}

	if config.Temperature != nil {
		protoConfig.Temperature = *config.Temperature
	}
	if config.MaxTokens != nil {
		protoConfig.MaxTokens = *config.MaxTokens
	}
	if config.TopP != nil {
		protoConfig.TopP = *config.TopP
	}
	if config.TopK != nil {
		protoConfig.TopK = *config.TopK
	}

	return protoConfig
}

func (s *GRPCServer) convertProtoConfigurationToInternal(pc *pb.APIConfiguration) *types.APIConfiguration {
	config := &types.APIConfiguration{
		ID:            pc.Id,
		VariationName: pc.VariationName,
		ModelName:     pc.ModelName,
		SystemPrompt:  pc.SystemPrompt,
	}

	if pc.Temperature > 0 {
		config.Temperature = &pc.Temperature
	}
	if pc.MaxTokens > 0 {
		config.MaxTokens = &pc.MaxTokens
	}
	if pc.TopP > 0 {
		config.TopP = &pc.TopP
	}
	if pc.TopK > 0 {
		config.TopK = &pc.TopK
	}

	return config
}

func (s *GRPCServer) convertFunctionToProto(function *types.FunctionDefinition) *pb.FunctionDefinition {
	// Create basic proto function
	protoFunction := &pb.FunctionDefinition{
		Id:          function.ID,
		UserId:      "current-user-1", // TODO: Get from actual user context
		Name:        function.Name,
		DisplayName: function.DisplayName,
		Description: function.Description,
		EndpointUrl: function.EndpointURL,
		HttpMethod:  function.HttpMethod,
		IsActive:    function.IsActive,
		CreatedAt:   timestamppb.New(function.CreatedAt),
		UpdatedAt:   timestamppb.New(function.UpdatedAt),
	}

	// TODO: Convert ParametersSchema and MockResponse from map to structpb.Struct
	if len(function.ParametersSchema) > 0 {
		if schema, err := structpb.NewStruct(function.ParametersSchema); err == nil {
			protoFunction.ParametersSchema = schema
		}
	}

	if len(function.MockResponse) > 0 {
		if response, err := structpb.NewStruct(function.MockResponse); err == nil {
			protoFunction.MockResponse = response
		}
	}

	return protoFunction
}

func (s *GRPCServer) convertProtoFunctionToInternal(pf *pb.FunctionDefinition) *types.FunctionDefinition {
	function := &types.FunctionDefinition{
		ID:          pf.Id,
		Name:        pf.Name,
		DisplayName: pf.DisplayName,
		Description: pf.Description,
		EndpointURL: pf.EndpointUrl,
		HttpMethod:  pf.HttpMethod,
		IsActive:    pf.IsActive,
	}

	// TODO: Convert structpb.Struct to map
	if pf.ParametersSchema != nil {
		function.ParametersSchema = pf.ParametersSchema.AsMap()
	}

	if pf.MockResponse != nil {
		function.MockResponse = pf.MockResponse.AsMap()
	}

	return function
}

func (s *GRPCServer) convertProtoConfigurations(protoConfigs []*pb.APIConfiguration) []types.APIConfiguration {
	var configs []types.APIConfiguration
	for _, pc := range protoConfigs {
		config := *s.convertProtoConfigurationToInternal(pc)
		config.CreatedAt = time.Now()
		configs = append(configs, config)
	}
	return configs
}

func (s *GRPCServer) convertExecutionResultToProto(result *types.ExecutionResult) (*pb.ExecutionResult, error) {
	// Convert execution run
	protoRun := s.convertExecutionRunToProto(&result.ExecutionRun)

	// Convert variation results
	var protoResults []*pb.VariationResult
	for _, vr := range result.Results {
		protoConfig := s.convertConfigurationToProto(&vr.Configuration)

		protoRequest := &pb.APIRequest{
			Id:              vr.Request.ID,
			ExecutionRunId:  vr.Request.ExecutionRunID,
			ConfigurationId: vr.Request.ConfigurationID,
			RequestType:     string(vr.Request.RequestType),
			Prompt:          vr.Request.Prompt,
			Context:         vr.Request.Context,
			FunctionName:    vr.Request.FunctionName,
			CreatedAt:       timestamppb.New(vr.Request.CreatedAt),
		}

		// Convert usage metadata
		usageStruct, _ := structpb.NewStruct(vr.Response.UsageMetadata)

		protoResponse := &pb.APIResponse{
			Id:             vr.Response.ID,
			RequestId:      vr.Response.RequestID,
			ResponseStatus: string(vr.Response.ResponseStatus),
			ResponseText:   vr.Response.ResponseText,
			FinishReason:   vr.Response.FinishReason,
			ErrorMessage:   vr.Response.ErrorMessage,
			ResponseTimeMs: vr.Response.ResponseTimeMs,
			UsageMetadata:  usageStruct,
			CreatedAt:      timestamppb.New(vr.Response.CreatedAt),
		}

		protoResult := &pb.VariationResult{
			Configuration: protoConfig,
			Request:       protoRequest,
			Response:      protoResponse,
			ExecutionTime: vr.ExecutionTime,
		}
		protoResults = append(protoResults, protoResult)
	}

	// Convert comparison result
	var protoComparison *pb.ComparisonResult
	if result.Comparison != nil {
		protoComparison = &pb.ComparisonResult{
			Id:                  result.Comparison.ID,
			ExecutionRunId:      result.Comparison.ExecutionRunID,
			ComparisonType:      result.Comparison.ComparisonType,
			MetricName:          result.Comparison.MetricName,
			BestConfigurationId: result.Comparison.BestConfigurationID,
			AnalysisNotes:       result.Comparison.AnalysisNotes,
			CreatedAt:           timestamppb.New(result.Comparison.CreatedAt),
		}
	}

	return &pb.ExecutionResult{
		ExecutionRun: protoRun,
		Results:      protoResults,
		Comparison:   protoComparison,
		TotalTime:    result.TotalTime,
		SuccessCount: int32(result.SuccessCount),
		ErrorCount:   int32(result.ErrorCount),
	}, nil
}

// =============================================================================
// SERVER STARTUP
// =============================================================================

// runGRPCServer starts the gRPC server
func runGRPCServer() {
	server, err := NewGRPCServer()
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}
	defer server.Close()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGogentServiceServer(grpcServer, server)

	fmt.Printf("🚀 GoGent gRPC Server starting on port %s\n", port)
	fmt.Printf("📡 Health check: use gRPC client to call Health method\n")
	fmt.Printf("🔧 Available gRPC methods:\n")
	fmt.Printf("   - Authentication: Login, Register, CreateTemporaryUser, etc.\n")
	fmt.Printf("   - Execution: Execute, GetExecutionStatus, ListExecutionRuns\n")
	fmt.Printf("   - Configuration: ListConfigurations, CreateConfiguration\n")
	fmt.Printf("   - Functions: ListFunctions, CreateFunction, TestFunction\n")
	fmt.Printf("   - Database: GetDatabaseStats, GetTableData\n")
	fmt.Printf("   - Health: Health\n")
	fmt.Println()

	log.Fatal(grpcServer.Serve(lis))
}
