#! /usr/bin/env python
# -*- coding:utf-8 -*-

import subprocess
import os
import time

server_addr = [["Tokyo", "18.179.143.193","1"],
["London", "52.56.118.10", "2"],
["Singapore", "13.228.206.188", "3"],
["Mumbai", "13.232.96.221", "4"],
["Frankfort", "18.184.239.232", "5"],
["Seoul", "13.124.172.86", "6"],
["Montreal", "52.60.163.60","7"]]

#fout = open("out.txt","w")

balance_Map = {1:{}}
for item in server_addr:
    rtn = subprocess.check_output("curl -s --connect-timeout 3 -XPOST {}:{}/scripts -d \"cmd=checkredis.sh\"".format(item[1], 30310), 
            shell=True, stderr=subprocess.DEVNULL)
    lines = rtn.decode().split("\n")
    for i in range(0, len(lines)-1, 2):
        key, val = lines[i:i+2]
        if not key in balance_Map[1]:
            balance_Map[1][key] = {}
        if not val in balance_Map[1][key]:
            balance_Map[1][key][val]= [] 
        balance_Map[1][key][val].append(item[2])

for key in balance_Map[1]:
    if len(balance_Map[1][key]) > 1:
        print(key, balance_Map[1][key])
        

    #print(rtn.decode(), file=fout)
