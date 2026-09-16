# 签到信息管理与账户凭证拆分 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把签到 Beacon 参数收到可引用的 `igo_checkin_infos`，把 TraceInt Cookie 与微信签到 Token 收到 `igo_accounts`，让 `/checkin` 与一条龙按 ID 引用，并支持占座号 ≠ 打卡号。

**Architecture:** 下游 `igo` 插件新增账户/签到信息两张表与 CRUD+sign API；一条龙只存 `account_id` / `checkin_account_id` / `checkin_info_id`；启动幂等回填旧会话、旧卡片和 settings profile。抢座链路继续读写 `igo_sessions`。前端 `/checkin` 改成账户+签到信息管理，一条龙向导改为选引用。

**Tech Stack:** Go (Gin, GORM, Goose SQLite/Postgres), Next.js 16 (React 19, shadcn/ui, next-intl, Playwright), snowflake `idgen`.

**Spec:** `docs/superpowers/specs/2026-09-16-checkin-info-and-accounts-design.md`

## Global Constraints

- 只改下游 `backend/igo-lib/plugins/igo` 与产品 `frontend/`。严禁改 `backend/core/`、`backend/pkg/`、上游 `backend/plugins/`。
- 表结构只用 Goose SQL（postgres + sqlite），禁止 GORM AutoMigrate；无物理外键；Go 零值与库默认值一致；表名 `igo_*`。
- 不改抢座/占座/捡漏/明日预约对 `igo_sessions` 的读写。
- 不删一条龙旧列，不删旧 `/checkin/session|from-code|devices|sign|profiles` 路由。
- 不覆盖工作区未提交的其它 pipeline 改动：只加引用字段与读账户/签到信息，不重写无关执行逻辑。
- Cookie / Token 不回 JSON 明文。创建资源 HTTP 201 + `response.Created` Location。删除成功 204。
- 前端 `bun` + Biome，文案走 `next-intl`（zh-CN / en）。
- 测试用 `t.TempDir()`，凭据占位符（`mock_cookie`、`mock_token`）。
- 每任务结束只 `git add` 本任务文件并 Conventional Commit，禁止 `git add .`，禁止 push。

## File map

| 文件 | 职责 |
|---|---|
| `migrations/{sqlite,postgres}/00003_accounts_and_checkin_infos.sql` | 建表 + pipeline 加列 |
| `consts/consts.go` | 表名、`CodeNeedAuth` |
| `model/entity/account.go` `model/entity/checkin_info.go` | GORM 映射 |
| `model/entity/pipeline.go` | 三个引用字段 |
| `model/do/account.go` `model/do/checkin_info.go` | API DTO |
| `dao/account.go` `dao/checkin_info.go` | 持久化 |
| `service/account.go` `service/checkin_info.go` `service/backfill.go` | 业务 |
| `controller/account.go` `controller/checkin_info.go` | HTTP |
| `plugin.go` | 路由 + DB Bind 后回填 |
| `bot/login_auth.go` `bot/checkin_auth.go` | 补录写账户 |
| `frontend/lib/services/igo/{account.ts,checkin.ts,types.ts}` | client |
| `frontend/app/(main)/checkin/` `components/igo/checkin/` | 新 UI |
| `frontend/app/(main)/pipeline/components/` | 向导改引用 |
| `frontend/e2e/checkin.spec.ts` `frontend/e2e/pipeline.spec.ts` | E2E |
| `.agents/notes/implemented/architecture/2026-09-16-checkin-info-and-accounts.md` | 决策笔记 |

---

### Task 1: Goose 迁移、常量、Entity

**Files:**
- Create: `backend/igo-lib/plugins/igo/migrations/sqlite/00003_accounts_and_checkin_infos.sql`
- Create: `backend/igo-lib/plugins/igo/migrations/postgres/00003_accounts_and_checkin_infos.sql`
- Create: `backend/igo-lib/plugins/igo/model/entity/account.go`
- Create: `backend/igo-lib/plugins/igo/model/entity/checkin_info.go`
- Modify: `backend/igo-lib/plugins/igo/consts/consts.go`
- Modify: `backend/igo-lib/plugins/igo/model/entity/pipeline.go`
- Modify: `backend/igo-lib/plugins/igo/model/entity/table_test.go`

**Interfaces:**
- Produces: `entity.Account`, `entity.CheckInInfo`；`entity.PipelineConfig` 增加 `AccountID` `CheckinAccountID` `CheckinInfoID uint64`；`consts.TableAccounts` `TableCheckInInfos` `CodeNeedAuth`

- [ ] **Step 1: 写会失败的表名测试**

`table_test.go` 的 `got` map 增加：

```go
entity.Account{}.TableName():    {},
entity.CheckInInfo{}.TableName(): {},
```

`consts.OwnedTables` 暂不改。运行：

```
cd backend && go test ./igo-lib/plugins/igo/model/entity -run TestTableNamesMatchOwnedTables -count=1
```

Expected: FAIL（缺类型或 OwnedTables 长度不匹配）。

- [ ] **Step 2: 写迁移与类型**

SQLite `00003` Up：

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS igo_accounts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    cookie TEXT NOT NULL DEFAULT '',
    cookie_expires_at DATETIME,
    cookie_source VARCHAR(32) NOT NULL DEFAULT '',
    checkin_token TEXT NOT NULL DEFAULT '',
    checkin_expires_at DATETIME,
    nickname VARCHAR(128) NOT NULL DEFAULT '',
    school VARCHAR(128) NOT NULL DEFAULT '',
    student_name VARCHAR(128) NOT NULL DEFAULT '',
    student_number VARCHAR(64) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_igo_accounts_user ON igo_accounts (user_id);

CREATE TABLE IF NOT EXISTS igo_checkin_infos (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    beacon_uuid VARCHAR(64) NOT NULL DEFAULT '',
    major INTEGER NOT NULL DEFAULT 0,
    minor INTEGER NOT NULL DEFAULT 0,
    latitude VARCHAR(32) NOT NULL DEFAULT '',
    longitude VARCHAR(32) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_igo_checkin_infos_user ON igo_checkin_infos (user_id);

ALTER TABLE igo_pipeline_configs ADD COLUMN account_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE igo_pipeline_configs ADD COLUMN checkin_account_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE igo_pipeline_configs ADD COLUMN checkin_info_id BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS igo_checkin_infos;
DROP TABLE IF EXISTS igo_accounts;
-- +goose StatementEnd
```

Postgres 同样结构，时间列用 `TIMESTAMPTZ`；Down 额外：

```sql
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS checkin_info_id;
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS checkin_account_id;
ALTER TABLE igo_pipeline_configs DROP COLUMN IF EXISTS account_id;
DROP TABLE IF EXISTS igo_checkin_infos;
DROP TABLE IF EXISTS igo_accounts;
```

`consts.go`：

```go
TableAccounts     = "igo_accounts"
TableCheckInInfos = "igo_checkin_infos"
CodeNeedAuth      = "need_auth"
```

`OwnedTables` 追加这两项。`NeedAuthError`：

```go
type NeedAuthError struct {
	Kind      string
	AccountID uint64
	AuthURL   string
	Msg       string
}

func (e *NeedAuthError) Error() string { return e.Msg }
```

Entity 字段与 spec 3.1 / 3.2 一一对应，`json:"id,string"`，凭证列 `json:"-"`。Pipeline 三个 `uint64`：`gorm:"not null;default:0"`。

- [ ] **Step 3: 跑测试**

```
cd backend && go test ./igo-lib/plugins/igo/model/entity -count=1 && go test ./igo-lib/plugins/igo/dao -run TestSQLFilesCoverOwnedTables -count=1
```

Expected: PASS。

- [ ] **Step 4: Commit**

```
git add backend/igo-lib/plugins/igo/migrations/sqlite/00003_accounts_and_checkin_infos.sql \
  backend/igo-lib/plugins/igo/migrations/postgres/00003_accounts_and_checkin_infos.sql \
  backend/igo-lib/plugins/igo/consts/consts.go \
  backend/igo-lib/plugins/igo/model/entity/account.go \
  backend/igo-lib/plugins/igo/model/entity/checkin_info.go \
  backend/igo-lib/plugins/igo/model/entity/pipeline.go \
  backend/igo-lib/plugins/igo/model/entity/table_test.go
git commit -m "feat(igo): add accounts and checkin_infos schema"
```

---

### Task 2: Account DAO

**Files:**
- Create: `backend/igo-lib/plugins/igo/dao/account.go`
- Create: `backend/igo-lib/plugins/igo/dao/account_test.go`

**Interfaces:**
- Consumes: `entity.Account`，`openMigratedDB`（`dao/dao_test.go`）
- Produces:

```go
func CreateAccount(ctx context.Context, row *entity.Account) error
func GetAccountByUser(ctx context.Context, id, userID uint64) (*entity.Account, error) // miss → nil, nil
func ListAccountsByUser(ctx context.Context, userID uint64) ([]entity.Account, error) // created_at ASC
func UpdateAccount(ctx context.Context, row *entity.Account) error
func DeleteAccount(ctx context.Context, id, userID uint64) error
func FindAccountByCookie(ctx context.Context, userID uint64, cookie string) (*entity.Account, error)
func ListPipelineIDsByAccount(ctx context.Context, userID, accountID uint64) ([]string, error)
```

`ListPipelineIDsByAccount`：`account_id = ? OR checkin_account_id = ?`，同用户。

- [ ] **Step 1: 写失败测试 `dao/account_test.go`**

```go
func TestAccountCRUDAndIsolation(t *testing.T) {
	openMigratedDB(t)
	ctx := context.Background()
	a := &entity.Account{UserID: 1, Name: "甲", Cookie: "mock_cookie_a"}
	require.NoError(t, dao.CreateAccount(ctx, a))
	require.NotZero(t, a.ID)

	got, err := dao.GetAccountByUser(ctx, a.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "甲", got.Name)

	cross, err := dao.GetAccountByUser(ctx, a.ID, 2)
	require.NoError(t, err)
	assert.Nil(t, cross)

	dup, err := dao.FindAccountByCookie(ctx, 1, "mock_cookie_a")
	require.NoError(t, err)
	assert.Equal(t, a.ID, dup.ID)
}
```

另测：空 cookie 的 `FindAccountByCookie` 返回 nil（不去重空串）。

Run: `cd backend && go test ./igo-lib/plugins/igo/dao -run TestAccount -count=1`

Expected: FAIL compile（无 `CreateAccount`）。

- [ ] **Step 2: 实现 `dao/account.go`**

`CreateAccount` 调 `ensureID(&row.ID)`，填 `CreatedAt`/`UpdatedAt` UTC。`GetAccountByUser` 用 `id = ? AND user_id = ?`，`ErrRecordNotFound` → `nil, nil`。`FindAccountByCookie` 若 `strings.TrimSpace(cookie)==""` 直接 `nil, nil`。

- [ ] **Step 3: 再跑 DAO 测试** Expected: PASS。

- [ ] **Step 4: Commit** `feat(igo): add account dao`

---

### Task 3: Check-in info DAO

**Files:**
- Create: `backend/igo-lib/plugins/igo/dao/checkin_info.go`
- Create: `backend/igo-lib/plugins/igo/dao/checkin_info_test.go`

**Interfaces:**
- Produces:

```go
func CreateCheckInInfo(ctx context.Context, row *entity.CheckInInfo) error
func GetCheckInInfoByUser(ctx context.Context, id, userID uint64) (*entity.CheckInInfo, error)
func ListCheckInInfosByUser(ctx context.Context, userID uint64) ([]entity.CheckInInfo, error)
func UpdateCheckInInfo(ctx context.Context, row *entity.CheckInInfo) error
func DeleteCheckInInfo(ctx context.Context, id, userID uint64) error
func ListPipelineIDsByCheckInInfo(ctx context.Context, userID, infoID uint64) ([]string, error)
```

- [ ] **Step 1: 失败测试** 创建、跨用户 Get 为 nil、`ListPipelineIDsByCheckInInfo` 在 pipeline `checkin_info_id` 指向该行时返回卡片 ID。先插 `entity.PipelineConfig{ID:"card1", UserID:1, Name:"n", Cookie:"c", LibraryID:1, SeatKey:"s", CheckinInfoID: info.ID}`。

Run: `cd backend && go test ./igo-lib/plugins/igo/dao -run TestCheckInInfo -count=1` Expected: FAIL。

- [ ] **Step 2: 实现** 与 Account DAO 同一模式。

- [ ] **Step 3: 再跑** Expected: PASS。

- [ ] **Step 4: Commit** `feat(igo): add checkin info dao`

---

### Task 4: Account service + DO

**Files:**
- Create: `backend/igo-lib/plugins/igo/model/do/account.go`
- Create: `backend/igo-lib/plugins/igo/service/account.go`
- Create: `backend/igo-lib/plugins/igo/service/account_test.go`

**Interfaces:**
- Consumes: `dao.CreateAccount` 等；`s.resolveCookie` `s.resolveCheckinToken` `s.api().GetCheckInDevices`
- Produces:

```go
func (s *Service) ListAccounts(ctx context.Context, userID uint64) ([]do.AccountDTO, error)
func (s *Service) GetAccount(ctx context.Context, userID, id uint64) (*do.AccountDTO, error)
func (s *Service) CreateAccount(ctx context.Context, userID uint64, req do.CreateAccountRequest) (*do.AccountDTO, error)
func (s *Service) UpdateAccount(ctx context.Context, userID, id uint64, req do.UpdateAccountRequest) (*do.AccountDTO, error)
func (s *Service) DeleteAccount(ctx context.Context, userID, id uint64) error
func (s *Service) LoginAccount(ctx context.Context, userID, id uint64, req do.AccountLoginRequest) (*do.AccountDTO, error)
func (s *Service) AuthorizeAccountCheckin(ctx context.Context, userID, id uint64, req do.AccountCheckinAuthRequest) (*do.AccountCheckinAuthResponse, error)
```

DTO：

```go
type AccountDTO struct {
	ID               uint64     `json:"id,string"`
	Name             string     `json:"name"`
	HasCookie        bool       `json:"has_cookie"`
	CookieMasked     string     `json:"cookie_masked,omitempty"`
	CookieExpiresAt  *time.Time `json:"cookie_expires_at,omitempty"`
	HasCheckinToken  bool       `json:"has_checkin_token"`
	CheckinExpiresAt *time.Time `json:"checkin_expires_at,omitempty"`
	Nickname         string     `json:"nickname"`
	School           string     `json:"school"`
	StudentName      string     `json:"student_name"`
	StudentNumber    string     `json:"student_number"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
type CreateAccountRequest struct {
	Name         string `json:"name" binding:"required"`
	Cookie       string `json:"cookie"`
	CheckinToken string `json:"checkin_token"`
}
type UpdateAccountRequest struct {
	Name string `json:"name"`
}
type AccountLoginRequest struct {
	Code string `json:"code" binding:"required"`
}
type AccountCheckinAuthRequest struct {
	Code string `json:"code" binding:"required"`
}
type AccountCheckinAuthResponse struct {
	Account AccountDTO             `json:"account"`
	Device  *CheckInDeviceResponse `json:"device,omitempty"`
}
```

- [ ] **Step 1: 失败测试**（复用 `service/pipeline_test.go` 的 sqlite + mock api 方式；若过重，在 `account_test.go` 用 `openMigratedDB` 模式：把 `dao.SetDBService` 接到 service 测试里，mock `s.api`。）

最小用例：

1. `CreateAccount` 空 name → validation_error。
2. 创建 `name:"甲"` 无凭证 → `HasCookie false`。
3. 用户 2 Get 用户 1 的 ID → not_found。
4. 插入 pipeline `account_id=a.ID` 后 Delete → CodedError 409 `conflict`，Msg 含卡片 ID。
5. `AuthorizeAccountCheckin`：mock `GetCheckInDevices` 返回学号，账户 `StudentName` 被回填。

Run: `cd backend && go test ./igo-lib/plugins/igo/service -run TestAccount -count=1` Expected: FAIL。

- [ ] **Step 2: 实现**

- name `TrimSpace` 空 → 400。
- Create：可选 Cookie 走 `resolveCookie`；可选 CheckinToken 走 `resolveCheckinToken`，成功则 `GetCheckInDevices` 填资料（失败不阻断创建，资料留空）。
- Delete：`ListPipelineIDsByAccount` 非空 → `consts.NewError(409, consts.CodeConflict, "账户仍被一条龙引用: "+join最多5个ID)`。
- Login/Checkin-auth：先 `GetAccountByUser`，miss → 404。Checkin-auth 组装 `AccountCheckinAuthResponse`。
- `toAccountDTO` 用 `traceint.MaskCookie`。

- [ ] **Step 3: 再跑** Expected: PASS。

- [ ] **Step 4: Commit** `feat(igo): add account service`

---

### Task 5: Check-in info service + sign

**Files:**
- Create: `backend/igo-lib/plugins/igo/model/do/checkin_info.go`
- Create: `backend/igo-lib/plugins/igo/service/checkin_info.go`
- Create: `backend/igo-lib/plugins/igo/service/checkin_info_test.go`

**Interfaces:**
- Produces:

```go
func (s *Service) ListCheckInInfos(ctx context.Context, userID uint64) ([]do.CheckInInfoDTO, error)
func (s *Service) GetCheckInInfo(ctx context.Context, userID, id uint64) (*do.CheckInInfoDTO, error)
func (s *Service) CreateCheckInInfo(ctx context.Context, userID uint64, req do.CreateCheckInInfoRequest) (*do.CheckInInfoDTO, error)
func (s *Service) UpdateCheckInInfo(ctx context.Context, userID, id uint64, req do.UpdateCheckInInfoRequest) (*do.CheckInInfoDTO, error)
func (s *Service) DeleteCheckInInfo(ctx context.Context, userID, id uint64) error
func (s *Service) SignCheckInInfo(ctx context.Context, userID, infoID uint64, req do.SignCheckInInfoRequest) (*do.CheckInSignResponse, error)
```

```go
type CheckInInfoDTO struct {
	ID         uint64    `json:"id,string"`
	Name       string    `json:"name"`
	BeaconUUID string    `json:"beacon_uuid"`
	Major      int       `json:"major"`
	Minor      int       `json:"minor"`
	Latitude   string    `json:"latitude"`
	Longitude  string    `json:"longitude"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type CreateCheckInInfoRequest struct {
	Name       string `json:"name" binding:"required"`
	BeaconUUID string `json:"beacon_uuid"`
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}
type UpdateCheckInInfoRequest struct {
	Name       string `json:"name"`
	BeaconUUID string `json:"beacon_uuid"`
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}
type SignCheckInInfoRequest struct {
	AccountID           uint64 `json:"account_id,string" binding:"required"`
	ExpectedLibraryID   int    `json:"expected_library_id" binding:"required"`
	ExpectedLibraryName string `json:"expected_library_name"`
}
```

保存时若 UUID 非空则 `traceint.NormalizeUUID`，失败 400；major/minor 非 0 时校验 0–65535。允许全空 Beacon。

`SignCheckInInfo`：

1. 取 info、account，任一跨用户 → 404。
2. account.checkin_token 空，或 `GetCheckInDevices` 失败 → `*consts.NeedAuthError{Kind:"CHECKIN", AccountID: account.ID, AuthURL: consts.WeChatCheckinAuthURL, Msg:"签到凭据已失效，请更新该账户的签到凭证"}`。
3. UUID normalize 失败或 lat/lng 空 → 400「请先补全签到信息的 Beacon 与坐标」。
4. 组 `do.CheckInSignRequest`，调现有 `s.api().SignCheckIn`（与 `service/checkin.go` 相同的 serverTime 流程）。

- [ ] **Step 1: 失败测试**

- CRUD + 删除被 pipeline 引用 409。
- sign：完整 Beacon + mock SignCheckIn 成功。
- sign：空 token → `NeedAuthError` Kind CHECKIN。
- sign：缺 UUID → validation_error。
- sign：account 属用户 2 → not_found。

Run: `cd backend && go test ./igo-lib/plugins/igo/service -run TestCheckInInfo -count=1` Expected: FAIL。

- [ ] **Step 2: 实现 `checkin_info.go`** 不要改旧 `SignCheckIn(userID, req)`。

- [ ] **Step 3: 再跑** Expected: PASS。

- [ ] **Step 4: Commit** `feat(igo): add checkin info service and sign-by-id`

---

### Task 6: HTTP 路由

**Files:**
- Create: `backend/igo-lib/plugins/igo/controller/account.go`
- Create: `backend/igo-lib/plugins/igo/controller/checkin_info.go`
- Modify: `backend/igo-lib/plugins/igo/plugin.go`（只加路由，不删旧 checkin 路由）
- Modify: `backend/igo-lib/plugins/igo/controller/controller.go`（parseUint64ID + replyNeedAuth）

**Interfaces:**
- Consumes: Task 4/5 的 Service 方法
- Produces: 下列路由（均在 `consts.APIPrefix` + authMW）

```
GET    /accounts
POST   /accounts
GET    /accounts/:id
PUT    /accounts/:id
DELETE /accounts/:id
POST   /accounts/:id/login
POST   /accounts/:id/checkin-auth
GET    /checkin/infos
POST   /checkin/infos
GET    /checkin/infos/:id
PUT    /checkin/infos/:id
DELETE /checkin/infos/:id
POST   /checkin/infos/:id/sign
```

- [ ] **Step 1: `parseUint64ID`**

```go
func parseUint64ID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.AbortBadRequestWithCode(c, consts.CodeInvalidID, "ID 格式错误")
		return 0, false
	}
	return id, true
}
```

`reply` 增加：

```go
var na *consts.NeedAuthError
if errors.As(err, &na) {
	c.JSON(http.StatusConflict, gin.H{
		"error": gin.H{"code": consts.CodeNeedAuth, "message": na.Msg},
		"data": gin.H{
			"need_auth":  na.Kind,
			"account_id": strconv.FormatUint(na.AccountID, 10),
			"auth_url":   na.AuthURL,
		},
	})
	return true
}
```

Create 成功：`response.Created(c, "/api/v1/igo/accounts/"+id, dto)`（infos 同理）。

Handler 风格照抄 `controller/pipeline.go` 的 `withUser` + swagger 注释。

- [ ] **Step 2: 用现有 controller 测试方式或 `plugin_test` 断言路由注册**（若 `plugin_test.go` 有路由表快照，补上新路径）。至少：

```
cd backend && go test ./igo-lib/plugins/igo -count=1
cd backend && go test ./igo-lib/plugins/igo/controller -count=1
```

Expected: PASS。旧 `/checkin/session` 等仍注册。

- [ ] **Step 3: Commit** `feat(igo): expose account and checkin-info APIs`

---

### Task 7: 幂等回填

**Files:**
- Create: `backend/igo-lib/plugins/igo/service/backfill.go`
- Create: `backend/igo-lib/plugins/igo/service/backfill_test.go`
- Modify: `backend/igo-lib/plugins/igo/plugin.go`

**Interfaces:**
- Produces: `func (s *Service) BackfillAccountsAndCheckinInfos(ctx context.Context) error`

规则（失败只打日志，返回 error 给测试，Apply 里吞掉不阻断启动）：

1. 列出所有 `igo_sessions`（新增 `dao.ListAllSessions`，或 SQL `Find`）。Cookie 非空且该 `user_id` 还没有任何账户 → 建 `name:"默认"`，拷 Cookie/Expires。
2. 每个 `igo_checkin_sessions`：找该用户名为「默认」的账户，没有则最早创建的，再没有则新建「默认」。若目标 `checkin_token==""` 则写入 Token/Expires。
3. `ListAllPipelineConfigs` 中 `AccountID==0` 的行：`FindAccountByCookie` 复用或 `CreateAccount{Name: card.Name, Cookie: card.Cookie}`；若 Beacon 任一非空则 `CreateCheckInInfo{Name: card.Name+" 签到", ...}` 并设 `CheckinInfoID`；`CheckinToken` 非空且占座账户 token 空则写入账户；`UpdatePipelineConfig` 写 `AccountID`（`AutoCheckin` 保持）。
4. settings payload 里 `checkin_profiles`：每个 profile 若不存在同用户同名签到信息则创建（名称 `library_name` 或 `"场馆 {id}"`）。不改 JSON。

- [ ] **Step 1: 失败测试**

- 插入 session+checkin_session+pipeline（account_id=0, beacon 有值）+ settings profile。
- 调 `Backfill` 两次。
- 断言：账户 1 条（默认与卡片 cookie 相同则复用，总数稳定）；签到信息含卡片 Beacon 与场馆名；pipeline.account_id≠0；第二次 ID 不变。

Run: `cd backend && go test ./igo-lib/plugins/igo/service -run TestBackfill -count=1` Expected: FAIL。

- [ ] **Step 2: 实现 + `plugin.go`**

`BackfillAccountsAndCheckinInfos` 内部用 `sync.Once`（测试可导出 `ResetBackfillForTest`）。`ctx.Effect` 是 `OnDispose`，不能用来启动回填。

在现有 `ctx.Bind(dao.SetDBService)` 旁再绑一次 DB 就绪钩子：

```go
ctx.Bind(func(_ contracts.DBService) {
	if err := p.svc.BackfillAccountsAndCheckinInfos(ctx.GoContext()); err != nil {
		fmt.Printf("[igo] account backfill: %v\n", err)
	}
})
```

`RunPipeline` 在 `account_id==0` 时再调一次（Once 保证不重复搬数据，步骤 3 对剩余 `account_id=0` 行仍要能跑——因此 Once 只包 session/checkin_session/settings，pipeline 行扫描每次执行都做，或 Once 里做完全部 4 步且 RunPipeline 对单行再跑步骤 3）。推荐：函数每次扫描 `account_id=0` 的卡片（幂等），session/settings 用 Once。

需要 `dao.ListAllSessions` / `ListAllCheckInSessions`：写在 `dao/extra.go`。

- [ ] **Step 3: 再跑** Expected: PASS。

- [ ] **Step 4: Commit** `feat(igo): backfill accounts and checkin infos`

---

### Task 8: 一条龙改为引用账户/签到信息

**Files:**
- Modify: `backend/igo-lib/plugins/igo/model/do/pipeline.go`
- Modify: `backend/igo-lib/plugins/igo/service/pipeline.go`
- Modify: `backend/igo-lib/plugins/igo/service/pipeline_test.go`
- Modify: `backend/igo-lib/plugins/igo/dao/pipeline.go`（仅在缺 Update 列时补 `account_id` 等；不要改无关 WIP）
- Modify: `backend/igo-lib/plugins/igo/controller/pipeline.go`（layout 增加 account_id，最小改动）

**Interfaces:**
- `CreatePipelineConfigRequest`：`AccountID uint64 \`json:"account_id,string" binding:"required"\``，`CheckinAccountID`，`CheckinInfoID`；**删除** Cookie/CheckinToken/Beacon 的 binding（结构体去掉这些字段）。
- `UpdatePipelineConfigRequest`：三个 ID + 座位字段 + AutoCheckin；去掉 Cookie/Token/Beacon。
- `PipelineConfigDTO`：加 `AccountID`、`CheckinAccountID`、`CheckinInfoID`、`Account *AccountDTO`、`CheckinAccount *AccountDTO`、`CheckinInfo *CheckInInfoDTO`；Beacon 只读从 info join。
- `HelperLibraryLayoutRequest` 加 `AccountID uint64 \`json:"account_id,string"\``。Cookie 仍可选。解析顺序：非空 Cookie → AccountID 的 cookie → 现有 WIP 的 session 回落（若已存在）。**不要删** WIP 里 cookie 可选逻辑。
- `RunPipeline`：占座用 occupy account.cookie；自动打卡用 checkin account（id=0 则 occupy）的 token + info Beacon + `row.LibraryID`。覆盖凭据写到对应**账户**，不再 `UpdatePipelineCookie`（旧函数可留着不被调用）。
- `account_id==0`：对该行调 backfill 步骤 3；仍为 0 → `need_auth: LOGIN` HTTP 200。
- Beacon 不完整：占座成功，`CheckinStatus` 说明去补签到信息。

- [ ] **Step 1: 改/补 `pipeline_test.go`**

新用例（mock api）：

```go
func TestRunPipeline_SeparateOccupyAndCheckinAccounts(t *testing.T) {
	// occupy account cookie=mock_occ, checkin account token=mock_chk, info beacon=FDA5...
	// RunPipeline → ReserveSeat 收到 mock_occ，SignCheckIn 收到 mock_chk
}
func TestRunPipeline_CheckinAccountIDZeroUsesOccupy(t *testing.T) {
	// checkin_account_id=0 → SignCheckIn 用 occupy 的 token
}
func TestCreatePipelineConfig_RequiresAccountID(t *testing.T) {
	// 无 account_id → 400
}
func TestCreatePipelineConfig_AutoCheckinRequiresInfo(t *testing.T) {
	// auto_checkin true, checkin_info_id=0 → 400
}
```

旧测试若仍传 `Cookie:` 创建请求，改为先插账户再传 `AccountID`。只改本文件里被编译失败的用例，不要顺手大重构。

Run: `cd backend && go test ./igo-lib/plugins/igo/service -run TestPipeline -count=1` 先红。

- [ ] **Step 2: 改 service**

`CreatePipelineConfig`：校验账户属于 userID；`auto_checkin && checkin_info_id==0` 400；`checkin_account_id!=0` 时校验该账户属于 userID。行上仍可把 occupy cookie **复制**进旧列（只写一次便于回滚观察），但执行路径禁止读旧列。

`validatePipelineAuth` / `checkAndReserveSeat` / `executeBeaconCheckin` 改为接收已解析的 `occupy *entity.Account`、`checkinAcc *entity.Account`、`info *entity.CheckInInfo`。不要整文件重排无关函数。

`applyPipelineOverrides`：有 Cookie 则 `dao.UpdateAccount` 占座账户；有 CheckinToken 则更新打卡账户。

`HelperGetLibraryLayout(ctx, userID, cookie, libraryID)` 若签名已是这样，加 accountID 参数或在 controller 里先取 cookie：

```go
cookie := req.Cookie
if cookie == "" && req.AccountID != 0 {
	acc, err := dao.GetAccountByUser(ctx, req.AccountID, userID)
	// miss → 404；cookie 空 → 400
	cookie = acc.Cookie
}
```

- [ ] **Step 3: 跑 pipeline 单测 + e2e 包**

```
cd backend && go test ./igo-lib/plugins/igo/service -count=1
cd backend && go test ./igo-lib/plugins/igo -run Pipeline -count=1
```

Expected: PASS。失败只修本任务引用逻辑。

- [ ] **Step 4: Commit** `feat(igo): pipeline cards reference accounts and checkin infos`

---

### Task 9: Bot 补录写账户

**Files:**
- Modify: `backend/igo-lib/plugins/igo/bot/login_auth.go`（仅当 RunPipeline 覆盖已写账户后，测试仍过则本任务可能无代码改动）
- Modify: `backend/igo-lib/plugins/igo/bot/checkin_auth.go`
- Modify: `backend/igo-lib/plugins/igo/bot/login_auth_test.go`
- Modify: `backend/igo-lib/plugins/igo/bot/checkin_auth_test.go`

**Interfaces:**
- Bot 仍调 `RunPipeline(..., &do.RunPipelineRequest{Cookie: text})` / `{CheckinToken: text}`。Task 8 已把 override 落到账户。本任务确认测试：`lastReq.Cookie` / `CheckinToken` 仍被传入即可。
- 若 stub 要编过，更新 DTO 零值，不要改对话状态机。

- [ ] **Step 1: 跑** `cd backend && go test ./igo-lib/plugins/igo/bot -count=1`
- [ ] **Step 2: 修编译/断言**
- [ ] **Step 3: Commit**（有改动才提交）`fix(igo): keep bot auth overrides targeting pipeline accounts`

---

### Task 10: 前端 types 与 service

**Files:**
- Create: `frontend/lib/services/igo/account.ts`
- Modify: `frontend/lib/services/igo/checkin.ts`
- Modify: `frontend/lib/services/igo/types.ts`
- Modify: `frontend/lib/services/igo/index.ts`
- Modify: `frontend/lib/services/igo/pipeline.ts`（layout 传 `account_id`）

**Interfaces:**

```ts
export interface AccountDTO {
  id: string;
  name: string;
  has_cookie: boolean;
  cookie_masked?: string;
  cookie_expires_at?: string;
  has_checkin_token: boolean;
  checkin_expires_at?: string;
  nickname: string;
  school: string;
  student_name: string;
  student_number: string;
  created_at: string;
  updated_at: string;
}
export interface CheckInInfoDTO {
  id: string;
  name: string;
  beacon_uuid: string;
  major: number;
  minor: number;
  latitude: string;
  longitude: string;
  created_at: string;
  updated_at: string;
}
```

`CreatePipelineConfigRequest`：`account_id: string`；`checkin_account_id?: string`；`checkin_info_id?: string`；`auto_checkin: boolean`；去掉 cookie/token/beacon。DTO 增加三个 id 与可选 `account` `checkin_account` `checkin_info`。

```ts
export class IGoAccountService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/accounts';
  static list() { return this.get<AccountDTO[]>(''); }
  static create(data: CreateAccountRequest) { return this.post<AccountDTO>('', data as never); }
  static update(id: string, data: { name: string }) { return this.put<AccountDTO>(`/${id}`, data as never); }
  static remove(id: string) { return this.delete<void>(`/${id}`); }
  static login(id: string, data: { code: string }) { return this.post<AccountDTO>(`/${id}/login`, data as never); }
  static checkinAuth(id: string, data: { code: string }) {
    return this.post<AccountCheckinAuthResponse>(`/${id}/checkin-auth`, data as never);
  }
}
```

`IGoCheckInService` 增加 `listInfos/createInfo/updateInfo/deleteInfo/signInfo`，路径 `/infos`、`/infos/${id}/sign`。旧 session 方法可留着但页面不再调用。

`index.ts` 注册 `account: IGoAccountService`。

- [ ] **Step 1: 改类型与 client**
- [ ] **Step 2: `cd frontend && bun run tsc --noEmit` 或项目现有 `bun run lint`** 修类型错误（pipeline 组件会红，Task 12 再改；若 lint 拦全部，本任务先让 service/types 自洽，组件放到 11/12）。
- [ ] **Step 3: Commit** `feat(igo): add frontend account and checkin-info clients`

---

### Task 11: `/checkin` 页面

**Files:**
- Modify: `frontend/app/(main)/checkin/page.tsx`
- Create: `frontend/components/igo/checkin/account-list.tsx`
- Create: `frontend/components/igo/checkin/checkin-info-list.tsx`
- Create: `frontend/components/igo/checkin/checkin-sign-bar.tsx`
- Modify: `frontend/messages/zh-CN.json` `frontend/messages/en.json` 的 `igo.checkin`
- 旧 `checkin-auth-card.tsx` 等可留文件但 page 不再引用（或抽授权对话框给账户列表复用手动录入 UI）

**行为：**

- 页标题仍「远程签到」，description 改为管理账户与签到信息并按引用打卡。
- 账户列表：名称、学号、Cookie/Token 徽章、录入登录、录入签到、删除（409 toast 引用 ID）。
- 签到信息列表：名称、UUID 摘要、编辑 Beacon、删除。
- 打卡条：Select 账户、Select 签到信息、场馆默认 `IGoService.venue.getBoundLibrary()` 的 `library_id`；提交 `signInfo`。409 且 `data.need_auth==='CHECKIN'` 打开该 `account_id` 的签到授权对话框，不改签到信息。
- `checkinAuth` 返回 `device.beacon_uuids` 且正在编辑的 info UUID 为空 → 填第一项。

复用现有手动录入对话框文案（`manualDialogTitle` 等），把目标从「全局 session」改成「该账户」。

- [ ] **Step 1: 实现组件与 i18n**
- [ ] **Step 2: `cd frontend && bun run lint` 修本页问题**
- [ ] **Step 3: Commit** `feat(igo): rebuild checkin page around accounts and infos`

---

### Task 12: 一条龙向导改引用

**Files:**
- Modify: `frontend/app/(main)/pipeline/components/pipeline-create-dialog.tsx`
- Modify: `frontend/app/(main)/pipeline/components/pipeline-edit-dialog.tsx`
- Modify: 其它 pipeline 组件若展示 Beacon 输入
- Modify: `frontend/messages/zh-CN.json` `en.json` 的 `igo.pipeline`

**行为：**

- Step 1：Select 已有账户，或粘贴 Cookie/链接 → `pipeline.helperVerifySession` → `account.create({name, cookie})` → 选中新账户。
- 自动打卡 Switch：打开后 Select 签到信息（空态链接到 `/checkin`）+ 可选另一打卡账户（默认「与占座相同」）。
- **禁止**再渲染 lat/lng/mac/major/minor 输入。
- Step 2 layout：`helperGetLibraryLayout({ account_id, library_id })`，不要依赖全局 session。
- 保存 payload：`{ id, name, account_id, library_*, seat_*, auto_checkin, checkin_info_id?, checkin_account_id? }`。
- 编辑弹窗同样。

- [ ] **Step 1: 改对话框**
- [ ] **Step 2: lint**
- [ ] **Step 3: Commit** `feat(igo): pipeline wizard selects accounts and checkin infos`

---

### Task 13: E2E

**Files:**
- Create: `frontend/e2e/checkin.spec.ts`
- Modify: `frontend/e2e/pipeline.spec.ts`

**Checkin E2E（mock API）：**

- mock `GET /api/v1/igo/accounts`、`GET /api/v1/igo/checkin/infos`、`GET /api/v1/igo/libraries/bound`。
- 创建签到信息 POST `/checkin/infos` body 含 `name`。
- 打卡 POST `/checkin/infos/:id/sign` body 含 `account_id` 与 `expected_library_id`。
- mock sign 409 `{error:{code:'need_auth'}, data:{need_auth:'CHECKIN', account_id:'1', auth_url:'https://example.test/auth'}}`，断言出现账户签到授权录入，而不是 Beacon 表单。

**Pipeline E2E：**

- 去掉「填写 beacon lat/mac/major」断言。
- 打开自动打卡后断言 `#create-beacon-lat` 不存在。
- 创建请求 `postData.account_id` 有值；`auto_checkin true` 时 `checkin_info_id` 有值；无 `beacon_uuid` 写入字段。

- [ ] **Step 1: 写/改 spec**
- [ ] **Step 2: 按仓库方式跑 Playwright**（`cd frontend && bunx playwright test e2e/checkin.spec.ts e2e/pipeline.spec.ts`）。Expected: PASS。
- [ ] **Step 3: Commit** `test(igo): e2e for checkin infos and pipeline account refs`

---

### Task 14: Swagger、笔记、全量校验

**Files:**
- Modify: 由 `make swagger` 生成的 `backend/docs/*`
- Create: `.agents/notes/implemented/architecture/2026-09-16-checkin-info-and-accounts.md`
- Modify: `backend/igo-lib/plugins/igo/plugin.go` 或 `service/account.go` 入口加反向注释

**Note 要点：**

- Problem：凭证绑在签到信息/卡片上，双号打卡会串 Token；Beacon 重复填写。
- Decision：`igo_accounts` 存 Cookie+Token+资料；`igo_checkin_infos` 只存 Beacon；pipeline 引用三个 ID；`igo_sessions` 仍给抢座用。
- Alternatives：扩 `igo_sessions` 为 1:N（回归面大）；只抽 Beacon 表（双号仍无账户库）。
- Consequences：Cookie 两处存储，以后用当前会话指向 account_id；回填不能自动拆历史双号卡片。

反向注释：

```go
// Note: 账户与签到信息拆表，pipeline 只存 ID — 见 .agents/notes/implemented/architecture/2026-09-16-checkin-info-and-accounts.md
```

放在 `entity.Account` 或 `plugin.go` Apply 路由附近。

- [ ] **Step 1:** `cd backend && make swagger`（在 backend 目录，按仓库 Makefile）。
- [ ] **Step 2:** `make code-check` 与 `make format`；`cd frontend && bun run lint && bun run format`。
- [ ] **Step 3:** `cd backend && go test ./igo-lib/plugins/igo/... -count=1`
- [ ] **Step 4: Commit** `docs(igo): note account and checkin-info split`（笔记+swagger；code-check 若只改格式并进同一提交仅限本任务文件）

---

## Spec coverage

| Spec | Task |
|---|---|
| `igo_accounts` / `igo_checkin_infos` 表 | 1 |
| pipeline 三列 | 1, 8 |
| 回填 4 步幂等 | 7 |
| 账户 CRUD/login/checkin-auth | 4, 6, 10, 11 |
| 签到信息 CRUD + sign by id | 5, 6, 10, 11 |
| sign 409 need_auth | 5, 6, 11, 13 |
| 一条龙引用、双号执行、override 写账户 | 8, 9 |
| layout 用账户 Cookie | 8, 12 |
| `/checkin` 新 UI | 11 |
| 向导无 Beacon 手填 | 12, 13 |
| 不改 igo_sessions / 不删旧路由 | 6, Global |
| 旧列只读 | 8 |
| Bot 补录 | 9 |
| E2E | 13 |
| Agent note | 14 |

Spec 未写、计划补上：`POST /pipeline/library-layout` 增加 `account_id`，否则选已有账户后前端没有 Cookie 可拉座位。
