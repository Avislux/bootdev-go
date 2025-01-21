module example.com/jt/hellogo

go 1.23.4

replace example.com/jt/mystrings v0.0.0 => ../mystrings

require (
	example.com/jt/mystrings v0.0.0
)
