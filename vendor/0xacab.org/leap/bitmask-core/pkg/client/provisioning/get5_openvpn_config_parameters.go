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

// NewGet5OpenvpnConfigParams creates a new Get5OpenvpnConfigParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGet5OpenvpnConfigParams() *Get5OpenvpnConfigParams {
	return &Get5OpenvpnConfigParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGet5OpenvpnConfigParamsWithTimeout creates a new Get5OpenvpnConfigParams object
// with the ability to set a timeout on a request.
func NewGet5OpenvpnConfigParamsWithTimeout(timeout time.Duration) *Get5OpenvpnConfigParams {
	return &Get5OpenvpnConfigParams{
		timeout: timeout,
	}
}

// NewGet5OpenvpnConfigParamsWithContext creates a new Get5OpenvpnConfigParams object
// with the ability to set a context for a request.
func NewGet5OpenvpnConfigParamsWithContext(ctx context.Context) *Get5OpenvpnConfigParams {
	return &Get5OpenvpnConfigParams{
		Context: ctx,
	}
}

// NewGet5OpenvpnConfigParamsWithHTTPClient creates a new Get5OpenvpnConfigParams object
// with the ability to set a custom HTTPClient for a request.
func NewGet5OpenvpnConfigParamsWithHTTPClient(client *http.Client) *Get5OpenvpnConfigParams {
	return &Get5OpenvpnConfigParams{
		HTTPClient: client,
	}
}

/*
Get5OpenvpnConfigParams contains all the parameters to send to the API endpoint

	for the get5 openvpn config operation.

	Typically these are written to a http.Request.
*/
type Get5OpenvpnConfigParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get5 openvpn config params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *Get5OpenvpnConfigParams) WithDefaults() *Get5OpenvpnConfigParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get5 openvpn config params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *Get5OpenvpnConfigParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) WithTimeout(timeout time.Duration) *Get5OpenvpnConfigParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) WithContext(ctx context.Context) *Get5OpenvpnConfigParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) WithHTTPClient(client *http.Client) *Get5OpenvpnConfigParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get5 openvpn config params
func (o *Get5OpenvpnConfigParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *Get5OpenvpnConfigParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
