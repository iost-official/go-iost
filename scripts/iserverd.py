#! /usr/bin/env python
# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys

HOME=os.environ['HOME']
GOPATH=os.environ['GOPATH']
pwd="$GOPATH/src/github.com/iost-official/prototype"

def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,close_fds=True,shell=True)
	stdoutdata, stderrdata = obj.communicate()
	return stdoutdata

def has_proc(pn):
	ret=wCommand("ps -e|grep "+pn+"|grep -v grep|grep -v iserverd.py")
	lines=ret.split("\n")
	for line in lines:
		if line.endswith("iserver"):
			return 1
	return 0 

def start():
	if has_proc("iserver")!=0:
		return 1
	wCommand("cd "+pwd+"/iserver;nohup ./iserver >> test.log 2>&1 &")
	return 0

#0:success
#1:fail
def restart():
	for i in range(0,3):
		if(start()!=0):
			a=wCommand("ps -ax|grep iserver|grep -v grep|grep -v iserverd.py|awk 'NR==1{print $1}'")
			wCommand("kill TERM "+a)

			time.sleep(1)
		else:
			return 0
	return 1

def stop():
	for i in range(0,3):
		if(has_proc("iserver")!=0):
			a=wCommand("ps -ax|grep iserver|grep -v grep|grep -v iserverd.py|awk 'NR==1{print $1}'")
			wCommand("kill TERM "+a)
			time.sleep(1)
		else:
			return 0
	return 1


def upgrade():
	if(stop()!=0):
		return 1
	wCommand("cd "+pwd+" && git checkout develop && git pull")
	wCommand("cd "+pwd+"/iserver && go build")
	#ret=wCommand("nohup redis-server &")
	#delete dump.rdb
	return start()

func={
	"start":start,
	"stop":stop,
	"restart":restart,
	"upgrade":upgrade,
}
if __name__ == "__main__":
	if(len(sys.argv)!=2):
		sys.exit(1)
	com=sys.argv[1]
	if com not in func.keys():
		sys.exit(2)
	
	print(func[com]())
	sys.exit(0)
