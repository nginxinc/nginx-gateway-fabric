HUGO?=hugo
# the officially recommended unofficial docker image
HUGO_IMG?=klakegg/hugo:0.115.3
# NGINX Hugo theme module location
THEME_MODULE=github.com/nginxinc/nginx-hugo-theme
## Pulls the current theme version from the Netlify settings
THEME_VERSION=$(NGINX_THEME_VERSION)


# if there's no local hugo, fallback to docker
ifeq (, $(shell ${HUGO} version 2> /dev/null))
ifeq (, $(shell docker version 2> /dev/null))
    $(error Docker and Hugo are not installed. Hugo (<0.91) or Docker are required to build the local preview.)
else
    HUGO=docker run --rm -it -v ${CURDIR}:/src -p 1313:1313 ${HUGO_IMG}
endif
endif

MARKDOWNLINT?=markdownlint
MARKDOWNLINT_IMG?=ghcr.io/igorshubovych/markdownlint-cli:latest

# if there's no local markdownlint, fallback to docker
ifeq (, $(shell ${MARKDOWNLINT} version 2> /dev/null))
ifeq (, $(shell docker version 2> /dev/null))
ifneq (, $(shell $(NETLIFY) "true"))
    $(error Docker and markdownlint are not installed. markdownlint or Docker are required to lint.)
else
    MARKDOWNLINT=docker run --rm -i -v ${CURDIR}:/src --workdir /src ${MARKDOWNLINT_IMG}
endif
endif
endif

MARKDOWNLINKCHECK?=markdown-link-check
MARKDOWNLINKCHECK_IMG?=ghcr.io/tcort/markdown-link-check:stable
# if there's no local markdown-link-check, fallback to docker
ifeq (, $(shell ${MARKDOWNLINKCHECK} --version 2> /dev/null))
ifeq (, $(shell docker version 2> /dev/null))
ifneq (, $(shell $(NETLIFY) "true"))
    $(error Docker and markdown-link-check are not installed. markdown-link-check or Docker are required to check links.)
else
    MARKDOWNLINKCHECK=docker run --rm -it -v ${CURDIR}:/hugo --workdir /hugo ${MARKDOWNLINKCHECK_IMG}
endif
endif
endif


.PHONY: docs clean hugo-mod docs-local docs-drafts netlify lint-grammar lint-markdown link-check all all-staging all-dev

## For use in Netlify CI only
all: hugo-mod build-production

all-staging: hugo-mod build-staging

all-dev: hugo-mod build-dev
## end for use in Netlify CI

docs: 
	${HUGO} server --disableFastRender

clean:
	rm -rf ./public

docs-local: clean 
	${HUGO}

docs-drafts:
	${HUGO} server -D --disableFastRender

lint-grammar:
	docker run --rm -it -v ${CURDIR}/content:/root/content --entrypoint "vale" --workdir /root/content artifactory.f5net.com/cylon-indigo-docker-dev/indigo/tools/docs/vale-lint:0.7.0 ./

lint-markdown:
	${MARKDOWNLINT} -c mdlint_conf.json  -- content

link-check:
	${MARKDOWNLINKCHECK} $(shell find content -name '*.md')

## commands for use in Netlify CI
hugo-mod:
	hugo mod get $(THEME_MODULE)@v$(THEME_VERSION)

build-production: hugo-mod
	hugo --gc -e production

build-staging: hugo-mod
	hugo --gc -e staging

build-dev: hugo-mod
	hugo --gc -e development

deploy-preview: hugo-mod
	hugo --gc -d public/nginx-gateway-fabric/ -b ${NETLIFY_DEPLOY_URL}/nginx-gateway-fabric/