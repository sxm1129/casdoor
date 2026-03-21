package object

import "github.com/casdoor/casdoor/util"

type UserIdentity struct {
	Id           int    `xorm:"int notnull pk autoincr" json:"id"`
	Owner        string `xorm:"varchar(100) notnull index" json:"owner"`
	Name         string `xorm:"varchar(100) notnull index" json:"name"`
	ProviderType string `xorm:"varchar(100) notnull index" json:"providerType"`
	ProviderId   string `xorm:"varchar(255)" json:"providerId"`
	AccessToken  string `xorm:"mediumtext" json:"accessToken"`
}

func GetUserIdentitiesByUser(owner, name string) ([]*UserIdentity, error) {
	identities := []*UserIdentity{}
	err := ormer.Engine.Where("owner = ? and name = ?", owner, name).Find(&identities)
	if err != nil {
		return identities, err
	}
	return identities, nil
}

func UpdateUserIdentity(owner, name, providerType, providerId, accessToken string) error {
	identity := UserIdentity{
		Owner:        owner,
		Name:         name,
		ProviderType: providerType,
	}

	exists, err := ormer.Engine.Get(&identity)
	if err != nil {
		return err
	}

	if exists {
		identity.ProviderId = providerId
		if accessToken != "" {
			identity.AccessToken = accessToken
		}
		_, err = ormer.Engine.ID(identity.Id).Update(&identity)
		return err
	} else {
		identity.ProviderId = providerId
		identity.AccessToken = accessToken
		_, err = ormer.Engine.Insert(&identity)
		return err
	}
}

func GetUserByProviderId(providerType, providerId string) (*User, error) {
	identity := UserIdentity{
		ProviderType: providerType,
		ProviderId:   providerId,
	}
	exists, err := ormer.Engine.Get(&identity)
	if err != nil || !exists {
		return nil, err
	}

	return GetUser(util.GetId(identity.Owner, identity.Name))
}
