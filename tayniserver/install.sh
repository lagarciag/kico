#!/bin/bash

echo "Installing tayniserver"
go install -v tayniserver.go

FILE=/usr/local/bin/tayniserver

if [ -f $FILE ]; then
    echo "File $FILE exists."
else
    echo "File $FILE does not exist."
    sudo ln -s $GOPATH/bin/tayniserver /usr/local/bin/tayniserver 
    fi

echo "Installing service..."
sudo cp $GOPATH/src/github.com/lagarciag/tayni/tayniserver/tayniserver.service /etc/systemd/system/
echo "Enabling service..."
sudo systemctl enable tayniserver.service

