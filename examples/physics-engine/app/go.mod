module demo/app

go 1.26.1

require (
	demo/mathcore v0.0.0
	demo/physics v0.0.0
	demo/renderer v0.0.0
	demo/simulation v0.0.0
	github.com/bradphelan/nuke-engine v0.0.0
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace (
	demo/mathcore => ../mathcore
	demo/physics => ../physics
	demo/renderer => ../renderer
	demo/simulation => ../simulation
	github.com/bradphelan/nuke-engine => ../../..
)
