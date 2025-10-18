package plug

import (
	"path/filepath"
	"plugin"
)

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

var builtinList = []string{
	"pkg/plug/plugins/builtin_list/cmd.so",
}

func BuiltinList() ([]string, error) {
	list := make([]string, len(builtinList))
	for i := 0; i < len(builtinList); i++ {
		absPath, err := filepath.Abs(builtinList[i])
		if err != nil {
			return nil, err
		}
		list[i] = absPath
	}
	return list, nil
}
