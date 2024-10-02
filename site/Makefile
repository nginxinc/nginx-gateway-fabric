HUGO?=hugo
HUGO_VERSION?=$(shell hugo version 2>/dev/null | awk '{print $$2}' | cut -d '.' -f 2)
HUGO_IMG?=hugomods/hugo:std-go-git-0.134.3

THEME_MODULE = github.com/nginxinc/nginx-hugo-theme

ifeq ($(shell [ $(HUGO_VERSION) -gt 133 2>/dev/null ] && echo true || echo false), true)
    $(info Hugo is available and has a version greater than 133. Proceeding with build.)
else
    $(warning Hugo is not available or using a version less than 134. Attempting to use docker. HUGO_VERSION=$(HUGO_VERSION))
    HUGO=docker run --rm -it -v ${CURDIR}:/src -p 1313:1313 ${HUGO_IMG} /src/hugo-entrypoint.sh
    ifeq (, $(shell docker version 2> /dev/null))
        $(error Hugo (>0.134) or Docker are required to build the local previews.)
    endif
endif

MARKDOWNLINT?=markdownlint
MARKDOWNLINT_IMG?=ghcr.io/igorshubovych/markdownlint-cli:latest

# if there's no local markdownlint, fallback to docker
ifeq (, $(shell ${MARKDOWNLINT} version 2> /dev/null))
ifeq (, $(shell docker version 2> /dev/null))
else
    MARKDOWNLINT=docker run --rm -i -v ${CURDIR}:/src --workdir /src ${MARKDOWNLINT_IMG}
endif
endif

MARKDOWNLINKCHECK?=markdown-link-check
MARKDOWNLINKCHECK_IMG?=ghcr.io/tcort/markdown-link-check:stable
# if there's no local markdown-link-check, fallback to docker
ifeq (, $(shell ${MARKDOWNLINKCHECK} --version 2> /dev/null))
ifeq (, $(shell docker version 2> /dev/null))
else
    MARKDOWNLINKCHECK=docker run --rm -it -v ${CURDIR}:/docs --workdir /docs ${MARKDOWNLINKCHECK_IMG}
endif
endif
endif


.PHONY: docs docs-draft docs-local clean hugo-get hugo-tidy lint-markdown link-check

docs:
	${HUGO}

clean:
	if [[ -d ${PWD}/public ]] ; then rm -rf ${PWD}/public && echo "Removed public directory" ; else echo "Did not find a public directory to remove" ; fi

watch:
	${HUGO} --bind 0.0.0.0 -p 1313 server --disableFastRender

watch-drafts:
	${HUGO} --bind 0.0.0.0 -p 1313 server -D --disableFastRender

clean:
	[ -d "public" ] && rm -rf "public"

hugo-get:
	hugo mod get -u github.com/nginxinc/nginx-hugo-theme

hugo-tidy:
	hugo mod tidy

hugo-update: hugo-get hugo-tidy

lint-markdown:
	${MARKDOWNLINT} -c .markdownlint.yaml  -- content

link-check:
	${MARKDOWNLINKCHECK} $(shell find content -name '*.md')

lint-markdown:
	${MARKDOWNLINT} -c .markdownlint.json  -- content

# Commands used by Netlify CI
hugo-mod:
	hugo mod get $(THEME_MODULE)@v$(THEME_VERSION)

build-production:
	hugo --gc -e production

build-staging:
	hugo --gc -e staging

build-dev:
	hugo --gc -e development

deploy-preview: hugo-mod
	hugo --gc -b ${NETLIFY_DEPLOY_URL}/nginx-gateway-fabric/
