# go-ian-api

An API that allows sending a zip or tar.gz to an HTTP endpoint and get back a Debian Package

The zip or tar.gz should contain the contents of a folder that is buildable by [go-ian](https://github.com/penguinpowernz/go-ian)

## Todo (in order of priority)

- [ ] actually extract the zip and build the debian package from it
- [ ] implement rate limiting
- [ ] devise a one liner to use for pushing zip and downloading package in one go
- [ ] accept params to set package version number, architecture, etc
- [ ] create debian package to install the service on a system
- [ ] run each package build inside a docker