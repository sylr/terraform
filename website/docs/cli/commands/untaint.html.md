---
layout: "docs"
page_title: "Command: untaint"
sidebar_current: "docs-commands-untaint"
description: |-
  The `terraform untaint` command tells Terraform that an object is functioning
  correctly, even though its creation failed or it was previously manually
  marked as degraded.
---

# Command: untaint

Terraform has a marker called "tainted" which it uses to track that an object
might be damaged and so a future Terraform plan ought to replace it.

Terraform automatically marks an object as "tainted" if an error occurs during
a multi-step "create" action, because Terraform can't be sure that the object
was left in a fully-functional state.

You can also manually mark an object as "tainted" using the deprecated command
[`terraform taint`](./taint.html), although we no longer recommend that
workflow.

If Terraform currently considers a particular object as tainted but you've
determined that it's actually functioning correctly and need _not_ be replaced,
you can use `terraform untaint` to remove the taint marker from that object.

This command _will not_ modify any real remote objects, but will modify the
state in order to remove the tainted status.

If you remove the taint marker from an object but then later discover that it
was degraded after all, you can create and apply a plan to replace it without
first re-tainting the object, by using a command like the following:

```
terraform apply -replace="aws_instance.example[0]"
```

## Usage

Usage: `terraform untaint [options] address`

The `address` argument is a [resource address](/docs/cli/state/resource-addressing.html)
identifying a particular resource instance which is currently tainted.

This command also accepts the following options:

* `-allow-missing` - If specified, the command will succeed (exit code 0)
  even if the resource is missing. The command might still return an error
  for other situations, such as if there is a problem reading or writing
  the state.

* `-lock=false` - Disables Terraform's default behavior of attempting to take
  a read/write lock on the state for the duration of the operation.

* `-lock-timeout=DURATION` - Unless locking is disabled with `-lock=false`,
  instructs Terraform to retry acquiring a lock for a period of time before
  returning an error. The duration syntax is a number followed by a time
  unit letter, such as "3s" for three seconds.

* `-no-color` - Disables terminal formatting sequences in the output. Use this
  if you are running Terraform in a context where its output will be
  rendered by a system that cannot interpret terminal formatting.

* `-ignore-remote-version` - When using the enhanced remote backend with
  Terraform Cloud, continue even if remote and local Terraform versions differ.
  This may result in an unusable Terraform Cloud workspace, and should be used
  with extreme caution.

For configurations using
[the `local` backend](/docs/language/settings/backends/local.html) only,
`terraform taint` also accepts the legacy options
[`-state`, `-state-out`, and `-backup`](/docs/language/settings/backends/local.html#command-line-arguments).
