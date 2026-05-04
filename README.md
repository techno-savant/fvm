# fvm

`fvm` is a Foundry version manager. It installs signed fvm binaries from GitHub Releases.

It helps you install multiple Foundry Virtual Tabletop versions on one machine and switch between them per project, per user, or for just the current shell session.

If you have used tools like `asdf`, `nvm`, or `pyenv`, this is the same basic idea for Foundry.

## Who this is for

`fvm` is meant for Foundry developers, module authors, system authors, and hobbyists who need to test against more than one Foundry release without manually renaming folders or juggling symlinks.

## What `fvm` does

With `fvm`, you can:

- install multiple Foundry versions side-by-side
- set a default version for your machine
- pin a version for one project with `.fvm-version`
- temporarily override the version for your current shell
- ask which version is active and where it lives on disk

## How version selection works

When `fvm` needs to decide which Foundry version to use, it checks in this order:

1. `FOUNDRY_VERSION` in your current shell
2. the nearest `.fvm-version` file in your current project
3. your global default in `~/.fvm/version`

If none of those exist, `fvm` will tell you no version is configured.

## Quick start

This is the shortest path from zero to working.

### 1. Install `fvm`

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

This installs the `fvm` binary itself. It does not install a Foundry version yet.

If you want more detail, custom install directories, or manual installation steps, see [docs/install.md](docs/install.md).

### 2. Make sure `fvm` works

```sh
fvm version
fvm help
```

You should see the CLI version and the available commands.

### 3. Enable shell integration

Pick your shell:

For bash:

```sh
eval "$(fvm init bash)"
```

For zsh:

```sh
eval "$(fvm init zsh)"
```

For fish:

```sh
fvm init fish | source
```

That only affects your current shell session. To make it permanent, add the same command to your shell config file:

- bash: `~/.bashrc` or `~/.bash_profile`
- zsh: `~/.zshrc`
- fish: `~/.config/fish/config.fish`

This step puts `~/.fvm/shims` on your `PATH`, which is how `fvm` intercepts `foundry` and `foundryvtt` and routes them to the selected version.

### 4. Configure Foundry authentication

Official Foundry downloads are authenticated. `fvm` supports two ways to do that:

- pass an existing Foundry web session cookie
- optionally give `fvm` your Foundry username and password so it can try to log in for you

Recommended option: cookie-based auth.

```sh
export FOUNDRY_COOKIE='sessionid=...; csrftoken=...'
```

Quick way to get it from your browser:

1. log in to https://foundryvtt.com in your browser
2. open DevTools
3. go to Application/Storage -> Cookies -> https://foundryvtt.com
4. copy the `sessionid` and `csrftoken` cookie values
5. format them like this:

```sh
export FOUNDRY_COOKIE='sessionid=YOUR_SESSION_ID; csrftoken=YOUR_CSRF_TOKEN'
```

You can also often find them in the browser Network tab by opening any request to `foundryvtt.com` and copying the `Cookie` request header.

Username/password login is best-effort only and may be rejected by Foundry's website flow.

```sh
export FOUNDRY_USERNAME='your-foundry-username-or-email'
export FOUNDRY_PASSWORD='***'
```

You can also put either form in `~/.fvm/config.yaml`:

```yaml
foundry_cookie: "sessionid=...; csrftoken=..."
# optional best-effort fallback
foundry_username: "your-foundry-username-or-email"
foundry_password: "your-password"
```

If both are present, `fvm` uses the cookie first.

Treat all of this like credentials. Do not commit it to git, paste it into screenshots, or leave it lying around in shared shell history.

### 5. Install a Foundry version

```sh
fvm install 13.346
```

By default, `fvm` requests the `node` platform build from Foundry.

That is deliberate. The `node` artifact is the archive format `fvm` can install cleanly across platforms today. Native desktop artifacts like `.dmg` are not the default path.

If you need to override the platform, set it in config:

```yaml
foundry_platform: node
```

Supported values are:

- `node`
- `linux`
- `mac`
- `windows`
- `windows_portable`

For most people, leave it on `node`.

### 6. Set a default version

```sh
fvm global 13.346
```

This writes your default version to `~/.fvm/version`.

### 7. Verify what is active

```sh
fvm current
fvm which
```

Typical output might look like:

```text
13.346 (global)
/Users/you/.fvm/versions/13.346/foundry/foundry
```

## Common workflows

### Use one version for all projects by default

```sh
fvm install 13.346
fvm global 13.346
```

### Pin a specific project to a version

From inside the project directory:

```sh
fvm install 12.331
fvm local 12.331
```

This creates a `.fvm-version` file in the current directory.

Anyone opening that project later can run `fvm current` and see what it expects.

### Temporarily switch just for this terminal

After enabling shell integration, use:

```sh
fvm_use 12.331
```

That sets `FOUNDRY_VERSION` for the current shell only. Close the shell and the override is gone.

Use this when you want to test quickly without changing project files or your global default.

### Check where a version is installed

```sh
fvm where 13.346
```

### Check whether your environment is healthy

```sh
fvm doctor
```

Use this first if something feels off.

## Command reference

### `fvm current`

Shows the active version and where it came from.

Examples:

```sh
fvm current
```

Possible output:

```text
13.346 (.fvm-version)
13.346 (global)
13.346 (FOUNDRY_VERSION)
```

### `fvm which`

Prints the path to the active Foundry executable.

```sh
fvm which
```

If the selected version is not installed yet, `fvm` tells you which install command to run.

Both default shim names resolve through the same executable selection logic:

```sh
foundry
foundryvtt
fvm which foundry
fvm which foundryvtt
```

`fvm` generates both `foundry` and `foundryvtt` shims by default.

### `fvm where <version>`

Shows the install directory for a specific version.

```sh
fvm where 13.346
```

### `fvm local <version>`

Pins the current project to a version by writing `.fvm-version`.

```sh
fvm local 13.346
```

### `fvm global <version>`

Sets your machine-wide default version.

```sh
fvm global 13.346
```

### `fvm list`

Lists versions currently installed on your machine.

```sh
fvm list
```

### `fvm list-remote`

Lists remote versions `fvm` knows about.

```sh
fvm list-remote
```

Remote listing is still evolving. Authenticated download support is the more complete path today.

### `fvm install <version>`

Downloads and installs a Foundry version.

```sh
fvm install 13.346
```

Credential-based example:

```sh
export FOUNDRY_USERNAME='your-foundry-username-or-email'
export FOUNDRY_PASSWORD='your-password'
fvm install 13.346
```

Cookie-based example:

```sh
export FOUNDRY_COOKIE='sessionid=...; csrftoken=...'
fvm install 13.346
```

Persistent machine-local example:

```yaml
# ~/.fvm/config.yaml
foundry_username: "your-foundry-username-or-email"
foundry_password: "your-password"
foundry_platform: node
```

Your Foundry auth data is sensitive. Keep it out of git, dotfile repos, screenshots, and shared terminals.

### `fvm shim regenerate`

Rebuilds the shim executables in `~/.fvm/shims`.

```sh
fvm shim regenerate
```

Run this if your shell integration is set up but the shim directory seems stale, or after upgrading from an older `fvm` build that only wrote one shim.

### `fvm doctor`

Checks for common setup problems.

```sh
fvm doctor
```

This is a good first step if:

- `foundry` is not resolving to the version you expect
- `fvm` works but shell switching does not
- your shim directory is missing from `PATH`
- you are not sure whether anything is installed yet

## Where `fvm` stores things

By default, `fvm` uses `~/.fvm/`.

Common paths:

- `~/.fvm/config.yaml` — machine-local configuration
- `~/.fvm/version` — your global default version
- `~/.fvm/versions/` — installed Foundry versions
- `~/.fvm/shims/` — shim executables added to your `PATH`
- `~/.fvm/tmp/` — temporary install working files

## Troubleshooting

### `fvm: command not found`

The binary is either not installed or not on your `PATH`.

Check:

```sh
which fvm
```

If that prints nothing, re-run the installer or add the install directory to your shell `PATH`.

### `foundry: command not found`

You probably installed `fvm` but did not enable shell integration.

Run:

```sh
eval "$(fvm init bash)"
```

Or the zsh/fish equivalent, then try again.

If `foundryvtt` fails the same way, the fix is the same. If shell integration is already enabled, run `fvm shim regenerate` to rebuild the shim directory.

### `fvm current` says no version is configured

That means you have not set any of the version selectors yet.

Do one of these:

```sh
fvm global 13.346
```

or:

```sh
fvm local 13.346
```

or:

```sh
export FOUNDRY_VERSION=13.346
```

### Install fails with an authentication error

That usually means one of these:

- your cookie expired
- your username/password is wrong
- your account is not allowed to download that build

Try again with fresh credentials.

If you are using a cookie, make sure it includes the Foundry session values, not some unrelated browser cookie.

### Install fails with an unsupported archive format

You probably forced a native desktop platform like `mac` and got a `.dmg`.

Use the default `node` platform unless you are deliberately experimenting with native artifacts.

```yaml
foundry_platform: node
```

### `fvm which` points somewhere unexpected

Run:

```sh
fvm current
fvm doctor
```

That will usually tell you whether a local `.fvm-version`, shell override, or missing `PATH` entry is causing the mismatch.

## Release process

`fvm` now uses GitHub Actions for CI and release automation.

### CI

On every push to `main` and every pull request, GitHub Actions runs:

- `go test ./...`
- a smoke build of `./cmd/fvm`

### Versioning

This repo is set up for conventional commits and `release-please`.

Use commit prefixes like:

- `fix:` for patch releases
- `feat:` for minor releases
- `feat!:` or a `BREAKING CHANGE:` footer for major releases
- `docs:`, `ci:`, `chore:` for non-feature maintenance work

`release-please` watches `main` and opens or updates a release PR based on those commits.

### Publishing

When the release PR is merged and a tag like `v0.1.2` is created, the release workflow automatically:

- builds darwin/amd64
- builds darwin/arm64
- builds linux/amd64
- builds linux/arm64
- packages `fvm_<tag>_<os>_<arch>.tar.gz`
- generates `checksums.sha256`
- uploads all assets to the matching GitHub Release

That means the installer script can rely on release assets existing, instead of someone having to remember to upload them manually.

## Status

`fvm` is still early. The current implementation is strongest at:

- installing authenticated official Foundry builds
- managing multiple installed versions locally
- switching versions with shims and shell integration

The current intentional bias is reliability over cleverness. That is why the default install path uses the `node` artifact instead of pretending desktop installer formats are solved.

## License

See [LICENSE](LICENSE).
