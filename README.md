# Sourceforge 2 Github

Move sourceforge tickets to github


## Build

```sh
# Download and install dependencies
go get -u github.com/Ajnasz/sf2gh

# Build main program
cd $GOPATH/src/github.com/Ajnasz/sf2gh
go build
```

## Usage

Generate a github token on https://github.com/settings/tokens

Edit the config.example.json, set the `userName` to your github username and
set the accessToken to the one you generated.

Create the github repository where you want to move the sourceforge tickets to. ( https://github.com/new )

```sh
./sf2gh -ghRepo github-repo-name -project sf-project-name
```

For example to move fluxbox tickets to your fluxbox github repo:

```sh
./sf2gh -ghRepo fluxbox -project fluxbox
```

### Available options

 - ghRepo: Github repository name (required)
 - project: Sourceforge project name (required)
 - sleepTime: Sleep between api calls in milliseconds. Github may stop you use the API if you call it too frequently (optional)
 - debug: Enable debug messages (optional)
