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

// NewGetAPI5OpenvpnConfigParams creates a new GetAPI5OpenvpnConfigParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPI5OpenvpnConfigParams() *GetAPI5OpenvpnConfigParams {
	return &GetAPI5OpenvpnConfigParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPI5OpenvpnConfigParamsWithTimeout creates a new GetAPI5OpenvpnConfigParams object
// with the ability to set a timeout on a request.
func NewGetAPI5OpenvpnConfigParamsWithTimeout(timeout time.Duration) *GetAPI5OpenvpnConfigParams {
	return &GetAPI5OpenvpnConfigParams{
		timeout: timeout,
	}
}

// NewGetAPI5OpenvpnConfigParamsWithContext creates a new GetAPI5OpenvpnConfigParams object
// with the ability to set a context for a request.
func NewGetAPI5OpenvpnConfigParamsWithContext(ctx context.Context) *GetAPI5OpenvpnConfigParams {
	return &GetAPI5OpenvpnConfigParams{
		Context: ctx,
	}
}

// NewGetAPI5OpenvpnConfigParamsWithHTTPClient creates a new GetAPI5OpenvpnConfigParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPI5OpenvpnConfigParamsWithHTTPClient(client *http.Client) *GetAPI5OpenvpnConfigParams {
	return &GetAPI5OpenvpnConfigParams{
		HTTPClient: client,
	}
}

/*
GetAPI5OpenvpnConfigParams contains all the parameters to send to the API endpoint

	for the get API 5 openvpn config operation.

	Typically these are written to a http.Request.
*/
type GetAPI5OpenvpnConfigParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API 5 openvpn config params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPI5OpenvpnConfigParams) WithDefaults() *GetAPI5OpenvpnConfigParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API 5 openvpn config params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPI5OpenvpnConfigParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) WithTimeout(timeout time.Duration) *GetAPI5OpenvpnConfigParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) WithContext(ctx context.Context) *GetAPI5OpenvpnConfigParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) WithHTTPClient(client *http.Client) *GetAPI5OpenvpnConfigParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API 5 openvpn config params
func (o *GetAPI5OpenvpnConfigParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPI5OpenvpnConfigParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
