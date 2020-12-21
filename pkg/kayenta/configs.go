package kayenta

//There is a decent change we may not need this entire module as we can embed the config in the analysis call

import (
	log "github.com/sirupsen/logrus"
)

//UpsertCanaryConfigs adds additional logic around the Kayenta service since it does not allow for upserts
func UpsertCanaryConfigs(d *DefaultClient, application string, cc CanaryConfig) (string, error) {
	cc.Applications = []string{application}

	configs, err := d.GetCanaryConfigs(application)
	if err != nil {
		return "", err
	}

	if len(configs) == 0 {
		log.Info("did not find existing config, creating one now")
		return d.CreateCanaryConfig(cc)
	}
	//TODO make sure this line is tested
	cc.Id = configs[0].Id
	log.Infof("found existing config with id: %s, updating canary configs", cc.Id)
	return d.UpdateCanaryConfig(cc)
}
