// Code generated by go-swagger; DO NOT EDIT.

package provisioning

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// Get5OpenvpnConfigReader is a Reader for the Get5OpenvpnConfig structure.
type Get5OpenvpnConfigReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *Get5OpenvpnConfigReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGet5OpenvpnConfigOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGet5OpenvpnConfigBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGet5OpenvpnConfigNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGet5OpenvpnConfigInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("[GET /5/openvpn/config] Get5OpenvpnConfig", response, response.Code())
	}
}

// NewGet5OpenvpnConfigOK creates a Get5OpenvpnConfigOK with default headers values
func NewGet5OpenvpnConfigOK() *Get5OpenvpnConfigOK {
	return &Get5OpenvpnConfigOK{}
}

/*
Get5OpenvpnConfigOK describes a response with status code 200, with default header values.

OK
*/
type Get5OpenvpnConfigOK struct {
	Payload string
}

// IsSuccess returns true when this get5 openvpn config o k response has a 2xx status code
func (o *Get5OpenvpnConfigOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get5 openvpn config o k response has a 3xx status code
func (o *Get5OpenvpnConfigOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 openvpn config o k response has a 4xx status code
func (o *Get5OpenvpnConfigOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get5 openvpn config o k response has a 5xx status code
func (o *Get5OpenvpnConfigOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 openvpn config o k response a status code equal to that given
func (o *Get5OpenvpnConfigOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get5 openvpn config o k response
func (o *Get5OpenvpnConfigOK) Code() int {
	return 200
}

func (o *Get5OpenvpnConfigOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigOK %s", 200, payload)
}

func (o *Get5OpenvpnConfigOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigOK %s", 200, payload)
}

func (o *Get5OpenvpnConfigOK) GetPayload() string {
	return o.Payload
}

func (o *Get5OpenvpnConfigOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5OpenvpnConfigBadRequest creates a Get5OpenvpnConfigBadRequest with default headers values
func NewGet5OpenvpnConfigBadRequest() *Get5OpenvpnConfigBadRequest {
	return &Get5OpenvpnConfigBadRequest{}
}

/*
Get5OpenvpnConfigBadRequest describes a response with status code 400, with default header values.

Bad Request
*/
type Get5OpenvpnConfigBadRequest struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 openvpn config bad request response has a 2xx status code
func (o *Get5OpenvpnConfigBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 openvpn config bad request response has a 3xx status code
func (o *Get5OpenvpnConfigBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 openvpn config bad request response has a 4xx status code
func (o *Get5OpenvpnConfigBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this get5 openvpn config bad request response has a 5xx status code
func (o *Get5OpenvpnConfigBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 openvpn config bad request response a status code equal to that given
func (o *Get5OpenvpnConfigBadRequest) IsCode(code int) bool {
	return code == 400
}

// Code gets the status code for the get5 openvpn config bad request response
func (o *Get5OpenvpnConfigBadRequest) Code() int {
	return 400
}

func (o *Get5OpenvpnConfigBadRequest) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigBadRequest %s", 400, payload)
}

func (o *Get5OpenvpnConfigBadRequest) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigBadRequest %s", 400, payload)
}

func (o *Get5OpenvpnConfigBadRequest) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5OpenvpnConfigBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5OpenvpnConfigNotFound creates a Get5OpenvpnConfigNotFound with default headers values
func NewGet5OpenvpnConfigNotFound() *Get5OpenvpnConfigNotFound {
	return &Get5OpenvpnConfigNotFound{}
}

/*
Get5OpenvpnConfigNotFound describes a response with status code 404, with default header values.

Not Found
*/
type Get5OpenvpnConfigNotFound struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 openvpn config not found response has a 2xx status code
func (o *Get5OpenvpnConfigNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 openvpn config not found response has a 3xx status code
func (o *Get5OpenvpnConfigNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 openvpn config not found response has a 4xx status code
func (o *Get5OpenvpnConfigNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get5 openvpn config not found response has a 5xx status code
func (o *Get5OpenvpnConfigNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get5 openvpn config not found response a status code equal to that given
func (o *Get5OpenvpnConfigNotFound) IsCode(code int) bool {
	return code == 404
}

// Code gets the status code for the get5 openvpn config not found response
func (o *Get5OpenvpnConfigNotFound) Code() int {
	return 404
}

func (o *Get5OpenvpnConfigNotFound) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigNotFound %s", 404, payload)
}

func (o *Get5OpenvpnConfigNotFound) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigNotFound %s", 404, payload)
}

func (o *Get5OpenvpnConfigNotFound) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5OpenvpnConfigNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGet5OpenvpnConfigInternalServerError creates a Get5OpenvpnConfigInternalServerError with default headers values
func NewGet5OpenvpnConfigInternalServerError() *Get5OpenvpnConfigInternalServerError {
	return &Get5OpenvpnConfigInternalServerError{}
}

/*
Get5OpenvpnConfigInternalServerError describes a response with status code 500, with default header values.

Internal Server Error
*/
type Get5OpenvpnConfigInternalServerError struct {
	Payload interface{}
}

// IsSuccess returns true when this get5 openvpn config internal server error response has a 2xx status code
func (o *Get5OpenvpnConfigInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get5 openvpn config internal server error response has a 3xx status code
func (o *Get5OpenvpnConfigInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get5 openvpn config internal server error response has a 4xx status code
func (o *Get5OpenvpnConfigInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get5 openvpn config internal server error response has a 5xx status code
func (o *Get5OpenvpnConfigInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get5 openvpn config internal server error response a status code equal to that given
func (o *Get5OpenvpnConfigInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the get5 openvpn config internal server error response
func (o *Get5OpenvpnConfigInternalServerError) Code() int {
	return 500
}

func (o *Get5OpenvpnConfigInternalServerError) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigInternalServerError %s", 500, payload)
}

func (o *Get5OpenvpnConfigInternalServerError) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /5/openvpn/config][%d] get5OpenvpnConfigInternalServerError %s", 500, payload)
}

func (o *Get5OpenvpnConfigInternalServerError) GetPayload() interface{} {
	return o.Payload
}

func (o *Get5OpenvpnConfigInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
