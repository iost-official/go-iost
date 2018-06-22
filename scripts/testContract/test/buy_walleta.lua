--- main 买1元
-- snatch treasure with 1 coin
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
    print("buy wallet")
    for i=0,9 do
        Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "BuyCoin", "gvCQNmkuA6AwdddRMSUg6jr8W7swKWAnhEY3cAthj9bX", 1) == 0)
    end
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "QueryWinner") == "gvCQNmkuA6AwdddRMSUg6jr8W7swKWAnhEY3cAthj9bX")
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "QueryClaimed") == "false")
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "Claim") ~= 0)
    Assert(Call("7tvZekXE5eQvupDf28nMJi2Dct8kMMT2hg3NUVrtZPrY", "QueryClaimed") == "false")

	return 0
end--f
