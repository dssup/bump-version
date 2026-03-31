# Contributing

This document is to help to understand some decisions and how to contribute.

## Read README.en.md

This may sound  obvious but this document  already has a lot  of the information
about this tool but it also has  the information about how you should commit and
what your commit message should look like. This tool uses itself for releases.

## Maintainers

You can ask the maintainers if you're  unsure about something or have a question
that is not answered here.

- Daniil Stepanov <dstepanov485@gmail.com>

## Stack

- Go 1.26 or newer

You may additionally need:

- golangci-lint (linter for Go)
- gofumpt (a stricter formatter for Go)

If you know  more code quality tools  for Go then feel free  to contribute them.
It's always great to make sure code is even more reliable and consistent.

## Languages

As for natural languages, this project must support both English and Russian.

As for programming languages and other Turing complete languages, do not use any
other language  besides Go. It helps  to keep this project  consistent and avoid
mixing many languages.

## Distribution

Make the distribution simple without  any installation required. Don't add other
languages other than Go. Only statically link the binaries.

## Questions and Answers

### Why There is No Shell Scripts or Makefiles?

Because you can do the same things  using just Go. The entire project, including
scripts is  written in Go. There's  no need to complicate  things by introducing
another language especially the one that is platform-dependant or even "It works
on my machine"-dependant such as shell scripting languages. If you remember, the
reason why  Perl was  created is  to solve  the problem  of shell  scripts being
platform-dependant. The  Go scripts  are approximately the  length but  are much
more readable, portable, reliable and  extensible for more complicated logic. It
allows to  use the  whole Go  tooling infrastructure  in scripts  too such  as ,
libraries  (although dependencies  are discouraged  in this  project), debugger,
formatter, linter,  everything basically. Plus,  `Go` has a convenient  `go run`
command  to  run  code  with  on-the-fly compilation,  so  there's  no  need  to
explicitly compile.

The `scripts/` directory is a conventional  name. All project scripts must be in
that directory. You can personally use  a global program, e.g., `sc`, across all
of the Go  projects you have. This  allows to just type `sc  check`, `sc build`,
`sc  clean`. You  can  do  it by  adding  your own  `sc`  program  to your  PATH
environment variable. It can be implemented as  `cd scripts && go run . "$@"` or
any other way you want. To make  this practice more convenient, keep the scripts
as a single program-launcher for many scripts.

As for incremental  builds, both Go compiler and modern  computers are extremely
fast. Moreover,  Go actually does use  the global build cache  for artifacts, so
the build  process is  mostly automated  by the Go  tooling itself  already, and
there's no  need to  introduce makefiles  on top of  it. It  would be  the third
language, platform-dependant one, and a tool that the user would have to install
in order  to simply  build this software.  Software must be  built by  the tools
provided by the programming language itself. Software must be built by the tools
of programming language itself, as well  as the software build process code must
be  written in  the same  language  that the  software  is written  in. This  is
inspired by the philosophy of the Zig programming language.

Keep things simple please. Don't  add complexity and polyglot-required things to
this project. This allows this project to  simply say "install only Go" in order
to do  everything here including building  and testing. This is  easier for both
developers and users.

### Can I Use Library X?

No. This  project strives for  portability and  avoids the dependency  hell. Use
_only_ the Go standard library. One of the coolest features of Go is the amazing
batteries-included  standard library  that  dramatically reduces  the number  of
third-party packages needed.
