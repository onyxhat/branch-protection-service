# Branch Protection Service

Branch Protection Service is a web service that listens for [organization events](https://docs.github.com/en/developers/webhooks-and-events/webhooks/about-webhooks#events) to protect the main/master branch from force pushing and branch deletion.

## Requirements
* A Github Personal Access Token
* A Github Organization
* Go (version 1.16+) if using binary
* Docker if using container

## Creating a Personal Access Token
Steps for creating a personal access token can be found [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

For scopes to the token, you can check the following boxes:
* public_repo
* admin:org
* admin:repo_hook
* admin:org_hook
* write:discussion

## Creating an Organization
Steps for creating an organization can be found [here](https://docs.github.com/en/organizations/collaborating-with-groups-in-organizations/creating-a-new-organization-from-scratch)

## Running the Web Service
Clone and build the repository.
```
go build 
```

The binary takes two inputs, `token` and `org`. `token` is the github personal access token string. `org` is the github organization name string.

Running the binary looks like this,
```
./branch-protection-service -token <github pat> -org <org name>
```

Alternatively, you can run this as a container. First build,
```
docker build -t branch-protection-service .
```

Run the container 
```
docker run --rm -d --name bps -e TOKEN=<github pat> -e ORG=<org name> branch-protection-service
```

## Design
The web service uses the personal access token of an authenticated user of a Github Organization to listen to Organization events. It polls events from the Organization every minute and checks to see if repositories have their main/master branch protected. If a new repository is created, and main/master branch established, it will protect that branch from force pushing and branch deletion. It will then open an issue and @mention the author, assigning the issue to them as notification.

## Go Library
The go-github library is used to access the [GitHub API](https://docs.github.com/en/rest).

[Github Go SDK (unofficial)](https://github.com/google/go-github)
[go-github godocs](https://pkg.go.dev/github.com/google/go-github/v40)
