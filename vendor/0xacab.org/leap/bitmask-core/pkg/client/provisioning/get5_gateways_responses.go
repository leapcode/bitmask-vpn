// Code generated by go-swagger; DO NOT EDIT.

package provisioning

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"0xacab.org/leap/bitmask-core/models"
)

// Get5GatewaysReader is a Reader for the Get5Gateways structure.
type Get5GatewaysReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *Get5GatewaysReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGet5GatewaysOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGet5GatewaysBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGet5GatewaysNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGet5GatewaysInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /5/gateways] Get5Gateways", response, response.Code())
	}
}

// NewGet5GatewaysOK creates a Get5GatewaysOK with default headers values
func NewGet5GatewaysOK() *Get5GatewaysOK {
	return &Get5GatewaysOK{}
}

/*
Get5GatewaysOK describes a response with status code 200, with default header values.

OK
*/
type Get5GatewaysOK struct {
	Payload []*models.ModelsGateway
}

// IsSuccess returns true when this get5 gateways o k response has a 2xx status code
func (o *Get5GatewaysOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get5 gateways o k response has a 3xx status code
func (o *Get5GatewaysOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 gateways o k response has a 4xx status code
func (o *Get5GatewaysOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get5 gateways o k response has a 5xx status code
func (o *Get5GatewaysOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 gateways o k response a status code equal to that given
func (o *Get5GatewaysOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get5 gateways o k response
func (o *Get5GatewaysOK) Code() int {
	return 200
}

func (o *Get5GatewaysOK) Error() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysOK  %+v", 200, o.Payload)
}

func (o *Get5GatewaysOK) String() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysOK  %+v", 200, o.Payload)
}

func (o *Get5GatewaysOK) GetPayload() []*models.ModelsGateway {
	return o.Payload
}

func (o *Get5GatewaysOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5GatewaysBadRequest creates a Get5GatewaysBadRequest with default headers values
func NewGet5GatewaysBadRequest() *Get5GatewaysBadRequest {
	return &Get5GatewaysBadRequest{}
}

/*
Get5GatewaysBadRequest describes a response with status code 400, with default header values.

Bad Request
*/
type Get5GatewaysBadRequest struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 gateways bad request response has a 2xx status code
func (o *Get5GatewaysBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 gateways bad request response has a 3xx status code
func (o *Get5GatewaysBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 gateways bad request response has a 4xx status code
func (o *Get5GatewaysBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this get5 gateways bad request response has a 5xx status code
func (o *Get5GatewaysBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 gateways bad request response a status code equal to that given
func (o *Get5GatewaysBadRequest) IsCode(code int) bool {
	return code == 400
}

// Code gets the status code for the get5 gateways bad request response
func (o *Get5GatewaysBadRequest) Code() int {
	return 400
}

func (o *Get5GatewaysBadRequest) Error() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysBadRequest  %+v", 400, o.Payload)
}

func (o *Get5GatewaysBadRequest) String() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysBadRequest  %+v", 400, o.Payload)
}

func (o *Get5GatewaysBadRequest) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5GatewaysBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5GatewaysNotFound creates a Get5GatewaysNotFound with default headers values
func NewGet5GatewaysNotFound() *Get5GatewaysNotFound {
	return &Get5GatewaysNotFound{}
}

/*
Get5GatewaysNotFound describes a response with status code 404, with default header values.

Not Found
*/
type Get5GatewaysNotFound struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 gateways not found response has a 2xx status code
func (o *Get5GatewaysNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 gateways not found response has a 3xx status code
func (o *Get5GatewaysNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 gateways not found response has a 4xx status code
func (o *Get5GatewaysNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get5 gateways not found response has a 5xx status code
func (o *Get5GatewaysNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 gateways not found response a status code equal to that given
func (o *Get5GatewaysNotFound) IsCode(code int) bool {
	return code == 404
}

// Code gets the status code for the get5 gateways not found response
func (o *Get5GatewaysNotFound) Code() int {
	return 404
}

func (o *Get5GatewaysNotFound) Error() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysNotFound  %+v", 404, o.Payload)
}

func (o *Get5GatewaysNotFound) String() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysNotFound  %+v", 404, o.Payload)
}

func (o *Get5GatewaysNotFound) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5GatewaysNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5GatewaysInternalServerError creates a Get5GatewaysInternalServerError with default headers values
func NewGet5GatewaysInternalServerError() *Get5GatewaysInternalServerError {
	return &Get5GatewaysInternalServerError{}
}

/*
Get5GatewaysInternalServerError describes a response with status code 500, with default header values.

Internal Server Error
*/
type Get5GatewaysInternalServerError struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 gateways internal server error response has a 2xx status code
func (o *Get5GatewaysInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 gateways internal server error response has a 3xx status code
func (o *Get5GatewaysInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 gateways internal server error response has a 4xx status code
func (o *Get5GatewaysInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get5 gateways internal server error response has a 5xx status code
func (o *Get5GatewaysInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get5 gateways internal server error response a status code equal to that given
func (o *Get5GatewaysInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the get5 gateways internal server error response
func (o *Get5GatewaysInternalServerError) Code() int {
	return 500
}

func (o *Get5GatewaysInternalServerError) Error() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysInternalServerError  %+v", 500, o.Payload)
}

func (o *Get5GatewaysInternalServerError) String() string {
	return fmt.Sprintf("[GET /5/gateways][%d] get5GatewaysInternalServerError  %+v", 500, o.Payload)
}

func (o *Get5GatewaysInternalServerError) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5GatewaysInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}