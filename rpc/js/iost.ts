import {grpc} from "grpc-web-client";

import {Cli} from "./generated/cli_pb_service";
import {Key, TransactionKey} from "./generated/cli_pb";

export class IOST {

    node_url: string;

    GetTransaction(publisher: string, nonce: number, callback: (t: Object) => void) {
        const tk = new TransactionKey();
        tk.setPublisher(publisher);
        tk.setNonce(nonce);
        grpc.unary(Cli.GetTransaction, {
            request: tk,
            host: this.node_url,
            onEnd: output => {
                callback(output.message.toObject())
            }
        })
    }


    GetBalance(account: string, callback: (res) => void) {
        const key = new Key();
        key.setS(account);
        grpc.unary(Cli.GetBalance, {
                request: key,
                host: this.node_url,
                onEnd: res => {
                    const {status, statusMessage, headers, message, trailers} = res;
                    callback(res)
                },
            }
        )
    }
}