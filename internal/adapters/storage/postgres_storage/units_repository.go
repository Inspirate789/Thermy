package postgres_storage

import (
	"errors"
	"github.com/Inspirate789/Thermy-backend/internal/domain/entities"
	"github.com/Inspirate789/Thermy-backend/internal/domain/interfaces"
	"github.com/Inspirate789/Thermy-backend/internal/domain/services/storage"
	"github.com/lib/pq"
)

type UnitsPgRepository struct{}

func (r *UnitsPgRepository) makeOutputUnitDTO(conn storage.ConnDB, layer string, lang string, unit entities.Unit) (interfaces.OutputUnitDTO, error) {
	args := map[string]any{
		"layer_name": layer,
		"lang":       lang,
		"unit_id":    unit.ID,
	}
	propertiesID, err := namedSelectSliceFromScript[[]int](conn, selectPropertiesIdByUnitId, args)
	if err != nil {
		return interfaces.OutputUnitDTO{}, err
	}

	contextsID, err := namedSelectSliceFromScript[[]int](conn, selectContextsIdByUnit, args)
	if err != nil {
		return interfaces.OutputUnitDTO{}, err
	}

	unitDTO := interfaces.OutputUnitDTO{
		ModelID:      unit.ModelID,
		RegDate:      unit.RegDate,
		Text:         unit.Text,
		PropertiesID: propertiesID,
		ContextsID:   contextsID,
	}

	return unitDTO, nil
}

func (r *UnitsPgRepository) combineUnlinkedUnits(conn storage.ConnDB, layer string, unlinkedUnits entities.UnitsMap, combinedUnits interfaces.UnitDtoMaps) (interfaces.UnitDtoMaps, []int, error) {
	uniqueContextsID := make(map[int]bool)

	for lang := range unlinkedUnits {
		for _, unit := range unlinkedUnits[lang] {
			unitDTO, err := r.makeOutputUnitDTO(conn, layer, lang, unit)
			if err != nil {
				return nil, nil, err
			}

			for _, contextID := range unitDTO.ContextsID {
				uniqueContextsID[contextID] = true
			}

			combinedUnits = append(combinedUnits, map[string]interfaces.OutputUnitDTO{
				lang: unitDTO,
			})
		}
	}

	contextsID := make([]int, 0, len(uniqueContextsID))
	for id := range uniqueContextsID {
		contextsID = append(contextsID, id)
	}

	return combinedUnits, contextsID, nil
}

func (r *UnitsPgRepository) combineLinkedUnits(conn storage.ConnDB, layer string, linkedUnits joinedUnitMaps, combinedUnits interfaces.UnitDtoMaps) (interfaces.UnitDtoMaps, []int, error) {
	uniqueContextsID := make(map[int]bool)

	for _, unitMap := range linkedUnits {
		unitMapDTO := make(map[string]interfaces.OutputUnitDTO)
		for lang := range unitMap {
			unit := unitMap[lang]
			unitDTO, err := r.makeOutputUnitDTO(conn, layer, lang, unit)
			if err != nil {
				return nil, nil, err
			}
			unitMapDTO[lang] = unitDTO

			for _, contextID := range unitDTO.ContextsID {
				uniqueContextsID[contextID] = true
			}
		}

		combinedUnits = append(combinedUnits, unitMapDTO)
	}

	contextsID := make([]int, 0, len(uniqueContextsID))
	for id := range uniqueContextsID {
		contextsID = append(contextsID, id)
	}

	return combinedUnits, contextsID, nil
}

func (r *UnitsPgRepository) combineUnits(conn storage.ConnDB, layer string, linkedUnits joinedUnitMaps, unlinkedUnits entities.UnitsMap) (interfaces.UnitDtoMaps, []int, error) {
	unlinkedUnitsLen := 0
	for lang := range unlinkedUnits {
		unlinkedUnitsLen += len(unlinkedUnits[lang])
	}

	combinedUnits := make(interfaces.UnitDtoMaps, 0, unlinkedUnitsLen+len(linkedUnits))

	combinedUnits, contextsID1, err := r.combineUnlinkedUnits(conn, layer, unlinkedUnits, combinedUnits)
	if err != nil {
		return nil, nil, err
	}

	combinedUnits, contextsID2, err := r.combineLinkedUnits(conn, layer, linkedUnits, combinedUnits)
	if err != nil {
		return nil, nil, err
	}

	return combinedUnits, append(contextsID1, contextsID2...), nil
}

type unitQueries struct {
	unlinkedUnitsQuery string
	linkedUnitsQuery   string
	optionalArgs       map[string]any
}

func (r *UnitsPgRepository) getUnitsByQueries(conn storage.ConnDB, layer string, langs []string, qData unitQueries) (joinedUnitMaps, entities.UnitsMap, error) {
	args1 := map[string]any{
		"layer_name": layer,
		"lang":       "",
	}
	args2 := map[string]any{
		"layer_name": layer,
	}
	for argName, arg := range qData.optionalArgs {
		args1[argName] = arg
		args2[argName] = arg
	}

	unlinkedUnitsMap := make(entities.UnitsMap)
	for _, lang := range langs {
		args1["lang"] = lang
		unlinkedUnits, err := namedSelectSliceFromScript[[]entities.Unit](conn, qData.unlinkedUnitsQuery, args1)
		if err != nil {
			return nil, nil, err
		}
		unlinkedUnitsMap[lang] = unlinkedUnits
	}

	linkedUnits, err := namedSelectSliceFromScript[joinedUnits](conn, qData.linkedUnitsQuery, args2)
	if err != nil {
		return nil, nil, err
	}

	return linkedUnits.toMaps(), unlinkedUnitsMap, nil
}

func (r *UnitsPgRepository) GetAllUnits(conn storage.ConnDB, layer string) (interfaces.OutputUnitsDTO, error) {
	linkedUnits, unlinkedUnits, err := r.getUnitsByQueries(conn, layer, []string{"ru", "en"}, unitQueries{
		unlinkedUnitsQuery: selectUnlinkedUnitsByLang,
		linkedUnitsQuery:   selectAllLinkedUnits,
		optionalArgs:       nil,
	})
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	combinedUnits, contextsID, err := r.combineUnits(conn, layer, linkedUnits, unlinkedUnits)
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	contexts, err := selectSliceFromScript[[]interfaces.ContextDTO](conn, selectContextsById, pq.Array(contextsID))
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	return interfaces.OutputUnitsDTO{Units: combinedUnits, Contexts: contexts}, nil
}

func (r *UnitsPgRepository) GetUnitsByModels(conn storage.ConnDB, layer string, modelsID []int) (interfaces.OutputUnitsDTO, error) {
	linkedUnits, unlinkedUnits, err := r.getUnitsByQueries(conn, layer, []string{"ru", "en"}, unitQueries{
		unlinkedUnitsQuery: selectUnlinkedUnitsByLangAndModelsId,
		linkedUnitsQuery:   selectLinkedUnitsByModelsId,
		optionalArgs: map[string]any{
			"models_id_array": pq.Array(modelsID),
		},
	})
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	combinedUnits, contextsID, err := r.combineUnits(conn, layer, linkedUnits, unlinkedUnits)
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	contexts, err := selectSliceFromScript[[]interfaces.ContextDTO](conn, selectContextsById, pq.Array(contextsID))
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	return interfaces.OutputUnitsDTO{Units: combinedUnits, Contexts: contexts}, nil
}

func (r *UnitsPgRepository) GetUnitsByProperties(conn storage.ConnDB, layer string, propertiesID []int) (interfaces.OutputUnitsDTO, error) {
	linkedUnits, unlinkedUnits, err := r.getUnitsByQueries(conn, layer, []string{"ru", "en"}, unitQueries{
		unlinkedUnitsQuery: selectUnlinkedUnitsByLangAndPropertiesId,
		linkedUnitsQuery:   selectLinkedUnitsByPropertiesId,
		optionalArgs: map[string]any{
			"properties_id_array": pq.Array(propertiesID),
		},
	})
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	combinedUnits, contextsID, err := r.combineUnits(conn, layer, linkedUnits, unlinkedUnits)
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	contexts, err := selectSliceFromScript[[]interfaces.ContextDTO](conn, selectContextsById, pq.Array(contextsID))
	if err != nil {
		return interfaces.OutputUnitsDTO{}, err
	}

	return interfaces.OutputUnitsDTO{Units: combinedUnits, Contexts: contexts}, nil
}

func (r *UnitsPgRepository) SaveUnits(conn storage.ConnDB, layer string, data interfaces.SaveUnitsDTO) error {
	return errors.New("postgres storage does not support function SaveUnits") // TODO: implement me
}

func (r *UnitsPgRepository) RenameUnit(conn storage.ConnDB, layer string, oldName string, newName string) error {
	return errors.New("postgres storage does not support function RenameUnit") // TODO: implement me
}

func (r *UnitsPgRepository) SetUnitProperties(conn storage.ConnDB, layer string, unitName string, propertiesID []int) error {
	return errors.New("postgres storage does not support function SetUnitProperties") // TODO: implement me
}
