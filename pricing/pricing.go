package pricing

// PricingTable maps memory size (in MB) to price per 1ms in dollars.
type PricingTable map[int]float64

// GetPricingTable returns the pricing table for the given architecture.
// For this example, we use the Ireland region pricing for ARM and the provided x86 pricing.
func GetPricingTable(arch string) PricingTable {
	if arch == "arm64" {
		return PricingTable{
			128:   0.0000000017,
			512:   0.0000000067,
			1024:  0.0000000133,
			1536:  0.0000000200,
			2048:  0.0000000267,
			3072:  0.0000000400,
			4096:  0.0000000533,
			5120:  0.0000000667,
			6144:  0.0000000800,
			7168:  0.0000000933,
			8192:  0.0000001067,
			9216:  0.0000001200,
			10240: 0.0000001333,
		}
	}
	// Default to x86 pricing
	return PricingTable{
		128:   0.0000000021,
		512:   0.0000000083,
		1024:  0.0000000167,
		1536:  0.0000000250,
		2048:  0.0000000333,
		3072:  0.0000000500,
		4096:  0.0000000667,
		5120:  0.0000000833,
		6144:  0.0000001000,
		7168:  0.0000001167,
		8192:  0.0000001333,
		9216:  0.0000001500,
		10240: 0.0000001667,
	}
}
