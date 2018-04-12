#### 概况

1. 搭建极小化测试环境
2. 迭代完善交易模块

#### 使用方式

单节点在命令行中交互运行，即输入一行指令，节点执行一个指令

数据存储在内存中

```
Usage:
  getbalance ADDRESS - Get balance of ADDRESS
  createblockchain ADDRESS - Create a blockchain and send genesis block reward to ADDRESS
  printchain - Print all the blocks of the blockchain
  send FROM TO AMOUNT - Send AMOUNT of coins from FROM address to TO
  exit - Exit the program
```

#### 使用样例
```
>> createblockchain yanxuecan
Done!
>> getbalance yanxuecan
Balance of 'yanxuecan': 10
>> send yanxuecan batman 4
Success!
>> getbalance batman
Balance of 'batman': 4
>> printchain
Prev hash: a8d887a4b3c6442c1b13db41c65dc81c5b1bfaf097b954685594b06070326f19
Hash: 9db4fd57a0397f32a97ab1f4bc738abc5ef22a80897dce7cbfdabc2f26297c29
--- Transaction eff2466f829bfa44dc4ded6c795512756bebc4ce282ac08ae24f40d6b02db2c2:
     Input 0:
       TXID:      1b1dc2af451892fa324af435eb275db6d2840d1f7ed5a66f4b4a3a9a216948c7
       Out:       0
       ScriptSig: 79616e78756563616e
     Output 0:
       Value:  4
       ScriptPubKey: 6261746d616e
     Output 1:
       Value:  6
       ScriptPubKey: 79616e78756563616e
Prev hash:
Hash: a8d887a4b3c6442c1b13db41c65dc81c5b1bfaf097b954685594b06070326f19
--- Transaction 1b1dc2af451892fa324af435eb275db6d2840d1f7ed5a66f4b4a3a9a216948c7:
     Input 0:
       TXID:
       Out:       -1
       ScriptSig: 5468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73
     Output 0:
       Value:  10
       ScriptPubKey: 79616e78756563616e
>> exit
```