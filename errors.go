package gz

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
)

///////////////////////
// Database error codes
///////////////////////

// ErrorNoDatabase is triggered when the database connection is unavailable
const ErrorNoDatabase = 1000

// ErrorDbDelete is triggered when the database was unable to delete a resource
const ErrorDbDelete = 1001

// ErrorDbSave is triggered when the database was unable to save a resource
const ErrorDbSave = 1002

// ErrorIDNotFound is triggered when a resource with the specified id is not
// found in the database
const ErrorIDNotFound = 1003

// ErrorNameNotFound is triggered when a resource with the specified name is not
// found in the database
const ErrorNameNotFound = 1004

// ErrorFileNotFound is triggered when a model's file with the specified name is not
// found
const ErrorFileNotFound = 1005

///////////////////
// JSON error codes
///////////////////

// ErrorMarshalJSON is triggered if there is an error marshalling data into JSON
const ErrorMarshalJSON = 2000

// ErrorUnmarshalJSON is triggered if there is an error unmarshalling JSON
const ErrorUnmarshalJSON = 2001

///////////////////
// Protobuf error codes
///////////////////

// ErrorMarshalProto is triggered if there is an error marshalling data into protobuf
const ErrorMarshalProto = 2500

//////////////////////
// Request error codes
//////////////////////

// ErrorIDNotInRequest is triggered when a id is not found in the request
const ErrorIDNotInRequest = 3000

// ErrorIDWrongFormat is triggered when an id is not in a valid format
const ErrorIDWrongFormat = 3001

// ErrorNameWrongFormat is triggered when a name is not in a valid format
const ErrorNameWrongFormat = 3002

// ErrorPayloadEmpty is triggered when payload is expected but is not found in
// the request
const ErrorPayloadEmpty = 3003

// ErrorForm is triggered when an expected field is missing in a multipart
// form request.
const ErrorForm = 3004

// ErrorUnexpectedID is triggered when the id of a file attached in a
// request is not expected. E.g.: When the attached world file does not end in
// ".world" during a world creation request.
const ErrorUnexpectedID = 3005

// ErrorUnknownSuffix is triggered when a suffix for content negotiation is not
// recognized.
const ErrorUnknownSuffix = 3006

// ErrorUserNotInRequest is triggered when the user/team is not found in
// the request.
const ErrorUserNotInRequest = 3007

// ErrorUserUnknown is triggered when the user/team does not exist on the
// server
const ErrorUserUnknown = 3008

// ErrorMissingField is triggered when the JSON contained in a request does
// not contain one or more required fields
const ErrorMissingField = 3009

// ErrorOwnerNotInRequest is triggered when an owner is not found in the request
const ErrorOwnerNotInRequest = 3010

// ErrorModelNotInRequest is triggered when a model is not found in the request
const ErrorModelNotInRequest = 3011

// ErrorFormMissingFiles is triggered when the expected "file" field is missing
// in the multipart form request.
const ErrorFormMissingFiles = 3012

// ErrorFormInvalidValue is triggered when a given form field has an invalid
// value.
const ErrorFormInvalidValue = 3013

// ErrorFormDuplicateFile is triggered when the POSTed model carries duplicate
// file entries.
const ErrorFormDuplicateFile = 3014

// ErrorFormDuplicateModelName is triggered when the POSTed model carries duplicate
// model name.
const ErrorFormDuplicateModelName = 3015

// ErrorInvalidPaginationRequest is triggered when the requested pagination is invalid.
// eg. invalid page or per_page argument values.
const ErrorInvalidPaginationRequest = 3016

// ErrorPaginationPageNotFound is triggered when the requested page is empty / not found.
const ErrorPaginationPageNotFound = 3017

// ErrorFormDuplicateWorldName is triggered when the POSTed world carries duplicate
// name.
const ErrorFormDuplicateWorldName = 3018

// ErrorWorldNotInRequest is triggered when a world is not found in the request
const ErrorWorldNotInRequest = 3019

// ErrorCollectionNotInRequest is triggered when a collection name is not found in the request
const ErrorCollectionNotInRequest = 3020

////////////////////////////
// Authorization error codes
////////////////////////////

// ErrorAuthNoUser is triggered when there's no user in the database with the
// claimed user ID.
const ErrorAuthNoUser = 4000

// ErrorAuthJWTInvalid is triggered when is not possible to get a user ID
// from the JWT in the request
const ErrorAuthJWTInvalid = 4001

// ErrorUnauthorized is triggered when a user is not authorized to perform a
// given action.
const ErrorUnauthorized = 4002

////////////////////
// Simulation error codes
////////////////////

// ErrorLaunchGazebo is triggered when we cannot launch gazebo in the context of a simulation
const ErrorLaunchGazebo = 5000

// ErrorShutdownGazebo is triggered when there is an error during the process of
// shutting down gazebo.
const ErrorShutdownGazebo = 5001

// ErrorSimGroupNotFound is triggered when a simulation Group ID is not found.
const ErrorSimGroupNotFound = 5002

// ErrorK8Create is triggered when there is an error creating a kubernetes resource
const ErrorK8Create = 5003

// ErrorK8Delete is triggered when there is an error deleting a kubernetes resource
const ErrorK8Delete = 5004

// ErrorLaunchingCloudInstance is triggered when there is an error launching a cloud
// instance (eg. aws ec2)
const ErrorLaunchingCloudInstance = 5005

// ErrorStoppingCloudInstance is triggered when there is an error stopping or
// terminating a cloud instance (eg. aws ec2)
const ErrorStoppingCloudInstance = 5006

// ErrorInvalidSimulationStatus is triggered when the simulation is not in a status
// suitable for the requested operation. This error usually has a status extra argument.
const ErrorInvalidSimulationStatus = 5007

// ErrorLaunchingCloudInstanceNotEnoughResources is triggered when there are not enough
// cloud instances to launch a simulation
const ErrorLaunchingCloudInstanceNotEnoughResources = 5008

////////////////////
// Queue error codes
////////////////////

// ErrorQueueEmpty is triggered when a queue is empty
const ErrorQueueEmpty = 6000

// ErrorQueueIndexOutOfBounds is triggered when there is an attempt to access an index that does not exist.
const ErrorQueueIndexOutOfBounds = 6001

// ErrorQueueInternalChannelClosed is triggered when the queue's internal channel is closed.
const ErrorQueueInternalChannelClosed = 6002

// ErrorQueueSwapIndexesMatch is triggered when there is an attempt to swap the same element.
const ErrorQueueSwapIndexesMatch = 6003

// ErrorQueueMoveIndexFrontPosition is triggered when there is an attempt to move the front element to the front position
const ErrorQueueMoveIndexFrontPosition = 6004

// ErrorQueueMoveIndexBackPosition is triggered when there is an attempt to move the back element to the back position
const ErrorQueueMoveIndexBackPosition = 6005

// ErrorQueueTooManyListeners is triggered when there are too many listeners waiting for the next element from the queue.
const ErrorQueueTooManyListeners = 6006

////////////////////
// Other error codes
////////////////////

// ErrorZipNotAvailable is triggered when the server does not have a zip file
// for the requested resource
const ErrorZipNotAvailable = 100000

// ErrorResourceExists is triggered when the server cannot create a new resource
// because the requested id already exists. E.g.: When the creation of a new
// model is requested but the server already has a model with the same id.
const ErrorResourceExists = 100001

// ErrorCreatingDir is triggered when the server was unable to create a new
// directory for a resource (no space on device or a temporary server problem).
const ErrorCreatingDir = 100002

// ErrorCreatingRepo is triggered when the server was unable to create a new
// repository for a resource (no space on device or a temporary server problem).
const ErrorCreatingRepo = 100003

// ErrorCreatingFile is triggered when the server was unable to create a new
// file for a resource (no space on device or a temporary server problem).
const ErrorCreatingFile = 100004

// ErrorUnzipping is triggered when the server was unable to unzip a zipped file
const ErrorUnzipping = 100005

// ErrorNonExistentResource is triggered when the server was unable to find a
// resource.
const ErrorNonExistentResource = 100006

// ErrorRepo is triggered when the server was unable to handle repo command.
const ErrorRepo = 100007

// ErrorRemovingDir is triggered when the server was unable to remove a
// directory.
const ErrorRemovingDir = 100008

// ErrorFileTree is triggered when there was a problem accessing the model's
// files.
const ErrorFileTree = 100009

// ErrorVersionNotFound is triggered when the requested version of a resource is
// not available
const ErrorVersionNotFound = 100010

// ErrorCastingID is triggered when casting an id fails.
const ErrorCastingID = 100011

// ErrorScheduler is triggered initializing a scheduler fails.
const ErrorScheduler = 100012

// ErrorUnexpected is used to represent unexpected or still uncategorized errors.
const ErrorUnexpected = 150000

// ErrMsg is serialized as JSON, and returned if the request does not succeed
// TODO: consider making ErrMsg an 'error'
type ErrMsg struct {
	// Internal error code.
	ErrCode int `json:"errcode"`
	// HTTP status code.
	StatusCode int `json:"-"`
	// Error message.
	Msg string `json:"msg"`
	// Extra information/arguments associated to Error message.
	Extra []string `json:"extra"`
	// The root cause error
	BaseError error `json:"-"`
	// Generated ID for easy tracking in server logs
	ErrID string `json:"errid"`
	// Associated request Route, if applicable
	Route string `json:"route"`
	// Associated request User-Agent, if applicable
	UserAgent string `json:"user-agent"`
	// Associated request remote address, if applicable
	RemoteAddress string `json:"remote-address"`
}

// LogString creates a verbose error string
func (e *ErrMsg) LogString() string {
	return fmt.Sprintf("[ErrID:%s][ErrCode:%d] %s. Extra: %v. [Route:%s]", e.ErrID, e.ErrCode, e.Msg, e.Extra, e.Route)
}

// NewErrorMessage is a convenience function that receives an error code
// and returns a pointer to an ErrMsg.
func NewErrorMessage(err int64) *ErrMsg {
	em := ErrorMessage(err)
	return &em
}

// WithStack wraps a given error with a stack trace if needed.
func WithStack(base error) error {
	// wrap the error with a stack trace, if needed.
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	if _, ok := base.(stackTracer); !ok {
		base = errors.WithStack(base)
	}
	return base
}

// NewErrorMessageWithBase receives an error code and a root error
// and returns a pointer to an ErrMsg.
func NewErrorMessageWithBase(err int64, base error) *ErrMsg {
	em := NewErrorMessage(err)
	em.BaseError = WithStack(base)
	return em
}

// NewErrorMessageWithArgs receives an error code, a root error, and a slice
// of extra arguments, and returns a pointer to an ErrMsg.
func NewErrorMessageWithArgs(err int64, base error, extra []string) *ErrMsg {
	em := NewErrorMessageWithBase(err, base)
	em.Extra = extra
	return em
}

// ErrorMessageOK creates an ErrMsg initialized with OK (default) values.
func ErrorMessageOK() ErrMsg {
	return ErrMsg{ErrCode: 0, StatusCode: http.StatusOK, Msg: ""}
}

// ErrorMessage receives an error code and generate an error message response
func ErrorMessage(err int64) ErrMsg {

	em := ErrorMessageOK()

	em.ErrID = uuid.NewV4().String()

	switch err {
	case ErrorNoDatabase:
		em.Msg = "Unable to connect to the database"
		em.ErrCode = ErrorNoDatabase
		em.StatusCode = http.StatusServiceUnavailable
	case ErrorDbDelete:
		em.Msg = "Unable to remove resource from the database"
		em.ErrCode = ErrorDbDelete
		em.StatusCode = http.StatusInternalServerError
	case ErrorDbSave:
		em.Msg = "Unable to save resource into the database"
		em.ErrCode = ErrorDbSave
		em.StatusCode = http.StatusInternalServerError
	case ErrorIDNotFound:
		em.Msg = "Requested id not found on server"
		em.ErrCode = ErrorIDNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorNameNotFound:
		em.Msg = "Requested name not found on server"
		em.ErrCode = ErrorNameNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorFileNotFound:
		em.Msg = "Requested file not found on server"
		em.ErrCode = ErrorFileNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorMarshalJSON:
		em.Msg = "Unable to marshal the response into a JSON"
		em.ErrCode = ErrorMarshalJSON
		em.StatusCode = http.StatusInternalServerError
	case ErrorUnmarshalJSON:
		em.Msg = "Unable to decode JSON payload included in the request"
		em.ErrCode = ErrorUnmarshalJSON
		em.StatusCode = http.StatusBadRequest
	case ErrorMarshalProto:
		em.Msg = "Unable to marshal the response into a protobuf"
		em.ErrCode = ErrorMarshalProto
		em.StatusCode = http.StatusInternalServerError
	case ErrorIDNotInRequest:
		em.Msg = "ID not present in request"
		em.ErrCode = ErrorIDNotInRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorOwnerNotInRequest:
		em.Msg = "Owner name not present in request"
		em.ErrCode = ErrorOwnerNotInRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorModelNotInRequest:
		em.Msg = "Model name not present in request"
		em.ErrCode = ErrorModelNotInRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorWorldNotInRequest:
		em.Msg = "World name not present in request"
		em.ErrCode = ErrorWorldNotInRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorIDWrongFormat:
		em.Msg = "ID in request is in an invalid format"
		em.ErrCode = ErrorIDWrongFormat
		em.StatusCode = http.StatusBadRequest
	case ErrorNameWrongFormat:
		em.Msg = "Name in request is in an invalid format"
		em.ErrCode = ErrorNameWrongFormat
		em.StatusCode = http.StatusBadRequest
	case ErrorPayloadEmpty:
		em.Msg = "Payload empty in the request"
		em.ErrCode = ErrorPayloadEmpty
		em.StatusCode = http.StatusBadRequest
	case ErrorForm:
		em.Msg = "Missing field in the multipart form"
		em.ErrCode = ErrorForm
		em.StatusCode = http.StatusBadRequest
	case ErrorFormMissingFiles:
		em.Msg = "Missing file field, or empty list of files, in the multipart form"
		em.ErrCode = ErrorFormMissingFiles
		em.StatusCode = http.StatusBadRequest
	case ErrorFormDuplicateFile:
		em.Msg = "Duplicate file in multipart form"
		em.ErrCode = ErrorFormDuplicateFile
		em.StatusCode = http.StatusBadRequest
	case ErrorFormDuplicateModelName:
		em.Msg = "Duplicate model name"
		em.ErrCode = ErrorFormDuplicateModelName
		em.StatusCode = http.StatusBadRequest
	case ErrorFormDuplicateWorldName:
		em.Msg = "Duplicate world name"
		em.ErrCode = ErrorFormDuplicateWorldName
		em.StatusCode = http.StatusBadRequest
	case ErrorInvalidPaginationRequest:
		em.Msg = "Invalid pagination request"
		em.ErrCode = ErrorInvalidPaginationRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorPaginationPageNotFound:
		em.Msg = "Page not found"
		em.ErrCode = ErrorPaginationPageNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorFormInvalidValue:
		em.Msg = "Invalid value in field."
		em.ErrCode = ErrorFormInvalidValue
		em.StatusCode = http.StatusBadRequest
	case ErrorUnexpectedID:
		em.Msg = "Unexpected id included in your request"
		em.ErrCode = ErrorUnexpectedID
		em.StatusCode = http.StatusBadRequest
	case ErrorUnknownSuffix:
		em.Msg = "Unknown suffix requested"
		em.ErrCode = ErrorUnknownSuffix
		em.StatusCode = http.StatusBadRequest
	case ErrorUserNotInRequest:
		em.Msg = "User or team not present in the request"
		em.ErrCode = ErrorUserNotInRequest
		em.StatusCode = http.StatusBadRequest
	case ErrorUserUnknown:
		em.Msg = "Provided user or team does not exist on the server"
		em.ErrCode = ErrorUserUnknown
		em.StatusCode = http.StatusBadRequest
	case ErrorMissingField:
		em.Msg = "One or more required fields are missing"
		em.ErrCode = ErrorMissingField
		em.StatusCode = http.StatusBadRequest
	case ErrorAuthNoUser:
		em.Msg = "No user in server with the claimed identity"
		em.ErrCode = ErrorAuthNoUser
		em.StatusCode = http.StatusForbidden
	case ErrorAuthJWTInvalid:
		em.Msg = "Unable to process user ID from the JWT included in request"
		em.ErrCode = ErrorAuthJWTInvalid
		em.StatusCode = http.StatusForbidden
	case ErrorUnauthorized:
		em.Msg = "Unauthorized request"
		em.ErrCode = ErrorUnauthorized
		em.StatusCode = http.StatusUnauthorized
	case ErrorQueueEmpty:
		em.Msg = "Queue is empty"
		em.ErrCode = ErrorQueueEmpty
		em.StatusCode = http.StatusBadRequest
	case ErrorQueueIndexOutOfBounds:
		em.Msg = "Queue index is out of bounds"
		em.ErrCode = ErrorQueueIndexOutOfBounds
		em.StatusCode = http.StatusBadRequest
	case ErrorQueueInternalChannelClosed:
		em.Msg = "Queue's internal channel is closed"
		em.ErrCode = ErrorQueueInternalChannelClosed
		em.StatusCode = http.StatusLocked
	case ErrorQueueSwapIndexesMatch:
		em.Msg = "Cannot swap the same element in the queue"
		em.ErrCode = ErrorQueueSwapIndexesMatch
		em.StatusCode = http.StatusBadRequest
	case ErrorQueueMoveIndexFrontPosition:
		em.Msg = "Cannot move the first element to the front"
		em.ErrCode = ErrorQueueMoveIndexFrontPosition
		em.StatusCode = http.StatusBadRequest
	case ErrorQueueMoveIndexBackPosition:
		em.Msg = "Cannot move the last element to the back"
		em.ErrCode = ErrorQueueMoveIndexBackPosition
		em.StatusCode = http.StatusBadRequest
	case ErrorQueueTooManyListeners:
		em.Msg = "Too many dequeue listeners"
		em.ErrCode = ErrorQueueTooManyListeners
		em.StatusCode = http.StatusServiceUnavailable
	case ErrorZipNotAvailable:
		em.Msg = "Zip file not available for this resource"
		em.ErrCode = ErrorZipNotAvailable
		em.StatusCode = http.StatusServiceUnavailable
	case ErrorResourceExists:
		em.Msg = "A resource with the same id already exists"
		em.ErrCode = ErrorResourceExists
		em.StatusCode = http.StatusConflict
	case ErrorCreatingDir:
		em.Msg = "Unable to create a new directory for the resource"
		em.ErrCode = ErrorCreatingDir
		em.StatusCode = http.StatusInternalServerError
	case ErrorCreatingRepo:
		em.Msg = "Unable to create a new repository for the resource"
		em.ErrCode = ErrorCreatingRepo
		em.StatusCode = http.StatusInternalServerError
	case ErrorCreatingFile:
		em.Msg = "Unable to create a new file for the resource"
		em.ErrCode = ErrorCreatingFile
		em.StatusCode = http.StatusInternalServerError
	case ErrorUnzipping:
		em.Msg = "Unable to unzip a file"
		em.ErrCode = ErrorUnzipping
		em.StatusCode = http.StatusBadRequest
	case ErrorNonExistentResource:
		em.Msg = "Unable to find the requested resource"
		em.ErrCode = ErrorNonExistentResource
		em.StatusCode = http.StatusNotFound
	case ErrorRepo:
		em.Msg = "Unable to process repository command"
		em.ErrCode = ErrorRepo
		em.StatusCode = http.StatusServiceUnavailable
	case ErrorRemovingDir:
		em.Msg = "Unable to remove a resource directory"
		em.ErrCode = ErrorRemovingDir
		em.StatusCode = http.StatusInternalServerError
	case ErrorFileTree:
		em.Msg = "Unable to get files from model"
		em.ErrCode = ErrorFileTree
		em.StatusCode = http.StatusInternalServerError
	case ErrorVersionNotFound:
		em.Msg = "Requested version not found on server"
		em.ErrCode = ErrorVersionNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorLaunchGazebo:
		em.Msg = "Could not launch gazebo"
		em.ErrCode = ErrorLaunchGazebo
		em.StatusCode = http.StatusInternalServerError
	case ErrorShutdownGazebo:
		em.Msg = "Could not shutdown gazebo"
		em.ErrCode = ErrorShutdownGazebo
		em.StatusCode = http.StatusInternalServerError
	case ErrorSimGroupNotFound:
		em.Msg = "Simulation GroupID not found"
		em.ErrCode = ErrorSimGroupNotFound
		em.StatusCode = http.StatusNotFound
	case ErrorK8Create:
		em.Msg = "Error creating kubernetes resource"
		em.ErrCode = ErrorK8Create
		em.StatusCode = http.StatusInternalServerError
	case ErrorK8Delete:
		em.Msg = "Error deleting kubernetes resource"
		em.ErrCode = ErrorK8Delete
		em.StatusCode = http.StatusInternalServerError
	case ErrorLaunchingCloudInstance:
		em.Msg = "Error launching ec2 instance"
		em.ErrCode = ErrorLaunchingCloudInstance
		em.StatusCode = http.StatusInternalServerError
	case ErrorStoppingCloudInstance:
		em.Msg = "Error stopping ec2 instance"
		em.ErrCode = ErrorStoppingCloudInstance
		em.StatusCode = http.StatusInternalServerError
	case ErrorInvalidSimulationStatus:
		em.Msg = "Invalid simulation status"
		em.ErrCode = ErrorInvalidSimulationStatus
		em.StatusCode = http.StatusBadRequest
	case ErrorLaunchingCloudInstanceNotEnoughResources:
		em.Msg = "Not enough ec2 instances available to launch simulation"
		em.ErrCode = ErrorLaunchingCloudInstanceNotEnoughResources
		em.StatusCode = http.StatusInternalServerError
	case ErrorCastingID:
		em.Msg = "Could not process the given ID"
		em.ErrCode = ErrorCastingID
		em.StatusCode = http.StatusInternalServerError
	case ErrorScheduler:
		em.Msg = "Could not initialize a scheduler"
		em.ErrCode = ErrorScheduler
		em.StatusCode = http.StatusInternalServerError
	case ErrorUnexpected:
		em.Msg = "Unexpected error"
		em.ErrCode = ErrorUnexpected
		em.StatusCode = http.StatusInternalServerError
	}

	em.BaseError = errors.New(em.Msg)
	return em
}
