package config

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	"io"

	"strings"

	"reflect"
)

//GetConfigurationService returns the ConfigurationService
func GetConfigurationService() ConfigurationService {
	return configurationImpl{}
}

type configurationImpl struct{}

func (c configurationImpl) Update(cfg Configuration) ([]UpdatedConfig, error) {
	mapSrc := make(map[string]interface{})
	mapDst := make(map[string]interface{})
	updatedConfig := make([]UpdatedConfig, 0)
	err := c.readFile(cfg.FilePath, &mapSrc)
	if err != nil {
		return nil, err
	}

	err = c.parseContent(strings.NewReader(cfg.Content), &mapDst)
	if err != nil {
		return nil, err
	}

	if cfg.PartialUpdate {
		updatedConfig, err = c.updateConfigurations(mapSrc, mapDst)
		if err != nil {
			return nil, err
		}
	} else {
		mapSrc = mapDst
	}

	err = c.writeFile(cfg.FilePath, &mapSrc)
	if err != nil {
		return nil, err
	}
	return updatedConfig, nil
}

func (c configurationImpl) updateConfigurations(mapSrc, mapDst map[string]interface{}) (updatedConf []UpdatedConfig, err error) {
	updatedConf = make([]UpdatedConfig, 0)
	for k, v := range mapDst {
		srcValue, srcHasKey := mapSrc[k]
		if !srcHasKey {
			mapSrc[k] = v
		} else {
			switch srcValue.(type) { //nolint:gocritic
			case map[string]interface{}:
				switch v.(type) { //nolint:gocritic
				case map[string]interface{}:
					prmSrc, _ := srcValue.(map[string]interface{})
					prmDst, _ := v.(map[string]interface{})
					uConf, err := c.updateConfigurations(prmSrc, prmDst)
					if err != nil {
						return updatedConf, errors.New(err.Error() + ":" + k)
					}
					if len(uConf) > 0 {
						u := UpdatedConfig{
							Key:     k,
							Updated: uConf,
						}
						updatedConf = append(updatedConf, u)
					}
				default:
					return updatedConf, errors.New("TypeMismatch: " + k)
				}
			default:
				if !reflect.DeepEqual(srcValue, v) {
					u := UpdatedConfig{
						Key:      k,
						Existing: srcValue,
						Updated:  v,
					}
					updatedConf = append(updatedConf, u)
					mapSrc[k] = v
				}
			}
		}
	}
	sort.Sort(byKey(updatedConf)) // Ensure deterministic results. Helpful for testing
	return updatedConf, nil
}

func (c configurationImpl) readFile(filePath string, maps *map[string]interface{}) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck
	return c.parseContent(file, maps)
}

func (c configurationImpl) parseContent(reader io.Reader, maps *map[string]interface{}) (err error) {
	deser := json.NewDecoder(reader)
	err = deser.Decode(maps)
	if err != nil {
		return
	}
	return
}

func (c configurationImpl) writeFile(filePath string, maps *map[string]interface{}) (err error) {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck

	ser := json.NewEncoder(file)
	ser.SetEscapeHTML(false)
	ser.SetIndent("", "\t")
	err = ser.Encode(maps)
	if err != nil {
		return
	}

	return
}

// For sorting an array of updated configs by their key
type byKey []UpdatedConfig

func (updates byKey) Len() int {
	return len(updates)
}
func (updates byKey) Swap(i, j int) {
	updates[i], updates[j] = updates[j], updates[i]
}
func (updates byKey) Less(i, j int) bool {
	return updates[i].Key < updates[j].Key
}
