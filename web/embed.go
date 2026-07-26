package web

import "embed"

//go:embed dist/*
// dist/assets/_* covers underscore-prefixed files (e.g. _plugin-vue_export-helper)
// which go:embed's wildcard pattern excludes by default.
//go:embed dist/assets/_*
var FS embed.FS
