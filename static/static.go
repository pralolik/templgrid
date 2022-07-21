package static

import "embed"

//go:embed emails/*
var emails embed.FS

//go:embed components/*
var components embed.FS

//go:embed i10n/*
var i10n embed.FS

func Emails() embed.FS {
	return emails
}

func Components() embed.FS {
	return components
}

func I10n() embed.FS {
	return i10n
}
