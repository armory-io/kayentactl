package kayenta

import "io"

//UpsertCanaryConfigs adds additional logic around the Kayenta service since it does not allow for upserts
func UpsertCanaryConfigs(d *DefaultClient, application string, canaryConfig io.Reader) error {
	configs, err := d.GetCanaryConfigs(application)
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		d.CreateCanaryConfig(canaryConfig)
	}
	return nil
}
