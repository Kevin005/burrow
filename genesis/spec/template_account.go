package spec

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

type TemplateAccount struct {
	// Template accounts sharing a name will be merged when merging genesis specs
	Name string `json:",omitempty" toml:",omitempty"`
	// Address  is convenient to have in file for reference, but otherwise ignored since derived from PublicKey
	Address     *crypto.Address   `json:",omitempty" toml:",omitempty"`
	NodeAddress *crypto.Address   `json:",omitempty" toml:",omitempty"`
	PublicKey   *crypto.PublicKey `json:",omitempty" toml:",omitempty"`
	Amount      *uint64           `json:",omitempty" toml:",omitempty"`
	Power       *uint64           `json:",omitempty" toml:",omitempty"`
	Permissions []string          `json:",omitempty" toml:",omitempty"`
	Roles       []string          `json:",omitempty" toml:",omitempty"`
}

func (ta TemplateAccount) Validator(keyClient keys.KeyClient, index int, generateNodeKeys bool) (*genesis.Validator, error) {
	var err error
	gv := new(genesis.Validator)
	gv.PublicKey, gv.Address, err = ta.RealisePubKeyAndAddress(keyClient)
	if err != nil {
		return nil, err
	}
	if generateNodeKeys && ta.NodeAddress == nil {
		// If neither PublicKey or Address set then generate a new one
		address, err := keyClient.Generate("nodekey-"+ta.Name, crypto.CurveTypeEd25519)
		if err != nil {
			return nil, err
		}
		ta.NodeAddress = &address
	}
	if ta.Power == nil {
		gv.Amount = DefaultPower
	} else {
		gv.Amount = *ta.Power
	}
	if ta.Name == "" {
		gv.Name = accountNameFromIndex(index)
	} else {
		gv.Name = ta.Name
	}

	gv.UnbondTo = []genesis.BasicAccount{{
		Address:   gv.Address,
		PublicKey: gv.PublicKey,
		Amount:    gv.Amount,
	}}
	gv.NodeAddress = ta.NodeAddress
	return gv, nil
}

func (ta TemplateAccount) AccountPermissions() (ptypes.AccountPermissions, error) {
	basePerms, err := permission.BasePermissionsFromStringList(ta.Permissions)
	if err != nil {
		return permission.ZeroAccountPermissions, nil
	}
	return ptypes.AccountPermissions{
		Base:  basePerms,
		Roles: ta.Roles,
	}, nil
}

func (ta TemplateAccount) Account(keyClient keys.KeyClient, index int) (*genesis.Account, error) {
	var err error
	ga := new(genesis.Account)
	ga.PublicKey, ga.Address, err = ta.RealisePubKeyAndAddress(keyClient)
	if err != nil {
		return nil, err
	}
	if ta.Amount == nil {
		ga.Amount = DefaultAmount
	} else {
		ga.Amount = *ta.Amount
	}
	if ta.Name == "" {
		ga.Name = accountNameFromIndex(index)
	} else {
		ga.Name = ta.Name
	}
	if ta.Permissions == nil {
		ga.Permissions = permission.DefaultAccountPermissions.Clone()
	} else {
		ga.Permissions, err = ta.AccountPermissions()
		if err != nil {
			return nil, err
		}
	}
	return ga, nil
}

// Adds a public key and address to the template. If PublicKey will try to fetch it by Address.
// If both PublicKey and Address are not set will use the keyClient to generate a new keypair
func (ta TemplateAccount) RealisePubKeyAndAddress(keyClient keys.KeyClient) (pubKey crypto.PublicKey, address crypto.Address, err error) {
	if ta.PublicKey == nil {
		if ta.Address == nil {
			// If neither PublicKey or Address set then generate a new one
			address, err = keyClient.Generate(ta.Name, crypto.CurveTypeEd25519)
			if err != nil {
				return
			}
		} else {
			address = *ta.Address
		}
		// Get the (possibly existing) key
		pubKey, err = keyClient.PublicKey(address)
		if err != nil {
			return
		}
	} else {
		address = (*ta.PublicKey).Address()
		if ta.Address != nil && *ta.Address != address {
			err = fmt.Errorf("template address %s does not match public key derived address %s", ta.Address,
				ta.PublicKey)
		}
		pubKey = *ta.PublicKey
	}
	return
}
