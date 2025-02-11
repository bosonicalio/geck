package validation

// ValidatorConfig is the configuration structure for a [Validator] instance.
type ValidatorConfig struct {
	StructFieldDriver StructFieldDriver `env:"VALIDATOR_STRUCT_SCANNER_DRIVER" envDefault:"json"`
}
