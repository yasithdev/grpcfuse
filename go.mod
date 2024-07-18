module main

go 1.22.5

replace grpcfs => ./grpcfs

require (
	github.com/jacobsa/fuse v0.0.0-20240626143436-8a36813dc074
	grpcfs v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.22.0 // indirect
