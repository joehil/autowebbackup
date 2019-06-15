# autowebbackup
Automatically backup your web pages in an ftps directory

The backups can optionally be encrypted.

If you have your own webserver, it is your own responsability to create backups of your web pages. I personally like to rent an ftp-server to copy my backups to.

At the moment I have an ftp-server that makes use of the ftps protocol to encrypt it's data streams. The server is a ProFTPd server.

I created this software to automate the creation of daily, weekly and monthly backups. You can backup as many directories as you like. The software reads it's parameters from the file /etc/autowebbackup.toml.

This software was written in go (golang). To use it on your server, do the following:

- install Go (Golang) on your server
- copy autowebbackup to a directory of your own choice
- change to that directory
- run "go build"
- now you are supposed to see the new file "autowebbackup"
- copy the file "autowebbackup.toml" to "/etc/autowebbackup.toml"
- modify /etc/autowebbackup.toml to your needs
- run autowebbackup

To use autowebbackup regularly, consider adding it to your crontab. It is best to run autowebbackup daily, best during the night.

After installation autowebbackup can perform various tasks:


Backup directories configured in the configuration file

```
autowebbackup backup
```

Get a list of backups

```
autowebbackup list
```

Fetch a backup to the configured temporary directory

```
autowebbackup fetch <name of backup file>
```

Decrypt the fetched backup

If your backup is encrypted, decrypt it this way:

```
autowebbackup decrypt
```

