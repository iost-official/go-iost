#! /usr/bin/env python
# -*- coding:utf-8 -*-
import subprocess
import os
import time
import sys
import random 

HOME=os.environ['HOME']
GOPATH=os.environ['GOPATH']
cur_path=GOPATH+"/src/github.com/iost-official/prototype/scripts/sendtx/"
def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,shell=True)
	obj.wait()
	ret=obj.stdout.read()
	obj.stdout.close()
	return ret

def has(fn):
	return os.path.exists(fn)

def Contract():
	#open acc_list.txt
	#pubkey
	#seckey
	#...
	fd=open(cur_path+"acc_list.txt")
	lines=fd.readlines()
	fd.close()
	money=random.random()
	id0=0;id1=0;
	while id0==id1:
		id0=random.randint(0,len(lines)-1)
		id0=id0-(id0&1)
		id1=random.randint(0,len(lines)-1)
		id1=id1-(id1&1)

	f=open(cur_path+"test/1to2.lua","w")
	f.writelines([
		'--- main 合约主入口\n',
		'-- server1转账server2\n',
		'-- @gas_limit 10000\n',
		'-- @gas_price 0.001\n',
		'-- @param_cnt 0\n',
		'-- @return_cnt 0\n',
		'function main()\n',
		'	Transfer("'+lines[id0][:-1]+'","'+lines[id1][:-1]+'",'+str(money)+')\n',
		'end\n',
	])
	f.close()
	#write pubkey and seckey to ~/.ssh/test_secp
	f1=open(HOME+"/.ssh/test_secp","w")
	f1.write(lines[id0+1][:-1])
	f1.close()	
	
	f2=open(HOME+"/.ssh/test_secp.pub","w")
	f2.write(lines[id0][:-1])
	f2.close()
	return True
#TODO iwallet 应该使用最新版本编译的
#TODO 所有文件路径都应该是绝对地址，用函数封装一下
def Compile():
	#print "[iwallet compile]:",
	wCommand("rm -f "+cur_path+"test/1to2.sc")
	wCommand(cur_path+"iwallet compile -n "+str(random.randint(0,sys.maxint))+" "+cur_path+"test/1to2.lua")
	if has(cur_path+"test/1to2.sc"):
		#print("ok")
		return True 
	#print("fail")
	return False

def Sign():
	#print "[iwallet sign]:",
	wCommand("rm -f "+cur_path+"test/1to2.sig")
	ret=wCommand(cur_path+"iwallet sign "+cur_path+"test/1to2.sc -k ~/.ssh/test_secp")
	if has(cur_path+"test/1to2.sig"):
		#print("ok")
		return True 
	#print("fail")
	return False

def Publish():
	#print "[iwallet publish]:",
	ret=wCommand(cur_path+"iwallet publish "+cur_path+"test/1to2.sc "+cur_path+"test/1to2.sig -k ~/.ssh/test_secp")
	if ret.startswith("ok"):
		#check balance here
		#print("ok")
		return True
	#print("fail")
	return False

if __name__ == "__main__":
	ans="SUCCESS"
	func_list=[Contract,Compile,Sign,Publish,]
	for func in func_list:
		if func()==False:
			ans="FAIL"
			break
	print(ans)
