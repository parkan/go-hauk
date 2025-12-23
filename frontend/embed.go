package frontend

import "embed"

//go:embed index.html main.js style.css lib assets
var Files embed.FS
