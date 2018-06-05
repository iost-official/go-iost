#! /usr/bin/env python
# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys

HOME=os.environ['HOME']

def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,shell=True)
	obj.wait()
	ret=obj.stdout.read()
	obj.stdout.close()
	return ret

def upgrade():
	pwd="$GOPATH/src/github.com/iost-official/prototype"
	ret=wCommand("cd "+pwd+" && git checkout develop && git pull")
	ret=wCommand("cd "+pwd+"/iserver && go build")
	ret=wCommand("nohup redis-server &")
	ret=wCommand("cd "+pwd+"/iserver && nohup ./iserver &")

	return "SUCCESS"		

if __name__ == "__main__":
	print(upgrade())
