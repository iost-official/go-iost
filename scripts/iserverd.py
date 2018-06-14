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

#need to be updated
#you can just check the ports owned by youself if you are not root
#0:port is occupied
#1:port isn't occupied
def check_port(port):
	return 0
	ret=wCommand("netstat -tunlp|grep "+str(port))
	ret=ret.split("\n")
	cnt=0
	for line in ret:
		li=line.strip()
		if li.endswith("iserver"):
			cnt+=1
			for status in li.split(" "):
				if status.endswith("LISTEN"):
					cnt+=1
	return (0 if cnt==2 else 1)


def has_proc(pn):
	ret=wCommand("ps -e|grep "+pn+"|grep -v grep|grep -v iserverd.py")
	lines=ret.split("\n")
	for line in lines:
		if line.strip().endswith("iserver"):
			if check_port(30301)+check_port(30303)==0:
				return 1
			else:
				return 0
	return 0

#1:proc exist
#0:proc not exist
def exist():
	if has_proc("iserver")==1:
		return 0
	return 1

def start():
	if exist()==0:
		return 1
	wCommand("nohup iserver --config /workdir/iserver.yml >> test.log 2>&1 &")
	return 0

#0:success
#1:fail
#todo 判断iserver是否存在，用PID
def restart():
	for i in range(0,3):
		if(start()!=0):
			a=wCommand("ps -ax|grep iserver|grep -v grep|grep -v iserverd.py|grep -v defunct|awk 'NR==1{print $1}'")
			wCommand("kill -9 "+a)

			time.sleep(1)
		else:
			return 0
	return 1

def stop():
	for i in range(0,3):
		if exist()==0:
			a=wCommand("ps -ax|grep iserver|grep -v grep|grep -v iserverd.py|grep -v defunct|awk 'NR==1{print $1}'")
			wCommand("kill -9 "+a)
			time.sleep(1)
		else:
			return 0
	return 1


def upgrade():
	wCommand("cd "+pwd+" && git checkout .")
	wCommand("cd "+pwd+" && git checkout testnet")
	wCommand("cd "+pwd+" && git checkout consensus")
	#wCommand("cd "+pwd+" && git reset --hard origin/testnet")
	wCommand("cd "+pwd+" && git reset --hard origin/consensus")
	wCommand("cd "+pwd+" && git pull")
	wCommand("cd "+pwd+"/iserver && go install")
#	wCommand("rm -rf /workdir/blockDB /workdir/txDB/ /workdir/netpath /workdir/test.log /workdir/dump.rdb")

	#stop iserver now
#	if(stop()!=0):
#		return 1
	#ret=wCommand("nohup redis-server &")
	#delete dump.rdb
#	return start()
	return 0

func={
	"start":start,
	"stop":stop,
	"restart":restart,
	"upgrade":upgrade,
	"exist":exist,
}
if __name__ == "__main__":
	if(len(sys.argv)!=2):
		sys.exit(1)
	com=sys.argv[1]
	if com not in func.keys():
		sys.exit(2)

	print("FAIL" if func[com]()!=0 else "SUCCESS")
	sys.exit(0)
