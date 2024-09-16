
![lumaca_logo](https://github.com/jmcharter/lumaca/assets/3820235/763295a9-8dc9-4e66-af8a-024a40f077bf)

### A simple static site generator written in Go.

## What is Lumaca?

Lumaca is a static site generator written in Go, intended to generate my personal blog.  
This is not a commercial project and is not intended to be distributed or used by others. With that said, feel free to use it if you feel it meets your needs, but understand that using it comes with no warranty, offer of support or any guarantees.

The project is MIT licensed, so feel free to fork it, copy it, modify and generally do as you please.

## Usage

### Installation
The easiest way to get started with Lumaca is to use the Go package manager:

```sh
go install github.com/jmcharter/lumaca@latest
```

Alternatively, you can clone this repo and generate your own binary.

### Get Started

Once it's installed, create a new directory for your blog and run the `init` command:

```sh
mkdir MyBlog
cd MyBlog
lumaca init --author "John Doe" --title "MyBlog"
```

This will create a config file for you, which you can customize if you wish. It will also generate all of the boilerplate and template you need to get going.

To create your first blog post, use the `new` command.

```sh
lumaca new --title "My First Post"
```

A markdown file with some Frontmatter will be generated in your `Posts` directory (`content/posts`). Edit as appropriate, and when you're finished writing your post, you can run `build` to generate the html.

```sh
lumaca build
```

And if you'd like to see how it looks locally before deploying somewhere, you can run `serve`.

```sh
lumaca serve
```
