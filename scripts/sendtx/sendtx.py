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
project_path=GOPATH+"/src/github.com/iost-official/prototype/"
def wCommand(com):
	obj = subprocess.Popen([com], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,shell=True)
	obj.wait()
	ret=obj.stdout.read()
	obj.stdout.close()
	return ret

def has(fn):
	return os.path.exists(fn)

def Buildwallet():
 	ret=wCommand("cd "+project_path+"iwallet;go build")
	ret=wCommand("cp "+project_path+"iwallet/iwallet "+cur_path)
	return True

def GenAccounts():
	if Buildwallet()==False:return False
	accounts=[]
	for i in range(0,100):
		wCommand(cur_path+"iwallet account -c test")
		f=open(HOME+"/.ssh/test_secp","r")
		seckey=f.read()
		f.close()
		f=open(HOME+"/.ssh/test_secp.pub","r")
		pubkey=f.read()
		f.close()
		accounts.extend([pubkey+'\n',seckey+'\n'])

	fd=open(cur_path+"acc_list.txt","w")
	fd.writelines(accounts)
	fd.close()
	return "SUCCESS"

def Sendmoney():
	fd=open(cur_path+"acc_list.txt","r")
	lines=fd.readlines()
	fd.close()
	lto=[]
	for i in range(0,len(lines),2):
		lto.append(lines[i][:-1])
	genesisPubkey="2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"
	genesisSeckey="BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"
	constructAll(genesisPubkey,lto,10000000000)
	f1=open(HOME+"/.ssh/test_secp","w")
	f1.write(genesisSeckey)
	f1.close()
	
	f2=open(HOME+"/.ssh/test_secp.pub","w")
	f2.write(genesisPubkey)
	f2.close()

	flist=[Compile,Sign,Publish,]
	for func in flist:
		if func()==False:
			return "FAIL"
	return "SUCCESS"



def construct(_from,_to,money):
	f=open(cur_path+"test/1to2.lua","w")
	f.writelines([
		'--- main 合约主入口\n',
		'-- server1转账server2\n',
		'-- @gas_limit 10000\n',
		'-- @gas_price 0.001\n',
		'-- @param_cnt 0\n',
		'-- @return_cnt 0\n',
		'function main()\n',
		'	Transfer("'+_from+'","'+_to+'",'+str(money)+')\n',
		'end\n',
	])
	f.close()

def constructAll(genesisPubkey,lto,money):
	f=open(cur_path+"test/1to2.lua","w")
	lines=[
		'--- main 合约主入口\n',
		'-- server1转账server2\n',
		'-- @gas_limit 10000\n',
		'-- @gas_price 0.001\n',
		'-- @param_cnt 0\n',
		'-- @return_cnt 0\n',
		'function main()\n',
	]
	for i in range(0,len(lto)):
		lines.append('	Transfer("'+genesisPubkey+'","'+lto[i]+'",'+str(money)+')\n')
	lines.append('end\n')
	f.writelines(lines)
	f.close()


def Contract():
	#open acc_list.txt
	#0.pubkey
	#1.seckey
	#...
	fd=open(cur_path+"acc_list.txt","r")
	lines=fd.readlines()
	fd.close()
	money=random.random()
	id0=0;id1=0;
	while id0==id1:
		id0=random.randint(0,len(lines)-1)
		id0=id0-(id0&1)
		id1=random.randint(0,len(lines)-1)
		id1=id1-(id1&1)

	construct(lines[id0][:-1],lines[id1][:-1],money)
	_f=open("/workdir/sendtx.log","a+")
	_f.write(lines[id0][:-1]+"  "+lines[id1][:-1]+"  "+str(money)+"\n")
	_f.close()

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
	if(len(sys.argv)!=2):
		sys.exit(1)

	com=sys.argv[1]
	if com=="genaccounts":
		print(GenAccounts())
		sys.exit(0)
	if com=="sendtoall":
		print(Sendmoney())
		sys.exit(0)
	if com=="sendtransaction":
		ans="SUCCESS"
		func_list=[Buildwallet,Contract,Compile,Sign,Publish,]
		while True:
			for func in func_list:
				if func()==False:
					ans="FAIL"
					break
			print(ans)
			time.sleep(0.1)

