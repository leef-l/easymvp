package engine

import providerutil "easymvp/utility/provider"

func decodeProviderProtocols(raw string, providerType string) []string {
	return providerutil.DecodeSupportedProtocols(raw, providerType)
}
