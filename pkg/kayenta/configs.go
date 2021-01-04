package kayenta

//There is a decent change we may not need this entire module as we can embed the config in the analysis call

//UpsertCanaryConfigs adds additional logic around the Kayenta service since it does not allow for upserts
func UpsertCanaryConfigs(d *DefaultClient, application string, cc CanaryConfig) (string, error) {
	cc.Applications = []string{application}

	configs, err := d.GetCanaryConfigs(application)
	if err != nil {
		return "", err
	}

	if len(configs) == 0 {
		return d.CreateCanaryConfig(cc)
	}
	//TODO make sure this line is tested
	cc.Id = configs[0].Id
	return d.UpdateCanaryConfig(cc)
}
