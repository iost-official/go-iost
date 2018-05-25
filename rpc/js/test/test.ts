import {IOST} from "../iost"

let iost = new IOST();
iost.node_url = "http://localhost:30303";
iost.GetBalance("tB4Bc8G7bMEJ3SqFPJtsuXXixbEUDXrYfE5xH4uFmHaV", function (res){
    console.log(res)
});