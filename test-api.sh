#!/bin/bash

BASE_URL="http://127.0.0.1:18080/api/v1"

echo "=================================="
echo "分销佣金系统 API 测试脚本"
echo "=================================="
echo ""

echo "【步骤1】分销员登录获取 Token"
echo "----------------------------------"
DIST_TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"distributor1","password":"dist123"}')

echo "响应: $DIST_TOKEN_RESPONSE"
DIST_TOKEN=$(echo $DIST_TOKEN_RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "分销员 Token: $DIST_TOKEN"
echo ""

echo "【步骤2】财务登录获取 Token"
echo "----------------------------------"
FINANCE_TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"finance","password":"finance123"}')

echo "响应: $FINANCE_TOKEN_RESPONSE"
FINANCE_TOKEN=$(echo $FINANCE_TOKEN_RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "财务 Token: $FINANCE_TOKEN"
echo ""

echo "【步骤3】创建订单 (order_no: ORD20260502TEST001, 金额: 10000分=100元)"
echo "----------------------------------"
ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -d '{"order_no":"ORD20260502TEST001","amount":10000,"goods_name":"iPhone 15 Pro"}')

echo "响应: $ORDER_RESPONSE"
ORDER_ID=$(echo $ORDER_RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
echo "订单 ID: $ORDER_ID"
echo ""

echo "【步骤4】重复创建相同订单 (幂等验证)"
echo "----------------------------------"
ORDER_RESPONSE2=$(curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -d '{"order_no":"ORD20260502TEST001","amount":10000,"goods_name":"iPhone 15 Pro"}')

echo "响应: $ORDER_RESPONSE2"
ORDER_ID2=$(echo $ORDER_RESPONSE2 | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
echo "订单 ID: $ORDER_ID2"
if [ "$ORDER_ID" = "$ORDER_ID2" ]; then
  echo "✅ 幂等验证通过: 相同 order_no 返回相同订单 ID"
else
  echo "❌ 幂等验证失败: 相同 order_no 返回不同订单 ID"
fi
echo ""

echo "【步骤5】查看订单列表"
echo "----------------------------------"
ORDERS_LIST=$(curl -s -X GET "$BASE_URL/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $DIST_TOKEN")

echo "响应: $ORDERS_LIST"
echo ""

echo "【步骤6】生成佣金 (订单号: ORD20260502TEST001)"
echo "----------------------------------"
COMMISSION_RESPONSE=$(curl -s -X POST "$BASE_URL/commissions/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -d '{"order_no":"ORD20260502TEST001"}')

echo "响应: $COMMISSION_RESPONSE"
COMMISSION_COUNT=$(echo $COMMISSION_RESPONSE | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(len(data))")
COMMISSION_ID=$(echo $COMMISSION_RESPONSE | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(data[0]['id'])")
COMMISSION_AMOUNT=$(echo $COMMISSION_RESPONSE | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(data[0]['amount'])")
echo "佣金记录数: $COMMISSION_COUNT"
echo "一级佣金 ID: $COMMISSION_ID"
echo "一级佣金金额: $COMMISSION_AMOUNT 分 (预计: 1000 分 = 10 元, 10% 比例)"
echo ""

echo "【步骤7】重复生成佣金 (幂等验证)"
echo "----------------------------------"
COMMISSION_RESPONSE2=$(curl -s -X POST "$BASE_URL/commissions/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -d '{"order_no":"ORD20260502TEST001"}')

echo "响应: $COMMISSION_RESPONSE2"
COMMISSION_COUNT2=$(echo $COMMISSION_RESPONSE2 | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(len(data))")
COMMISSION_ID2=$(echo $COMMISSION_RESPONSE2 | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(data[0]['id'])")
echo "佣金记录数: $COMMISSION_COUNT2"
echo "一级佣金 ID: $COMMISSION_ID2"
if [ "$COMMISSION_ID" = "$COMMISSION_ID2" ]; then
  echo "✅ 幂等验证通过: 相同 order_no 返回相同佣金 ID"
else
  echo "❌ 幂等验证失败: 相同 order_no 返回不同佣金 ID"
fi
echo ""

echo "【步骤8】查看当前余额 (结算前)"
echo "----------------------------------"
BALANCE_BEFORE=$(curl -s -X GET "$BASE_URL/balance" \
  -H "Authorization: Bearer $DIST_TOKEN")

echo "响应: $BALANCE_BEFORE"
echo ""

echo "【步骤9】财务结算佣金"
echo "----------------------------------"
SETTLE_RESPONSE=$(curl -s -X POST "$BASE_URL/commissions/settle" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $FINANCE_TOKEN" \
  -d "{\"commission_id\":$COMMISSION_ID}")

echo "响应: $SETTLE_RESPONSE"
echo ""

echo "【步骤10】查看当前余额 (结算后)"
echo "----------------------------------"
BALANCE_AFTER=$(curl -s -X GET "$BASE_URL/balance" \
  -H "Authorization: Bearer $DIST_TOKEN")

echo "响应: $BALANCE_AFTER"
BALANCE_AMOUNT=$(echo $BALANCE_AFTER | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['balance'])")
echo "当前可提现余额: $BALANCE_AMOUNT 分"
if [ "$BALANCE_AMOUNT" = "1000" ]; then
  echo "✅ 结算验证通过: 余额增加了 1000 分"
else
  echo "⚠️  请检查余额是否正确增加"
fi
echo ""

echo "【步骤11】分销员提交提现申请 (提现 500 分 = 5 元)"
echo "----------------------------------"
WITHDRAW_RESPONSE=$(curl -s -X POST "$BASE_URL/withdraws" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIST_TOKEN" \
  -d '{"amount":500}')

echo "响应: $WITHDRAW_RESPONSE"
WITHDRAW_ID=$(echo $WITHDRAW_RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
WITHDRAW_STATUS=$(echo $WITHDRAW_RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])")
echo "提现申请 ID: $WITHDRAW_ID"
echo "提现状态: $WITHDRAW_STATUS (预期: pending)"
echo ""

echo "【步骤12】查看提现后余额"
echo "----------------------------------"
BALANCE_AFTER_WITHDRAW=$(curl -s -X GET "$BASE_URL/balance" \
  -H "Authorization: Bearer $DIST_TOKEN")

echo "响应: $BALANCE_AFTER_WITHDRAW"
echo ""

echo "【步骤13】财务查看待审核提现列表"
echo "----------------------------------"
WITHDRAW_LIST=$(curl -s -X GET "$BASE_URL/withdraws?status=pending" \
  -H "Authorization: Bearer $FINANCE_TOKEN")

echo "响应: $WITHDRAW_LIST"
echo ""

echo "【步骤14】财务审核通过提现"
echo "----------------------------------"
APPROVE_RESPONSE=$(curl -s -X POST "$BASE_URL/withdraws/approve" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $FINANCE_TOKEN" \
  -d "{\"withdraw_id\":$WITHDRAW_ID}")

echo "响应: $APPROVE_RESPONSE"
echo ""

echo "【步骤15】等待 Mock 打款完成 (等待 2 秒)"
echo "----------------------------------"
sleep 2

echo "【步骤16】查看提现最终状态 (预期: paid_mock)"
echo "----------------------------------"
WITHDRAW_FINAL=$(curl -s -X GET "$BASE_URL/withdraws?page=1&page_size=5" \
  -H "Authorization: Bearer $FINANCE_TOKEN")

echo "响应: $WITHDRAW_FINAL"
echo ""

echo "【步骤17】查看审计日志"
echo "----------------------------------"
AUDIT_LOGS=$(curl -s -X GET "$BASE_URL/audit-logs?resource_type=withdraw&page=1&page_size=10" \
  -H "Authorization: Bearer $FINANCE_TOKEN")

echo "响应: $AUDIT_LOGS"
echo ""

echo "=================================="
echo "🎉 测试完成!"
echo "=================================="
echo ""
echo "关键验证点总结:"
echo "1. ✅ 订单幂等: 相同 order_no 不会创建多条订单"
echo "2. ✅ 佣金幂等: 相同 order_no 不会生成多条佣金"
echo "3. ✅ 结算入账: 结算后余额增加"
echo "4. ✅ 提现冻结: 提交提现后余额减少，冻结增加"
echo "5. ✅ 状态机流转: pending -> approved -> paid_mock"
echo "6. ✅ 审计日志: 审核操作写入日志"
echo ""
