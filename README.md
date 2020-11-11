# go-ian-api

An API that allows sending a zip or tar.gz to an HTTP endpoint and get back a Debian Package

The zip or tar.gz should contain the contents of a folder that is buildable by [go-ian](https://github.com/penguinpowernz/go-ian)

Made as part of the Level1Techs Devember https://www.youtube.com/watch?v=Daf5C52s124

## Building

(theorectically, nothing to build yet)

Run `make`

## Usage

(theorectically, nothing to build yet)

Run `bin/go-ian-api`

Then call it with something like:

```
curl -X POST http://localhost:8080/upload \
  -F "file=@test.zip" \
  -H "Content-Type: multipart/form-data"
  --max-time 30
```

## Todo (in order of priority)

- [ ] devise a one liner to use for pushing zip and downloading package in one go
- [ ] actually extract the zip and build the debian package from it
- [ ] use UUIDs to uniquefy different requests
- [ ] delete package files after 15 mins
- [ ] implement rate limiting
- [ ] accept params to set package version number, architecture, etc
- [ ] create debian package to install the service on a system
- [ ] run each package build inside a docker
