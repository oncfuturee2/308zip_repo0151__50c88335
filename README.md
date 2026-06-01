# 分销佣金结算与提现审核系统 MVP

基于 Go 1.22 + Gin + GORM + PostgreSQL + Redis 的社交电商分销佣金系统 MVP 版本。

## 核心业务闭环

```
创建已完成订单 → 按固定比例生成佣金 → 结算后增加可提现余额 → 分销员提现 → 财务审核通过并 mock 打款
```

## 技术栈

- **Go 1.22** - 后端语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **PostgreSQL** - 关系型数据库
- **Redis** - 缓存（用于幂等控制）
- **JWT** - 身份认证
- **简单 RBAC** - 角色权限控制

## 项目结构

```
.
├── cmd/
│   └── main.go              # 程序入口
├── internal/
│   ├── config/              # 配置管理
│   │   └── config.go
│   ├── database/            # 数据库连接与迁移
│   │   ├── postgres.go
│   │   ├── redis.go
│   │   ├── migration.go
│   │   └── seed.go
│   ├── models/              # 数据模型
│   │   ├── user.go
│   │   ├── distributor.go
│   │   ├── order.go
│   │   ├── commission_record.go
│   │   ├── withdraw_request.go
│   │   └── audit_log.go
│   ├── middleware/          # 中间件
│   │   └── auth.go          # JWT认证 + RBAC
│   ├── service/             # 业务逻辑层
│   │   ├── user_service.go
│   │   ├── distributor_service.go
│   │   ├── order_service.go
│   │   ├── commission_service.go
│   │   ├── withdraw_service.go
│   │   ├── audit_log_service.go
│   │   ├── idempotent.go    # Redis幂等服务
│   │   └── payout_provider.go # Mock打款服务
│   ├── handler/             # HTTP处理器
│   │   ├── auth_handler.go
│   │   ├── order_handler.go
│   │   ├── commission_handler.go
│   │   ├── withdraw_handler.go
│   │   └── audit_log_handler.go
│   ├── router/              # 路由配置
│   │   └── router.go
│   └── pkg/                 # 公共包
│       ├── jwt/
│       ├── response/
│       └── utils/
├── Dockerfile               # Docker构建配置
├── supervisord.conf         # 服务管理配置
├── wait-for-services.sh     # 服务等待脚本
├── .dockerignore            # Docker忽略文件
├── go.mod
└── README.md
```

## 数据模型（6张表）

### 1. users - 用户表
- `id` - 主键
- `username` - 用户名（唯一）
- `password` - 密码（加密）
- `role` - 角色：`distributor` / `finance` / `admin`

### 2. distributors - 分销员表
- `id` - 主键
- `user_id` - 用户ID（唯一）
- `parent_id` - 上级分销员ID（支持三级分销层级）
- `real_name` - 真实姓名
- `phone` - 手机号
- `bank_card_no` - 银行卡号
- `bank_name` - 银行名称
- `balance` - 可提现余额（分）
- `frozen_balance` - 冻结余额（分）
- `total_commission` - 累计佣金（分）
- `status` - 状态

### 3. orders - 订单表
- `id` - 主键
- `order_no` - 订单号（唯一索引）
- `distributor_id` - 分销员ID
- `amount` - 订单金额（分）
- `goods_name` - 商品名称
- `status` - 订单状态

### 4. commission_records - 佣金记录表
- `id` - 主键
- `order_no` - 订单号（与 distributor_id 组成联合唯一索引，保证同订单同分销员幂等）
- `distributor_id` - 分销员ID
- `level` - 佣金层级：1-一级（直接） 2-二级（上级） 3-三级（上上级）
- `order_amount` - 订单金额（分）
- `rate` - 佣金比例（%）
- `amount` - 佣金金额（分）
- `status` - 状态：`pending` / `settled`

### 5. withdraw_requests - 提现申请表
- `id` - 主键
- `request_no` - 申请单号
- `distributor_id` - 分销员ID
- `amount` - 提现金额（分）
- `bank_card_no` - 银行卡号
- `bank_name` - 银行名称
- `real_name` - 真实姓名
- `status` - 状态：`pending` / `approved` / `rejected` / `paid_mock`
- `reject_reason` - 驳回原因

### 6. audit_logs - 审计日志表
- `id` - 主键
- `distributor_id` - 分销员ID
- `operator_id` - 操作人ID
- `resource_type` - 资源类型：`withdraw` / `commission`
- `resource_id` - 资源ID
- `action` - 操作类型
- `old_status` - 原状态
- `new_status` - 新状态
- `change_reason` - 变更原因

## API 接口（12个）

### 公共接口
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/v1/login | 登录 | 无 |
| GET | /api/v1/profile | 获取当前用户信息 | 登录 |

### 分销员接口
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/v1/orders | 创建mock完成订单 | distributor |
| GET | /api/v1/orders | 订单列表 | distributor |
| GET | /api/v1/commissions | 佣金列表 | distributor |
| GET | /api/v1/balance | 可提现余额 | distributor |
| POST | /api/v1/withdraws | 提交提现申请 | distributor |
| GET | /api/v1/withdraws | 提现列表（查看自己的） | distributor |

### 财务/管理员接口
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/v1/distributors | 创建新分销员 | admin/finance |
| POST | /api/v1/commissions/generate | 生成佣金（订单归因） | 登录 |
| POST | /api/v1/commissions/settle | 结算佣金 | finance/admin |
| POST | /api/v1/withdraws/approve | 审核提现通过 | finance/admin |
| POST | /api/v1/withdraws/reject | 审核提现驳回 | finance/admin |
| GET | /api/v1/withdraws | 提现列表（查看全部） | finance/admin |
| GET | /api/v1/audit-logs | 审计日志查询 | finance/admin |

## 企业级特性

### 1. 幂等订单归因

**订单幂等 Key（Redis）：**
```
idempotent:order:{order_no}   - 订单创建幂等锁
idempotent:commission:{order_no} - 佣金生成幂等锁
```

**幂等机制说明：**
- 使用 Redis `SETNX` 实现分布式锁
- Key 过期时间：24 小时
- 重复创建相同 `order_no` 的订单或重复生成佣金，不会产生两条记录

### 2. 提现状态机

**状态流转图：**
```
pending → approved → paid_mock
    ↓
  rejected
```

**合法的状态跳转：**
| 当前状态 | 允许跳转的目标状态 |
|----------|-------------------|
| pending | approved, rejected |
| approved | paid_mock |
| rejected | 无 |
| paid_mock | 无 |

**非法跳转会返回错误：** "invalid state transition"

### 3. Mock Provider

- **MockPayoutProvider** - 模拟打款服务，审核通过后自动调用
- **Mock 银行卡校验** - 银行卡信息校验（始终返回成功）

## Seed 账号

系统启动时自动创建以下测试账号：

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| 财务 | finance | finance123 | 可审核提现、查看审计日志 |
| 分销员 | distributor1 | dist123 | 可创建订单、提交提现 |

## 快速启动

### Docker 单文件启动

```bash
# 构建镜像
docker build -t review-system .

# 运行容器（使用分配的宿主机端口 18080）
docker run -d -p 18080:8080 --name docker-question-080 review-system
```

服务地址：
- API: http://127.0.0.1:18080

### 本地运行

1. 启动 PostgreSQL 和 Redis
2. 复制环境变量文件：
```bash
cp .env.example .env
```
3. 修改 `.env` 中的数据库连接配置
4. 运行：
```bash
go run cmd/main.go
```

## 验收步骤

### 步骤 1：登录并获取 Token

**分销员登录：**
```bash
curl -X POST http://127.0.0.1:18080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"distributor1","password":"dist123"}'
```

**财务登录：**
```bash
curl -X POST http://127.0.0.1:18080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"finance","password":"finance123"}'
```

### 步骤 2：创建订单

使用分销员 Token 创建已完成订单：
```bash
curl -X POST http://127.0.0.1:18080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <distributor_token>" \
  -d '{"order_no":"ORD20260502TEST001","amount":10000,"goods_name":"测试商品"}'
```
> `amount: 10000` 表示 100.00 元（整数分）

### 步骤 3：重复创建订单（幂等验证）

使用相同的 `order_no` 再次创建订单：
```bash
curl -X POST http://127.0.0.1:18080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <distributor_token>" \
  -d '{"order_no":"ORD20260502TEST001","amount":10000,"goods_name":"测试商品"}'
```
> 应该返回已存在的订单信息，不会创建新订单

### 步骤 4：生成佣金

```bash
curl -X POST http://127.0.0.1:18080/api/v1/commissions/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <distributor_token>" \
  -d '{"order_no":"ORD20260502TEST001"}'
```
> 默认三级佣金比例：一级10%（直接分销员）、二级5%（上级）、三级3%（上上级），
> 10000 分订单将生成多条佣金记录（若有上级层级关系）

### 步骤 5：重复生成佣金（幂等验证）

再次调用生成佣金接口：
```bash
curl -X POST http://127.0.0.1:18080/api/v1/commissions/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <distributor_token>" \
  -d '{"order_no":"ORD20260502TEST001"}'
```
> 应该返回已存在的佣金记录，不会重复入账

### 步骤 6：财务结算佣金

使用财务 Token 结算佣金：
```bash
# 先查询佣金列表获取 commission_id
curl -X GET "http://127.0.0.1:18080/api/v1/commissions?page=1&page_size=10" \
  -H "Authorization: Bearer <distributor_token>"

# 结算佣金（需要财务/管理员权限）
curl -X POST http://127.0.0.1:18080/api/v1/commissions/settle \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <finance_token>" \
  -d '{"commission_id":1}'
```

### 步骤 7：查看可提现余额

```bash
curl -X GET http://127.0.0.1:18080/api/v1/balance \
  -H "Authorization: Bearer <distributor_token>"
```
> 结算后余额应该增加

### 步骤 8：分销员提交提现

```bash
curl -X POST http://127.0.0.1:18080/api/v1/withdraws \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <distributor_token>" \
  -d '{"amount":500}'
```
> 提现 500 分（5元），提交后余额减少，冻结余额增加

### 步骤 9：财务审核通过

```bash
# 先查询提现列表获取 withdraw_id
curl -X GET "http://127.0.0.1:18080/api/v1/withdraws?status=pending" \
  -H "Authorization: Bearer <finance_token>"

# 审核通过
curl -X POST http://127.0.0.1:18080/api/v1/withdraws/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <finance_token>" \
  -d '{"withdraw_id":1}'
```
> 审核通过后：
> 1. 状态变为 `approved`
> 2. 自动调用 Mock Payout Provider
> 3. 状态变为 `paid_mock`
> 4. 写入审计日志

### 步骤 10：查看审计日志

```bash
curl -X GET "http://127.0.0.1:18080/api/v1/audit-logs?resource_type=withdraw" \
  -H "Authorization: Bearer <finance_token>"
```

## 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| DB_HOST | PostgreSQL 主机 | localhost |
| DB_PORT | PostgreSQL 端口 | 5432 |
| DB_USER | PostgreSQL 用户 | postgres |
| DB_PASSWORD | PostgreSQL 密码 | postgres123 |
| DB_NAME | PostgreSQL 数据库名 | distribution |
| REDIS_HOST | Redis 主机 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| REDIS_PASSWORD | Redis 密码 | 空 |
| JWT_SECRET | JWT 密钥 | your_jwt_secret_key_here |
| JWT_EXPIRE_HOURS | JWT 过期时间（小时） | 24 |
| COMMISSION_LEVEL1_RATE | 一级佣金比例（%） | 10 |
| COMMISSION_LEVEL2_RATE | 二级佣金比例（%） | 5 |
| COMMISSION_LEVEL3_RATE | 三级佣金比例（%） | 3 |

## 注意事项

1. **金额单位**：所有金额均使用**整数分**存储，避免浮点数精度问题
2. **Mock 打款**：`paid_mock` 状态表示已模拟打款，实际生产环境需要对接真实支付渠道
3. **幂等 Key**：Redis 幂等 Key 24小时后过期，确保相同 order_no 在短时间内不会重复处理
4. **状态机**：提现状态流转严格限制，非法跳转返回错误

## 不包含的功能（MVP 简化）

- ❌ 退款冲减
- ❌ 风控人工复核
- ❌ 真实打款（仅 Mock）
- ❌ 月度大批量锁账
- ❌ 银行卡真实校验（仅 Mock）
