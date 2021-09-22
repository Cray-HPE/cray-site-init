NAME ?= hms-securestorage 
VERSION ?= $(shell cat .version)

all : unittest coverage

unittest: 
		docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .
		./runUnitTest.sh

coverage:
		./runCoverage.sh	
