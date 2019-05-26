# autowebbackup
Automatically backup your web pages in an ftps directory

If you have your own webserver, it is your own responsability to create backups of your web pages. I personally like to rent an ftp-server to copy my backups to.

At the moment I have an ftp-server that makes use of the ftps protocol to encrypt it's data streams.

I created this software to automate the creation of daily, weekly and monthly backups. You can backup as many directories as you like. The software reads it's parameters from the file /etc/autowebbackup.toml.


