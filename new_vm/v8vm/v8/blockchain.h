#ifndef IOST_V8_BLOCKCHAIN_H
#define IOST_V8_BLOCKCHAIN_H

#include "v8.h"
#include "vm.h"
#include "stddef.h"

using namespace v8;

void InitBlockChain(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args);

class IOSTBlockchain {
private:
    SandboxPtr sbx;
public:
    IOSTBlockchain(SandboxPtr ptr): sbx(ptr) {}

    int Transfer(char *from, char *to, char *amount) {
        size_t gasCount = 0;
        int ret = goTransfer(sbx, from, to, amount, &gasCount);
        return ret;
    }
    int Withdraw(char *to, char *amount) {
        size_t gasCount = 0;
        int ret = goWithdraw(sbx, to, amount, &gasCount);
        return ret;
    }
    int Deposit(char *from, char *amount) {
        size_t gasCount = 0;
        int ret = goDeposit(sbx, from, amount, &gasCount);
        return ret;
    }
    int TopUp(char *contract, char *from, char *amount) {
        size_t gasCount = 0;
        int ret = goTopUp(sbx, contract, from, amount, &gasCount);
        return ret;
    }
    int Countermand(char *contract, char *to, char *amount) {
        size_t gasCount = 0;
        int ret = goCountermand(sbx, contract, to, amount, &gasCount);
        return ret;
    }
    char *BlockInfo() {
        size_t gasCount = 0;
        char *blkInfo = goBlockInfo(sbx, &gasCount);
        return blkInfo;
    }
    char *TxInfo() {
        size_t gasCount = 0;
        char *txInfo = goTxInfo(sbx, &gasCount);
        return txInfo;
    }
    char *Call(char *contract, char *api, char *args) {
        size_t gasCount = 0;
        char *result = goCall(sbx, contract, api, args, &gasCount);
        return result;
    }
};

#endif // IOST_V8_BLOCKCHAIN_H