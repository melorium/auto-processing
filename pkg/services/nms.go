package services

import (
	"context"
	"fmt"

	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"

	"github.com/jinzhu/gorm"
)

type NmsService struct {
	db *gorm.DB
}

func NewNmsService(db *gorm.DB) NmsService {
	return NmsService{db: db}
}

func (s NmsService) Apply(ctx context.Context, r api.NmsApplyRequests) (*api.NmsApplyResponse, error) {
	// Start transaction to fall back
	// if we get an error
	tx := s.db.BeginTx(ctx, nil)

	// Iterate over the requested nm-servers
	// and append them to the response
	var resp api.NmsApplyResponse
	for _, nms := range r.Nms {
		// Check if the requested NMS exists (in that case update it)
		var newNms api.Nms
		if err := tx.Preload("Licences").Where("address = ?", nms.Address).First(&newNms).Error; err != nil {
			// return the error if it isn't a "record not found"-error
			if !gorm.IsRecordNotFoundError(err) {
				return nil, err
			}
		}

		// Set data to the new Nms-model
		newNms.Address = nms.Address
		newNms.Port = nms.Port
		newNms.Username = nms.Username
		newNms.Password = nms.Password
		newNms.Workers = nms.Workers

		// Create hash-map for the existing licences
		hashMap := make(map[string]api.Licence)
		for _, lic := range newNms.Licences {
			hashMap[lic.Type] = lic
		}

		// append the requested licences to the new nms-model
		for _, lic := range nms.Licences {
			var newLicence api.Licence

			// Check if the licencetype exists in the hashmap and update the amount
			// else create a new licence
			if existing, found := hashMap[lic.Licence.Type]; found {
				newLicence = existing
			} else {
				newLicence.Type = lic.Licence.Type
			}
			newLicence.Amount = lic.Licence.Amount
			newNms.Licences = append(newNms.Licences, newLicence)
		}

		// Save the new NMS to the DB
		if err := tx.Save(&newNms).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to apply server %s : %v", newNms.Address, err)
		}

		// Append the new nms to the response
		resp.Nms = append(resp.Nms, newNms)
	}

	return &resp, tx.Commit().Error
}
func (s NmsService) List(ctx context.Context, r api.NmsListRequest) (*api.NmsListResponse, error) {
	var nms []api.Nms
	err := s.db.Preload("Licences").Find(&nms).Error
	return &api.NmsListResponse{Nms: nms}, err
}

func (s NmsService) ListLicences(ctx context.Context, r api.NmsListLicencesRequest) (*api.NmsListLicencesResponse, error) {
	return nil, nil
}
