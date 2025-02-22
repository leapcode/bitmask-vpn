// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// MainGateways main gateways
//
// swagger:model main.Gateways
type MainGateways struct {

	// all
	All string `json:"all,omitempty"`

	// location
	Location string `json:"location,omitempty"`
}

// Validate validates this main gateways
func (m *MainGateways) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this main gateways based on context it is used
func (m *MainGateways) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *MainGateways) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *MainGateways) UnmarshalBinary(b []byte) error {
	var res MainGateways
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
