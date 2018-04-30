@echo off

@go get github.com/mitchellh/gox
gox -osarch="linux/amd64"
del admin-service
ren admin-service_linux_amd64 admin-service
