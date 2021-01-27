# Theory of operation

- [Theory of operation](#theory-of-operation)
  - [Introduction](#introduction)
  - [Interfacing with restic key management](#interfacing-with-restic-key-management)
    - [Basic principles](#basic-principles)
    - [Identifying generated keys versus Vault tokens](#identifying-generated-keys-versus-vault-tokens)
    - [Integrated calls to Vault with restic](#integrated-calls-to-vault-with-restic)
  - [Reference architecture / Assumptions](#reference-architecture--assumptions)
  - [Operational scenario](#operational-scenario)

## Introduction

This plugins bridges the **restic** backup application with the **HashiCorp Vault** secrets management tool.

Using the **database** secrets engine, it entends to generate ephemeral repository keys applicable only to a single backup job.

Mostly, it is a wrapper to execute shell commands that will add and remove keys to a defined restic repository.

## Interfacing with restic key management

### Basic principles

*Taken from https://restic.readthedocs.io/en/stable/070_encryption.html*

It is possible to set multiple keys on a restic repository using the `key` command:

```
$ restic -r /srv/restic-repo key list
enter password for repository:
 ID          User        Host        Created
----------------------------------------------------------------------
*eb78040b    username    kasimir   2015-08-12 13:29:57

$ restic -r /srv/restic-repo key add
enter password for repository:
enter password for new key:
enter password again:
saved new key as <Key of username@kasimir, created on 2015-08-12 13:35:05.316831933 +0200 CEST>

$ restic -r /srv/restic-repo key list
enter password for repository:
 ID          User        Host        Created
----------------------------------------------------------------------
 5c657874    username    kasimir   2015-08-12 13:35:05
*eb78040b    username    kasimir   2015-08-12 13:29:57
```

Assuming the repository has already been created with an initial password, all that needs to be done is have a plugin to create and delete keys using the standard `database` paradigm of Vault.

At this point in time, the matter of rotating this initial password (the *root* password in the context of Vault) is not handled yet.

### Identifying generated keys versus Vault tokens

The additional key will be easy to identify: it will originate from the user and hostname that runs `Vault`.
We can thus identify our key by matching on the most recent key created from this source.

This could however be a problem: if several tokens are created due to several backup jobs, it would be difficult to distinguish them on the `restic` side.

An analysis of the `cmd_key.go` file from the `restic` source code shows there is a simple solution: the user and host can be passed as flags to the `key` command:

```
func init() {
var cmdKey = &cobra.Command{
	Use:   "key [flags] [list|add|remove|passwd] [ID]",
	Short: "Manage keys (passwords)",
	Long: `
The "key" command manages keys (passwords) for accessing the repository.

...

	cmdRoot.AddCommand(cmdKey)

	flags := cmdKey.Flags()
	flags.StringVarP(&newPasswordFile, "new-password-file", "", "", "`file` from which to read the new password")
	flags.StringVarP(&keyUsername, "user", "", "", "the username for new keys")
	flags.StringVarP(&keyHostname, "host", "", "", "the hostname for new keys")
}
```

This also solves another problem: the generated password can be passed as an argument through a temporary file.  Since we can pass an existing key as an environment variable, this solves all the problems: each token will be uniquely identified by being used as the username.

### Integrated calls to Vault with restic

The scripts that make use of restic can then be wrapped with **Consul Template** (as explained in https://learn.hashicorp.com/tutorials/vault/application-integration), and that's job done.

## Reference architecture / Assumptions

Let us consider three nodes:

* `VaultServer` runs HashiCorp `Vault` with our database secrets plugin.
* `SourceNode` is an entity that has data it wants to backup with `restic` installed.
* `ResticRepo` is a target `restic` repository, already initialised.  This can be a local folder on `SourceNode`, a remote SCP target, etc.  It does not really matter.

The following assumptions are made over this plugin:

* The `restic` binary is available on the `VaultServer` node.
* The `VaultServer` node can, running `restic`, reach `ResticRepo`.
* We consider `consulenv` is used in the rest of this document.  It is up to the users to adapt if needed.

TODO: Define the permissions needed for the `SourceNode` tokens.

## Operational scenario

1. `SourceNode` requests a password with a REST call to `VaultServer` through `consulenv`.
2. `VaultServer` generates a new ephemeral password
3. `VaultServer` writes the password to a volatile file and, using the `restic` command, adds it to `ResticRepo`
4. `SourceNode` receives the key and can use it.
5. After the TTL has passed, `VaultServer` deletes the key from `ResticRepo`.