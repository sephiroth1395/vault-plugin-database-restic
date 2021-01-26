# Theory of operation

## Introduction

This plugins bridges the **restic** backup application with the **HashiCorp Vault** secrets management tool.

Using the **database** secrets engine, it entends to generate ephemeral repository keys applicable only to a single backup job.

Mostly, it is a wrapper to execute shell commands that will add and remove keys to a defined restic repository.

## Interfacing with restic key management

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

Assuming the repository has already been created with an initial password, all that needs to be done is have a plugin to create and delete keys using the standard `database` paradigm of vault.

The scripts that make use of restic can then be wrapped with **Consul Template** (as explained in https://learn.hashicorp.com/tutorials/vault/application-integration), and that's job done.

At this point in time, the matter of rotating this initial password (the *root* password in the context of Vault) is not handled yet.