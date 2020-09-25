// Code generated by oto; DO NOT EDIT.

package api

import (
	"context"
	"log"
	"net/http"

	"github.com/pacedotdev/oto/otohttp"

	datastore "github.com/avian-digital-forensics/auto-processing/pkg/datastore"
)

// NmsService handles the Nuix Management Servers
type NmsService interface {
	Apply(context.Context, NmsApplyRequests) (*NmsApplyResponse, error)
	List(context.Context, NmsListRequest) (*NmsListResponse, error)
	ListLicences(context.Context, NmsListLicencesRequest) (*NmsListLicencesResponse, error)
}

// RunnerService handles all the runners
type RunnerService interface {

	// Apply applies the configuration to the backend
	Apply(context.Context, RunnerApplyRequest) (*RunnerApplyResponse, error)
	// FailedStage sets a stage to Failed
	FailedStage(context.Context, StageRequest) (*StageResponse, error)
	// FinishStage sets a stage to Finished
	FinishStage(context.Context, StageRequest) (*StageResponse, error)
	// Get returns the requested Runner
	Get(context.Context, RunnerGetRequest) (*RunnerGetResponse, error)
	// List returns the runners from the backend
	List(context.Context, RunnerListRequest) (*RunnerListResponse, error)
	// LogDebug logs a debug-message
	LogDebug(context.Context, LogRequest) (*LogResponse, error)
	// LogError logs an error-message
	LogError(context.Context, LogRequest) (*LogResponse, error)
	// LogInfo logs an info-message
	LogInfo(context.Context, LogRequest) (*LogResponse, error)
	// LogItem logs an item
	LogItem(context.Context, LogItemRequest) (*LogResponse, error)
	// StartStage sets a stage to Active
	StartStage(context.Context, StageRequest) (*StageResponse, error)
}

// ServerService handles all the servers
type ServerService interface {
	Apply(context.Context, ServerApplyRequest) (*ServerApplyResponse, error)
	List(context.Context, ServerListRequest) (*ServerListResponse, error)
}

type nmsServiceServer struct {
	server     *otohttp.Server
	nmsService NmsService
}

// Register adds the NmsService to the otohttp.Server.
func RegisterNmsService(server *otohttp.Server, nmsService NmsService) {
	handler := &nmsServiceServer{
		server:     server,
		nmsService: nmsService,
	}
	server.Register("NmsService", "Apply", handler.handleApply)
	server.Register("NmsService", "List", handler.handleList)
	server.Register("NmsService", "ListLicences", handler.handleListLicences)
}

func (s *nmsServiceServer) handleApply(w http.ResponseWriter, r *http.Request) {
	var request NmsApplyRequests
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.nmsService.Apply(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *nmsServiceServer) handleList(w http.ResponseWriter, r *http.Request) {
	var request NmsListRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.nmsService.List(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *nmsServiceServer) handleListLicences(w http.ResponseWriter, r *http.Request) {
	var request NmsListLicencesRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.nmsService.ListLicences(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

type runnerServiceServer struct {
	server        *otohttp.Server
	runnerService RunnerService
}

// Register adds the RunnerService to the otohttp.Server.
func RegisterRunnerService(server *otohttp.Server, runnerService RunnerService) {
	handler := &runnerServiceServer{
		server:        server,
		runnerService: runnerService,
	}
	server.Register("RunnerService", "Apply", handler.handleApply)
	server.Register("RunnerService", "FailedStage", handler.handleFailedStage)
	server.Register("RunnerService", "FinishStage", handler.handleFinishStage)
	server.Register("RunnerService", "Get", handler.handleGet)
	server.Register("RunnerService", "List", handler.handleList)
	server.Register("RunnerService", "LogDebug", handler.handleLogDebug)
	server.Register("RunnerService", "LogError", handler.handleLogError)
	server.Register("RunnerService", "LogInfo", handler.handleLogInfo)
	server.Register("RunnerService", "LogItem", handler.handleLogItem)
	server.Register("RunnerService", "StartStage", handler.handleStartStage)
}

func (s *runnerServiceServer) handleApply(w http.ResponseWriter, r *http.Request) {
	var request RunnerApplyRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.Apply(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleFailedStage(w http.ResponseWriter, r *http.Request) {
	var request StageRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.FailedStage(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleFinishStage(w http.ResponseWriter, r *http.Request) {
	var request StageRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.FinishStage(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleGet(w http.ResponseWriter, r *http.Request) {
	var request RunnerGetRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.Get(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleList(w http.ResponseWriter, r *http.Request) {
	var request RunnerListRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.List(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleLogDebug(w http.ResponseWriter, r *http.Request) {
	var request LogRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.LogDebug(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleLogError(w http.ResponseWriter, r *http.Request) {
	var request LogRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.LogError(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleLogInfo(w http.ResponseWriter, r *http.Request) {
	var request LogRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.LogInfo(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleLogItem(w http.ResponseWriter, r *http.Request) {
	var request LogItemRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.LogItem(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *runnerServiceServer) handleStartStage(w http.ResponseWriter, r *http.Request) {
	var request StageRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.runnerService.StartStage(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

type serverServiceServer struct {
	server        *otohttp.Server
	serverService ServerService
}

// Register adds the ServerService to the otohttp.Server.
func RegisterServerService(server *otohttp.Server, serverService ServerService) {
	handler := &serverServiceServer{
		server:        server,
		serverService: serverService,
	}
	server.Register("ServerService", "Apply", handler.handleApply)
	server.Register("ServerService", "List", handler.handleList)
}

func (s *serverServiceServer) handleApply(w http.ResponseWriter, r *http.Request) {
	var request ServerApplyRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.serverService.Apply(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

func (s *serverServiceServer) handleList(w http.ResponseWriter, r *http.Request) {
	var request ServerListRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.serverService.List(r.Context(), request)
	if err != nil {
		log.Println("TODO: oto service error:", err)
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

type Base struct {
	ID    uint   `json:"id" yaml:"id"`
	CTime int64  `json:"cTime" yaml:"cTime"`
	MTime int64  `json:"mTime" yaml:"mTime"`
	DTime *int64 `json:"dTime" yaml:"dTime"`
}

// Case holds the information for a case
type Case struct {
	datastore.Base
	// Name of the case
	Name string `json:"name" yaml:"name"`
	// Directory of the case
	Directory string `json:"directory" yaml:"directory"`
	// Description of the case
	Description string `json:"description" yaml:"description"`
	// Investigator of the case
	Investigator string `json:"investigator" yaml:"investigator"`
}

// CaseSettings holds information about the cases if Processing-stage is used for a
// Runner
type CaseSettings struct {
	datastore.Base
	// CaseLocation is the parent-folder for all cases
	CaseLocation string `json:"caseLocation" yaml:"caseLocation"`
	// Case holds the information for the single-case
	CaseID uint  `json:"caseID" yaml:"caseID"`
	Case   *Case `json:"case" yaml:"case"`
	// CompoundCase holds the information for the compound-case
	CompoundCaseID uint  `json:"compoundCaseID" yaml:"compoundCaseID"`
	CompoundCase   *Case `json:"compoundCase" yaml:"compoundCase"`
	// ReviewCompound holds the information for the review-compound
	ReviewCompoundID uint  `json:"reviewCompoundID" yaml:"reviewCompoundID"`
	ReviewCompound   *Case `json:"reviewCompound" yaml:"reviewCompound"`
}

// Evidence holds information about a specific evidence
type Evidence struct {
	datastore.Base
	// ProcessID foreign-key for process-table
	ProcessID uint `json:"processID" yaml:"processID"`
	// Name of the evidence
	Name string `json:"name" yaml:"name"`
	// Directory of where the evidence is located
	Directory string `json:"directory" yaml:"directory"`
	// Description of the evidence
	Description string `json:"description" yaml:"description"`
	// Encoding for the evidence (used when processing)
	Encoding string `json:"encoding" yaml:"encoding"`
	// TimeZone for the evidence
	TimeZone string `json:"timeZone" yaml:"timeZone"`
	// Custodian for the evidence
	Custodian string `json:"custodian" yaml:"custodian"`
	// Locale for the evidence (used when processing)
	Locale string `json:"locale" yaml:"locale"`
}

// Exclude excludes items in a Nuix-case based on a search
type Exclude struct {
	datastore.Base
	// StageID foreign-key for stage-table
	StageID uint `json:"stageID" yaml:"stageID"`
	// Search query in the case
	Search string `json:"search" yaml:"search"`
	// Reason to exclude the items from the search
	Reason string `json:"reason" yaml:"reason"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// File holds information about a file
type File struct {
	datastore.Base
	// SearchAndTagID foreign-key for searchandtag-table
	SearchAndTagID uint `json:"searchAndTagID" yaml:"searchAndTagID"`
	// Path for where the file is located at
	Path string `json:"path" yaml:"path"`
}

// Licence holds information about licences in Nuix Management Server
type Licence struct {
	datastore.Base
	// Foreign-key for the NMS-server
	NmsID uint `json:"nmsID" yaml:"nmsID"`
	// Type of licence
	Type string `json:"type" yaml:"type"`
	// Amount of licences for this type
	Amount int64 `json:"amount" yaml:"amount"`
	// Amount of licenses in use for this type
	InUse int64 `json:"inUse" yaml:"inUse"`
}

// LicenceApplyRequest is the input-object for applying NMS-licence
type LicenceApplyRequest struct {
	// Type of licence
	Type string `json:"type" yaml:"type"`
	// Amount of licences for this type
	Amount int64 `json:"amount" yaml:"amount"`
}

// Licences is a holder for Licence
type Licences struct {
	Licence LicenceApplyRequest `json:"licence" yaml:"licence"`
}

type LogItemRequest struct {
	Runner       string `json:"runner" yaml:"runner"`
	Stage        string `json:"stage" yaml:"stage"`
	StageID      int    `json:"stageID" yaml:"stageID"`
	Message      string `json:"message" yaml:"message"`
	Count        int    `json:"count" yaml:"count"`
	MimeType     string `json:"mimeType" yaml:"mimeType"`
	GUID         string `json:"gUID" yaml:"gUID"`
	ProcessStage string `json:"processStage" yaml:"processStage"`
}

type LogRequest struct {
	Runner    string `json:"runner" yaml:"runner"`
	Stage     string `json:"stage" yaml:"stage"`
	StageID   int    `json:"stageID" yaml:"stageID"`
	Message   string `json:"message" yaml:"message"`
	Exception string `json:"exception" yaml:"exception"`
}

type LogResponse struct {
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Nms is the main struct for the Nuix Management Servers
type Nms struct {
	datastore.Base
	// Address of the nms-server
	Address string `json:"address" yaml:"address"`
	// Port for the nms-server
	Port int64 `json:"port" yaml:"port"`
	// Username for the nms-server
	Username string `json:"username" yaml:"username"`
	// Password for the nms-server
	Password string `json:"password" yaml:"password"`
	// amount of workers licensed to the server
	Workers int64 `json:"workers" yaml:"workers"`
	// Amount of workers in use
	InUse int64 `json:"inUse" yaml:"inUse"`
	// Licences available at the server
	Licences []Licence `json:"licences" yaml:"licences"`
}

// NmsApplyRequest is the input-object for Apply in the NMS-service
type NmsApplyRequest struct {
	// Address of the nms-server
	Address string `json:"address" yaml:"address"`
	// Port for the nms-server
	Port int64 `json:"port" yaml:"port"`
	// Username for the nms-server
	Username string `json:"username" yaml:"username"`
	// Password for the nms-server
	Password string `json:"password" yaml:"password"`
	// amount of workers licensed to the server
	Workers int64 `json:"workers" yaml:"workers"`
	// Licences available at the server
	Licences []Licences `json:"licences" yaml:"licences"`
}

type NmsApplyRequests struct {
	Nms []NmsApplyRequest `json:"nms" yaml:"nms"`
}

// NmsApplyResponse is the output-object for Apply in the NMS-service
type NmsApplyResponse struct {
	Nms []Nms `json:"nms" yaml:"nms"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// NmsListLicencesRequest is the input-object for listing licences for a specific
// NMS
type NmsListLicencesRequest struct {
	// ID for the nms-server to list the licences for
	NmsID uint `json:"nmsID" yaml:"nmsID"`
}

// NmsListLicencesResponse is the output-object for listing licences for a specific
// NMS
type NmsListLicencesResponse struct {
	Licences []Licence `json:"licences" yaml:"licences"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// NmsListRequest is the input-object for List in the NMS-service
type NmsListRequest struct {
}

// NmsListResponse is the output-object for List in the NMS-service
type NmsListResponse struct {
	Nms []Nms `json:"nms" yaml:"nms"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// NuixSwitch is a command argument for nuix-console
type NuixSwitch struct {
	datastore.Base
	RunnerID uint   `json:"runnerID" yaml:"runnerID"`
	Value    string `json:"value" yaml:"value"`
}

// Ocr performs OCR based on a search in a Nuix-case
type Ocr struct {
	datastore.Base
	// StageID foreign-key for stage-table
	StageID uint `json:"stageID" yaml:"stageID"`
	// Profile for the ocr-processor
	Profile     string `json:"profile" yaml:"profile"`
	ProfilePath string `json:"profilePath" yaml:"profilePath"`
	// Search query in the case
	Search string `json:"search" yaml:"search"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// Populate populates data based on a search in a Nuix-case
type Populate struct {
	datastore.Base
	// StageID foreign-key for stage-table
	StageID uint `json:"stageID" yaml:"stageID"`
	// Search query in the case
	Search string `json:"search" yaml:"search"`
	// Types for the items to populate
	Types []*Type `json:"types" yaml:"types"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// Process -stage processes data into a Nuix-case
type Process struct {
	datastore.Base
	// Foreign-key for stage
	StageID uint `json:"stageID" yaml:"stageID"`
	// Profile for the processor
	Profile     string `json:"profile" yaml:"profile"`
	ProfilePath string `json:"profilePath" yaml:"profilePath"`
	// EvidenceStore to process to the nuix-case
	EvidenceStore []*Evidence `json:"evidenceStore" yaml:"evidenceStore"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// Reload reloads items in a Nuix-case based on a search
type Reload struct {
	datastore.Base
	// StageID foreign-key for stage-table
	StageID uint `json:"stageID" yaml:"stageID"`
	// Profile for the reload-processing
	Profile     string `json:"profile" yaml:"profile"`
	ProfilePath string `json:"profilePath" yaml:"profilePath"`
	// Search query in the case
	Search string `json:"search" yaml:"search"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// Runner holds the information for a specific runner
type Runner struct {
	datastore.Base
	// Name for the runner
	Name string `json:"name" yaml:"name"`
	// Server to use for the runner
	Hostname string `json:"hostname" yaml:"hostname"`
	// Nms to use for the runner
	Nms string `json:"nms" yaml:"nms"`
	// Licence to use for the runner
	Licence string `json:"licence" yaml:"licence"`
	// Xmx to use for the runner
	Xmx string `json:"xmx" yaml:"xmx"`
	// Amount of workers to use for the runner
	Workers int64 `json:"workers" yaml:"workers"`
	// Active - if the runner is active or not
	Active bool `json:"active" yaml:"active"`
	// Finished - if the runner has finished or not
	Finished bool `json:"finished" yaml:"finished"`
	// CaseSettings for the cases to use
	CaseSettingsID uint          `json:"caseSettingsID" yaml:"caseSettingsID"`
	CaseSettings   *CaseSettings `json:"caseSettings" yaml:"caseSettings"`
	// Stages for the runner
	Stages []*Stage `json:"stages" yaml:"stages"`
	// Switches to use for nuix-console
	Switches []*NuixSwitch `json:"switches" yaml:"switches"`
}

// RunnerApplyRequest is the input-object for applying a runner-configuration to
// the Runner-service
type RunnerApplyRequest struct {
	// Name for the runner
	Name string `json:"name" yaml:"name"`
	// Server to use for the runner
	Hostname string `json:"hostname" yaml:"hostname"`
	// Nms to use for the runner
	Nms string `json:"nms" yaml:"nms"`
	// Licence to use for the runner
	Licence string `json:"licence" yaml:"licence"`
	// Xmx to use for the runner
	Xmx string `json:"xmx" yaml:"xmx"`
	// Amount of workers to use for the runner
	Workers int64 `json:"workers" yaml:"workers"`
	// CaseSettings is the settings for the cases that should be processed if
	// Process-stage is used
	CaseSettings *CaseSettings `json:"caseSettings" yaml:"caseSettings"`
	// Stages for the runner
	Stages []*Stage `json:"stages" yaml:"stages"`
	// Switches to use for nuix-console
	Switches []string `json:"switches" yaml:"switches"`
}

// RunnerApplyResponse is the output-object for applying a runner-configuration to
// the backend
type RunnerApplyResponse struct {
	Runner Runner `json:"runner" yaml:"runner"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// RunnerGetRequest is the input-object for requesting a runner by name
type RunnerGetRequest struct {
	Name string `json:"name" yaml:"name"`
}

// RunnerGetResponse is the output-object for requesting a runner by name
type RunnerGetResponse struct {
	Runner Runner `json:"runner" yaml:"runner"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// RunnerListRequest is the input-object for listing the runners from the backend
type RunnerListRequest struct {
}

// RunnerListResponse is the input-object for listing the runners from the backend
type RunnerListResponse struct {
	Runners []Runner `json:"runners" yaml:"runners"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

type StageRequest struct {
	Runner  string `json:"runner" yaml:"runner"`
	StageID uint   `json:"stageID" yaml:"stageID"`
}

// Stage holds different types of stages for a Runner
type Stage struct {
	datastore.Base
	// Foreign-key for runners
	RunnerID uint `json:"runnerID" yaml:"runnerID"`
	// Process-stage processes data into a Nuix-case
	Process *Process `json:"process" yaml:"process"`
	// SearchAndTag searches and tags data in a Nuix-case
	SearchAndTag *SearchAndTag `json:"searchAndTag" yaml:"searchAndTag"`
	// Populate populates data based on a search in a Nuix-case
	Populate *Populate `json:"populate" yaml:"populate"`
	// Ocr performs OCR based on a search in a Nuix-case
	Ocr *Ocr `json:"ocr" yaml:"ocr"`
	// Exclude excludes items in a Nuix-case based on a search
	Exclude *Exclude `json:"exclude" yaml:"exclude"`
	// Reload reloads items in a Nuix-case based on a search
	Reload *Reload `json:"reload" yaml:"reload"`
}

type StageResponse struct {
	Stage Stage `json:"stage" yaml:"stage"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// SearchAndTag searches and tags data in a Nuix-case
type SearchAndTag struct {
	datastore.Base
	// StageID foreign-key for stage-table
	StageID uint `json:"stageID" yaml:"stageID"`
	// Search query in the case
	Search string `json:"search" yaml:"search"`
	// Tag for the items from the search
	Tag string `json:"tag" yaml:"tag"`
	// Files for the search-and-tag
	Files []*File `json:"files" yaml:"files"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}

// Server is the main-struct for the servers
type Server struct {
	datastore.Base
	// Hostname of the server
	Hostname string `json:"hostname" yaml:"hostname"`
	// Port for the server
	Port int64 `json:"port" yaml:"port"`
	// OperatingSystem the server is running
	OperatingSystem string `json:"operatingSystem" yaml:"operatingSystem"`
	// Username for connection to the server
	Username string `json:"username" yaml:"username"`
	// Password for connection to the server
	Password string `json:"password" yaml:"password"`
	// NuixPath to know where to run Nuix
	NuixPath string `json:"nuixPath" yaml:"nuixPath"`
	// Active - if the server has an active job
	Active bool `json:"active" yaml:"active"`
}

// ServerApplyRequest is the input-object for Apply in the server-service
type ServerApplyRequest struct {
	Hostname        string `json:"hostname" yaml:"hostname"`
	Port            int64  `json:"port" yaml:"port"`
	OperatingSystem string `json:"operatingSystem" yaml:"operatingSystem"`
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	NuixPath        string `json:"nuixPath" yaml:"nuixPath"`
}

// ServerApplyResponse is the output-object for Apply in the server-service
type ServerApplyResponse struct {
	Server Server `json:"server" yaml:"server"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ServerListRequest is the input-object for List in the server-service
type ServerListRequest struct {
}

// ServerListResponse is the output-object for List in the server-service
type ServerListResponse struct {
	Servers []Server `json:"servers" yaml:"servers"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Type holds information for a type
type Type struct {
	datastore.Base
	// PopulateID foreign-key for populate-table
	PopulateID uint `json:"populateID" yaml:"populateID"`
	// Type-name
	Type string `json:"type" yaml:"type"`
	// Status for the stage
	Status int64 `json:"status" yaml:"status"`
}
