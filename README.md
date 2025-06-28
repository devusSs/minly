# minly

minly is a Go CLI tool designed to combine [MinIO](https://min.io/) and [YOURLS](https://yourls.org/) to create a data store with an integrated url shortener.

It is designed to be easy to use and setup and uses the named services as backend applications to avoid further configuration and development issues.

## Setup

If you are using some flavours of Linux or [WSL(2)](https://learn.microsoft.com/en-us/windows/wsl/install) (proven to be an issue), keyring might not work properly or might not be installed in the first place.

To fix this you can run the following commands to install missing components and make them work. Make sure every command is entered invidiually and completes
before running the next one.

```
sudo apt-get update

sudo apt-get install gnome-keyring libsecret-tools dbus-x11

sudo killall gnome-keyring-daemon

eval "$(printf '\n' | gnome-keyring-daemon --unlock)"

eval "$(printf '\n' | /usr/bin/gnome-keyring-daemon --start)"
```

Visit this little [Github issue](https://github.com/XeroAPI/xoauth/issues/25) for further information.

If more help is required and / or problems arise consult the search machine of your choice.

## Usage

TBA...

# Disclaimer

This tool is still in active development and does not guarantee bug-free usage. There may also be unintended consequences or issues created by the tool. The developer (devusSs) also is not a professional software developer and simply maintains this tool for fun. Use this tool at your own responsibility.

Also do not use this tool if you do not understand what it does or which purpose it serves.

This tool is in no way associated with [MinIO](https://min.io/) or [YOURLS](https://yourls.org/). It simply uses their awesome tools.

## LICENSE

Licensed under [MIT License](./LICENSE).
