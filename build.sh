protoc --proto_path ../../../ -I=./proto --go_out=plugins=grpc:./proto proto/datastore.proto
mv proto/github.com/brotherlogic/datastore/proto/* ./proto
