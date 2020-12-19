package kayenta

func UpsertCanaryConfigs(d *DefaultClient, application string) error {
	d.GetCanaryConfigs()
	return nil
}
