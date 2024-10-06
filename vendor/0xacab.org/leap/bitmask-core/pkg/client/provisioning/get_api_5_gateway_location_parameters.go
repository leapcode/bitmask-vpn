// Code generated by go-swagger; DO NOT EDIT.

package provisioning

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetAPI5GatewayLocationParams creates a new GetAPI5GatewayLocationParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPI5GatewayLocationParams() *GetAPI5GatewayLocationParams {
	return &GetAPI5GatewayLocationParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPI5GatewayLocationParamsWithTimeout creates a new GetAPI5GatewayLocationParams object
// with the ability to set a timeout on a request.
func NewGetAPI5GatewayLocationParamsWithTimeout(timeout time.Duration) *GetAPI5GatewayLocationParams {
	return &GetAPI5GatewayLocationParams{
		timeout: timeout,
	}
}

// NewGetAPI5GatewayLocationParamsWithContext creates a new GetAPI5GatewayLocationParams object
// with the ability to set a context for a request.
func NewGetAPI5GatewayLocationParamsWithContext(ctx context.Context) *GetAPI5GatewayLocationParams {
	return &GetAPI5GatewayLocationParams{
		Context: ctx,
	}
}

// NewGetAPI5GatewayLocationParamsWithHTTPClient creates a new GetAPI5GatewayLocationParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPI5GatewayLocationParamsWithHTTPClient(client *http.Client) *GetAPI5GatewayLocationParams {
	return &GetAPI5GatewayLocationParams{
		HTTPClient: client,
	}
}

/*
GetAPI5GatewayLocationParams contains all the parameters to send to the API endpoint

	for the get API 5 gateway location operation.

	Typically these are written to a http.Request.
*/
type GetAPI5GatewayLocationParams struct {

	/* Location.

	   Location ID
	*/
	Location string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API 5 gateway location params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPI5GatewayLocationParams) WithDefaults() *GetAPI5GatewayLocationParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API 5 gateway location params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPI5GatewayLocationParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) WithTimeout(timeout time.Duration) *GetAPI5GatewayLocationParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) WithContext(ctx context.Context) *GetAPI5GatewayLocationParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) WithHTTPClient(client *http.Client) *GetAPI5GatewayLocationParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithLocation adds the location to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) WithLocation(location string) *GetAPI5GatewayLocationParams {
	o.SetLocation(location)
	return o
}

// SetLocation adds the location to the get API 5 gateway location params
func (o *GetAPI5GatewayLocationParams) SetLocation(location string) {
	o.Location = location
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPI5GatewayLocationParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param location
	if err := r.SetPathParam("location", o.Location); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}