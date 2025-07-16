# minly

minly is a Go CLI application that allows you to manage and interact with [MinIO](https://min.io/) and [YOURLS](https://yourls.org/).
It provides commands to upload files to [MinIO](https://min.io/) and create short URLs using [YOURLS](https://yourls.org/).
It also supports managing and deleting files uploaded to [MinIO](https://min.io/).

## Usage

The usage of this tool is designed to be easy. Download the [latest release](https://github.com/devusSs/minly/releases/latest)
and unpack it to a place of your choice.

### Initializing the config and secrets

You may then run `minly init` to set up the config and secrets for the tool.

Since the secrets will be managed by keyring this will also need to be set up properly. Usually that works out of the box for Linux, macOS and Windows,
however some flavors of Linux and also WSL(2) have their issues with it. Refer to [the keyring section](./README.md#keyring) for more information.

### Keyring

For some flavors of Linux or also WSL(2) you might need to set up keyring properly first. To do so run each
of the following commands **individually** and **wait for them to complete successfully**.

In case keyring does not work properly for you or e.g. prompts you for a password in a GUI (e.g. on WSL) you can simply run the [fix_keyring.sh script](./scripts/fix_keyring.sh) in the [scripts folder](./scripts/). The easiest way to do this is probably running `/bin/bash scripts/fix_keyring.sh` while being in the repository's directory.

Thanks to [this little GitHub issue](https://github.com/XeroAPI/xoauth/issues/25) which also helped me resolve those issues.

### Running the app

After [setting up](./README.md#initializing-the-config-and-secrets) you can use the app as you wish and also automatically via scripts.

To see implemented commands use `minly` or `minly -h`. These commands and subcommands may be subject to change in the future. So please refer to the `help` function for more information.

## Building the app yourself

Although it is highly recommended to download [the latest release](https://github.com/devusSs/minly/releases/latest) and simply unpack that to which ever path of your choice and then run the app, you can also build it yourself.

To do so it is highly recommended to run the [included buildscript](./scripts/build.sh) by running `/bin/bash scripts/build.sh` while being in the repository's directory.
That script will set all needed build information if possible and makes the version work / print things properly.

## Disclaimer

This tool is still in active development and does not guarantee bug-free usage. There may also be unintended consequences
or issues created by the tool.
The developer also is not a professional software developer and simply maintains this tool for fun and personal use.
Use this tool at your own responsibility.

This tool is in no way affiliated with [MinIO](https://min.io/) or [YOURLS](https://yourls.org/).

Make sure to use it with caution and at your own risk. Do not use it for malicious purposes.

## LICENSE

Licensed under [MIT License](./LICENSE).