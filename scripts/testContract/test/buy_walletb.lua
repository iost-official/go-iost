--- main 买1元
-- snatch treasure with 1 coin
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walletb
function main()
    for i=0,9 do
        Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "BuyCoin", "2538yUDuKTLaXqCTFS1tfVmMEL4dVnzLDWChoMdoxgCa4", 1) == 0)
    end
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "QueryClaimed") == "false")
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "Claim") == 0)
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "QueryClaimed") == "true")

    return "success";
end--f
