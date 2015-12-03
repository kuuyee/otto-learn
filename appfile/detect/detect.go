package detect

// App 会根据给定的目录发现application类型
func App(dir string, c *Config) (string, error) {
	for _, d := range c.Detectors {
		check, err := d.Detect(dir)
		if err != nil {
			return "", err
		}
		if check {
			return d.Type, nil
		}
	}
	return "", nil
}
