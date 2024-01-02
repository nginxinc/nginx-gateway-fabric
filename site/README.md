# NGINX Gateway Fabric Documentation

This directory contains all of the user documentation for NGINX Gateway Fabric, as well as the requirements for building and publishing the documentation.

Documentation is written in Markdown, built using [Hugo](https://gohugo.io) and deployed with [Netlify](https://www.netlify.com/).

## Setup

Hugo is the only requirement for building documentation.

To install Hugo locally, follow the [official Hugo instructions](https://gohugo.io/getting-started/installing/).

> **Note**: We are currently running [Hugo v0.115.3](https://github.com/gohugoio/hugo/releases/tag/v0.115.3) in production.

If you have [Docker](https://www.docker.com/get-started/) installed, there is a fallback in the [Makefile](Makefile) which means you do need to install Hugo locally.

## Developing documentation locally

To build the docs locally, run the `make` command inside this `/docs` directory:

```text
make clean          -   Removes the local `public` directory, which is the default output path used by Hugo
make docs           -   Start a local Hugo server for live previews while editing
make docs-drafts    -   Start a local Hugo server for live previews, including documentation marked with `draft: true`
make hugo-mod       -   Cleans the Hugo module cache and fetches the latest version of the theme
make hugo-mod-tidy  -   Removes unused entries in go.mod and go.sum, then verifies the dependencies
```

## Adding new documentation

### Using Hugo to generate a new documentation file

To create a new documentation file with the pre-configured Hugo front-matter for the task template, run the following command inside this `/docs` directory:

`hugo new <SECTIONNAME>/<FILENAME>.<FORMAT>`

For example:

```shell
hugo new getting-started/install.md
```

The default template (task) should be used for most pages. For other content templates, you can use the `--kind` flag:

```shell
hugo new tutorials/deploy.md --kind tutorial
```

The available content templates (`kind`) are:

- concept: Help a user learn about a specific feature or feature set.
- tutorial: Walk a user through an example use case scenario.
- reference: Describes an API, command line tool, configuration options, etc.
- troubleshooting: Guide a user towards solving a specific problem.
- openapi: A template with the requirements to render an openapi.yaml specification.

## Documentation formatting

### Basic markdown formatting

There are multiple ways to format text: for consistency and clarity, these are our conventions:

- Bold: Two asterisks on each side - `**Bolded text**`.
- Italic: One underscore on each side - `_Italicized text_`.
- Unordered lists: One dash - `- Unordered list item`.
- Ordered lists: The 1 character followed by a stop - `1. Ordered list item`.

> **Note**: The ordered notation automatically enumerates lists when built by Hugo.

Close every section with a horizontal line by using three dashes: `---`.

### How to format internal links

Internal links should use Hugo [ref and relref shortcodes](https://gohugo.io/content-management/cross-references/).

- Although file extensions are optional for Hugo, we include them as best practice for page anchors.
- Relative paths are preferred, but just the filename is permissible.
- Paths without a leading forward slash (`/`) are first resolved relative to the current page, then the remainder of the website.

Here are two examples:

```md
To install <software>, refer to the [installation instructions]({{< ref "install.md" >}}).
To install <integation>, refer to the [integration instructions]({{< relref "/integration/thing.md#section" >}}).
```

### How to add images

Use the `img` [shortcode](#using-hugo-shortcodes) to add images into your documentation.

1. Add the image to the `/static/img` directory.
1. Add the `img` shortcode:
    `{{< img src="<img-file.png>" >}}`
   - **Do not include a forward slash at the beginning of the file path.**
   - This will break the image when it's rendered: read about the  [Hugo relURL Function](https://gohugo.io/functions/relurl/#input-begins-with-a-slash) to learn more.


> **Note**: The `img` shortcode accepts all of the same parameters as the Hugo [figure shortcode](https://gohugo.io/content-management/shortcodes/#figure).

### Using Hugo shortcodes

[Hugo shortcodes](/docs/themes/f5-hugo/layouts/shortcodes/) are used to format callouts, add images, and reuse content across different docs.

For example, to use the `note` callout:

```md
{{< note >}}Provide the text of the note here.{{< /note >}}
```

The callout shortcodes support multi-line blocks:

```md
{{< caution >}}
You should probably never do this specific thing in a production environment.

If you do, and things break, don't say we didn't warn you.
{{< /caution >}}
```

Supported callouts:

- `caution`
- `important`
- `note`
- `see-also`
- `tip`
- `warning`

Here are some other shortcodes:

- `fa`: inserts a Font Awesome icon
- `img`: include an image and define things like alt text and dimensions
- `include`: include the content of a file in another file; the included file must be present in the content/includes directory
- `link`: makes it possible to link to a file and prepend the path with the Hugo baseUrl
- `openapi`: loads an OpenAPI spec and renders as HTML using ReDoc
- `raw-html`: makes it possible to include a block of raw HTML
- `readfile`: includes the content of another file in the current file; does not require the included file to be in a specific location.
