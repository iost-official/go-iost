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
	ret=wCommand("cd $GOPATH/src/github.com/iost-official/prototype && git checkout develop && git pull")
	ret=wCommand("cd $GOPATH/src/github.com/iost-official/prototype/iserver && go build")
	return "SUCCESS"		

if __name__ == "__main__":
	print(upgrade())
