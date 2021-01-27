package resticplugin

var (
	_ dbplugin.Database = &resticRepo{}
)
