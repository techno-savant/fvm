# Installing fvm

`fvm` is distributed as a self-contained binary for Linux and macOS on AMD64 and ARM64.

This page focuses on the install experience for normal users, especially people who do not spend all day in shell configs.

## What this installs

The installer installs the `fvm` CLI itself.

It does not automatically install a Foundry version. After installing `fvm`, you still need to configure Foundry authentication and install a version:

```sh
export FOUNDRY_COOKIE='sessionid=...; csrftoken=...'
fvm install 13.346
fvm global 13.346
```

`FOUNDRY_COOKIE` is the recommended auth method. Username/password login exists as a best-effort fallback and may be rejected by Foundry's website flow.

## Recommended install

Run the installer script directly from GitHub:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

The script will:

1. detect your operating system and CPU architecture
2. fetch the latest release tag from the GitHub API
3. download the correct release archive from GitHub Releases
4. verify the SHA-256 checksum
5. extract and install the `fvm` binary
6. tell you if you need to update your `PATH`

## After installation

Run these commands first:

```sh
fvm version
fvm help
```

If those work, `fvm` is installed correctly.

Then set up shell integration:

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

This adds `~/.fvm/shims` to your `PATH` so both `foundry` and `foundryvtt` resolve through `fvm`.

Then authenticate and install a Foundry version:

```sh
export FOUNDRY_COOKIE='sessionid=...; csrftoken=...'
fvm install 13.346
fvm global 13.346
fvm current
```

If you already had shims from an older `fvm` build, regenerate them after upgrading:

```sh
fvm shim regenerate
```

Username/password login exists as a best-effort fallback, but the cookie path is the reliable one.

## Authenticated Foundry downloads

Official Foundry packages are authenticated. `fvm install <version>` supports two approaches:

1. pass a Foundry web session cookie
2. optionally give `fvm` a Foundry username and password so it can try to log in for you

If both are configured, `fvm` uses the cookie first.

### Recommended option: session cookie

This is the reliable path.

If you already have a valid Foundry session and want to reuse it directly:

```sh
export FOUNDRY_COOKIE='sessionid=...; csrftoken=...'
fvm install 13.346
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

Or in `~/.fvm/config.yaml`:

```yaml
foundry_cookie: "sessionid=...; csrftoken=..."
```

### Optional fallback: username and password

This is best-effort only and may be rejected by Foundry's website flow.

Current-shell example:

```sh
export FOUNDRY_USERNAME='your-foundry-username-or-email'
export FOUNDRY_PASSWORD='***'
fvm install 13.346
```

Persistent machine-local config example:

```yaml
# ~/.fvm/config.yaml
foundry_username: "your-foundry-username-or-email"
foundry_password: "your-password"
```

If username/password login fails, use `FOUNDRY_COOKIE` instead.

### Safety notes

- treat your cookie and password like credentials
- do not commit them to git
- do not paste them into bug reports or screenshots
- prefer machine-local config over project config for anything sensitive
- if you keep shell history, be aware that inline exports may be recorded

### When installs stop working

If installs suddenly fail after working before, one of these is probably true:

- your Foundry session cookie expired
- your username/password changed or was entered wrong
- your Foundry account does not have access to that build

Refresh the cookie or retry with known-good credentials.

## Foundry platform choice

By default, `fvm` requests the `node` platform build from Foundry.

That is intentional.

The `node` artifact is the format `fvm` can install reliably today across platforms. Native desktop artifacts like `.dmg` are not the default path because they require installer-specific extraction logic that `fvm` does not pretend to support yet.

If you want to force a different platform, set it in `~/.fvm/config.yaml`:

```yaml
foundry_platform: node
```

Supported values:

- `node`
- `linux`
- `mac`
- `windows`
- `windows_portable`

Unless you have a very specific reason, leave this on `node`.

## Choose a specific `fvm` release

If you want to install a specific `fvm` release instead of the latest one:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | VERSION=v0.2.0 sh
```

Or download the script first:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  -o install.sh
VERSION=v0.2.0 sh install.sh
```

Important: this does not install Foundry `v0.2.0`. It installs `fvm` release `v0.2.0`.

Also important: this does not work the way many people expect:

```sh
VERSION=v0.2.0 curl ... | sh
```

That sets the variable on `curl`, not on the installer shell. Put the variable on `sh` instead.

## Choose an install directory

By default, the script installs to `/usr/local/bin` if it can write there. If not, it falls back to `$HOME/.local/bin`.

To choose your own install directory:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | INSTALL_DIR=/opt/bin sh
```

Or:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  -o install.sh
INSTALL_DIR=/opt/bin sh install.sh
```

Again, this does not work the way people often assume:

```sh
INSTALL_DIR=/opt/bin curl ... | sh
```

That applies the variable to `curl`, not to the installer.

## PATH setup

If the installer says the binary directory is not on your `PATH`, add it to your shell profile.

The most common fallback install directory is `$HOME/.local/bin`.

### bash

Add this to `~/.bashrc` or `~/.bash_profile`:

```sh
export PATH="$PATH:$HOME/.local/bin"
```

Then reload your shell:

```sh
source ~/.bashrc
```

If your system uses `~/.bash_profile` instead, reload that file instead.

### zsh

Add this to `~/.zshrc`:

```sh
export PATH="$PATH:$HOME/.local/bin"
```

Then reload your shell:

```sh
source ~/.zshrc
```

### fish

Use fish's universal path support or add the directory to your fish config.

A common approach is:

```fish
fish_add_path $HOME/.local/bin
```

If you want it in your config file, add it to `~/.config/fish/config.fish`.

## Shell integration setup

Installing the `fvm` binary is not the same thing as enabling version switching for `foundry`.

For switching to work smoothly, you usually want the shell integration too.

### What shell integration does

It mainly does two things:

1. puts `~/.fvm/shims` on your `PATH`
2. gives you the `fvm_use` helper for temporary shell-only switching

### Temporary setup for this shell only

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

### Permanent setup

Add the same command to your shell startup file.

bash:

```sh
eval "$(fvm init bash)"
```

zsh:

```sh
eval "$(fvm init zsh)"
```

fish:

```fish
fvm init fish | source
```

Put them in:

- `~/.bashrc` or `~/.bash_profile` for bash
- `~/.zshrc` for zsh
- `~/.config/fish/config.fish` for fish

## Manual install

If you do not want to use the installer script, download a release archive from the [Releases page][releases].

### Pick the correct archive

| Platform    | Archive                          |
| ----------- | -------------------------------- |
| macOS Intel | `fvm_Darwin_x86_64.tar.gz`       |
| macOS Apple | `fvm_Darwin_arm64.tar.gz`        |
| Linux Intel | `fvm_Linux_x86_64.tar.gz`        |
| Linux ARM64 | `fvm_Linux_arm64.tar.gz`         |
|

### Install manually

1. download the archive for your platform
2. extract it
3. move the `fvm` binary into a directory on your `PATH`
4. run `fvm version`
5. run `eval "$(fvm init <shell>)"`

Example:

```sh
tar -xzf fvm_Linux_x86_64.tar.gz
chmod +x fvm
mv fvm ~/.local/bin/
fvm version
```

Then continue with the auth and install flow described above.

## If something feels broken

Start here:

```sh
fvm doctor
fvm current
```

If `install` fails with an unsupported archive format, you probably forced a native platform and got something like a `.dmg`. Put `foundry_platform` back to `node`.

If `fvm` works but `foundry` does not, shell integration is probably missing.

If auth fails, verify whether you are using the right cookie or the right Foundry account credentials.

[releases]: https://github.com/foundry/fvm/releases
