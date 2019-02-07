# Terraform-secure-backend [![Build Status](https://travis-ci.org/orange-cloudfoundry/terraform-secure-backend.svg?branch=master)](https://travis-ci.org/orange-cloudfoundry/terraform-secure-backend)

An [http backend](https://www.terraform.io/docs/backends/types/http.html) which stores and retrieves tfstates files in a secure and encrypted way through [credhub](https://github.com/cloudfoundry-incubator/credhub).

This backend supports [state locking](https://www.terraform.io/docs/state/locking.html).

## Installation

Installer will place the latest release binary in your current working directory.

### On *nix system

You can install this via the command-line with either `curl` or `wget`.

#### via curl

```bash
$ bash -c "$(curl -fsSL https://raw.github.com/orange-cloudfoundry/terraform-secure-backend/master/bin/install.sh)"
```

#### via wget

```bash
$ bash -c "$(wget https://raw.github.com/orange-cloudfoundry/terraform-secure-backend/master/bin/install.sh -O -)"
```

### On windows

You can install it by downloading the `.exe` corresponding to your cpu from releases page: https://github.com/orange-cloudfoundry/terraform-secure-backend/releases .
Alternatively, if you have a terminal interpreting shell you can also use command line script above, it will download file in your current working dir.

## Commands

```
NAME:
   terraform-secure-backend - An http server to store terraform state file securely

USAGE:
   terraform-secure-backend [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config-path value, -c value  Path to the config file (default: "backend-config.yml")
   --help, -h                     show help
   --version, -v                  print the version
```

## Run it

There is two different ways to run the server:
1. [In local](#in-local)
2. [In a cloud](#in-a-cloud) through [gautocloud](https://github.com/cloudfoundry-community/gautocloud) (Run with ease this server on: Kubernetes, CloudFoundry or Heroku)

### In local

1. Create a `backend-config.yml` file where you want to run your server, following this schema:

```yaml
host: 0.0.0.0 # an be 127.0.0.1 too
port: 8080 # port to listen
name: terraform-secure # this name inside credhub to create an unique path for your tfstate
cert: ~ # Set a path or pem cert string certificate to run your senver in tls (ignored if lets_encrypt_domains is set)
key: ~ # Set a path or pem key string certificate to run your senver in tls (ignored if lets_encrypt_domains is set)
log_level: ~ # Verbosity, can be info, debug, warning, error
log_json: false # set to true to see logs as json instead of plain text (useful for logsearch)
no_color: false # set to true to not have color (this cannot be use when log_json is to true)
lets_encrypt_domains: [] # Set a or multiple domains name to acquire a certificate from let's encrypt
username: user # basic auth username to secure access to this app
password: password # basic auth password to secure access to this app
show_error: true # If true, if an error occurred details will be shown in the web page as json 

credhub_server: path.to.my.credhub.com # path to your credhub server (note https is enforced)
credhub_username: credhub_user # an UAA username with credhub.read and credhub.write scopes (this can be empty if credhub_client and credhub_secret are set)
credhub_password: credhub_password # an UAA password with credhub.read and credhub.write scopes  (this can be empty if credhub_client and credhub_secret are set)
credhub_client: ~ # an UAA client_id with credhub.read and credhub.write scopes (this can be empty if credhub_username and credhub_password are set)
credhub_secret: ~ # an UAA client_id with credhub.read and credhub.write scopes (this can be empty if credhub_username and credhub_password are set)
credhub_ca_cert: ~ # You can set the credhub ca_cert here if it's a self signed certificate
skip_ssl_validation: false # set to true to skip ssl validation when connecting to your credhub (prefer use credhub_ca_cert for security reasons)
cef: false # set to true to enable security event in common event format 
cef-file: ~ # set a path to a file to store security event in common event format to a file
dry-run: false # set to true to not sent to credhub state file
```

2. Run `./terraform-secure-backend` in your terminal and server is now started.

### In a cloud
  
#### On CloudFoundry

1. Create a cups service named `.*config` with the same credentials set in yaml, example:
```json
{
  "name": "terraform-secure",
  "credhub_server": "path.to.my.credhub.com",
  "credhub_username": "credhub_user",
  "credhub_password": "credhub_password"
}
```
2. Bind it to your terraform-secure-backend instance

#### On heroku or kubernetes

Add env var following this format: `.*CONFIG_OPTION`, example:

```bash
BACKEND_CONFIG_NAME="terraform-secure"
BACKEND_CONFIG_CREDHUB_SERVER="path.to.my.credhub.com"
BACKEND_CONFIG_CREDHUB_USERNAME="username"
BACKEND_CONFIG_CREDHUB_PASSWORD="password"
BACKEND_CONFIG_LETS_ENCRYPT_DOMAINS="mydomain1.com,mydomain2.com"
```

## Usage in your terraform

Add in your `.tf` file a new http backend (**Note**: `<deployment name>` is whatever you want, better a name which represent the name of your deployment):

```hcl
terraform {
  backend "http" {
    address = "https://path.to.my.secure.backend.com/states/<deployment name>"
    lock_address = "https://path.to.my.secure.backend.com/states/<deployment name>"
    unlock_address = "https://path.to.my.secure.backend.com/states/<deployment name>"
    username = "user"
    password = "password"
  }
}
```

## Api

The Api implements the terraform [http backend API](https://www.terraform.io/docs/backends/types/http.html) on each `https://path.to.my.secure.backend.com/states/<deployment name>`.

You can list all tfstates stored by calling: `https://path.to.my.secure.backend.com/states`
