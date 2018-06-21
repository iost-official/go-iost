--- main 买1元
-- snatch treasure with 1 coin
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walletb
function main()
    print("buy wallet")
    for i=0,9 do
        print(Call("main", "BuyCoin", "walletb", 1))
    end
    print(Call("main", "QueryWinner"))
    print(Call("main", "QueryClaimed"))
    print(Call("main", "Claim"))
    print(Call("main", "Claim"))
    print(Call("main", "QueryClaimed"))

    return "success";
end--f
