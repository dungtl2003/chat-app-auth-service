DOCKER_USERNAME ?= ilikeblue
DOCKER_FOLDER ?= ./docker
APPLICATION_NAME ?= chat-app-auth-service
GIT_HASH ?= $(shell git log --format="%h" -n 1) 
SERVER_PORT ?= 8020
ENV_FILE ?= .env

_BUILD_ARGS_TAG ?= ${GIT_HASH}
_BUILD_ARGS_RELEASE_TAG ?= latest
_BUILD_ARGS_DOCKERFILE ?= Dockerfile

clean:
	$(info ==================== cleaning project ====================)
	ENV_FILE=${ENV_FILE} ./mvnw_wrapper.sh spring-javaformat:apply

test: clean
	$(info ==================== running tests ====================)
	ENV_FILE=${ENV_FILE} ./mvnw_wrapper.sh clean verify

_builder: test
	$(info ==================== building dockerfile ====================)
	docker buildx build --platform linux/amd64 --tag ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_TAG} -f ${DOCKER_FOLDER}/${_BUILD_ARGS_DOCKERFILE} .

_pusher:
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_TAG}

_releaser:
	docker pull ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_TAG}
	docker tag  ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_TAG} ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_RELEASE_TAG}
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_RELEASE_TAG}

_server:
	docker container run --rm --env-file ${ENV_FILE} ${DOCKER_USERNAME}/${APPLICATION_NAME}:${_BUILD_ARGS_TAG}

run:
	ENV_FILE=${ENV_FILE} ./mvnw_wrapper.sh exec:java

server:
	$(MAKE) _server

server_%: 
	$(MAKE) _server \
		-e _BUILD_ARGS_TAG="$*-${GIT_HASH}" \
		-e _BUILD_ARGS_DOCKERFILE="Dockerfile.$*"

build:
	$(MAKE) _builder
 
push:
	$(MAKE) _pusher
 
release:
	$(MAKE) _releaser

build_%: 
	$(MAKE) _builder \
		-e _BUILD_ARGS_TAG="$*-${GIT_HASH}" \
		-e _BUILD_ARGS_DOCKERFILE="Dockerfile.$*"
 
push_%:
	$(MAKE) _pusher \
		-e _BUILD_ARGS_TAG="$*-${GIT_HASH}"
 
release_%:
	$(MAKE) _releaser \
		-e _BUILD_ARGS_TAG="$*-${GIT_HASH}" \
		-e _BUILD_ARGS_RELEASE_TAG="$*-latest"

.PHONY:
	clean test server server_% \
	build run push release build_% push_% release_%
