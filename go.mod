module github.com/openconfig/ondatra

go 1.17

require (
	github.com/golang/glog v1.0.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/kne v0.1.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/openconfig/gnmi v0.0.0-20210914185457-51254b657b7d
	github.com/openconfig/gnoi v0.0.0-20211102203610-1ece8ed91a0d
	github.com/openconfig/goyang v0.3.1
	github.com/openconfig/gribigo v0.0.0-20211117144123-9bcf5960a05d
	github.com/openconfig/ygot v0.13.1
	github.com/p4lang/p4runtime v1.3.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/openconfig/grpctunnel v0.0.0-20210610163803-fde4a9dc048d // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211129164237-f09f9a12af12 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace github.com/openconfig/gnmi => github.com/wenovus/gnmi v0.0.0-20220114231359-cfbc69541310

replace github.com/openconfig/gribigo => github.com/wenovus/gribigo v0.0.0-20220114232223-ee81f9935622

replace github.com/google/kne => /usr/local/google/home/wenbli/gocode/src/github.com/wenovus/kne
