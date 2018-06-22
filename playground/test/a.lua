--- main
-- 输出hello world
-- @gas_limit 10000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
    Transfer("walleta", "walletb", 100)
    return "success"
end--f

--- hello
-- 输出hello
-- @gas_limit 10000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
function hello()
    print("world")
    return true
end--f
