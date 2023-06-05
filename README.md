# GoPie
This is the repo of Go-pie, a concurrency testing tool for Golang.
This document is to introduce the structure of GoPie project.

## Project Structure
- cmd: files under `cmd` is the command line tools of GoPie, include the instrument tools and the fuzzing tools.
- patch: files under `patch` are the hacks of the runtime of Golang, these files will be replace before compiling the target binary.
- pkg: the packages of GoPie, contain all the implment details in the form of source code.
- script: shell scripts used during develop.
- Dockerfile: dev environment.

## Usage
1. Build the binary under `cmd` with `go build ./cmd/...`. There will be two binaries after compilation, the `inst` and `fuzz`. 
2. Then run the script/pre.sh to inst and complie your project which is to be tested. Then use `fuzz --task full --dir your_project` to start testing.
