# NGINX Gateway Fabric Docs

This directory contains the user documentation for NGINX Gateway Fabric and the requirements for linting, building, and publishing the docs.

We use [Hugo](https://gohugo.io/) to build the docs for NGINX, with the [nginx-hugo-theme](https://github.com/nginxinc/nginx-hugo-theme).

Docs should be written in Markdown.

In this directory, you will find the following files:

- a [Netlify](https://netlify.com) configuration file;
- configuration files for [markdownlint](https://github.com/DavidAnson/markdownlint/) and [markdown-link-check](https://github.com/tcort/markdown-link-check)
- a `./config` directory that contains the [Hugo](https://gohugo.io) configuration.

## Git Guidelines

- Keep a clean, concise and meaningful git commit history on your branch (within reason), rebasing locally and squashing before submitting a PR.
- Use the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format when writing a commit message, so that changelogs can be automatically generated
- Follow the guidelines of writing a good commit message as described here <https://chris.beams.io/posts/git-commit/> and summarised in the next few points:
  - In the subject line, use the present tense ("Add feature" not "Added feature").
  - In the subject line, use the imperative mood ("Move cursor to..." not "Moves cursor to...").
  - Limit the subject line to 72 characters or less.
  - Reference issues and pull requests liberally after the subject line.
  - Add more detailed description in the body of the git message (`git commit -a` to give you more space and time in your text editor to write a good message instead of `git commit -am`).

### Forking and Pull Requests

This repo uses a [forking workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/forking-workflow). Take the steps below to fork the repo, check out a feature branch, and open a pull request with your changes.

1. In the GitHub UI, select the **Fork** button.
   
    - On the **Create a new fork** page, select the **Owner** (the account where the fork of the repo will be placed).
    - Select the **Create fork** button.

2. If you plan to work on docs in your local development environment, clone your fork. 
   For example, to clone the repo using SSH, you would run the following command:
    
    ```shell
    git clone git@github.com:<your-account>/nginx-gateway-fabric.git
    ```

3. Check out a new feature branch in your fork. This is where you will work on your docs. 

   To do this via the command line, you would run the following command:

    ```shell
    git checkout -b <branch-name>
    ```

    **CAUTION**: Do not work on the main branch in your fork. This can cause issues when the NGINX Docs team needs to check out your feature branch for editing work.

4. Make atomic, [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) on your feature branch. 

5. When ready, open a pull request into the **main** branch in the **nginxinc/nginx-gateway-fabric** repo.
    
    - Fill in [our pull request template](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/.github/PULL_REQUEST_TEMPLATE.md) when opening your PR.
    - Tag the appropriate reviewers for your subject area.  
      Technical reviewers should be able to verify that the information provided is accurate.  
      Documentation reviewers ensure that the content conforms to the NGINX Style Guide, is grammatically correct, and adheres to the NGINX content templates. 

## Release Management and Publishing

**`main`** is the default branch in this repo. All the latest content updates are merged into this branch. 

The documentation is published from the latest public release branch, (for example, `release-4.0`). Work on your docs in feature branches off of the main branch. Open pull requests into the `main` when you are ready to merge your work.

If you are working on content for immediate publication in the docs site, cherrypick your changes to the current public release branch.

If you are working on content for a future release, make sure that you **do not** cherrypick them to the current public release branch, as this will publish them automatically.


## Setup

### Golang

Follow the instructions here to install Go: https://golang.org/doc/install

> To support the use of Hugo mods, you need to install Go v1.15 or newer.

### Hugo

Follow the instructions here to install Hugo: [Hugo Installation](https://gohugo.io/installation/)

> **NOTE:** We are currently running [Hugo v0.115.3](https://github.com/gohugoio/hugo/releases/tag/v0.115.3) in production.

### Markdownlint

We use markdownlint to check that Markdown files are correctly formatted. You can use `npm` to install markdownlint-cli:

```
npm install -g markdownlint-cli   
```

## How to write docs with Hugo

### Add a new doc

- To create a new doc that contains all of the pre-configured Hugo front-matter and the docs task template:

    `hugo new <SECTIONNAME>/<FILENAME>.<FORMAT>`

  e.g.,

    hugo new install.md

  > The default template -- task -- should be used in most docs.
- To create other types of docs, you can add the `--kind` flag:
    `hugo new tutorials/deploy.md --kind tutorial`


The available kinds are:

- Task: Enable the customer to achieve a specific goal, based on use case scenarios.
- Concept: Help a customer learn about a specific feature or feature set.
- Reference: Describes an API, command line tool, config options, etc.; should be generated automatically from source code. 
- Troubleshooting: Helps a customer solve a specific problem.
- Tutorial: Walk a customer through an example use case scenario; results in a functional PoC environment.

### Format internal links

Format links as [Hugo relrefs](https://gohugo.io/content-management/cross-references/). 

> Note: Using file extensions when linking to internal docs with `relref` is optional.
  
- You can use relative paths or just the filename. 
- Paths without a leading `/` are first resolved relative to the current page, then to the remainder of the site.
- Anchors are supported.

For example:

```md
To install NGINX Controller, refer to the [installation instructions]({{< ref "install" >}}).
```

### Add images

You can use the `img` [shortcode](#shortcodes) to insert images into your documentation.

1. Add the image to the static/img directory, or to the same directory as the doc you want to use it in.
   DO NOT include a forward slash at the beginning of the file path. This will break the image when it's rendered.
   See the docs for the [Hugo relURL Function](https://gohugo.io/functions/relurl/#input-begins-with-a-slash) to learn more.

1. Add the img shortcode:

    {{< img src="<img-file.png>" >}}
 
> Note: The shortcode accepts all of the same parameters as the [Hugo figure shortcode](https://gohugo.io/content-management/shortcodes/#figure). 

### Use Hugo shortcodes
You can use Hugo [shortcodes](https://gohugo.io/content-management/shortcodes) to do things like format callouts, add images, and reuse content across different docs. 

For example, to use the note callout:

```md
{{< note >}}Provide the text of the note here. {{< /note >}}
``` 

The callout shortcodes also support multi-line blocks:

```md
{{< caution >}}
You should probably never do this specific thing in a production environment. If you do, and things break, don't say we didn't warn you.
{{< /caution >}}
```

Supported callouts:
- caution
- important
- note
- see-also
- tip
- warning

A few more useful shortcodes:

- collapse: makes a section collapsible
- table: adds scrollbars to wide tables when viewed in small browser windows or mobile browsers
- fa: inserts a Font Awesome icon
- include: include the content of a file in another file (requires the included file to be in the /includes directory)
- link: makes it possible to link to a static file and prepend the path with the Hugo baseUrl
- openapi: loads an OpenAPI spec and renders as HTML using ReDoc
- raw-html: makes it possible to include a block of raw HTML
- readfile: includes the content of another file in the current file; useful for adding code examples

## How to build docs locally

To view the docs in a browser, run the Hugo server. This will reload the docs automatically so you can view updates as you work.

> Note: The docs use build environments to control the baseURL that will be used for things like internal references and resource (CSS and JS) loading. 
> You can view the config for each environment in the [config](./config) directory of this repo.
When running the Hugo server, you can specify the environment and baseURL if desired, but it's not necessary.

For example:

```
hugo server
```

```
hugo server -e development -b "http://127.0.0.1/nginx-gateway-fabric/"
```
