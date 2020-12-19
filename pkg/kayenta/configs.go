package kayenta

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

//UpsertCanaryConfigs adds additional logic around the Kayenta service since it does not allow for upserts
func UpsertCanaryConfigs(d *DefaultClient, application string, canaryConfig io.Reader) (string, error) {
	if application == "" {
		return "", fmt.Errorf("Application name cannot be empty")
	}
	configs, err := d.GetCanaryConfigs(application)
	if err != nil {
		return "", err
	}

	if len(configs) == 0 {
		log.Info("did not find existing config, creating one now")
		return d.CreateCanaryConfig(canaryConfig)
	}
	log.Infof("found existing config with id: %v,  updating canary configs", configs[0].Id)
	return d.UpdateCanaryConfig(configs[0].Id, canaryConfig)
}
