package storage

import (
	"context"
	"github.com/Inspirate789/Thermy-backend/internal/domain/entities"
	"github.com/Inspirate789/Thermy-backend/internal/domain/errors"
	"github.com/Inspirate789/Thermy-backend/internal/domain/interfaces"
	log "github.com/sirupsen/logrus"
)

type StorageService struct {
	storage         Storage
	logger          *log.Logger
	repeatOnFailure int
}

func NewStorageService(storage Storage, logger *log.Logger, repeatOnFailure int) *StorageService {
	return &StorageService{
		storage:         storage,
		logger:          logger,
		repeatOnFailure: repeatOnFailure,
	}
}

func (ss *StorageService) OpenConn(request *entities.AuthRequest, ctx context.Context) (ConnDB, string, error) {
	conn, role, err := ss.storage.OpenConn(request, ctx)
	if err != nil {
		ss.logger.Error(err)
		return conn, role, errors.IdentifyStorageError(err)
	}
	ss.logger.Infof("storage connection for %s opened", role)

	return conn, role, nil
}

func (ss *StorageService) CloseConn(conn ConnDB) error {
	err := ss.storage.CloseConn(conn)
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}
	ss.logger.Info("storage connection closed") // TODO: make them more informative?

	return nil
}

func (ss *StorageService) GetAllUnits(conn ConnDB, layer string) (interfaces.OutputUnitsDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, err
	}

	units, err := ss.storage.GetAllUnits(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}

	return units, nil
}

func (ss *StorageService) GetUnitsByModels(conn ConnDB, layer string, modelsDTO interfaces.ModelsIdDTO) (interfaces.OutputUnitsDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, err
	}

	if len(modelsDTO.Models) == 0 {
		return interfaces.OutputUnitsDTO{
			Units:    make([]map[string]interfaces.OutputUnitDTO, 0),
			Contexts: make([]interfaces.ContextDTO, 0),
		}, nil
	}

	units, err := ss.storage.GetUnitsByModels(conn, layer, modelsDTO.Models)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}

	return units, nil
}

func (ss *StorageService) GetUnitsByProperties(conn ConnDB, layer string, propertiesDTO interfaces.PropertiesIdDTO) (interfaces.OutputUnitsDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, err
	}

	if len(propertiesDTO.Properties) == 0 {
		return interfaces.OutputUnitsDTO{
			Units:    make([]map[string]interfaces.OutputUnitDTO, 0),
			Contexts: make([]interfaces.ContextDTO, 0),
		}, nil
	}

	units, err := ss.storage.GetUnitsByProperties(conn, layer, propertiesDTO.Properties)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputUnitsDTO{}, errors.IdentifyStorageError(err)
	}

	return units, nil
}

func (ss *StorageService) GetModels(conn ConnDB, layer string) (interfaces.OutputModelsDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputModelsDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.OutputModelsDTO{}, err
	}

	models, err := ss.storage.GetAllModels(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputModelsDTO{}, errors.IdentifyStorageError(err)
	}

	modelsDTO := make([]interfaces.OutputModelDTO, 0, len(models))
	for i := range models {
		modelsDTO = append(modelsDTO, interfaces.OutputModelDTO(models[i]))
	}

	return interfaces.OutputModelsDTO{Models: modelsDTO}, nil
}

func (ss *StorageService) GetModelElements(conn ConnDB, layer string) (interfaces.OutputModelElementsDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputModelElementsDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.OutputModelElementsDTO{}, err
	}

	modelElements, err := ss.storage.GetAllModelElements(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputModelElementsDTO{}, errors.IdentifyStorageError(err)
	}

	modelElementsDTO := make([]interfaces.OutputModelElementDTO, 0, len(modelElements))
	for i := range modelElements {
		modelElementsDTO = append(modelElementsDTO, interfaces.OutputModelElementDTO(modelElements[i]))
	}

	return interfaces.OutputModelElementsDTO{Elements: modelElementsDTO}, nil
}

func (ss *StorageService) GetProperties(conn ConnDB) (interfaces.OutputPropertiesDTO, error) {
	properties, err := ss.storage.GetAllProperties(conn)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputPropertiesDTO{}, errors.IdentifyStorageError(err)
	}

	propertiesDTO := make([]interfaces.OutputPropertyDTO, 0, len(properties))
	for i := range properties {
		propertiesDTO = append(propertiesDTO, interfaces.OutputPropertyDTO(properties[i]))
	}

	return interfaces.OutputPropertiesDTO{Properties: propertiesDTO}, nil
}

func (ss *StorageService) GetPropertiesByUnit(conn ConnDB, layer string, unit interfaces.SearchUnitDTO) (interfaces.OutputPropertiesDTO, error) {
	properties, err := ss.storage.GetPropertiesByUnit(conn, layer, unit)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.OutputPropertiesDTO{}, errors.IdentifyStorageError(err)
	}

	propertiesDTO := make([]interfaces.OutputPropertyDTO, 0, len(properties))
	for i := range properties {
		propertiesDTO = append(propertiesDTO, interfaces.OutputPropertyDTO(properties[i]))
	}

	return interfaces.OutputPropertiesDTO{Properties: propertiesDTO}, nil
}

func (ss *StorageService) GetLayers(conn ConnDB) (interfaces.LayersDTO, error) {
	layers, err := ss.storage.GetAllLayers(conn)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.LayersDTO{}, errors.IdentifyStorageError(err)
	}

	return interfaces.LayersDTO{Layers: layers}, nil
}

func (ss *StorageService) SaveUnits(conn ConnDB, layer string, unitsDTO interfaces.SaveUnitsDTO) error {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return err
	}

	for i := 0; i < ss.repeatOnFailure; i++ {
		err = ss.storage.SaveUnits(conn, layer, unitsDTO)
		if err == nil || errors.IdentifyStorageError(err) != errors.ErrConnectDatabase {
			break
		}
	}
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}

	return nil
}

func (ss *StorageService) UpdateUnits(conn ConnDB, layer string, unitsDTO interfaces.UpdateUnitsDTO) error {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return err
	}

	for _, unit := range unitsDTO.Units {
		name := unit.OldText

		if unit.NewText != nil {
			err = ss.storage.RenameUnit(conn, layer, unit.Lang, unit.OldText, *unit.NewText)
			if err != nil {
				ss.logger.Error(err)
				return errors.IdentifyStorageError(err)
			}
			name = *unit.NewText
		}

		if len(unit.PropertiesID) != 0 {
			err = ss.storage.SetUnitProperties(conn, layer, unit.Lang, name, unit.PropertiesID)
			if err != nil {
				ss.logger.Error(err)
				return errors.IdentifyStorageError(err)
			}
		}
	}

	return nil
}

func (ss *StorageService) SaveProperties(conn ConnDB, propertiesDTO interfaces.PropertyNamesDTO) (interfaces.PropertiesIdDTO, error) {
	propertiesID, err := ss.storage.SaveProperties(conn, propertiesDTO.Properties)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.PropertiesIdDTO{}, errors.IdentifyStorageError(err)
	}

	return interfaces.PropertiesIdDTO{Properties: propertiesID}, nil
}

func (ss *StorageService) SaveModels(conn ConnDB, layer string, modelsDTO interfaces.ModelNamesDTO) (interfaces.ModelsIdDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.ModelsIdDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.ModelsIdDTO{}, err
	}

	for _, modelName := range modelsDTO.Models {
		model := entities.Model{Name: modelName}
		err = model.IsValidName()
		if err != nil {
			ss.logger.Error(err)
			return interfaces.ModelsIdDTO{}, err
		}
	}

	modelsID, err := ss.storage.SaveModels(conn, layer, modelsDTO.Models)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.ModelsIdDTO{}, errors.IdentifyStorageError(err)
	}

	return interfaces.ModelsIdDTO{Models: modelsID}, nil
}

func (ss *StorageService) SaveModelElements(conn ConnDB, layer string, modelElementsDTO interfaces.ModelElementNamesDTO) (interfaces.ModelElementsIdDTO, error) {
	exist, err := ss.storage.LayerExist(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.ModelElementsIdDTO{}, errors.IdentifyStorageError(err)
	}
	if !exist {
		err = errors.ErrUnknownLayerWrap(layer)
		ss.logger.Error(err)
		return interfaces.ModelElementsIdDTO{}, err
	}

	modelElementsID, err := ss.storage.SaveModelElements(conn, layer, modelElementsDTO.ModelElements)
	if err != nil {
		ss.logger.Error(err)
		return interfaces.ModelElementsIdDTO{}, errors.IdentifyStorageError(err)
	}

	return interfaces.ModelElementsIdDTO{ModelElements: modelElementsID}, nil
}

func (ss *StorageService) SaveLayer(conn ConnDB, layer string) error {
	err := ss.storage.SaveLayer(conn, layer)
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}

	return nil
}

func (ss *StorageService) AddUser(conn ConnDB, user interfaces.UserDTO) error {
	err := ss.storage.AddUser(conn, user)
	if err != nil {
		ss.logger.Error(err)
		return errors.IdentifyStorageError(err)
	}
	ss.logger.Infof("storage user (name: %s, role: %s) inserted to database", user.Name, user.Role)

	return nil
}

//func (ss *StorageService) GetUserPassword(conn ConnDB, username string) (string, error) {
//	password, err := ss.storage.GetUserPassword(conn, username)
//	if err != nil {
//		ss.logger.Error(err)
//		return "", errors.IdentifyStorageError(err)
//	}
//
//	return password, nil
//}
