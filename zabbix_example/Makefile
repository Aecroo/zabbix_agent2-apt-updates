.ONESHELL:

PACKAGE=zabbix-agent2-plugin-example
TOPDIR := $(CURDIR)
SHELL := /bin/bash

ifeq ($(OS),Windows_NT)
GOOS := windows
SHELL := cmd
TOPDIR := $(subst /,\,$(CURDIR))
PACKAGE:=$(PACKAGE).exe
WINDRES = windres.exe
ifneq ("$(shell findstr ZABBIX_RC_NUM $(TOPDIR)\windres\resource.h)","")
ifeq ("$(WINDRES_FLAGS)","")
WINDRES_FLAGS := \
	-D ZABBIX_LICENSE_YEARS='\"$(word 4,$(shell findstr Copyright $(TOPDIR)\main.go | findstr 2001-20))\"' \
	-D ZABBIX_VERSION_MAJOR=$(lastword $(shell findstr VERSION_MAJOR $(TOPDIR)\main.go | findstr =)) \
	-D ZABBIX_VERSION_MINOR=$(lastword $(shell findstr VERSION_MINOR $(TOPDIR)\main.go | findstr =)) \
	-D ZABBIX_VERSION_PATCH=$(lastword $(subst //nolint:revive,,$(shell findstr VERSION_PATCH $(TOPDIR)\main.go | findstr =))) \
	-D ZABBIX_VERSION_RC='\"$(lastword $(subst //nolint:revive,,$(shell findstr VERSION_RC $(TOPDIR)\main.go | findstr =)))\"' \
	-D ZABBIX_VERSION_RC_NUM=1000
endif
endif

RFLAGS := $(RFLAGS) --input-format=rc -O coff

ifeq ("$(ARCH)", "")
ifdef PROCESSOR_ARCHITECTURE
	ARCH := $(PROCESSOR_ARCHITECTURE)
else
	ARCH := x86
endif
endif

ifeq ($(ARCH), x86)
	RFLAGS := $(RFLAGS) --target=pe-i386
else ifeq ($(ARCH), AMD64)
	RFLAGS := $(RFLAGS) --target=pe-x86-64
else ifeq (,$(findstring ARM,$(ARCH)))
ifneq ($(ARCH), $(PROCESSOR_ARCHITECTURE))
$(error Unsupported CPU architecture: $(ARCH))
endif
endif
endif

ifeq ($(ARCH), x86)
	GOARCH := 386
else ifeq ($(ARCH), AMD64)
	GOARCH := amd64
else ifeq ($(ARCH), ARM)
	GOARCH := arm
else ifeq ($(ARCH), ARM64)
	GOARCH := arm64
endif

ifndef GOOS
GOOS := $(shell go env GOOS)
endif

ifndef GOARCH
GOARCH := $(shell go env GOARCH)
endif

DISTFILES = \
	ChangeLog \
	go.mod \
	go.sum \
	LICENSE \
	main.go \
	Makefile \
	example.conf \
	README.md

DIST_SUBDIRS = \
	plugin \
	windres \
	vendor

.build_rc:
ifneq ("$(WINDRES)","")
	$(WINDRES) $(TOPDIR)\windres\resource.rc $(WINDRES_FLAGS) $(RFLAGS) \
		-D VER_FILEDESCRIPTION_STR='\"$(PACKAGE)\"' \
		-D _WINDOWS -o "$(TOPDIR)\$(PACKAGE).syso"
endif

build: .build_rc
ifeq ($(OS),Windows_NT)
	set GOOS=$(GOOS)
	set GOARCH=$(GOARCH)
	go build -o "$(TOPDIR)/$(PACKAGE)"
else
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -o "$(TOPDIR)/$(PACKAGE)"
endif

clean:
ifeq ($(OS),Windows_NT)
	if exist "$(TOPDIR)\vendor" rmdir /S /Q "$(TOPDIR)\vendor"
	del /F "$(TOPDIR)\$(PACKAGE)*"
else
	rm -rf "$(TOPDIR)/vendor"
	rm -rf "$(TOPDIR)/$(PACKAGE)"*
endif
	go clean "$(TOPDIR)/..."

check:
	go test -v "$(TOPDIR)/..."

style:
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) "$(TOPDIR)/..."

format:
	go fmt "$(TOPDIR)/..."

dist:
ifneq ($(OS),Windows_NT)
	cd $(TOPDIR); \
	go mod vendor; \
	[[ "$$(head -1 ChangeLog)" =~ ^Changes[[:space:]]for[[:space:]]([0-9]+)\.([0-9]+)\.([0-9]+)((alpha|beta|rc)([0-9]+))? ]]; \
	major_verison=$${BASH_REMATCH[1]}; \
	minor_verison=$${BASH_REMATCH[2]}; \
	patch_verison=$${BASH_REMATCH[3]}; \
	alphatag=$${BASH_REMATCH[4]}; \
	lic_years=$(word 4, $(shell grep ' Copyright (C) 2001-' ./main.go)); \
	distdir="$(PACKAGE)-$${major_verison}.$${minor_verison}.$${patch_verison}$${alphatag}"; \
	dist_archive="$${distdir}.tar.gz"; \
	mkdir -p ./$${distdir}; \
	for distfile in '$(DISTFILES)'; do \
		cp -fp ./$${distfile} ./$${distdir}/; \
	done; \
	for subdir in '$(DIST_SUBDIRS)'; do \
		cp -fpR ./$${subdir} ./$${distdir}; \
	done; \
# File revision number must be numeric (Git commit hash cannot be used).
# Therefore to make it numeric and meaningful it is artificially composed from:
#    - branch (development or release),
#    - type (alpha, beta, rc or release),
#    - number of alpha, beta or rc.
# 'branch' expression tries to find out is it a development branch or release branch.
#      Result is encoded as: 1 - dev branch, release branch or error occurred, 2 - tag.
# 'type_name' expression tries to find out what type of release it is.
#      Expected result is: "alpha", "beta", "rc" or "" (empty string).
# 'type_num' expression encodes 'type_name' as numeric value:
#      1 - alpha, 2 - beta, 3 - rc, 4 - release, 0 - unknown.
# 'type_count' expression tries to find out number of "alpha", "beta" or "rc" (e.g. 1 from "rc1").
	branch=`(git symbolic-ref -q HEAD > /dev/null && echo 1) || (git tag -l --points-at HEAD| grep "."| grep -q -v "-" && echo 2) || echo 1`; \
	type_name=`cat ./main.go| sed -n -e '/AGENT_VERSION_RC/s/.*"\([a-z]*\)[0-9]*"/\1/p'`; \
	type_num=`(test "x$$type_name" = "xalpha" && echo "1") || echo ""`; \
	type_num=`(test -z $$type_num && test "x$$type_name" = "xbeta" && echo "2") || echo "$$type_num"`; \
	type_num=`(test -z $$type_num && test "x$$type_name" = "xrc" && echo "3") || echo "$$type_num"`; \
	type_num=`(test -z $$type_num && test -z $$type_name && echo "4") || echo "$$type_num"`; \
	type_num=`(test -z $$type_num && echo "0") || echo "$$type_num"`; \
	type_count=`cat ./main.go|sed -n -e '/ZABBIX_VERSION_RC/s/.*"[a-z]*\([0-9]*\)"/\1/p'`; \
	type_count=`printf '%02d' $$type_count`; \
	cat ./$${distdir}/windres/resource.h|sed "s/{ZABBIX_VERSION_MAJOR}/$${major_verison}/g"| \
	sed "s/{ZABBIX_VERSION_MINOR}/$${minor_verison}/g"| sed "s/{ZABBIX_VERSION_PATCH}/$${patch_verison}/g"| \
	sed "s/{ZABBIX_VERSION_RC}/\"$${alphatag}\"/g"| sed "s/{ZABBIX_RC_NUM}/$$branch$$type_num$$type_count/g"| \
	sed "s/{ZABBIX_LICENSE_YEARS}/\"$${lic_years}\"/g"| sed "s/{VER_FILEDESCRIPTION_STR}/$(PACKAGE)/g" \
	> ./$${distdir}/windres/resource.h.new; mv ./$${distdir}/windres/resource.h.new ./$${distdir}/windres/resource.h; \
	tar -czvf ./$${dist_archive} ./$${distdir}; \
	rm -rf ./$${distdir}
endif

sbom.json:
	CGO_CFLAGS="${CGO_CFLAGS}" CGO_LDFLAGS="${CGO_LDFLAGS}" cyclonedx-gomod mod \
		   -output-version 1.4 \
		   -licenses -assert-licenses -json -output "$@"

sbom.xml:
	CGO_CFLAGS="${CGO_CFLAGS}" CGO_LDFLAGS="${CGO_LDFLAGS}" cyclonedx-gomod mod \
		   -output-version 1.4 \
		   -licenses -assert-licenses -output "$@"

sbom: sbom.json

.PHONY: sbom
