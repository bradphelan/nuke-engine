module demo/simulation

go 1.26.1

require github.com/bradphelan/nuke-engine v0.0.0

require github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect

replace (
	demo/mathcore => ../mathcore
	demo/physics => ../physics
	demo/renderer => ../renderer
	github.com/bradphelan/nuke-engine => ../../..
)
