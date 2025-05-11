package response

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SUCCESS = "success"
)

type (
	ErrorResponse struct {
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}

	ValidatorFieldError struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
)

func OKResponse() (int, any) {
	return http.StatusOK, map[string]any{
		"message": "SUCCESS",
	}
}

func BadRequestMsg(msg any) (int, any) {
	return http.StatusBadRequest, map[string]any{
		"error":   http.StatusText(http.StatusBadRequest),
		"message": msg,
	}
}

func NotFoundMsg(msg any) (int, any) {
	return http.StatusNotFound, map[string]any{
		"error":   http.StatusText(http.StatusNotFound),
		"message": msg,
	}
}

func Forbidden() (int, any) {
	return http.StatusForbidden, map[string]any{
		"error":   "Do not have permission for the request.",
		"message": http.StatusText(http.StatusForbidden),
	}
}

func Unauthorized() (int, any) {
	return http.StatusUnauthorized, map[string]any{
		"error":   http.StatusText(http.StatusUnauthorized),
		"message": http.StatusText(http.StatusUnauthorized),
	}
}

func ServiceUnavailableMsg(msg any) (int, any) {
	return http.StatusServiceUnavailable, map[string]any{
		"error":   http.StatusText(http.StatusServiceUnavailable),
		"message": msg,
	}
}

func Created(data any) (int, any) {
	result := map[string]any{
		"message": "SUCCESS",
		"data":    data,
	}

	return http.StatusCreated, result
}

func Pagination(data, total, limit, offset any) (int, any) {
	return http.StatusOK, map[string]any{
		"data":   data,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}
}

func OK(data any) (int, any) {
	return http.StatusOK, data
}

const (
	ERR_TOKEN_IS_EMPTY            = "token is empty"
	ERR_TOKEN_IS_INVALID          = "token is invalid"
	ERR_TOKEN_IS_EXPIRED          = "token is expired"
	ERR_EMPTY_CONN                = "empty connection"
	ERR_DATA_NOT_FOUND            = "data not found"
	ERR_DATA_INVALID              = "data is invalid"
	ERR_EXAMPLE_NOT_FOUND         = "example not found"
	ERR_EXAMPLE_INVALID           = "example is invalid"
	ERR_MASTERDATA_NOT_FOUND      = "masterdata not found"
	ERR_MASTERDATA_INVALID        = "masterdata is invalid"
	ERR_INSERT_FAILED             = "insert failed"
	ERR_GET_FAILED                = "get failed"
	ERR_PUT_FAILED                = "put failed"
	ERR_PATCH_FAILED              = "patch failed"
	ERR_DELETE_FAILED             = "delete failed"
	ERR_VALIDATION_FAILED         = "validation failed"
	ERR_INVALID_USERNAME_PASSWORD = "invalid username or password"

	ERR_REQUEST_INVALID         = "request invalid"
	ERR_REQUEST_IS_EXISTED      = "request is existed"
	ERR_REQUEST_NOTFOUND        = "request not found"
	ERR_REQUEST_IS_EXPIRED      = "request is expired"
	ERR_REQUEST_CODE_IS_INVALID = "request code is invalid"

	ERR_PERMISSION_DENIED = "permission denied"
)

var MAP_ERR_RESPONSE = map[string]struct {
	GRPC_Code codes.Code
	Message   string
}{
	ERR_TOKEN_IS_EMPTY: {
		GRPC_Code: codes.Unauthenticated,
		Message:   ERR_TOKEN_IS_EMPTY,
	},
	ERR_TOKEN_IS_INVALID: {
		GRPC_Code: codes.Unauthenticated,
		Message:   ERR_TOKEN_IS_INVALID,
	},
	ERR_TOKEN_IS_EXPIRED: {
		GRPC_Code: codes.Unauthenticated,
		Message:   ERR_TOKEN_IS_EXPIRED,
	},
	ERR_EMPTY_CONN: {
		GRPC_Code: codes.OK,
		Message:   ERR_EMPTY_CONN,
	},
	ERR_EXAMPLE_NOT_FOUND: {
		GRPC_Code: codes.OK,
		Message:   ERR_EXAMPLE_NOT_FOUND,
	},
	ERR_EXAMPLE_INVALID: {
		GRPC_Code: codes.OK,
		Message:   ERR_EXAMPLE_INVALID,
	},
	ERR_DATA_NOT_FOUND: {
		GRPC_Code: codes.NotFound,
		Message:   ERR_DATA_NOT_FOUND,
	},
	ERR_DATA_INVALID: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_DATA_INVALID,
	},
	ERR_INSERT_FAILED: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_INSERT_FAILED,
	},
	ERR_GET_FAILED: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_GET_FAILED,
	},
	ERR_PUT_FAILED: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_PUT_FAILED,
	},
	ERR_PATCH_FAILED: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_PATCH_FAILED,
	},
	ERR_DELETE_FAILED: {
		GRPC_Code: codes.Unavailable,
		Message:   ERR_DELETE_FAILED,
	},
}

func HandleGRPCErrResponse(err error) (code codes.Code, response any) {
	if e, ok := status.FromError(err); ok {
		if e, ok := MAP_ERR_RESPONSE[e.Message()]; ok {
			return e.GRPC_Code, ErrorResponse{
				Message: e.Message,
				Error:   err.Error(),
			}
		}
	}

	return codes.Unavailable, map[string]any{
		"code":    http.StatusText(http.StatusServiceUnavailable),
		"message": "internal server error",
		"error":   err.Error(),
	}
}
