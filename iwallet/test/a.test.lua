--- main 合约主入口
-- 输出hello world
-- @gas_limit 11
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
function main()
    Put("hello", "world")
    return "success"
end
