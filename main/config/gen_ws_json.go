// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package config

import (
	"encoding/json"
	"errors"

	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*RpcMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (r RpcWS) MarshalJSON() ([]byte, error) {
	type RpcWS struct {
		Disable    bool           `json:"disable"`
		Port       hexutil.Uint32 `json:"port"  gencodec:"required"`
		CorsDomain string         `json:"corsDomain"`
	}
	var enc RpcWS
	enc.Disable = r.Disable
	enc.Port = hexutil.Uint32(r.Port)
	enc.CorsDomain = r.CorsDomain
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (r *RpcWS) UnmarshalJSON(input []byte) error {
	type RpcWS struct {
		Disable    *bool           `json:"disable"`
		Port       *hexutil.Uint32 `json:"port"  gencodec:"required"`
		CorsDomain *string         `json:"corsDomain"`
	}
	var dec RpcWS
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Disable != nil {
		r.Disable = *dec.Disable
	}
	if dec.Port == nil {
		return errors.New("missing required field 'port' for RpcWS")
	}
	r.Port = uint32(*dec.Port)
	if dec.CorsDomain != nil {
		r.CorsDomain = *dec.CorsDomain
	}
	return nil
}
