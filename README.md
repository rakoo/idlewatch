idlewatch should eventually watch the IMAP folder of your choice with
the IDLE command and run a custom command when an event appears.

For the moment it scratches my itch: watching the "All Mail" folder of
my GMail account, and run [OfflineIMAP](http://offlineimap.org/) when a
new mail appears.

# Why

Polling sucks, and IDLE is the solution. I couldn't make OfflineIMAP's
own IDLE work for various reasons, here's my attempt at making it work.

# Installation

## OfflineIMAP
First you need a proper installation of OfflineIMAP. Here's mine, elided
with only the only important parts:

```ini
[general]
accounts = GMail

[Account GMail]
localrepository = GMailLocal
remoterepository = GMailRemote

[Repository GMailLocal]
type = Maildir
localfolders = ~/mails
folderfilter = lambda folder : folder not in [\
               "[Gmail].Important",\
               ]

nametrans = lambda folder : folder == 'archive' and '[Gmail]/All Mail' \
                         or folder == 'inbox' and 'INBOX' \
                         or folder == 'spam' and '[Gmail]/Spam' \
                         or folder == 'trash' and '[Gmail]/Trash' \
                         or folder == 'drafts' and '[Gmail]/Drafts' \
                         or folder == 'starred' and '[Gmail]/Starred' \
                         or folder == 'sent' and '[Gmail]/Sent Mail' \
                         or folder

[Repository GMailRemote]
type = Gmail
remoteuser = myuser
remotepass = mypassword
sslcacertfile = /etc/ssl/certs/ca-certificates.crt

nametrans = lambda folder : folder == '[Gmail]/All Mail' and 'archive' \
                         or folder == 'INBOX' and 'inbox' \
                         or folder == '[Gmail]/Spam' and 'spam' \
                         or folder == '[Gmail]/Trash' and 'trash' \
                         or folder == '[Gmail]/Drafts' and 'drafts' \
                         or folder == '[Gmail]/Starred' and 'starred' \
                         or folder == '[Gmail]/Sent Mail' and 'sent' \
                         or folder
```

You _must_ have a `[general]` section, and the first of your `accounts`
_must_ be `GMail`.

If after that you can run

```shell
$ offlineimap
```

in your shell, you're good to go

## idlewatch
* install the Go toolchain
* compile with
```shell
$ go build
```
* run it (preferably in the background, with a tmux or with a proper
  service management)
```shell
$ ./idlewatch
```
