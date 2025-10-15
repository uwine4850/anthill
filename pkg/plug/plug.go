package plug

import "plugin"

func OpenPlugin(pluginPath string) (*plugin.Symbol, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, err
	}
	return &sym, nil
}
