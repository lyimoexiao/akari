package web

import "embed"

//go:embed dist/*
//go:embed dist/assets/*
var FS embed.FS
