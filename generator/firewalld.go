package generator

import (
	"fmt"
	"github.com/HouzuoGuo/HANA-Firewall/model"
	"io/ioutil"
	"os"
	"path"
)

// Firewalld takes input from existing service configuration to install HANA firewall configuration.
type Firewalld struct {
	// HANAGlobal is the global configuration of HANA services.
	HANAGlobal model.HANAGlobalParameters
	// HANAServiceDefinition has association between short name of HANA services and their definitions.
	HANAServices []model.HANAServiceDefinition
}

// GenerateConfig takes HANA configuration as input returns generated XML file paths vs firewalld service definition.
func (fw *Firewalld) GenerateConfig() (ret map[string]model.FirewalldService, err error) {
	ret = make(map[string]model.FirewalldService)
	for _, def := range fw.HANAServices {
		shortName, svc, err := fw.HANAGlobal.MakeFirewalldService(&def)
		if err != nil {
			return nil, err
		}
		ret[shortName] = svc
	}
	return
}

// WriteConfig serialises firewalld service definition into XML files and place them under the directory.
func (fw *Firewalld) WriteConfig(destDir string, services map[string]model.FirewalldService) error {
	if info, err := os.Stat(destDir); err != nil || !info.IsDir() {
		return fmt.Errorf("Firewalld.WriteConfig: destination directory \"%s\" does not exist or it is not a directory", destDir)
	}
	for shortName, svc := range services {
		filePath := path.Join(destDir, shortName+".xml")
		if err := ioutil.WriteFile(filePath, []byte(svc.ToXML()), 0600); err != nil {
			return err
		}
	}
	return nil
}
