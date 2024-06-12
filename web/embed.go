package web

import "embed"

//go:embed app/** static/**
var Embed embed.FS
