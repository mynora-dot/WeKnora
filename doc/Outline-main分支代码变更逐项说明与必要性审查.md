# Outline 分支相对 main 的逐项代码说明与必要性审查

生成日期：2026-09-13。用途：本地审阅附件，不提交。

## 1. 对比依据

- 仓库：/Users/zhayinggang/Documents/github-mynora-dot/WeKnora
- 当前分支：`codex/outline-datasource`；完整 SHA：`e79e2b45ace3fa8e62248a5fb01f17a19bade356`。
- **本地 main**：`462999ec3f5c1467ef0ccf5cf8c422393f40a0e6`。本轮未 fetch，不能把它当成远端此刻最新 main。
- 比较命令：`git diff main HEAD`。本地 main 恰为 merge-base，故本次与 `main...HEAD` 相同。
- 开始检查时工作区干净。净变化 **41 个文件，3489 行新增、47 行删除**。逐文件正文与附录覆盖同一集合。
- 本次只做静态审阅、生成报告；没有运行测试/构建、没有改业务代码、没有提交或推送。已有测试存在不代表本轮测试通过。
- 解释单位是每个功能性变更/函数/配置块；import、类型引用和排版随对应功能说明。附录保留全部 diff，便于逐行核对，不遗漏零散变更。

## 2. 先看结论与需要注意的点

核心接入、分页、身份去重、逐条确认、完整基线和保守删除属于需求必要变更；文案、图标、指标属于配套增强。不能把“必要”理解为当前实现已证明没有缺陷，也不能把共享代码修改称为完全不影响旧逻辑。

| 项目 | 当前结论 | 建议 |
|---|---|---|
| 分页 9/25 残留 nextPath | 已按 total/短页终止，并校验下一跳/重复/总数变化 | 保留；执行三个列表端点边界回归 |
| 旧连接器去重 | Outline channel 才注入身份参数，其他调用零值保留旧规则 | 保留限制；回归普通上传/语雀/飞书 |
| 共享同步入口 | strict、两小时 timeout、租约主要限定 Outline；检查点与部分错误处理有共享变化 | 不能承诺零影响，见逐项说明 |
| 任意 body 的旧同步请求 | 现在解析 JSON，空 body 兼容，垃圾 body 从忽略变成拒绝 | 合理收紧，但发布说明应注明 |
| 原文入口 | 所有带合法 source_url 的文档都出现 | 有益 UI 扩展，不仅 Outline |
| 测试归档 | 相对 main 仍有 **6 个新增 *_test.go 文件**，共 37 个顶层 Test 函数 | 与此前只放本地附件的要求不完全一致；本轮不擅自迁移 |
| SQLite 迁移 | main 与 HEAD 均无本次 000015 索引文件，净 diff 无迁移 | 无须恢复；索引性能验收仍未完成 |
| 旧失效索引测试 | e79e2b45 已移除，相对 main 的最终文件只剩有效场景 | 不再按缺失文件测试，不能记成索引测试通过 |
| 更新/向量化 | 沿用整篇删除后重建，不是块级重算 | 接受替换窗口，验收检索新版本/失败恢复 |
| 性能与互斥 | 元数据 O(N)、每页完整游标序列化、租约校验/元数据查询增加开销 | 万篇/失锁/多实例真实验收仍必要 |

### 2.1 审阅发现的改进候选（本轮未修改）

- **防御性身份校验**：Outline 正常导入会填身份，但两项都空时 Repository 不报错；若期望所有 Outline 内部调用都严格要求身份，可在文件创建入口补拒绝。
- **预留字段**：ApplyResult.KnowledgeID 目前未使用，可精简，不影响当前确认逻辑。
- **长期资源占用**：客户端 limiter 全局 map 不回收，频繁创建/轮换凭据会累积。
- **租约不等于原子写入 fencing**：检查后到副作用之间可能失锁。现有取消与前置检查降低风险，不构成严格线性一致保证。
- **错误分类变化**：共享流式 PartialFetchError 现在转 partial，而非一般错误路径；对旧连接器队列重试语义需专门测试。
- **统计边界**：同版本幂等命中可返回 updated=true，更新数不能独立证明发生了解析；检查点时长指标不是 p95 性能结论。
- **请求与性能成本**：增量仍遍历源列表并比较指纹；不会按修改时间只抓差量。元数据恢复查询、逐条租约/存在性校验和完整游标落库需要容量评估。

## 3. 逐文件、逐功能说明

必要性判断区分“需求不可缺少”“当前实现的配套”“可选增强/可简化”，而不是把每行都评为必须。对比中的删除行也在相应功能的前后行为中解释。

### 3.1 `frontend/src/api/datasource/index.ts`

`triggerSync(id, forceFull=false)` 增加可选参数，请求由 `{}` 变为 `{force_full:false/true}`。必要：实现一次性全量，不修改长期同步模式。旧前端调用签名兼容；共享请求体确实变化，旧后端通常忽略字段，但混合版本不能视为已验收。

### 3.2 `frontend/src/i18n/locales/en-US.ts`

新增 29 行：12 类 `outlineError`（认证、权限、限流、响应、请求、实例、大小、游标、重试、删除、范围、运行冲突）；一次性全量、原文入口、实例地址；接入说明及文档入口；Outline 专用删除文案；连接器名称/描述与 Collection 类型。必要性：本地化及明确删除范围所需，接入说明属于易用性增强。保留原有键值，正常不改变旧来源文案。五种语言键一致；本轮未做人工作为母语的译文质量验收。

### 3.3 `frontend/src/i18n/locales/ja-JP.ts`

新增 29 行：12 类 `outlineError`（认证、权限、限流、响应、请求、实例、大小、游标、重试、删除、范围、运行冲突）；一次性全量、原文入口、实例地址；接入说明及文档入口；Outline 专用删除文案；连接器名称/描述与 Collection 类型。必要性：本地化及明确删除范围所需，接入说明属于易用性增强。保留原有键值，正常不改变旧来源文案。五种语言键一致；本轮未做人工作为母语的译文质量验收。

### 3.4 `frontend/src/i18n/locales/ko-KR.ts`

新增 29 行：12 类 `outlineError`（认证、权限、限流、响应、请求、实例、大小、游标、重试、删除、范围、运行冲突）；一次性全量、原文入口、实例地址；接入说明及文档入口；Outline 专用删除文案；连接器名称/描述与 Collection 类型。必要性：本地化及明确删除范围所需，接入说明属于易用性增强。保留原有键值，正常不改变旧来源文案。五种语言键一致；本轮未做人工作为母语的译文质量验收。

### 3.5 `frontend/src/i18n/locales/ru-RU.ts`

新增 29 行：12 类 `outlineError`（认证、权限、限流、响应、请求、实例、大小、游标、重试、删除、范围、运行冲突）；一次性全量、原文入口、实例地址；接入说明及文档入口；Outline 专用删除文案；连接器名称/描述与 Collection 类型。必要性：本地化及明确删除范围所需，接入说明属于易用性增强。保留原有键值，正常不改变旧来源文案。五种语言键一致；本轮未做人工作为母语的译文质量验收。

### 3.6 `frontend/src/i18n/locales/zh-CN.ts`

新增 29 行：12 类 `outlineError`（认证、权限、限流、响应、请求、实例、大小、游标、重试、删除、范围、运行冲突）；一次性全量、原文入口、实例地址；接入说明及文档入口；Outline 专用删除文案；连接器名称/描述与 Collection 类型。必要性：本地化及明确删除范围所需，接入说明属于易用性增强。保留原有键值，正常不改变旧来源文案。五种语言键一致；本轮未做人工作为母语的译文质量验收。

### 3.7 `frontend/src/utils/datasourceError.ts`

新增错误码→翻译类别映射与 `localizeDatasourceError`。只替换 Outline/租约相关码，未知码保留；非字符串返回空串由调用方兜底。建议保留：让配置和日志可读，不是抓取算法硬依赖。共享界面的错误显示会经过此函数，旧普通字符串保持原样；按正则替换而非结构化错误解析，复杂日志可能出现多段翻译。

### 3.8 `frontend/src/views/knowledge/KnowledgeBase.vue`

来源筛选增加 Outline 选项。必要于按来源筛选；不改变原选项。品牌名直接写 Outline 不需要翻译。

### 3.9 `frontend/src/views/knowledge/components/DocumentActionMenu.vue`

KnowledgeItem 增加可选 metadata；兼容 JSON 字符串和对象，提取 source_url；仅允许 HTTP(S)、无用户名密码的 URL；新增新窗口原文链接及 noopener/noreferrer，阻止点击冒泡。原文追溯所需。注意该入口未限定 channel=outline：所有带有效 source_url 的文档都会显示，属于明确的共享 UI 扩展；若要求完全不改变旧来源 UI，可加类型限制，但本轮不改代码。

### 3.10 `frontend/src/views/knowledge/components/DocumentListView.vue`

`getSourceInfo` 新增 outline 分支，显示来源图标和品牌名称。展示所需，不涉及同步/向量化。

### 3.11 `frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue`

1. 增加 Outline 定义、API 文档链接、六个接口权限提示、必填实例地址与密钥字段：接入所需。
2. 新增 resourceLoadError，并在选类型/加载时清空；错误展示重试入口；加载中/失败/零选中时禁止 Outline 继续：防止误保存范围，必要。
3. 选择 Outline 强制 overwrite，隐藏 skip：与后端仅支持覆盖更新一致，必要。
4. 新建临时连接采用 paused、空计划、关闭删除；仅非编辑临时记录更新时套用这些覆盖：防止选范围前调度或要求删除权限。不会在正常编辑时由 loadResources 强制暂停已有连接。
5. 资源图标/标签增加 collection：多选 Collection 的显示适配。
6. 凭据删除、测试连接、资源加载、保存等错误使用本地化函数：体验增强，普通旧错误原样保留。
7. Outline 删除开关使用专用文案，明确清理取消选择集合里的文档：必须与当前行为配套，否则用户不知清理范围。

### 3.12 `frontend/src/views/knowledge/settings/DataSourceSettings.vue`

`handleSync` 透传 forceFull；仅管理员可见且仅 Outline 增加重新全量菜单，运行中禁用；同步错误本地化。必要于手动恢复本地副本和强制重建，原“立即同步”保持默认 false。禁用只是 UI 提示，真实互斥由后端保证。

### 3.13 `frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue`

单条原因与整轮错误调用本地化函数。建议保留，方便识别分页不完整/权限/重试，非同步正确性硬依赖。仅显示变更，不修改持久化错误。

### 3.14 `frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue`

为 outline 增加 root-list 图标兜底。可选视觉增强，删除不会影响同步功能。

### 3.15 `go.mod`

增加 goldmark v1.8.2 直接依赖，用于识别 Markdown 链接和图片目标而不修改代码块。按当前实现必要，可替换实现但不宜用简单正则替代。go.sum 已有该版本模块及内容校验，因而没有 go.sum 净变化。

### 3.16 `internal/application/repository/datasource_repo.go`

Update 为 Outline 用显式 Select 写入字段，允许 false、空字符串、空计划/配置等零值生效：必要，否则 GORM struct Updates 会跳过零值。旧来源仍走原更新分支。
UpdateSyncState 对 Outline 用 CASE 保留数据库中的 paused，避免在途任务写回 active；Where 加 deleted_at IS NULL 是共享条件，意图不更新已删除连接。
新增 Pause 只更新状态，避免旧对象覆盖新检查点：Outline 暂停所需。暂停阻止后续计划，不等于立即停止正在运行的一轮。

### 3.17 `internal/application/repository/knowledge.go`

CheckKnowledgeExists 要求数据源 ID 与外部 ID 成对出现；提供时追加两个 metadata 条件，然后沿用原文件哈希/类型等条件。必要于同正文不同 Outline 身份不误判重复。它是通用参数机制，但生产文件入口只为 Outline 填值；两项都为空仍用 main 规则。注意同时为空不会被此函数当成身份缺失报错。
ListDataSourceKnowledgeMetadata 校验租户/知识库/数据源/页大小，只取 id/channel/metadata，以 id 游标分页，排除已删除行。必要于找回旧范围副本，避免全量加载正文。当前没有新增专用索引，万篇查询成本需要实测。

### 3.18 `internal/application/repository/outline_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestOutlineFileIdentityDeduplication`：身份条件使相同正文按数据源/文档独立；空身份保留旧去重；不成对报错。
- `TestOutlinePauseDoesNotLoseCheckpoint`：暂停后更新检查点不覆盖 paused。
- `TestOutlineUpdateCanClearScheduleAndCredentials`：可清空计划/凭据并关闭删除。

### 3.19 `internal/application/service/datasource_outline_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestAcknowledgingHandlerReportsFailureAndDeferred`：确认结果区分失败和关闭删除。
- `TestOutlineOptionsAllowOnlyPausedEmptyDraft`：空范围仅允许无计划暂停草稿、拒绝 skip/五字段 Cron。
- `TestOutlineRequiresParseAcceptanceAndStrictLookup`：拒绝创建返回 failed 和查询失败，pending 接收可确认。
- `TestOutlineSameVersionIsIdempotentUnlessNewFullRun`：同版本幂等但新 full run 强制 delete/create。
- `TestOutlineEnablingDeletionRevalidatesUnchangedCredentials`：启用删除即使凭据未变也重新验证能力。
- `TestFailedStreamCheckpointKeepsLastPersistedCursor`：检查点写库失败保留旧内存游标。
- `TestAcknowledgementDistinguishesMissingIdentityFromAlreadyDeleted`：缺外部身份与已不存在副本有不同确认结果。
- `TestOutlineDuplicateCompletedTaskPreservesResult`：已终结任务重复投递不覆盖状态和数量。

### 3.20 `internal/application/service/datasource_service.go`

此文件是共享同步链路的主要变更，按函数解释：

| 改动位置 | 行为与必要性 | 对旧逻辑的影响/限制 |
|---|---|---|
| UpdateDataSource | Outline 编辑先加锁并重新读记录，禁止改连接器类型；保留旧游标，校验模式/计划/范围；删除策略或状态变化也重新验证权限 | 必要，防止边同步边改凭据/范围、丢基线；Outline 分支限定 |
| UpdateDataSourceCredentials / ClearDataSourceCredentials | 写凭据前取得同一互斥锁 | 必要，避免运行中途切换身份；旧来源不加此锁 |
| ValidateConnection | Outline 加锁后只校验，不借测试连接改状态或覆盖游标 | 必要于草稿/暂停安全；注意前端编辑时测试连接仍会先调用原有 update 接口，不能宣称整个 UI 操作绝对只读 |
| ManualSync / ManualSyncWithOptions | 保留旧方法包装 false，新方法透传 forceFull；Outline 提交任务前锁定 | 必要于一次性全量与防重复；接口新能力不是硬改所有 service 实现 |
| PauseDataSource | Outline 使用可选 Pause 单列更新，缺能力仍走 Update | 必要于避免检查点被旧对象覆盖 |
| ResumeDataSource | Outline 锁后重读再置 active，校验连接再恢复 | 必要于避免无效配置恢复调度 |
| ProcessSync | 仅 Outline 增加两小时 context、执行租约和锁内重读；success/partial/canceled 日志重复投递直接返回 | 必要于互斥和结果幂等；两小时只是上限，不能延长调用方更短的 deadline；其他连接器不新增此上限 |
| config.SyncDeletions | 把实际策略传入运行时 config | 必要于删除能力验证和对账；旧连接器一般不消费新增字段 |
| applyFetchedItem | 原计数/删除/写入分支增加 ApplyResult 返回：关闭删除 deferred、失败 failed、已不存在或成功 applied | 必要于逐条确认；旧 Emit 忽略返回值，基本沿用原计数/写入分支 |
| streamSyncHandler.Emit | 新增 checkAlive 后执行原路径 | Outline 检查租约及连接存在；其他来源仅检查 ctx |
| EmitWithResult | 关闭删除不计数，返回确认状态；其他操作计数并透传错误 | 必要；确认代表已接收，不代表异步解析/检索已完成 |
| WalkSyncedItems | 可选 Repository 读取器每页 200 条；按 channel/datasource_id/external_id 校验并送出元数据；防游标不推进 | 必要于历史范围副本恢复；缺能力明确失败 |
| Checkpoint | 检查存活；读可选 metrics；先用 ds 副本写库，成功才更新内存游标，保留同步日志实时统计 | 写入顺序修复是共享行为，避免数据库失败后错误推进；metrics 转换错误现在会中止该次检查点，旧连接器若碰巧使用不兼容 metrics 字段需关注 |
| checkAlive | Outline 校验租约及数据源还存在 | 必要的失锁/删除防护；每次操作多一次查询，有吞吐成本，且不是事务级 fencing |
| lockOutlineMutation | Acquire + HasRunningSync + 锁内重读；失败释放 | 必要于排队任务也阻止配置变更；依赖运行日志终结/恢复，陈旧 running 日志的恢复需验证 |
| processSyncStreaming | 优先 SyncRunCursorPreparer，其次 FullSyncCursorPreparer，最后旧游标选择；兼容队列重试上下文 | 必要于重试不丢全量基线；未实现能力的旧连接器保留 fallback，但新增重试元数据读取会影响使用该上下文的路径 |
| PartialFetchError | 流式部分失败转 partial 日志，追加有限错误展示 | 有益共享增强；之前返回该错误的旧流式连接器也会改变错误分类/队列重试语义，必须回归 |
| 删除失败提示 | Outline 不显示“仅下次全量重试” | 必要，Outline 自身 pending 下轮可重试，旧来源提示保留 |
| updateSyncRunResult | Outline 使用脱离取消的 5 秒清理上下文，但仍检查租约 | 必要于超时后尽量落状态；失锁则不写，日志恢复仍依赖外围机制 |
| validateDataSourceConfig | 注入策略、Validate 后可选 BindIdentity，存绑定游标和 Outline 规范化 config | 必要于阻止跨实例复用历史身份；新增能力为可选 |
| validateOutlineOptions | 仅 overwrite、incremental/full、六字段 Cron；空范围仅无计划暂停草稿允许 | 功能约束所需；这是产品选择，不是 Outline API 本身要求 |
| ingestItem strict | Outline 查询/删除/硬删除失败立即返回；同指纹且可接受解析状态幂等跳过；新全量 run 强制重建；创建返回 nil/failed 不确认 | 必要于不能把失败版本记为成功；旧来源继续原宽松行为。幂等跳过返回 isUpdate=true，统计的“更新”不总等于重建任务数 |

没有新增块级差异 Embedding：仍复用先删除旧知识、再创建文件的既有方式，有替换窗口，删除成功而创建失败时旧副本已不存在。

### 3.21 `internal/application/service/knowledge_create.go`

仅 channel==ChannelOutline 从 metadata 读取 datasource_id/external_id，并传入去重查询；其他来源传空。必要于来源身份去重且限制对语雀/飞书/普通上传的影响。没有改解析、分块、Embedding 流程；小幅编辑仍整篇重建。防御性不足点：若内部调用声称 Outline 却两项都缺失，会回退普通去重；正常 ingestItem 会补齐身份，但尚无入口严格拒绝两项皆空。

### 3.22 `internal/container/container.go`

import 并实际 Register Outline：不可缺少，只有前端/元数据配置而没有实例注册会报 connector type not found in registry。
初始化 DefaultSyncLocks，使用容器 Redis Client（nil 时 Lite 进程锁）：必要于调度、编辑和 worker 共用同一锁。全局可变变量属于工程折中，未来可用依赖注入替代；多个容器并发初始化的行为未验证。

### 3.23 `internal/datasource/connector.go`

新增 ApplyOutcome 三态 applied/deferred/failed、ApplyResult、AcknowledgingStreamHandler：只有接收确认后推进版本，必要。KnowledgeID 目前没有实际写入/消费，是可删的预留字段。
SyncedItemReader：恢复历史身份，必要于兼容旧游标丢失范围记录。
FullSyncCursorPreparer 和 SyncRunCursorPreparer：保留基线并区分新全量与队列重试。当前 Outline 优先走 run preparer；full preparer 也被批量适配器调用，服务侧兼容分支属于扩展支持。
元数据注册增加 Outline、API key 认证、增量/删除能力、优先级 3：前端发现与能力声明必要。
原 StreamHandler/StreamingConnector 必需方法未扩大，旧连接器无须新增确认方法。

### 3.24 `internal/datasource/connector/outline/client.go`

全新文件，按实现单元：
1. apiError/fatal：只暴露稳定错误码，取消/超时/401 为全局致命；403/404 不凭错误推断删除。必要。
2. newClient：要求实例根地址和密钥、补 HTTPS、规范化 host，拒绝用户名密码/query/fragment/子路径；调用既有 SSRF 校验；HTTP 只允许白名单私网；30 秒客户端，拒绝重定向。必要于认证和凭据安全；根路径限定/速率参数是可调整策略。
3. limiter：按 base+key 的 SHA-256 共享每秒一次限速器，避免同凭据并发请求风暴。进程内限速，不是多副本全局配额；sync.Map 不清理旧 key，频繁轮换大量密钥有长期增长风险。
4. envelope：保留 total/offset/limit/nextPath 的缺失信息（指针），必要于区分 total=0 和无 total。
5. wait/retryDelay/call：可取消等待，最多 4 次尝试，POST Bearer/JSON/API v2；429/5xx/传输异常重试；读取最多 32 MiB+1，超限拒绝；校验 data/ok。必要于鲁棒性。数字 Retry-After 限 7200 秒，HTTP 日期形式未同样限幅（任务 context 可截断），参数一致性可改进。
6. walk：三个列表共用，默认 25；响应超大原 offset 缩页到最小 1；保持 total 的有无及值不变化，拒绝页内/跨页重复 ID、无 ID；结束检查唯一 ID 总数。分页正确性必要。seen 占 O(N) 身份内存，不能声称常量内存。
7. page 回调按页执行，后续页失败不会回滚此前已接收文档，但整轮不完整不进行删除；这是流式设计选择。

### 3.25 `internal/datasource/connector/outline/connector.go`

Type/NewConnector 接入接口；authenticate 读取 auth.info 的用户/工作区 ID；selectedIDs UUID 规范化、排序去重；collectionInfo/documentInfo 校验返回身份。基础必要。
Validate 校验凭据、列集合、逐个选中集合状态/文档接口，抽取首文档详情确认 text 能力，开启删除时探测 documents.deleted。必要于提前报不支持；探测/重复认证增加请求成本，非全量内容验收。
ListResources 用共用分页，只展示未归档/未删除 Collection，无父层级；ResolveResourceAncestors 返回空：复用现有资源界面必要。
collector/FetchAll/FetchIncremental 用流式实现提供批量接口适配；collector 的 applied 只代表收集入内存，不代表入库。符合现有 Connector 接口所需，生产服务优先 FetchStream。
BindIdentity 拒绝跨实例/工作区换凭据；同工作区换账号清当前 run，保留历史基线：必要于避免身份串用。

### 3.26 `internal/datasource/connector/outline/cursor.go`

version 保存归属/指纹/时间；run 保存 ID、full、完成集合、见过文档和已应用版本；state 保存实例、选中范围、历史基线、全部 pending、任务 ID、最后完整扫描、指标。必要于恢复与保守删除。
parseCursor 严格验证版本/必需字段/map/文档身份，拒绝外来损坏游标；cursor 转通用 SyncCursor。必要，不能把坏游标当空基线。
PrepareFullSyncCursor 新建 full run 但保留历史；PrepareSyncRunCursor 用任务 ID 区分新全量和同任务重试。必要，不能全量前直接清空 cursor。
当前 state.cursor 忽略内部 Marshal/Unmarshal 返回值（结构受控）；可改成显式错误但不是本次核心功能必需。游标随文档量线性增长，没有本文件级硬性 10 MiB 拒绝。

### 3.27 `internal/datasource/connector/outline/markdown.go`

absoluteLinks 用 goldmark AST 定位 Link/Image 目标，跳过锚点/已有绝对地址，按源文档 URL 解析相对路径；按原始切片地址找到位置，去重排序后只替换目标，转义括号/尖括号，保留代码和格式。为“链接可用且不损坏代码块”建议保留；不是同步游标硬依赖。依赖解析器目标引用原始切片的行为，有单测但没有覆盖所有 Markdown 扩展/原始 HTML；大量链接反复搜索原文的最坏成本需实测。

### 3.28 `internal/datasource/connector/outline/outline_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestAcknowledgementRetryAndIncremental`：确认失败恢复与无变化增量。
- `TestDeletionFailureAndDisabledDeletionRetainIdentity`：删除失败/关闭删除保留身份。
- `TestUnknownMissingAndIncompleteScanNeverDelete`：未知缺失和不完整扫描不删除。
- `TestMoveAcrossCollectionsAndDeselect`：跨 Collection 移动与取消选择后的清理。
- `TestFullSyncPreservesBaselineAndResumes`：全量保留基线和断点恢复。
- `TestConfigurationAndCursorValidation`：配置/游标非法拒绝。
- `TestDocumentShapesAndMarkdown`：文档形态和 Markdown 转换。
- `TestCursorTenThousandSize`：万篇游标序列化大小（不是万篇吞吐验收）。
- `TestRetryAfter`：Retry-After 数值/日期行为。
- `TestFullRunPreparationSurvivesAdmissionRetries`：入场前重试不重置同一次全量 run。
- `TestPaginationAcceptsTerminalNextPath`：0/1/9/24/25/26/50 等末页 nextPath 边界。
- `TestPaginationRejectsUntrustedOrIncompletePages`：非可信下一跳/不完整页面拒绝。
- `TestHTTPRetryBoundAndCancellation`：HTTP 重试次数上限及取消。
- `TestInstanceBindingRejectsWorkspaceChanges`：工作区改变拒绝绑定。

### 3.29 `internal/datasource/connector/outline/pagination.go`

nextPage 核验响应 offset 和 limit、行数不超页；有 total 用 offset+count==total 结束，提前空页报错；无 total 用短页结束，整页继续额外请求。nextPath 仅校验，不直接请求其 URL；必须同实例同端点、offset 唯一递增且连续，limit 如存在必须一致；拒绝用户信息/fragment/畸形参数。
必要性最高：解决 total=9/limit=25 仍返回 nextPath 的循环问题。合法末页可广告 offset+limit，但确认终止后不跟随。即使末页 nextPath 非法仍失败，属于保守策略。额外未知 query 键目前没有统一拒绝，但不会带入请求；不应写成“拒绝任何额外参数”。

### 3.30 `internal/datasource/connector/outline/pagination_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestCollectionDiscoveryUsesTerminalPagination`：真实日志 9/25 残留 nextPath 的集合发现回归。
- `TestPaginationMetadataValidation`：offset/limit/重复参数/错误端点等元数据拒绝。
- `TestDeletionRetainsDocumentWhenDetailStateIsMissing`：详情状态缺失不删除，即使删除列表有记录。
- `TestMissingListStateRequiresDetailBeforeAcceptance`：列表缺状态时先补详情。
- `TestSourceTimestampsRemainInKnowledgeMetadata`：源时间/来源元数据保留。
- `TestOversizedListReducesLimitAtSameOffset`：超大列表同 offset 缩页。
- `TestPendingFailuresAreNotTruncatedWithErrorSamples`：150 个失败身份完整保留而错误样本最多 100，并可全部恢复。

### 3.31 `internal/datasource/connector/outline/stream.go`

全新同步状态机，按阶段解释：
1. 要求确认型 handler，解析 cursor/配置、认证；拒绝跨实例；账号/选中范围变化重置当前 run，保留历史 Documents。必要于正确恢复和范围切换。
2. 可选 WalkSyncedItems 找回不在当前选中范围且游标缺失的历史项，验证工作区、同源链接、指纹后只加入候选基线。必要于修复老版本“取消范围但丢掉身份”的遗留副本；缺关键信息会保留副本而非冒险删除。
3. 清除已不在范围的 pending upserts，创建 run；checkpoint 更新 pending/大小/耗时指标。全部失败身份保留，warnings 样本限制 100。必要；指标属于增强，checkpoint_duration 是最近一次记录，不等于 p95。
4. 探测删除能力，不可用时保留导入功能、关闭本轮删除并报告 partial；认证失败/取消立即中止。必要于不因删除权限不足误删。
5. 检查全部集合：权限/删除不明导致完整性失败，归档集合标记不活跃。必要于跨集合整体对账。
6. apply 先标记 seen，缺状态/正文或时间倒退补 details；不活跃不导入；pending 在确认前存在；指纹不变或本轮已经 applied 则跳过；失败不推进 Documents。必要于重试无需源端再编辑。
7. 50 个处理或 30 秒尝试检查点，同时每页/每集合末尾写检查点；不是独立定时器保证每 30 秒一定写。单条正文流式入库，但 seen/documents/applied 仍 O(N) 元数据。
8. 先重试 pending，再扫描未完成集合；使用 createdAt ASC/已发布过滤；成功集合可续跑跳过；任一列表不完整禁止本轮 reconcile。必要于恢复和删除安全，不是源端快照一致性保证。
9. complete 后 reconcile，记 LastComplete、清 run，最终保存；warnings/pending 转 PartialFetchError。必要于显式区分部分成功。
10. reconcile 候选来自全体历史 Documents 减 Seen，涵盖取消选中的集合。开启删除先完整遍历 documents.deleted，再逐一 documentInfo 复核；若仍在选中活跃范围，取消删除并置 pending upsert；状态完整且不活跃可确认；详情错误本身不能推断删除，但已在认证删除列表中可作为证据。单纯未知 403/404 保留。
11. 关闭删除记 deferred 保留历史；删除确认 applied 才清 Documents/pending，失败保留 pending 下轮重试，每条删除后检查点。必要于开关重开/删除失败恢复。
限制：若删除列表显示已删而详情现在权限失败，会沿用删除列表证据，恢复与权限并发变化仍需真实实例测试；API 多次调用间不是原子快照。

### 3.32 `internal/datasource/connector/outline/types.go`

document/collection/identity 适配 API；Text 指针区分空正文和缺失；stateKnown 检查四个状态字段是否出现，decodeDocument 支持直接对象/嵌套 document。必要于不把状态缺失当删除证据。
active 要求选中、已发布、未归档删除；sourceURL 仅允许同实例同协议原文；fingerprint 包含更新时间/revision、标题、集合、父级、URL及状态，无有效时间且无 revision 才加入规范化正文哈希。必要于文档级增量；依赖源端正确推进版本，不能检测“时间/revision 不变但正文变”的异常源行为。
fileName 过滤路径/控制字符并按 UTF-8 rune 截到 200 字节加 .md；mappedItem 加转义标题、保留正文语义、转绝对链接、构建来源/版本/时间元数据。必要于既有文件导入和可追溯；空正文会生成标题文档，私有附件只保留链接，不下载。
sourceTimestamp 统一 UTC RFC3339Nano，零时间为空：合理数据规范化。

### 3.33 `internal/datasource/scheduler.go`

Outline 计划触发前加锁并重读 active 状态；查询 HasRunningSync 失败时 Outline 保守不入队。必要于 Cron 与手动/编辑一致互斥。旧连接器查询失败仍保留原行为，未给它们强加锁。锁异常分支直接 return，没有专门错误日志，是可观测性改进点。

### 3.34 `internal/datasource/sync_lock.go`

新增 SyncLocks：key 包含 tenant/id，随机 token 区分持有人。标准模式 Redis SET NX 120 秒租期；每 30 秒 Lua 校验 token 后续租，释放同样比对 token；丢租约取消 context。Lite 用互斥保护的进程内 owner map。CheckSyncLease 通过 context 查当前持有人。互斥必要；Lite 只支持单进程/单实例，不能据此承诺多副本互斥。每次副作用前校验与实际写入间仍有时间间隙，不能将它描述为数据库事务级 fencing；需故障注入验证长操作时失锁。

### 3.35 `internal/datasource/sync_lock_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestSyncLocksMutualExclusion`：Lite/Redis 相同数据源互斥、不同租户独立、释放后可重获。
- `TestLostRedisLeaseDoesNotReleaseNewOwner`：旧租约失效后不能删除新持有人锁。

### 3.36 `internal/handler/datasource.go`

CreateDataSource 用可选 bool/string 指针区分没传和显式 false/空计划；仅 Outline 默认 active、incremental、overwrite、删除开启、六小时计划：必要于零值语义和当前默认策略，默认删除开启属于可调整产品选择。
Update/Validate/ManualSync 识别 ErrSyncRunning 返回 HTTP 409 和 code：必要于表达冲突。
ManualSync 读取可选 force_full，通过可选 service 方法调用；旧实现 false 回退旧方法、true 报不支持。
decodeForceFull 接受空 body/{}，拒绝 null、数组、非法布尔与多段 JSON；未知对象字段不拒绝。必要于接口正确性，但主分支原来不解析 body，因此旧客户端发送任意垃圾 body 的行为会从可接受变为 400。这是实际共享行为变化，不能称绝对无影响。
原租户/知识库归属校验保留。

### 3.37 `internal/handler/datasource_credentials.go`

Put 与 DeleteField 遇同步互斥错误统一返回 409；其余错误继续原分支。必要于凭据编辑和普通编辑反馈一致。

### 3.38 `internal/handler/datasource_outline_test.go`

新增测试文件；测试保护必要，但不是生产运行依赖。当前仍被 Git 跟踪并位于原包，与此前“全部新增测试仅做本地附件”的目标不完全一致。本次只说明，不移动。测试夹具/桩/辅助类型仅服务下列场景。

- `TestDecodeForceFull`：空 body/对象/true/false 兼容，非法类型/多段 JSON 拒绝。
- `TestOutlineCreateDefaultsAndExplicitDisabledOptions`：默认策略和显式 false/空计划/暂停草稿。
- `TestOutlineManualSyncOptionsAndConflict`：force_full 透传及 409 冲突。

### 3.39 `internal/types/datasource.go`

新增 ConnectorTypeOutline；DataSourceConfig 增加运行时 SyncDeletions（json:"-"，不复制进持久化配置）；HasConfiguredCredentials 为 Outline 同时检查 base_url/api_key；SyncResult 增加可选 metrics。前三项必要；metrics 是观测增强。旧结果没有 metrics 时序列化不变；原连接器凭据分支保留。

### 3.40 `internal/types/interfaces/knowledge.go`

新增独立可选 DataSourceKnowledgeMetadataReader 接口，读取已导入元数据。历史范围清理恢复所需；没有往 KnowledgeRepository 必需接口塞新方法，避免破坏旧 mock/第三方实现。生产 Outline 路径需要实现该可选能力；缺失时报告错误。

### 3.41 `internal/types/knowledge.go`

新增 ChannelOutline 和 KnowledgeCheckParams.DataSourceID/ExternalID。必要于来源展示和身份隔离。相邻 RSS/IMA 注释对齐是 gofmt 排版变化，没有逻辑改变；新字段零值使旧调用保持原规则。

## 4. 是否影响 main 原有逻辑

| 链路 | 判断 | 最小回归场景 |
|---|---|---|
| 普通文件上传 | 没有改去重核心哈希条件；新身份参数为空 | 相同/不同文件类型、同正文重复上传、失败文档再上传 |
| 语雀/飞书/GitLab | 无新增强制确认接口或 Outline strict/timeout；会经过共享 Checkpoint/错误分类 | 无变化增量、更新、删除失败、检查点失败、部分抓取错误、任务重试 |
| 手动同步 API | 旧空 body、{} 调用保留；新增验证与全量字段 | 老客户端空 body、新 true/false、非法 JSON、可选 service fallback |
| 计划同步 | Outline 加锁，其他保留原调度路径 | Cron 重叠、暂停、running 查询失败 |
| 文件解析/向量化 | 无相应算法文件变化，仍走共享创建流程 | 更新后旧内容替换、新短语命中、异步失败重解析 |
| Repository 第三方实现 | 新读取能力独立可选接口，旧实现可继续编译 | Outline 缺读取器明确报错，旧来源不误调用 |
| 多语言/UI | 新功能与错误翻译；原文入口可作用旧文档 | 五语言、旧来源菜单、资源失败重试、凭据不回显 |
| 数据库 | 本次无 schema 净变更，有新 JSON 条件读取和零值更新 | PostgreSQL/SQLite 的身份查询、清空配置、暂停竞态、查询耗时 |

本轮不将任何场景标记为通过：这是代码审查报告，不是执行后的验收报告。建议受影响 Go 包、race、前端类型/国际化/构建由最终版本执行；真实权限、删除恢复、容量验收另留截图与日志。

## 5. 本地附件与历史范围说明

本报告只反映两个冻结 SHA 的最终净差异。此前先加后删的两份 SQLite 迁移不在净差异中；Workflow 修复提交删除了相应失效测试。此前对测试归档“没有新增测试留在包目录”的结论不能套用到当前分支：下面附录可直接看到仍存在的六份测试文件。

本次不执行任何提交/暂存，不改 .gitignore；用户移动报告后不会留下新的跟踪配置。若使用 git add .，仍需自行排除本地 doc 附件。本报告与原文 diff 都放在本文件内，移动一个文件即可保留审查依据。

## 附录 A：完整逐文件差异（冻结版本）

以下内容由 `git diff --no-ext-diff --no-color main HEAD -- <file>` 自动生成，每个文件对应正文同序号；包含全部新增/删除行。大型新文件按完整文件展示，具体职责见正文。测试源代码也保留用于逐条审查，但不构成本轮已执行证明。

### A.1 `frontend/src/api/datasource/index.ts`

````diff
diff --git a/frontend/src/api/datasource/index.ts b/frontend/src/api/datasource/index.ts
index ee6eb20d..c30080d9 100644
--- a/frontend/src/api/datasource/index.ts
+++ b/frontend/src/api/datasource/index.ts
@@ -134,8 +134,8 @@ export function resolveResourceAncestors(id: string, resourceIds: string[]) {
   return post(`/api/v1/datasource/${id}/resource-ancestors`, { resource_ids: resourceIds }, { timeout: 120000 })
 }
 
-export function triggerSync(id: string) {
-  return post(`/api/v1/datasource/${id}/sync`, {})
+export function triggerSync(id: string, forceFull = false) {
+  return post(`/api/v1/datasource/${id}/sync`, { force_full: forceFull })
 }
 
 export function pauseDataSource(id: string) {
````

### A.2 `frontend/src/i18n/locales/en-US.ts`

````diff
diff --git a/frontend/src/i18n/locales/en-US.ts b/frontend/src/i18n/locales/en-US.ts
index f556d162..91500141 100755
--- a/frontend/src/i18n/locales/en-US.ts
+++ b/frontend/src/i18n/locales/en-US.ts
@@ -6291,6 +6291,31 @@ export default {
     daysAgo: '{days} days ago'
   },
   datasource: {
+    outlineError: {
+      auth: 'Authentication failed. Replace the API key.',
+      permission: 'Access denied. Check the account and collection permissions.',
+      rate: 'Rate limit reached. Retry later.',
+      response: 'Unsupported Outline response. Check the deployment version and Markdown API support.',
+      request: 'The scan could not finish. Check connectivity and retry; existing copies are retained.',
+      instance: 'The instance or workspace changed. Create a new connection.',
+      size: 'The document exceeds the 32 MiB limit.',
+      cursor: 'Invalid sync checkpoint. Contact an administrator; do not clear the baseline.',
+      pending: 'Some documents are pending retry on the next sync.',
+      deletion: 'Deletion could not be confirmed. Copies are retained; check documents.deleted permissions.',
+      scope: 'Select at least one accessible, active collection.',
+      running: 'A sync is queued or running. Try again after it finishes.',
+    },
+    fullSyncNow: 'Run full sync',
+    openSource: 'Open original',
+    outlineBaseUrl: 'Outline instance URL',
+    prereqBarText_outline: 'Outline connection requirements',
+    prereqStep1Brief_outline: 'Create a read-only API key',
+    prereqStep1Desc_outline: 'Use a dedicated account with access to the selected collections.',
+    prereqStep2Brief_outline: 'Allow the required endpoints',
+    prereqStep2Desc_outline: 'Allow auth.info, collections.list/info and documents.list/info. Deletion sync also requires documents.deleted.',
+    prereqStep3Brief_outline: 'Review access and retention',
+    prereqStep3Desc_outline: 'Imported content uses target knowledge base permissions. Disabling deletion keeps copies. Private attachments are not downloaded. Use HTTPS; HTTP is limited to trusted private hosts.',
+    prereqOpenConsole_outline: 'Outline API documentation',
     title: 'Data Sources',
     description: 'Configure external data sources to sync content into this knowledge base',
     add: 'Add Data Source',
@@ -6360,6 +6385,7 @@ export default {
       skip: 'Skip existing'
     },
     syncDeletions: 'Sync deletions (remove knowledge when deleted at source)',
+    syncDeletionsOutline: 'Sync deletions (remove documents deleted at source or outside the selected collections, including deselected collections, during sync)',
     createAndSync: 'Create & Sync Now',
     createAndSyncSuccess: 'Data source created and sync task submitted',
     createButSyncFailed: 'Data source created, but failed to trigger sync',
@@ -6400,6 +6426,7 @@ export default {
       docsFailedSummary: '{n} document(s) failed to sync'
     },
     connector: {
+      outline: 'Outline',
       feishu: 'Feishu',
       lark: 'Lark',
       feishu_drive: 'Feishu Drive',
@@ -6411,6 +6438,7 @@ export default {
       gitlab: 'GitLab'
     },
     connectorDesc: {
+      outline: 'Sync Markdown documents from Outline collections',
       feishu: 'Sync documents, spreadsheets and files from Feishu Wiki',
       lark: 'Sync documents, spreadsheets and files from Lark Wiki (Feishu international)',
       feishu_drive: 'Sync documents, spreadsheets and files from a Feishu Drive folder',
@@ -6514,6 +6542,7 @@ export default {
       '24h': 'Daily'
     },
     resourceType: {
+      collection: 'Collection',
       wikiSpace: 'Wiki Space',
       docCategory: 'Document Tag',
       book: 'Yuque Book'
````

### A.3 `frontend/src/i18n/locales/ja-JP.ts`

````diff
diff --git a/frontend/src/i18n/locales/ja-JP.ts b/frontend/src/i18n/locales/ja-JP.ts
index 2a12bc75..1ba09719 100644
--- a/frontend/src/i18n/locales/ja-JP.ts
+++ b/frontend/src/i18n/locales/ja-JP.ts
@@ -6291,6 +6291,31 @@ export default {
     daysAgo: '{days}日前'
   },
   datasource: {
+    outlineError: {
+      auth: '認証に失敗しました。API キーを更新してください。',
+      permission: 'アクセスが拒否されました。アカウントとコレクションの権限を確認してください。',
+      rate: 'レート制限に達しました。後で再試行してください。',
+      response: 'Outline の応答に対応していません。バージョンと Markdown API を確認してください。',
+      request: 'スキャンを完了できませんでした。接続を確認して再試行してください。コピーは保持されます。',
+      instance: 'インスタンスまたはワークスペースが変わりました。新しい接続を作成してください。',
+      size: '文書が 32 MiB の上限を超えています。',
+      cursor: '同期チェックポイントが無効です。履歴を消去せず管理者に連絡してください。',
+      pending: '一部の文書は次回の同期で再試行されます。',
+      deletion: '削除を確認できません。コピーは保持されます。documents.deleted の権限を確認してください。',
+      scope: 'アクセス可能で有効なコレクションを選択してください。',
+      running: '同期が待機中または実行中です。完了後に再試行してください。',
+    },
+    fullSyncNow: '完全同期を実行',
+    openSource: '原文を開く',
+    outlineBaseUrl: 'Outline インスタンス URL',
+    prereqBarText_outline: 'Outline 接続の要件',
+    prereqStep1Brief_outline: '読み取り専用 API キーを作成',
+    prereqStep1Desc_outline: '選択したコレクションにアクセスできる専用アカウントを使用してください。',
+    prereqStep2Brief_outline: '必要なエンドポイントを許可',
+    prereqStep2Desc_outline: 'auth.info、collections.list/info、documents.list/info が必要です。削除同期には documents.deleted も必要です。',
+    prereqStep3Brief_outline: '権限と保持設定を確認',
+    prereqStep3Desc_outline: '取り込み後は対象ナレッジベースの権限を使用します。削除を無効にするとコピーを保持します。非公開添付は取得しません。HTTPS を推奨し、HTTP は信頼済み内部ホストに限定します。',
+    prereqOpenConsole_outline: 'Outline API ドキュメント',
     title: 'データソース',
     description: 'このナレッジベースにコンテンツを同期する外部データソースを設定します',
     add: 'データソースを追加',
@@ -6360,6 +6385,7 @@ export default {
       skip: '既存をスキップ'
     },
     syncDeletions: '削除を同期（同期元で削除されたナレッジをナレッジベースからも削除）',
+    syncDeletionsOutline: '削除を同期（同期時に、同期元で削除された文書や選択範囲外の文書を削除。選択解除したコレクションも対象）',
     createAndSync: '作成して今すぐ同期',
     createAndSyncSuccess: 'データソースを作成し、同期タスクを送信しました',
     createButSyncFailed: 'データソースを作成しましたが、同期の開始に失敗しました',
@@ -6400,6 +6426,7 @@ export default {
       docsFailedSummary: '{n}件のドキュメントの同期に失敗しました'
     },
     connector: {
+      outline: 'Outline',
       feishu: 'Feishu',
       lark: 'Lark',
       feishu_drive: 'Feishu Drive',
@@ -6411,6 +6438,7 @@ export default {
       gitlab: 'GitLab'
     },
     connectorDesc: {
+      outline: 'Outline コレクションの Markdown 文書を同期',
       feishu: 'Feishu Wikiからドキュメント、スプレッドシート、ファイルを同期します',
       lark: 'Lark Wiki（Feishu国際版）からドキュメント、スプレッドシート、ファイルを同期します',
       feishu_drive: 'Feishu Driveのフォルダからドキュメント、スプレッドシート、ファイルを同期します',
@@ -6514,6 +6542,7 @@ export default {
       '24h': '毎日'
     },
     resourceType: {
+      collection: 'コレクション',
       wikiSpace: 'Wikiスペース',
       docCategory: 'ドキュメントタグ',
       book: 'Yuqueナレッジベース'
````

### A.4 `frontend/src/i18n/locales/ko-KR.ts`

````diff
diff --git a/frontend/src/i18n/locales/ko-KR.ts b/frontend/src/i18n/locales/ko-KR.ts
index a9e59315..700ffbe0 100755
--- a/frontend/src/i18n/locales/ko-KR.ts
+++ b/frontend/src/i18n/locales/ko-KR.ts
@@ -688,6 +688,31 @@ export default {
     }
   },
   datasource: {
+    outlineError: {
+      auth: '인증에 실패했습니다. API 키를 교체하세요.',
+      permission: '접근이 거부되었습니다. 계정 및 컬렉션 권한을 확인하세요.',
+      rate: '요청 제한에 도달했습니다. 나중에 다시 시도하세요.',
+      response: 'Outline 응답이 호환되지 않습니다. 버전과 Markdown API 지원을 확인하세요.',
+      request: '스캔을 완료하지 못했습니다. 연결을 확인하고 재시도하세요. 기존 사본은 보존됩니다.',
+      instance: '인스턴스 또는 작업공간이 변경되었습니다. 새 연결을 만드세요.',
+      size: '문서가 32 MiB 제한을 초과합니다.',
+      cursor: '동기화 체크포인트가 유효하지 않습니다. 기록을 지우지 말고 관리자에게 문의하세요.',
+      pending: '일부 문서는 다음 동기화에서 재시도됩니다.',
+      deletion: '삭제를 확인할 수 없어 사본을 보존합니다. documents.deleted 권한을 확인하세요.',
+      scope: '접근 가능한 활성 컬렉션을 하나 이상 선택하세요.',
+      running: '동기화가 대기 중이거나 실행 중입니다. 완료 후 다시 시도하세요.',
+    },
+    fullSyncNow: '전체 동기화 실행',
+    openSource: '원문 열기',
+    outlineBaseUrl: 'Outline 인스턴스 URL',
+    prereqBarText_outline: 'Outline 연결 요구 사항',
+    prereqStep1Brief_outline: '읽기 전용 API 키 생성',
+    prereqStep1Desc_outline: '선택한 컬렉션에 접근 가능한 전용 계정을 사용하세요.',
+    prereqStep2Brief_outline: '필요한 엔드포인트 허용',
+    prereqStep2Desc_outline: 'auth.info, collections.list/info, documents.list/info가 필요합니다. 삭제 동기화에는 documents.deleted도 필요합니다.',
+    prereqStep3Brief_outline: '권한 및 보존 정책 확인',
+    prereqStep3Desc_outline: '가져온 콘텐츠는 대상 지식베이스 권한을 따릅니다. 삭제를 끄면 사본을 보존합니다. 비공개 첨부 파일은 다운로드하지 않습니다. HTTPS를 권장하며 HTTP는 신뢰하는 내부 호스트로 제한됩니다.',
+    prereqOpenConsole_outline: 'Outline API 문서',
     title: '데이터 소스 관리',
     description: '외부 데이터 소스를 구성하여 콘텐츠를 지식베이스에 자동 동기화',
     add: '데이터 소스 추가',
@@ -744,6 +769,7 @@ export default {
     syncScheduleLabel: '동기화 주기',
     conflictLabel: '충돌 전략',
     syncDeletions: '삭제 동기화 (소스에서 삭제 시 지식베이스에서도 삭제)',
+    syncDeletionsOutline: '삭제 동기화 (동기화 시 소스에서 삭제되거나 선택 범위에서 제외된 문서를 정리합니다. 선택 해제한 컬렉션도 포함됩니다)',
     createAndSync: '생성 후 즉시 동기화',
     createAndSyncSuccess: '데이터 소스가 생성되었으며 동기화 작업이 제출되었습니다',
     createButSyncFailed: '데이터 소스가 생성되었으나 동기화 트리거에 실패했습니다',
@@ -816,6 +842,7 @@ export default {
     hoursAgo: '{n}시간 전',
     daysAgo: '{n}일 전',
     resourceType: {
+      collection: '컬렉션',
       wikiSpace: '위키 공간',
       docCategory: '문서 태그',
       book: 'Yuque 지식베이스'
@@ -842,6 +869,7 @@ export default {
       authHeadersHint: '비공개 피드 접근용. 한 줄에 하나씩 「이름: 값」 형식으로 입력하세요. 예: Authorization: Bearer xxxx'
     },
     connectorDesc: {
+      outline: 'Outline 컬렉션의 Markdown 문서 동기화',
       feishu: '페이슈 위키에서 문서, 스프레드시트, 파일 동기화',
       lark: 'Lark 위키에서 문서, 스프레드시트, 파일 동기화',
       feishu_drive: "페이슈 드라이브 폴더에서 문서, 스프레드시트, 파일 동기화",
@@ -853,6 +881,7 @@ export default {
       gitlab: 'GitLab 프로젝트의 파일 동기화'
     },
     connector: {
+      outline: 'Outline',
       feishu: '페이슈 (Feishu)',
       lark: 'Lark (Feishu 글로벌)',
       feishu_drive: "페이슈 드라이브",
````

### A.5 `frontend/src/i18n/locales/ru-RU.ts`

````diff
diff --git a/frontend/src/i18n/locales/ru-RU.ts b/frontend/src/i18n/locales/ru-RU.ts
index d9f9b552..50f736b6 100755
--- a/frontend/src/i18n/locales/ru-RU.ts
+++ b/frontend/src/i18n/locales/ru-RU.ts
@@ -688,6 +688,31 @@ export default {
     }
   },
   datasource: {
+    outlineError: {
+      auth: 'Ошибка аутентификации. Замените API-ключ.',
+      permission: 'Доступ запрещён. Проверьте права учётной записи и коллекций.',
+      rate: 'Достигнут лимит запросов. Повторите позже.',
+      response: 'Несовместимый ответ Outline. Проверьте версию и поддержку Markdown API.',
+      request: 'Сканирование не завершено. Проверьте подключение и повторите. Копии сохранены.',
+      instance: 'Экземпляр или рабочее пространство изменены. Создайте новое подключение.',
+      size: 'Документ превышает лимит 32 MiB.',
+      cursor: 'Некорректная контрольная точка. Обратитесь к администратору, не удаляйте историю.',
+      pending: 'Некоторые документы будут повторно обработаны при следующей синхронизации.',
+      deletion: 'Удаление не подтверждено. Копии сохранены; проверьте права documents.deleted.',
+      scope: 'Выберите хотя бы одну доступную активную коллекцию.',
+      running: 'Синхронизация ожидает или выполняется. Повторите после завершения.',
+    },
+    fullSyncNow: 'Полная синхронизация',
+    openSource: 'Открыть оригинал',
+    outlineBaseUrl: 'URL экземпляра Outline',
+    prereqBarText_outline: 'Требования подключения Outline',
+    prereqStep1Brief_outline: 'Создайте API-ключ только для чтения',
+    prereqStep1Desc_outline: 'Используйте отдельную учётную запись с доступом к выбранным коллекциям.',
+    prereqStep2Brief_outline: 'Разрешите необходимые методы',
+    prereqStep2Desc_outline: 'Нужны auth.info, collections.list/info и documents.list/info. Для удаления также нужен documents.deleted.',
+    prereqStep3Brief_outline: 'Проверьте доступ и хранение',
+    prereqStep3Desc_outline: 'Импорт использует права целевой базы знаний. Отключение удаления сохраняет копии. Закрытые вложения не загружаются. Рекомендуется HTTPS; HTTP доступен только для доверенных внутренних узлов.',
+    prereqOpenConsole_outline: 'Документация Outline API',
     title: 'Источники данных',
     description: 'Настройте внешние источники данных для автоматической синхронизации контента',
     add: 'Добавить источник',
@@ -744,6 +769,7 @@ export default {
     syncScheduleLabel: 'Расписание синхронизации',
     conflictLabel: 'Стратегия конфликтов',
     syncDeletions: 'Синхронизировать удаления (удалять знания при удалении в источнике)',
+    syncDeletionsOutline: 'Синхронизировать удаления (при синхронизации удалять документы, удалённые в источнике или вне выбранных коллекций, включая коллекции со снятым выбором)',
     createAndSync: 'Создать и синхронизировать',
     createAndSyncSuccess: 'Источник данных создан, задача синхронизации поставлена в очередь',
     createButSyncFailed: 'Источник данных создан, но не удалось запустить синхронизацию',
@@ -816,6 +842,7 @@ export default {
     hoursAgo: '{n} ч назад',
     daysAgo: '{n} д назад',
     resourceType: {
+      collection: 'Коллекция',
       wikiSpace: 'Пространство вики',
       docCategory: 'Тег документа',
       book: 'База знаний Yuque'
@@ -842,6 +869,7 @@ export default {
       authHeadersHint: 'Для приватных лент. По одному в строке в формате «Имя: Значение», например Authorization: Bearer xxxx'
     },
     connectorDesc: {
+      outline: 'Синхронизация Markdown-документов коллекций Outline',
       feishu: 'Синхронизация документов, таблиц и файлов из Feishu Wiki',
       lark: 'Синхронизация документов, таблиц и файлов из Lark Wiki',
       feishu_drive: 'Синхронизация документов, таблиц и файлов из папки Feishu Drive',
@@ -853,6 +881,7 @@ export default {
       gitlab: 'Синхронизация файлов из проектов GitLab'
     },
     connector: {
+      outline: 'Outline',
       feishu: 'Feishu (Фэйшу)',
       lark: 'Lark',
       feishu_drive: 'Feishu Drive',
````

### A.6 `frontend/src/i18n/locales/zh-CN.ts`

````diff
diff --git a/frontend/src/i18n/locales/zh-CN.ts b/frontend/src/i18n/locales/zh-CN.ts
index b8690560..f8634401 100755
--- a/frontend/src/i18n/locales/zh-CN.ts
+++ b/frontend/src/i18n/locales/zh-CN.ts
@@ -688,6 +688,31 @@ export default {
     }
   },
   datasource: {
+    outlineError: {
+      auth: '认证失败，请替换 API Key。',
+      permission: '访问被拒绝，请检查账号及 Collection 权限。',
+      rate: '触发限流，请稍后重试。',
+      response: 'Outline 响应不兼容，请检查部署版本及 Markdown API 支持。',
+      request: '扫描未能完成，请检查网络后重试；已有副本会保留。',
+      instance: '实例或工作区已变更，请创建新的连接。',
+      size: '文档超过 32 MiB 上限。',
+      cursor: '同步检查点无效，请联系管理员，不要直接清空历史基线。',
+      pending: '部分文档将在下次同步时重试。',
+      deletion: '无法确认删除，已保留副本；请检查 documents.deleted 权限。',
+      scope: '请至少选择一个可访问且有效的 Collection。',
+      running: '同步已排队或正在运行，请完成后再操作。',
+    },
+    fullSyncNow: '重新全量同步',
+    openSource: '打开原文',
+    outlineBaseUrl: 'Outline 实例地址',
+    prereqBarText_outline: 'Outline 接入要求',
+    prereqStep1Brief_outline: '创建只读 API Key',
+    prereqStep1Desc_outline: '建议使用专用同步账号，授权访问需要同步的 Collection。',
+    prereqStep2Brief_outline: '允许必要的读取端点',
+    prereqStep2Desc_outline: '需要 auth.info、collections.list/info、documents.list/info；同步删除还需 documents.deleted。',
+    prereqStep3Brief_outline: '确认访问权限和保留策略',
+    prereqStep3Desc_outline: '导入后按目标知识库权限访问。关闭删除会保留副本，私有附件不随正文下载。建议 HTTPS，HTTP 仅限可信内网。',
+    prereqOpenConsole_outline: 'Outline API 文档',
     title: '数据源管理',
     description: '配置外部数据源，自动同步内容到知识库',
     add: '添加数据源',
@@ -746,6 +771,7 @@ export default {
     syncScheduleLabel: '同步频率',
     conflictLabel: '冲突策略',
     syncDeletions: '同步删除（源端删除时同步删除知识库中的条目）',
+    syncDeletionsOutline: '同步删除（同步时清理源端已删除或已移出所选 Collection 的文档，含取消选择的 Collection）',
     createAndSync: '创建并立即同步',
     createAndSyncSuccess: '数据源创建成功，同步任务已提交',
     createButSyncFailed: '数据源已创建，但触发同步失败',
@@ -818,6 +844,7 @@ export default {
     hoursAgo: '{n} 小时前',
     daysAgo: '{n} 天前',
     resourceType: {
+      collection: '集合',
       wikiSpace: '知识库空间',
       docCategory: '文档标签',
       book: '语雀知识库'
@@ -844,6 +871,7 @@ export default {
       authHeadersHint: '用于访问私有订阅源，每行一个，格式为「名称: 值」，例如 Authorization: Bearer xxxx'
     },
     connectorDesc: {
+      outline: '同步 Outline Collection 中的 Markdown 文档',
       feishu: '同步飞书知识库中的文档、表格、文件',
       lark: '同步 Lark 知识库中的文档、表格、文件（飞书国际版）',
       feishu_drive: "同步飞书云盘文件夹中的文档、表格、文件",
@@ -855,6 +883,7 @@ export default {
       gitlab: '同步 GitLab 项目中的文件'
     },
     connector: {
+      outline: 'Outline',
       feishu: '飞书',
       lark: 'Lark（飞书国际版）',
       feishu_drive: "飞书云盘",
````

### A.7 `frontend/src/utils/datasourceError.ts`

````diff
diff --git a/frontend/src/utils/datasourceError.ts b/frontend/src/utils/datasourceError.ts
new file mode 100644
index 00000000..4669bf34
--- /dev/null
+++ b/frontend/src/utils/datasourceError.ts
@@ -0,0 +1,32 @@
+import i18n from '@/i18n'
+
+const categories: Record<string, string> = {
+  outline_auth_failed: 'auth',
+  outline_acknowledgement_required: 'response',
+  outline_permission_denied: 'permission',
+  outline_rate_limited: 'rate',
+  outline_response_invalid: 'response',
+  outline_format_unsupported: 'response',
+  outline_request_failed: 'request',
+  outline_scan_incomplete: 'request',
+  outline_instance_changed: 'instance',
+  outline_document_too_large: 'size',
+  outline_cursor_invalid: 'cursor',
+  outline_pending_upserts: 'pending',
+  outline_deletion_unconfirmed: 'deletion',
+  outline_deletion_capability_unavailable: 'deletion',
+  outline_collection_inactive: 'scope',
+  outline_collection_required: 'scope',
+  outline_invalid_collection_id: 'scope',
+  outline_redirect_rejected: 'request',
+  datasource_sync_running: 'running',
+  datasource_lease_lost: 'request',
+}
+
+export function localizeDatasourceError(message: unknown): string {
+  if (typeof message !== 'string') return ''
+  return message.replace(/outline_[a-z_]+|datasource_sync_running|datasource_lease_lost/g, (code) => {
+    const category = categories[code]
+    return category ? i18n.global.t(`datasource.outlineError.${category}`) : code
+  })
+}
````

### A.8 `frontend/src/views/knowledge/KnowledgeBase.vue`

````diff
diff --git a/frontend/src/views/knowledge/KnowledgeBase.vue b/frontend/src/views/knowledge/KnowledgeBase.vue
index 86eda162..68d13c50 100644
--- a/frontend/src/views/knowledge/KnowledgeBase.vue
+++ b/frontend/src/views/knowledge/KnowledgeBase.vue
@@ -601,6 +601,7 @@ const sourceOptions = computed(() => [
   { label: t('knowledgeBase.channelFeishuDrive'), value: 'feishu_drive' },
   { label: t('knowledgeBase.channelNotion'), value: 'notion' },
   { label: t('knowledgeBase.channelYuque'), value: 'yuque' },
+  { label: 'Outline', value: 'outline' },
   { label: t('knowledgeBase.channelGitLab'), value: 'gitlab' },
   { label: t('knowledgeBase.channelIma'), value: 'ima' },
   { label: t('knowledgeBase.channelWechat'), value: 'wechat' },
````

### A.9 `frontend/src/views/knowledge/components/DocumentActionMenu.vue`

````diff
diff --git a/frontend/src/views/knowledge/components/DocumentActionMenu.vue b/frontend/src/views/knowledge/components/DocumentActionMenu.vue
index e9e115a3..72778a58 100644
--- a/frontend/src/views/knowledge/components/DocumentActionMenu.vue
+++ b/frontend/src/views/knowledge/components/DocumentActionMenu.vue
@@ -8,6 +8,7 @@ interface KnowledgeItem {
   title?: string;
   type?: string;
   parse_status?: string;
+  metadata?: Record<string, unknown> | string;
 }
 
 const props = defineProps<{
@@ -40,9 +41,25 @@ const isParseInFlight = computed(() =>
 );
 
 const fileName = computed(() => props.item.file_name || props.item.title || props.item.id);
+const sourceUrl = computed(() => {
+  try {
+    const metadata = typeof props.item.metadata === 'string'
+      ? JSON.parse(props.item.metadata) : props.item.metadata;
+    if (typeof metadata?.source_url !== 'string') return '';
+    const url = new URL(metadata.source_url);
+    return ['https:', 'http:'].includes(url.protocol) && !url.username && !url.password ? url.href : '';
+  } catch {
+    return '';
+  }
+});
 </script>
 
 <template>
+  <a v-if="sourceUrl" :href="sourceUrl" target="_blank" rel="noopener noreferrer"
+    class="doc-action-menu-item" @click.stop>
+    <t-icon class="icon" name="link" />
+    <span>{{ t('datasource.openSource') }}</span>
+  </a>
   <!-- 下载原始文档 -->
   <div
     v-if="canDownload && (item.type === 'file' || item.type === 'manual')"
````

### A.10 `frontend/src/views/knowledge/components/DocumentListView.vue`

````diff
diff --git a/frontend/src/views/knowledge/components/DocumentListView.vue b/frontend/src/views/knowledge/components/DocumentListView.vue
index 2fb5fff8..a86116d0 100644
--- a/frontend/src/views/knowledge/components/DocumentListView.vue
+++ b/frontend/src/views/knowledge/components/DocumentListView.vue
@@ -114,6 +114,7 @@ const getSourceInfo = (item: KnowledgeItem): { icon: string; label: string } =>
   if (ch === 'lark_drive') return { icon: 'cloud-download', label: t('knowledgeBase.channelLarkDrive') };
   if (ch === 'notion') return { icon: 'cloud-download', label: t('knowledgeBase.channelNotion') };
   if (ch === 'yuque') return { icon: 'cloud-download', label: t('knowledgeBase.channelYuque') };
+  if (ch === 'outline') return { icon: 'cloud-download', label: 'Outline' };
   if (ch === 'gitlab') return { icon: 'cloud-download', label: t('knowledgeBase.channelGitLab') };
   if (ch === 'ima') return { icon: 'cloud-download', label: t('knowledgeBase.channelIma') };
   if (ch === 'wechat') return { icon: 'cloud-download', label: t('knowledgeBase.channelWechat') };
````

### A.11 `frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue`

````diff
diff --git a/frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue b/frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue
index c3cf2c42..42363b18 100644
--- a/frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue
+++ b/frontend/src/views/knowledge/settings/DataSourceEditorDialog.vue
@@ -1,4 +1,5 @@
 <script setup lang="ts">
+import { localizeDatasourceError } from '@/utils/datasourceError'
 import { ref, computed, watch } from 'vue'
 import { MessagePlugin } from 'tdesign-vue-next'
 import { useI18n } from 'vue-i18n'
@@ -93,7 +94,7 @@ async function confirmRemoveCredentials() {
     form.value.config.credentials = {}
     MessagePlugin.success(t('credential.removedToast'))
   } catch (e: any) {
-    MessagePlugin.error(e?.message || t('credential.removeFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message) || t('credential.removeFailed'))
   } finally {
     removingCredentials.value = false
   }
@@ -184,6 +185,7 @@ const form = ref({
 // Step 2: Resources
 const resources = ref<Resource[]>([])
 const loadingResources = ref(false)
+const resourceLoadError = ref('')
 const selectedResourceIds = ref<string[]>([])
 const expandedResourceIds = ref(new Set<string>())
 // Lazy loading: parents whose children have already been fetched, and parents
@@ -428,7 +430,7 @@ async function ensureChildrenLoaded(id: string) {
     }
     loadedChildrenIds.value = new Set(loadedChildrenIds.value).add(id)
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('datasource.resourceLoadFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message || e?.error) || t('datasource.resourceLoadFailed'))
     // Collapse again so the user can retry the expand.
     const next = new Set(expandedResourceIds.value)
     next.delete(id)
@@ -600,6 +602,18 @@ const connectorDefs = computed<ConnectorDef[]>(() => [
       { key: 'base_url', labelKey: 'datasource.field.baseUrl', placeholder: 'https://www.yuque.com', optional: true, hintKey: 'datasource.field.baseUrlHint' },
     ],
   },
+  {
+    type: 'outline',
+    available: true,
+    docUrl: 'https://docs.getoutline.com/s/guide/doc/api-1rEIXDfLF6',
+    permissionDocUrl: 'https://docs.getoutline.com/s/guide/doc/api-1rEIXDfLF6',
+    permissionPageUrl: '',
+    requiredPermissions: ['auth.info', 'collections.list', 'collections.info', 'documents.list', 'documents.info', 'documents.deleted'],
+    fields: [
+      { key: 'base_url', labelKey: 'datasource.outlineBaseUrl', placeholder: 'https://outline.example.com' },
+      { key: 'api_key', labelKey: 'datasource.field.apiToken', placeholder: '', secret: true },
+    ],
+  },
   {
     // Tencent IMA (ima.qq.com). Uses the OpenAPI at /openapi/wiki/v1 with two
     // static headers (ima-openapi-clientid + ima-openapi-apikey); no OAuth.
@@ -767,6 +781,8 @@ function selectType(def: ConnectorDef) {
   form.value.type = def.type
   form.value.name = t(`datasource.connector.${def.type}`)
   form.value.config.credentials = {}
+  resourceLoadError.value = ''
+  if (def.type === 'outline') form.value.conflict_strategy = 'overwrite'
   if (isGitLabConnector(def.type)) addGitLabProject()
   rssAuthHeaders.value = []
   step.value = 1
@@ -809,7 +825,7 @@ async function testConnection() {
     MessagePlugin.success(t('datasource.testSuccess'))
   } catch (e: any) {
     testResult.value = 'error'
-    testErrorMsg.value = e?.message || e?.error || ''
+    testErrorMsg.value = localizeDatasourceError(e?.message || e?.error)
     MessagePlugin.error(t('datasource.testFailed'))
   }
   testing.value = false
@@ -818,12 +834,15 @@ async function testConnection() {
 // --- Load resources ---
 async function loadResources() {
   loadingResources.value = true
+  resourceLoadError.value = ''
   try {
     if (!tempDsId.value) {
       const res = await createDataSource({
         ...form.value,
         knowledge_base_id: props.kbId,
         status: 'paused',
+        sync_schedule: form.value.type === 'outline' ? '' : form.value.sync_schedule,
+        sync_deletions: form.value.type === 'outline' ? false : form.value.sync_deletions,
       } as any)
       const created = res?.data || res
       tempDsId.value = created.id
@@ -831,6 +850,7 @@ async function loadResources() {
       await updateDataSource(tempDsId.value, {
         ...form.value,
         knowledge_base_id: props.kbId,
+        ...(form.value.type === 'outline' ? { status: 'paused', sync_schedule: '', sync_deletions: false } : {}),
       } as any)
     }
 
@@ -862,7 +882,8 @@ async function loadResources() {
       if (hidden.length > 0) void revealExistingSelections(hidden)
     }
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('datasource.resourceLoadFailed'))
+    resourceLoadError.value = localizeDatasourceError(e?.message || e?.error) || t('datasource.resourceLoadFailed')
+    MessagePlugin.error(resourceLoadError.value)
   }
   loadingResources.value = false
 }
@@ -883,7 +904,7 @@ async function revealExistingSelections(hiddenIds: string[]) {
     // selection itself); calls are independent and dedup on merge.
     await Promise.all(ancestors.map(id => ensureChildrenLoaded(id)))
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('datasource.resourceLoadFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message || e?.error) || t('datasource.resourceLoadFailed'))
   }
 }
 
@@ -992,6 +1013,13 @@ async function nextStep() {
       if ((testResult.value as string) !== 'success') return
     }
   }
+  if (step.value === 2 && form.value.type === 'outline') {
+    if (loadingResources.value || resourceLoadError.value) return
+    if (selectedResourceIds.value.length === 0) {
+      MessagePlugin.warning(t('datasource.outlineError.scope'))
+      return
+    }
+  }
   if (step.value === 2 && isDriveConnector(form.value.type)) {
     // folder_token 是 Drive 连接器的必填项：为空就地标错并留在本步,
     // 不允许带着空 token 进入同步策略。
@@ -1065,7 +1093,7 @@ async function commitCredentialsIfNeeded(dsId: string): Promise<boolean> {
     rssAuthHeaders.value = []
     return true
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('credential.saveFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message || e?.error) || t('credential.saveFailed'))
     return false
   }
 }
@@ -1111,7 +1139,7 @@ async function handleSubmit() {
         await triggerSync(dataSourceId)
         MessagePlugin.success(t('datasource.createAndSyncSuccess'))
       } catch (e: any) {
-        MessagePlugin.warning(e?.message || e?.error || t('datasource.createButSyncFailed'))
+        MessagePlugin.warning(localizeDatasourceError(e?.message || e?.error) || t('datasource.createButSyncFailed'))
       }
     }
 
@@ -1122,7 +1150,7 @@ async function handleSubmit() {
     tempDsId.value = ''
     visible.value = false
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('datasource.saveFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message || e?.error) || t('datasource.saveFailed'))
   }
   submitting.value = false
 }
@@ -1155,6 +1183,7 @@ function resourceIconName(r: Resource): string {
     case 'wiki_space':
       return 'root-list'
     case 'book':
+    case 'collection':
       return 'book'
     case 'doc_category':
       return 'folder-open'
@@ -1180,6 +1209,7 @@ const resourceTypeLabelMap: Record<string, string> = {
   wiki_space: 'datasource.resourceType.wikiSpace',
   doc_category: 'datasource.resourceType.docCategory',
   book: 'datasource.resourceType.book',
+  collection: 'datasource.resourceType.collection',
 }
 
 function resourceTypeLabel(type: string): string {
@@ -1621,6 +1651,13 @@ const drawerConfirmText = computed(() => {
       </div>
 
       <div v-else-if="loadingResources" class="ds-loading-center"><t-loading /></div>
+      <div v-else-if="form.type === 'outline' && resourceLoadError" class="ds-resource-empty" role="alert">
+        <p class="ds-empty-title">{{ resourceLoadError }}</p>
+        <t-button variant="text" @click="loadResources">
+          <template #icon><t-icon name="refresh" /></template>
+          {{ t('datasource.retryLoadResources') }}
+        </t-button>
+      </div>
       <div v-else-if="resources.length > 0" class="resource-picker">
         <div class="resource-picker__toolbar">
           <span class="resource-picker__count">
@@ -1793,6 +1830,7 @@ const drawerConfirmText = computed(() => {
               {{ t('datasource.conflict.overwrite') }}
             </button>
             <button
+              v-if="form.type !== 'outline'"
               type="button"
               class="option-pill"
               :class="{ 'is-active': form.conflict_strategy === 'skip' }"
@@ -1806,7 +1844,7 @@ const drawerConfirmText = computed(() => {
         </div>
 
         <div class="form-item form-item--flat">
-          <t-checkbox v-model="form.sync_deletions">{{ t('datasource.syncDeletions') }}</t-checkbox>
+          <t-checkbox v-model="form.sync_deletions">{{ t(form.type === 'outline' ? 'datasource.syncDeletionsOutline' : 'datasource.syncDeletions') }}</t-checkbox>
         </div>
       </section>
     </template>
````

### A.12 `frontend/src/views/knowledge/settings/DataSourceSettings.vue`

````diff
diff --git a/frontend/src/views/knowledge/settings/DataSourceSettings.vue b/frontend/src/views/knowledge/settings/DataSourceSettings.vue
index d8ed8ed5..39b9d417 100644
--- a/frontend/src/views/knowledge/settings/DataSourceSettings.vue
+++ b/frontend/src/views/knowledge/settings/DataSourceSettings.vue
@@ -1,4 +1,5 @@
 <script setup lang="ts">
+import { localizeDatasourceError } from '@/utils/datasourceError'
 import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
 import { MessagePlugin } from 'tdesign-vue-next'
 import { useI18n } from 'vue-i18n'
@@ -95,13 +96,13 @@ async function removeDataSource(ds: DataSource) {
   }
 }
 
-async function handleSync(ds: DataSource) {
+async function handleSync(ds: DataSource, forceFull = false) {
   try {
-    await triggerSync(ds.id)
+    await triggerSync(ds.id, forceFull)
     MessagePlugin.success(t('datasource.syncTriggered'))
     await loadList(true)
   } catch (e: any) {
-    MessagePlugin.error(e?.message || e?.error || t('datasource.syncFailed'))
+    MessagePlugin.error(localizeDatasourceError(e?.message || e?.error) || t('datasource.syncFailed'))
   }
 }
 
@@ -235,6 +236,13 @@ onBeforeUnmount(stopPolling)
                         <t-icon name="refresh" :class="{ 'ds-icon-spin': isSyncRunning(ds) }" />
                         {{ isSyncRunning(ds) ? t('datasource.logStatus.running') : t('datasource.syncNow') }}
                       </t-dropdown-item>
+                      <t-dropdown-item
+                        v-if="canManageDataSource && ds.type === 'outline'"
+                        :disabled="isSyncRunning(ds)"
+                        @click="handleSync(ds, true)"
+                      >
+                        <t-icon name="refresh" /> {{ t('datasource.fullSyncNow') }}
+                      </t-dropdown-item>
                       <t-dropdown-item @click="openLogs(ds)">
                         <t-icon name="root-list" /> {{ t('datasource.logs') }}
                       </t-dropdown-item>
````

### A.13 `frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue`

````diff
diff --git a/frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue b/frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue
index b613aa40..bcd2ffe3 100644
--- a/frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue
+++ b/frontend/src/views/knowledge/settings/DataSourceSyncLogs.vue
@@ -2,6 +2,7 @@
 import { ref, watch, computed } from 'vue'
 import { useI18n } from 'vue-i18n'
 import { getSyncLogs, type SyncLog, type SyncItemError } from '@/api/datasource'
+import { localizeDatasourceError } from '@/utils/datasourceError'
 
 const props = defineProps<{
   dataSourceId: string
@@ -152,6 +153,7 @@ function formatSyncError(e: SyncItemError): string {
   } else {
     reason = e.message || ''
   }
+  reason = localizeDatasourceError(reason)
   return e.title ? (reason ? `${e.title} — ${reason}` : e.title) : reason
 }
 
@@ -283,7 +285,7 @@ const groupedLogs = computed(() => {
                   {{ t('datasource.logDetail.docsFailedSummary', { n: log.items_failed }) }}
                 </div>
                 <div v-else-if="log.error_message" class="tl-error">
-                  {{ log.error_message }}
+                  {{ localizeDatasourceError(log.error_message) }}
                 </div>
 
                 <!-- Per-item failures: which documents failed and why.
````

### A.14 `frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue`

````diff
diff --git a/frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue b/frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue
index a3cf664a..3b5de796 100644
--- a/frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue
+++ b/frontend/src/views/knowledge/settings/DataSourceTypeIcon.vue
@@ -44,6 +44,7 @@ function fallbackText(type: string) {
       class="ds-type-icon__img"
       :style="variant === 'inline' ? { width: `${size}px`, height: `${size}px` } : undefined"
     >
+    <t-icon v-else-if="type === 'outline'" name="root-list" :size="size" />
     <span v-else class="ds-type-icon-fallback">{{ fallbackText(type) }}</span>
   </span>
 </template>
````

### A.15 `go.mod`

````diff
diff --git a/go.mod b/go.mod
index c70e42eb..1289570f 100644
--- a/go.mod
+++ b/go.mod
@@ -72,6 +72,7 @@ require (
 	github.com/weaviate/weaviate-go-client/v5 v5.7.3
 	github.com/xuri/excelize/v2 v2.11.0
 	github.com/yanyiwu/gojieba v1.4.7
+	github.com/yuin/goldmark v1.8.2
 	go.opentelemetry.io/otel v1.43.0
 	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.43.0
 	go.opentelemetry.io/otel/sdk v1.43.0
````

### A.16 `internal/application/repository/datasource_repo.go`

````diff
diff --git a/internal/application/repository/datasource_repo.go b/internal/application/repository/datasource_repo.go
index c3b2f58e..7de5f887 100644
--- a/internal/application/repository/datasource_repo.go
+++ b/internal/application/repository/datasource_repo.go
@@ -86,6 +86,12 @@ func (r *DataSourceRepository) Update(ctx context.Context, ds *types.DataSource)
 		return errors.New("data source id is empty")
 	}
 	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
+		if ds.Type == types.ConnectorTypeOutline {
+			return tx.Model(ds).Where("deleted_at IS NULL").Select(
+				"name", "config", "sync_schedule", "sync_mode", "status", "conflict_strategy",
+				"sync_deletions", "last_sync_cursor", "error_message", "sync_log_retention_days",
+			).Updates(ds).Error
+		}
 		if err := tx.Model(ds).Updates(ds).Error; err != nil {
 			return err
 		}
@@ -107,11 +113,15 @@ func (r *DataSourceRepository) UpdateSyncState(ctx context.Context, ds *types.Da
 	if ds.ID == "" {
 		return errors.New("data source id is empty")
 	}
+	var status interface{} = ds.Status
+	if ds.Type == types.ConnectorTypeOutline {
+		status = gorm.Expr("CASE WHEN status = ? THEN status ELSE ? END", types.DataSourceStatusPaused, ds.Status)
+	}
 	if err := r.db.WithContext(ctx).
 		Model(&types.DataSource{}).
-		Where("id = ?", ds.ID).
+		Where("id = ? AND deleted_at IS NULL", ds.ID).
 		Updates(map[string]interface{}{
-			"status":           ds.Status,
+			"status":           status,
 			"last_sync_at":     ds.LastSyncAt,
 			"last_sync_cursor": ds.LastSyncCursor,
 			"last_sync_result": ds.LastSyncResult,
@@ -123,6 +133,12 @@ func (r *DataSourceRepository) UpdateSyncState(ctx context.Context, ds *types.Da
 	return nil
 }
 
+func (r *DataSourceRepository) Pause(ctx context.Context, id string) error {
+	return r.db.WithContext(ctx).Model(&types.DataSource{}).
+		Where("id = ? AND deleted_at IS NULL", id).
+		Update("status", types.DataSourceStatusPaused).Error
+}
+
 // Delete performs a soft delete
 func (r *DataSourceRepository) Delete(ctx context.Context, id string) error {
 	if id == "" {
````

### A.17 `internal/application/repository/knowledge.go`

````diff
diff --git a/internal/application/repository/knowledge.go b/internal/application/repository/knowledge.go
index db94133e..13a0b1a1 100644
--- a/internal/application/repository/knowledge.go
+++ b/internal/application/repository/knowledge.go
@@ -362,8 +362,15 @@ func (r *knowledgeRepository) CheckKnowledgeExists(
 	kbID string,
 	params *types.KnowledgeCheckParams,
 ) (bool, *types.Knowledge, error) {
+	if (params.DataSourceID == "") != (params.ExternalID == "") {
+		return false, nil, errors.New("datasource_id and external_id must be supplied together")
+	}
 	query := r.db.WithContext(ctx).Model(&types.Knowledge{}).
 		Where("tenant_id = ? AND knowledge_base_id = ? AND parse_status <> ?", tenantID, kbID, "failed")
+	if params.DataSourceID != "" {
+		query = query.Where("metadata->>'datasource_id' = ? AND metadata->>'external_id' = ?",
+			params.DataSourceID, params.ExternalID)
+	}
 
 	switch params.Type {
 	case "file":
@@ -805,6 +812,24 @@ func (r *knowledgeRepository) FindByDataSourceExternalID(
 	return &knowledge, nil
 }
 
+func (r *knowledgeRepository) ListDataSourceKnowledgeMetadata(
+	ctx context.Context, tenantID uint64, kbID, dataSourceID, afterID string, limit int,
+) ([]*types.Knowledge, error) {
+	if tenantID == 0 || kbID == "" || dataSourceID == "" || limit <= 0 || limit > 1000 {
+		return nil, errors.New("invalid data source metadata query scope or limit")
+	}
+	var rows []*types.Knowledge
+	query := r.db.WithContext(ctx).Model(&types.Knowledge{}).
+		Select("id", "channel", "metadata").
+		Where("tenant_id = ? AND knowledge_base_id = ? AND deleted_at IS NULL", tenantID, kbID).
+		Where("metadata->>'datasource_id' = ?", dataSourceID)
+	if afterID != "" {
+		query = query.Where("id > ?", afterID)
+	}
+	err := query.Order("id ASC").Limit(limit).Find(&rows).Error
+	return rows, err
+}
+
 // HardDeleteKnowledge physically removes a knowledge row. Call it AFTER
 // DeleteKnowledge's soft-delete cascade so sync-internal deletions never
 // become tombstones that block a later re-sync of the same external item.
````

### A.18 `internal/application/repository/outline_test.go`

````diff
diff --git a/internal/application/repository/outline_test.go b/internal/application/repository/outline_test.go
new file mode 100644
index 00000000..c9680435
--- /dev/null
+++ b/internal/application/repository/outline_test.go
@@ -0,0 +1,75 @@
+package repository
+
+import (
+	"context"
+	"fmt"
+	"testing"
+
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/google/uuid"
+	"github.com/stretchr/testify/require"
+)
+
+func TestOutlineFileIdentityDeduplication(t *testing.T) {
+	db := setupKnowledgeTestDB(t)
+	repo := NewKnowledgeRepository(db)
+	ctx := context.Background()
+	kb, id := uuid.NewString(), uuid.NewString()
+	require.NoError(t, db.Exec(`INSERT INTO knowledges
+		(id, tenant_id, knowledge_base_id, type, title, parse_status, file_hash, file_type, metadata)
+		VALUES (?, 1, ?, 'file', 'same', 'completed', 'hash', 'md', ?)`,
+		id, kb, `{"datasource_id":"ds-a","external_id":"a"}`).Error)
+	for _, tc := range []struct {
+		ds, external string
+		duplicate    bool
+	}{
+		{"ds-a", "a", true}, {"ds-a", "b", false}, {"ds-b", "a", false}, {"", "", true},
+	} {
+		t.Run(fmt.Sprintf("%s/%s", tc.ds, tc.external), func(t *testing.T) {
+			found, _, err := repo.CheckKnowledgeExists(ctx, 1, kb, &types.KnowledgeCheckParams{
+				Type: "file", FileHash: "hash", FileType: "md", DataSourceID: tc.ds, ExternalID: tc.external,
+			})
+			require.NoError(t, err)
+			require.Equal(t, tc.duplicate, found)
+		})
+	}
+	_, _, err := repo.CheckKnowledgeExists(ctx, 1, kb, &types.KnowledgeCheckParams{DataSourceID: "ds-a"})
+	require.Error(t, err)
+}
+
+func TestOutlinePauseDoesNotLoseCheckpoint(t *testing.T) {
+	db := setupDataSourceRepoTestDB(t)
+	repo := NewDataSourceRepository(db)
+	ctx := context.Background()
+	ds := &types.DataSource{ID: "outline-pause", TenantID: 1, KnowledgeBaseID: "kb", Name: "Outline",
+		Type: types.ConnectorTypeOutline, Status: types.DataSourceStatusActive}
+	require.NoError(t, repo.Create(ctx, ds))
+	require.NoError(t, repo.(*DataSourceRepository).Pause(ctx, ds.ID))
+	ds.LastSyncCursor = types.JSON(`{"checkpoint":1}`)
+	require.NoError(t, repo.UpdateSyncState(ctx, ds))
+	stored, err := repo.FindByID(ctx, ds.ID)
+	require.NoError(t, err)
+	require.Equal(t, types.DataSourceStatusPaused, stored.Status)
+	require.JSONEq(t, `{"checkpoint":1}`, stored.LastSyncCursor.ToString())
+}
+
+func TestOutlineUpdateCanClearScheduleAndCredentials(t *testing.T) {
+	db := setupDataSourceRepoTestDB(t)
+	repo := NewDataSourceRepository(db)
+	ctx := context.Background()
+	ds := &types.DataSource{ID: "outline-clear", TenantID: 1, KnowledgeBaseID: "kb", Type: types.ConnectorTypeOutline,
+		Name: "Outline", Status: types.DataSourceStatusActive, SyncSchedule: "0 0 */6 * * *", SyncDeletions: true,
+		Config: types.JSON(`{"type":"outline","credentials":{"base_url":"https://outline.example.com","api_key":"test-only"}}`)}
+	require.NoError(t, repo.Create(ctx, ds))
+	ds.SyncSchedule = ""
+	ds.SyncDeletions = false
+	ds.Config = types.JSON(`{"type":"outline","resource_ids":[]}`)
+	require.NoError(t, repo.Update(ctx, ds))
+	stored, err := repo.FindByID(ctx, ds.ID)
+	require.NoError(t, err)
+	require.Empty(t, stored.SyncSchedule)
+	require.False(t, stored.SyncDeletions)
+	cfg, err := stored.ParseConfig()
+	require.NoError(t, err)
+	require.False(t, cfg.HasConfiguredCredentials(types.ConnectorTypeOutline))
+}
````

### A.19 `internal/application/service/datasource_outline_test.go`

````diff
diff --git a/internal/application/service/datasource_outline_test.go b/internal/application/service/datasource_outline_test.go
new file mode 100644
index 00000000..f134d1be
--- /dev/null
+++ b/internal/application/service/datasource_outline_test.go
@@ -0,0 +1,162 @@
+package service
+
+import (
+	"context"
+	"encoding/json"
+	"errors"
+	"github.com/hibiken/asynq"
+	"mime/multipart"
+	"testing"
+
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/stretchr/testify/require"
+)
+
+func TestAcknowledgingHandlerReportsFailureAndDeferred(t *testing.T) {
+	result := &types.SyncResult{}
+	h := newStreamHandler(&DataSourceService{}, &types.DataSource{SyncDeletions: false}, result, &types.SyncLog{})
+	outcome, err := h.EmitWithResult(context.Background(), types.FetchedItem{ExternalID: "a", IsDeleted: true})
+	require.NoError(t, err)
+	require.Equal(t, datasource.ApplyDeferred, outcome.Outcome)
+	require.Zero(t, result.Total)
+	outcome, err = h.EmitWithResult(context.Background(), types.FetchedItem{
+		ExternalID: "a", Metadata: map[string]string{"error": "fetch failed"},
+	})
+	require.NoError(t, err)
+	require.Equal(t, datasource.ApplyFailed, outcome.Outcome)
+	require.Equal(t, 1, result.Total)
+	require.Equal(t, 1, result.Failed)
+}
+
+func TestOutlineOptionsAllowOnlyPausedEmptyDraft(t *testing.T) {
+	svc := &DataSourceService{}
+	ds := &types.DataSource{Type: types.ConnectorTypeOutline, Status: types.DataSourceStatusPaused}
+	cfg := &types.DataSourceConfig{}
+	require.NoError(t, svc.validateOutlineOptions(ds, cfg))
+	ds.SyncSchedule = "0 0 */6 * * *"
+	require.Error(t, svc.validateOutlineOptions(ds, cfg))
+	cfg.ResourceIDs = []string{"collection"}
+	require.NoError(t, svc.validateOutlineOptions(ds, cfg))
+	ds.ConflictStrategy = "skip"
+	require.Error(t, svc.validateOutlineOptions(ds, cfg))
+	ds.ConflictStrategy = "overwrite"
+	ds.SyncSchedule = "* * * * *"
+	require.Error(t, svc.validateOutlineOptions(ds, cfg))
+}
+
+type outlineAcceptanceKS struct {
+	*sweepFakeKS
+	created *types.Knowledge
+}
+
+func (k *outlineAcceptanceKS) CreateKnowledgeFromFile(context.Context, string, *multipart.FileHeader,
+	map[string]string, *bool, string, []string, string, *types.KnowledgeProcessOverrides) (*types.Knowledge, error) {
+	return k.created, nil
+}
+
+func TestOutlineRequiresParseAcceptanceAndStrictLookup(t *testing.T) {
+	repo := &deletionLookupKnowledgeRepo{}
+	ks := &outlineAcceptanceKS{sweepFakeKS: &sweepFakeKS{repo: repo},
+		created: &types.Knowledge{ID: "a", ParseStatus: types.ParseStatusFailed}}
+	svc := &DataSourceService{knowledgeService: ks}
+	ds := &types.DataSource{ID: "ds", Type: types.ConnectorTypeOutline}
+	item := &types.FetchedItem{ExternalID: "a", FileName: "a.md", Content: []byte("# a")}
+	_, err := svc.ingestItem(context.Background(), ds, item, nil)
+	require.Error(t, err)
+	ks.created.ParseStatus = types.ParseStatusPending
+	_, err = svc.ingestItem(context.Background(), ds, item, nil)
+	require.NoError(t, err)
+	repo.lookupErr = errors.New("lookup unavailable")
+	_, err = svc.ingestItem(context.Background(), ds, item, nil)
+	require.ErrorIs(t, err, repo.lookupErr)
+}
+
+func TestOutlineSameVersionIsIdempotentUnlessNewFullRun(t *testing.T) {
+	repo := &deletionLookupKnowledgeRepo{knowledge: &types.Knowledge{ID: "old", ParseStatus: types.ParseStatusCompleted,
+		Metadata: types.JSON(`{"source_fingerprint":"fp","source_sync_run_id":"run-a"}`)}}
+	ks := &sweepFakeKS{repo: repo}
+	svc := &DataSourceService{knowledgeService: ks}
+	ds := &types.DataSource{ID: "ds", Type: types.ConnectorTypeOutline}
+	item := &types.FetchedItem{ExternalID: "a", FileName: "a.md", Content: []byte("# a"),
+		Metadata: map[string]string{"source_fingerprint": "fp"}}
+	updated, err := svc.ingestItem(context.Background(), ds, item, nil)
+	require.NoError(t, err)
+	require.True(t, updated)
+	require.Empty(t, ks.events)
+	item.Metadata["source_full_sync"] = "true"
+	item.Metadata["source_sync_run_id"] = "run-b"
+	_, err = svc.ingestItem(context.Background(), ds, item, nil)
+	require.NoError(t, err)
+	require.Equal(t, []string{"delete:old", "create:a.md"}, ks.events)
+}
+
+type deletionPolicyConnector struct{ deletedItemConnector }
+
+func (deletionPolicyConnector) Type() string { return types.ConnectorTypeOutline }
+func (deletionPolicyConnector) Validate(_ context.Context, cfg *types.DataSourceConfig) error {
+	if cfg.SyncDeletions {
+		return errors.New("documents.deleted permission required")
+	}
+	return nil
+}
+
+func TestOutlineEnablingDeletionRevalidatesUnchangedCredentials(t *testing.T) {
+	cfg := &types.DataSourceConfig{Type: types.ConnectorTypeOutline, ResourceIDs: []string{"collection"},
+		Credentials: map[string]interface{}{"base_url": "https://outline.example.com", "api_key": "key"}}
+	blob, err := cfg.ToJSON()
+	require.NoError(t, err)
+	existing := &types.DataSource{ID: "outline-policy", TenantID: 1, KnowledgeBaseID: "kb",
+		Type: types.ConnectorTypeOutline, Status: types.DataSourceStatusActive, Config: blob}
+	registry := datasource.NewConnectorRegistry()
+	require.NoError(t, registry.Register(deletionPolicyConnector{}))
+	svc := &DataSourceService{dsRepo: newKBDeleteDSRepo("kb", existing), connectorRegistry: registry,
+		syncLogRepo: &processSyncSyncLogRepo{logs: map[string]*types.SyncLog{}}}
+	incoming := *existing
+	incoming.SyncDeletions = true
+	_, err = svc.UpdateDataSource(context.Background(), &incoming)
+	require.ErrorContains(t, err, "documents.deleted permission required")
+	require.False(t, existing.SyncDeletions)
+}
+
+type failedOutlineCheckpointRepo struct{ recordingDSRepo }
+
+func (r *failedOutlineCheckpointRepo) UpdateSyncState(context.Context, *types.DataSource) error {
+	return errors.New("checkpoint database unavailable")
+}
+
+func TestFailedStreamCheckpointKeepsLastPersistedCursor(t *testing.T) {
+	ds := &types.DataSource{ID: "ds", LastSyncCursor: types.JSON(`{"connector_cursor":{"saved":true}}`)}
+	h := newStreamHandler(&DataSourceService{dsRepo: &failedOutlineCheckpointRepo{}}, ds, &types.SyncResult{}, &types.SyncLog{})
+	err := h.Checkpoint(context.Background(), &types.SyncCursor{ConnectorCursor: map[string]interface{}{"saved": false}})
+	require.ErrorContains(t, err, "checkpoint database unavailable")
+	require.JSONEq(t, `{"connector_cursor":{"saved":true}}`, string(ds.LastSyncCursor))
+}
+
+func TestAcknowledgementDistinguishesMissingIdentityFromAlreadyDeleted(t *testing.T) {
+	repo := &deletionLookupKnowledgeRepo{}
+	ks := &sweepFakeKS{repo: repo}
+	ds := &types.DataSource{ID: "ds", SyncDeletions: true}
+	h := newStreamHandler(&DataSourceService{knowledgeService: ks}, ds, &types.SyncResult{}, &types.SyncLog{})
+	invalid, err := h.EmitWithResult(context.Background(), types.FetchedItem{IsDeleted: true})
+	require.NoError(t, err)
+	require.Equal(t, datasource.ApplyFailed, invalid.Outcome)
+	gone, err := h.EmitWithResult(context.Background(), types.FetchedItem{ExternalID: "gone", IsDeleted: true})
+	require.NoError(t, err)
+	require.Equal(t, datasource.ApplyApplied, gone.Outcome)
+}
+
+func TestOutlineDuplicateCompletedTaskPreservesResult(t *testing.T) {
+	for _, status := range []string{types.SyncLogStatusSuccess, types.SyncLogStatusPartial, types.SyncLogStatusCanceled} {
+		t.Run(status, func(t *testing.T) {
+			ds := &types.DataSource{ID: "outline-completed", TenantID: 1, Type: types.ConnectorTypeOutline, KnowledgeBaseID: "kb"}
+			log := &types.SyncLog{ID: "log", DataSourceID: ds.ID, Status: status, ItemsCreated: 9}
+			svc := &DataSourceService{dsRepo: newKBDeleteDSRepo("kb", ds), syncLogRepo: &processSyncSyncLogRepo{logs: map[string]*types.SyncLog{log.ID: log}}}
+			payload, err := json.Marshal(types.DataSourceSyncPayload{DataSourceID: ds.ID, TenantID: ds.TenantID, SyncLogID: log.ID})
+			require.NoError(t, err)
+			require.NoError(t, svc.ProcessSync(context.Background(), asynq.NewTask(types.TypeDataSourceSync, payload)))
+			require.Equal(t, status, log.Status)
+			require.Equal(t, 9, log.ItemsCreated)
+		})
+	}
+}
````

### A.20 `internal/application/service/datasource_service.go`

````diff
diff --git a/internal/application/service/datasource_service.go b/internal/application/service/datasource_service.go
index 950c0aa3..e7d4c628 100644
--- a/internal/application/service/datasource_service.go
+++ b/internal/application/service/datasource_service.go
@@ -21,6 +21,7 @@ import (
 	"github.com/Tencent/WeKnora/internal/types/interfaces"
 	secutils "github.com/Tencent/WeKnora/internal/utils"
 	"github.com/hibiken/asynq"
+	"github.com/robfig/cron/v3"
 )
 
 // DataSourceService implements the DataSourceService interface
@@ -155,6 +156,17 @@ func (s *DataSourceService) UpdateDataSource(ctx context.Context, ds *types.Data
 	if err != nil {
 		return nil, err
 	}
+	if existing.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, existing)
+		if err != nil {
+			return nil, err
+		}
+		defer release()
+		if ds.Type != existing.Type {
+			return nil, datasource.ErrInvalidConfig
+		}
+	}
 
 	if ds.KnowledgeBaseID == "" {
 		ds.KnowledgeBaseID = existing.KnowledgeBaseID
@@ -209,7 +221,15 @@ func (s *DataSourceService) UpdateDataSource(ctx context.Context, ds *types.Data
 		configActuallyChanged = !reflect.DeepEqual(*mergedCfg, *existingParsedCfg)
 	}
 	hasCreds := mergedCfg != nil && mergedCfg.HasConfiguredCredentials(ds.Type)
-	if hasCreds && (ds.Type != existing.Type || configActuallyChanged) {
+	if ds.Type == types.ConnectorTypeOutline {
+		ds.LastSyncCursor = existing.LastSyncCursor
+		if err := s.validateOutlineOptions(ds, mergedCfg); err != nil {
+			return nil, err
+		}
+	}
+	outlinePolicyChanged := ds.Type == types.ConnectorTypeOutline &&
+		(ds.SyncDeletions != existing.SyncDeletions || ds.Status != existing.Status)
+	if hasCreds && (ds.Type != existing.Type || configActuallyChanged || outlinePolicyChanged) {
 		if err := s.validateDataSourceConfig(ctx, ds); err != nil {
 			return nil, err
 		}
@@ -247,6 +267,14 @@ func (s *DataSourceService) UpdateDataSourceCredentials(
 	if err != nil {
 		return nil, err
 	}
+	if existing.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, existing)
+		if err != nil {
+			return nil, err
+		}
+		defer release()
+	}
 	parsed, err := existing.ParseConfig()
 	if err != nil {
 		return nil, err
@@ -288,6 +316,14 @@ func (s *DataSourceService) ClearDataSourceCredentials(ctx context.Context, id s
 	if err != nil {
 		return err
 	}
+	if existing.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, existing)
+		if err != nil {
+			return err
+		}
+		defer release()
+	}
 	parsed, err := existing.ParseConfig()
 	if err != nil {
 		return err
@@ -354,6 +390,25 @@ func (s *DataSourceService) ValidateConnection(ctx context.Context, dsID string)
 	if err != nil {
 		return err
 	}
+	if ds.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, ds)
+		if err != nil {
+			return err
+		}
+		defer release()
+		// Read-only validation must not unpause a draft or overwrite its cursor.
+		config, err := ds.ParseConfig()
+		if err != nil {
+			return err
+		}
+		config.SyncDeletions = ds.SyncDeletions
+		connector, err := s.connectorRegistry.Get(ds.Type)
+		if err != nil {
+			return err
+		}
+		return connector.Validate(ctx, config)
+	}
 
 	// Get connector
 	connector, err := s.connectorRegistry.Get(ds.Type)
@@ -454,10 +509,22 @@ func (s *DataSourceService) ResolveResourceAncestors(
 
 // ManualSync triggers an immediate sync for a data source
 func (s *DataSourceService) ManualSync(ctx context.Context, dsID string) (*types.SyncLog, error) {
+	return s.ManualSyncWithOptions(ctx, dsID, false)
+}
+
+func (s *DataSourceService) ManualSyncWithOptions(ctx context.Context, dsID string, forceFull bool) (*types.SyncLog, error) {
 	ds, err := s.GetDataSource(ctx, dsID)
 	if err != nil {
 		return nil, err
 	}
+	if ds.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, ds)
+		if err != nil {
+			return nil, err
+		}
+		defer release()
+	}
 
 	if ds.Status != types.DataSourceStatusActive &&
 		ds.Status != types.DataSourceStatusError &&
@@ -483,7 +550,7 @@ func (s *DataSourceService) ManualSync(ctx context.Context, dsID string) (*types
 		DataSourceID: dsID,
 		TenantID:     ds.TenantID,
 		SyncLogID:    syncLog.ID,
-		ForceFull:    false,
+		ForceFull:    forceFull,
 		Initiator:    types.TaskInitiatorFromContext(ctx),
 		Trigger:      "manual",
 	}
@@ -529,7 +596,14 @@ func (s *DataSourceService) PauseDataSource(ctx context.Context, id string) erro
 	}
 
 	ds.Status = types.DataSourceStatusPaused
-	if err := s.dsRepo.Update(ctx, ds); err != nil {
+	if repo, ok := s.dsRepo.(interface {
+		Pause(context.Context, string) error
+	}); ok && ds.Type == types.ConnectorTypeOutline {
+		err = repo.Pause(ctx, id)
+	} else {
+		err = s.dsRepo.Update(ctx, ds)
+	}
+	if err != nil {
 		logger.Errorf(ctx, "failed to pause data source: %v", err)
 		return err
 	}
@@ -551,6 +625,18 @@ func (s *DataSourceService) ResumeDataSource(ctx context.Context, id string) err
 	}
 
 	ds.Status = types.DataSourceStatusActive
+	if ds.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = s.lockOutlineMutation(ctx, ds)
+		if err != nil {
+			return err
+		}
+		defer release()
+		ds.Status = types.DataSourceStatusActive
+		if err := s.validateDataSourceConfig(ctx, ds); err != nil {
+			return err
+		}
+	}
 	if err := s.dsRepo.Update(ctx, ds); err != nil {
 		logger.Errorf(ctx, "failed to resume data source: %v", err)
 		return err
@@ -612,6 +698,26 @@ func (s *DataSourceService) ProcessSync(ctx context.Context, task *asynq.Task) e
 		return nil
 	}
 
+	if ds.Type == types.ConnectorTypeOutline {
+		var cancel context.CancelFunc
+		ctx, cancel = context.WithTimeout(ctx, 2*time.Hour)
+		defer cancel()
+		var release func()
+		ctx, release, err = datasource.DefaultSyncLocks.Acquire(ctx, ds.TenantID, ds.ID)
+		if errors.Is(err, datasource.ErrSyncRunning) {
+			return err
+		}
+		if err != nil {
+			return err
+		}
+		defer release()
+		// Reload under the lease; a queued job must not use stale credentials.
+		ds, err = s.GetDataSource(ctx, payload.DataSourceID)
+		if err != nil {
+			return err
+		}
+	}
+
 	// Get sync log
 	syncLog, err := s.syncLogRepo.FindByID(ctx, payload.SyncLogID)
 	if err != nil {
@@ -619,6 +725,14 @@ func (s *DataSourceService) ProcessSync(ctx context.Context, task *asynq.Task) e
 		return nil
 	}
 
+	if ds.Type == types.ConnectorTypeOutline {
+		switch syncLog.Status {
+		case types.SyncLogStatusSuccess, types.SyncLogStatusPartial, types.SyncLogStatusCanceled:
+			// A duplicate queue delivery must not overwrite the completed run's counts.
+			return nil
+		}
+	}
+
 	kb, kbErr := s.kbService.GetKnowledgeBaseByID(ctx, ds.KnowledgeBaseID)
 	if kbErr != nil {
 		logger.Warnf(ctx, "knowledge base not found (likely deleted), cancelling sync: kb=%s ds=%s err=%v",
@@ -670,6 +784,7 @@ func (s *DataSourceService) ProcessSync(ctx context.Context, task *asynq.Task) e
 	// Surface the KB's multimodal/VLM state to the connector so it only extracts
 	// embedded images for OCR when the KB can actually ingest them (never persisted).
 	config.MultimodalEnabled = kb.IsMultimodalEnabled()
+	config.SyncDeletions = ds.SyncDeletions
 
 	// Streaming path: connectors that support it interleave fetch→ingest→
 	// checkpoint so a large sync bounds memory and resumes after a timeout
@@ -860,16 +975,16 @@ func fetchFailureSyncError(item *types.FetchedItem, rawMsg string) types.SyncIte
 func (s *DataSourceService) applyFetchedItem(
 	ctx context.Context, ds *types.DataSource, item *types.FetchedItem,
 	tagIDs []string, result *types.SyncResult,
-) {
+) datasource.ApplyResult {
 	if item.IsDeleted {
 		if !ds.SyncDeletions {
 			// Sync deletion disabled: neither count nor delete.
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyDeferred}
 		}
 		if item.ExternalID == "" {
 			logger.Warnf(ctx, "skipping deletion for item %q: empty external_id", item.Title)
 			result.Skipped++
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 		}
 		// Perform real KB deletion, scoped to items owned by this data source
 		// so identical external IDs from different data sources cannot collide.
@@ -887,13 +1002,13 @@ func (s *DataSourceService) applyFetchedItem(
 				Code:    "deletion_lookup_failed",
 				Message: "Failed to look up the item before deletion; see server logs",
 			})
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 		}
 		if existing == nil {
 			// Deletion is idempotent: the source item may already have been
 			// removed manually or by an earlier sync.
 			result.Skipped++
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyApplied}
 		}
 		if deleteErr := s.knowledgeService.DeleteKnowledge(ctx, existing.ID); deleteErr != nil {
 			// The cursor is already past this item, so a failed deletion normally
@@ -908,7 +1023,7 @@ func (s *DataSourceService) applyFetchedItem(
 				Code:    "deletion_failed",
 				Message: "Deletion failed; see server logs",
 			})
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 		}
 		if herr := repo.HardDeleteKnowledge(ctx, ds.TenantID, existing.ID); herr != nil {
 			result.Failed++
@@ -920,10 +1035,10 @@ func (s *DataSourceService) applyFetchedItem(
 				Code:    "deletion_failed",
 				Message: "Deletion failed; see server logs",
 			})
-			return
+			return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 		}
 		result.Deleted++
-		return
+		return datasource.ApplyResult{Outcome: datasource.ApplyApplied}
 	}
 
 	if len(item.Content) == 0 && item.URL == "" {
@@ -936,7 +1051,7 @@ func (s *DataSourceService) applyFetchedItem(
 			logger.Infof(ctx, "skipping item %q (external_id=%s): no content or URL", item.Title, item.ExternalID)
 			result.Skipped++
 		}
-		return
+		return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 	}
 
 	isUpdate, err := s.ingestItem(ctx, ds, item, tagIDs)
@@ -966,11 +1081,13 @@ func (s *DataSourceService) applyFetchedItem(
 				Message: "Ingest failed; see server logs",
 			})
 		}
+		return datasource.ApplyResult{Outcome: datasource.ApplyFailed}
 	} else if isUpdate {
 		result.Updated++
 	} else {
 		result.Created++
 	}
+	return datasource.ApplyResult{Outcome: datasource.ApplyApplied}
 }
 
 // streamStartCursor decides which cursor a streaming fetch should resume from.
@@ -1004,26 +1121,107 @@ func (h *streamSyncHandler) Emit(ctx context.Context, item types.FetchedItem) er
 	if err := ctx.Err(); err != nil {
 		return err
 	}
+	if err := h.checkAlive(ctx); err != nil {
+		return err
+	}
 	h.result.Total++
 	h.svc.applyFetchedItem(withKBActivitySuppressed(ctx), h.ds, &item, h.tagIDs, h.result)
 	return nil
 }
 
+func (h *streamSyncHandler) EmitWithResult(ctx context.Context, item types.FetchedItem) (datasource.ApplyResult, error) {
+	if err := h.checkAlive(ctx); err != nil {
+		return datasource.ApplyResult{}, err
+	}
+	if item.IsDeleted && !h.ds.SyncDeletions {
+		return datasource.ApplyResult{Outcome: datasource.ApplyDeferred}, nil
+	}
+	h.result.Total++
+	outcome := h.svc.applyFetchedItem(withKBActivitySuppressed(ctx), h.ds, &item, h.tagIDs, h.result)
+	return outcome, ctx.Err()
+}
+
+// WalkSyncedItems recovers source identities even if a previous connector
+// discarded their cursor entries after the user changed its selected scope.
+// Page through metadata only; never read all stored document bodies into memory.
+func (h *streamSyncHandler) WalkSyncedItems(ctx context.Context, visit func(types.FetchedItem) error) error {
+	const pageSize = 200
+	afterID := ""
+	repo := h.svc.knowledgeService.GetRepository()
+	reader, ok := repo.(interfaces.DataSourceKnowledgeMetadataReader)
+	if !ok {
+		return errors.New("data source metadata reader unavailable")
+	}
+	for {
+		if err := h.checkAlive(ctx); err != nil {
+			return err
+		}
+		rows, err := reader.ListDataSourceKnowledgeMetadata(
+			ctx, h.ds.TenantID, h.ds.KnowledgeBaseID, h.ds.ID, afterID, pageSize)
+		if err != nil {
+			return err
+		}
+		for _, row := range rows {
+			if err := ctx.Err(); err != nil {
+				return err
+			}
+			if row.Channel != h.ds.Type {
+				continue
+			}
+			var metadata map[string]string
+			if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
+				return fmt.Errorf("invalid synced document metadata: %w", err)
+			}
+			if metadata["datasource_id"] != h.ds.ID || metadata["external_id"] == "" {
+				continue
+			}
+			if err := visit(types.FetchedItem{
+				ExternalID: metadata["external_id"], SourceResourceID: metadata["collection_id"],
+				Metadata: metadata,
+			}); err != nil {
+				return err
+			}
+		}
+		if len(rows) < pageSize {
+			return nil
+		}
+		nextID := rows[len(rows)-1].ID
+		if nextID <= afterID {
+			return errors.New("synced document metadata pagination did not advance")
+		}
+		afterID = nextID
+	}
+}
+
 // Checkpoint persists the connector cursor onto the data source and mirrors the
 // running counts into the sync log so progress survives a crash and the UI can
 // reflect a long sync mid-flight instead of jumping from 0 to done.
 func (h *streamSyncHandler) Checkpoint(ctx context.Context, cursor *types.SyncCursor) error {
+	if err := h.checkAlive(ctx); err != nil {
+		return err
+	}
 	if cursor == nil {
 		return nil
 	}
+	if metrics, ok := cursor.ConnectorCursor["metrics"]; ok {
+		raw, err := json.Marshal(metrics)
+		if err != nil {
+			return err
+		}
+		if err := json.Unmarshal(raw, &h.result.Metrics); err != nil {
+			return err
+		}
+	}
 	cursorJSON, err := cursor.ToJSON()
 	if err != nil {
 		return err
 	}
-	h.ds.LastSyncCursor = cursorJSON
-	if err := h.svc.dsRepo.UpdateSyncState(ctx, h.ds); err != nil {
+	snapshot := *h.ds
+	snapshot.LastSyncCursor = cursorJSON
+	if err := h.svc.dsRepo.UpdateSyncState(ctx, &snapshot); err != nil {
 		return err
 	}
+	h.ds.LastSyncCursor = cursorJSON
 
 	// Best-effort live progress; a failure here must not abort the sync.
 	h.syncLog.ItemsTotal = h.result.Total
@@ -1038,6 +1236,44 @@ func (h *streamSyncHandler) Checkpoint(ctx context.Context, cursor *types.SyncCu
 	return nil
 }
 
+func (h *streamSyncHandler) checkAlive(ctx context.Context) error {
+	if h.ds.Type != types.ConnectorTypeOutline {
+		return ctx.Err()
+	}
+	if err := datasource.CheckSyncLease(ctx); err != nil {
+		return err
+	}
+	if _, err := h.svc.dsRepo.FindByID(ctx, h.ds.ID); err != nil {
+		return err
+	}
+	return nil
+}
+
+func (s *DataSourceService) lockOutlineMutation(ctx context.Context, ds *types.DataSource) (context.Context, func(), error) {
+	locked, release, err := datasource.DefaultSyncLocks.Acquire(ctx, ds.TenantID, ds.ID)
+	if err != nil {
+		return ctx, nil, err
+	}
+	running, err := s.syncLogRepo.HasRunningSync(locked, ds.ID)
+	if err != nil || running {
+		release()
+		if err != nil {
+			return ctx, nil, err
+		}
+		return ctx, nil, datasource.ErrSyncRunning
+	}
+	fresh, err := s.dsRepo.FindByID(locked, ds.ID)
+	if err != nil || fresh == nil {
+		release()
+		if err != nil {
+			return ctx, nil, err
+		}
+		return ctx, nil, datasource.ErrDataSourceNotFound
+	}
+	*ds = *fresh
+	return locked, release, nil
+}
+
 // processSyncStreaming runs a sync through a StreamingConnector, ingesting each
 // item as it arrives and checkpointing progress so the run is memory-bounded and
 // resumable after a timeout.
@@ -1062,7 +1298,24 @@ func (s *DataSourceService) processSyncStreaming(
 
 	forceFull := payload.ForceFull || ds.SyncMode == types.SyncModeFull
 	attempt, _ := asynq.GetRetryCount(ctx)
-	startCursor, err := streamStartCursor(ds, forceFull, attempt)
+	if retry, _, ok := types.TaskRetryMetadataFromContext(ctx); ok {
+		attempt = retry
+	}
+	var startCursor *types.SyncCursor
+	if preparer, ok := sc.(datasource.SyncRunCursorPreparer); ok {
+		startCursor, err = ds.ParseSyncCursor()
+		if err == nil {
+			startCursor, err = preparer.PrepareSyncRunCursor(startCursor, syncLog.ID, forceFull)
+		}
+	} else if preparer, ok := sc.(datasource.FullSyncCursorPreparer); ok && forceFull && attempt == 0 {
+		startCursor, err = ds.ParseSyncCursor()
+		if err == nil {
+			startCursor, err = preparer.PrepareFullSyncCursor(startCursor)
+		}
+	} else {
+		startCursor, err = streamStartCursor(ds, forceFull, attempt)
+	}
+
 	if err != nil {
 		logger.Errorf(ctx, "failed to parse sync cursor: %v", err)
 		s.updateSyncRunResult(ctx, ds, syncLog, &types.SyncResult{}, nil,
@@ -1074,6 +1327,12 @@ func (s *DataSourceService) processSyncStreaming(
 	handler := &streamSyncHandler{svc: s, ds: ds, tagIDs: autoTagIDs, result: result, syncLog: syncLog}
 
 	nextCursor, fetchErr := sc.FetchStream(ctx, config, startCursor, handler)
+	var partial *datasource.PartialFetchError
+	var warnings []string
+	if errors.As(fetchErr, &partial) {
+		warnings = partial.Details
+		fetchErr = nil
+	}
 	if fetchErr != nil {
 		// Progress so far is already checkpointed onto ds.LastSyncCursor; leave
 		// it in place so the Asynq retry resumes from there. Persist counts.
@@ -1108,10 +1367,18 @@ func (s *DataSourceService) processSyncStreaming(
 	// applyFetchedItem).
 	status := types.SyncLogStatusSuccess
 	errMsg := ""
+	if len(warnings) > 0 {
+		status = types.SyncLogStatusPartial
+		errMsg = strings.Join(warnings, "; ")
+		for _, warning := range warnings {
+			recordSyncError(result, types.SyncItemError{Message: warning})
+		}
+		resultJSON, _ = result.ToJSON()
+	}
 	if result.Failed > 0 {
 		status = types.SyncLogStatusPartial
 		errMsg = fmt.Sprintf("%d document(s) failed to sync", result.Failed)
-		if result.DeletionFailed > 0 {
+		if result.DeletionFailed > 0 && ds.Type != types.ConnectorTypeOutline {
 			errMsg += fmt.Sprintf("; %d deletion failure(s) will only retry on the next full sync", result.DeletionFailed)
 		}
 	}
@@ -1131,6 +1398,14 @@ func (s *DataSourceService) updateSyncRunResult(
 	errorMessage string,
 	wasPaused bool,
 ) {
+	if ds.Type == types.ConnectorTypeOutline {
+		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
+		defer cancel()
+		if err := datasource.CheckSyncLease(cleanupCtx); err != nil {
+			return
+		}
+		ctx = cleanupCtx
+	}
 	syncLog.ItemsTotal = result.Total
 	syncLog.ItemsCreated = result.Created
 	syncLog.ItemsUpdated = result.Updated
@@ -1228,8 +1503,56 @@ func (s *DataSourceService) validateDataSourceConfig(ctx context.Context, ds *ty
 	if err != nil {
 		return datasource.ErrInvalidConfig
 	}
+	config.SyncDeletions = ds.SyncDeletions
+	if err := s.validateOutlineOptions(ds, config); err != nil {
+		return err
+	}
+	if err := connector.Validate(ctx, config); err != nil {
+		return err
+	}
+	if binder, ok := connector.(interface {
+		BindIdentity(context.Context, *types.DataSourceConfig, *types.SyncCursor) (*types.SyncCursor, error)
+	}); ok {
+		cursor, err := ds.ParseSyncCursor()
+		if err != nil {
+			return err
+		}
+		cursor, err = binder.BindIdentity(ctx, config, cursor)
+		if err != nil {
+			return err
+		}
+		ds.LastSyncCursor, err = cursor.ToJSON()
+		if err == nil && ds.Type == types.ConnectorTypeOutline {
+			ds.Config, err = config.ToJSON()
+		}
+		return err
+	}
+	return nil
+}
 
-	return connector.Validate(ctx, config)
+func (s *DataSourceService) validateOutlineOptions(ds *types.DataSource, config *types.DataSourceConfig) error {
+	if ds.Type != types.ConnectorTypeOutline {
+		return nil
+	}
+	if ds.ConflictStrategy != "" && ds.ConflictStrategy != "overwrite" {
+		return fmt.Errorf("%w: Outline requires overwrite", datasource.ErrInvalidConfig)
+	}
+	if ds.SyncMode != "" && ds.SyncMode != "incremental" && ds.SyncMode != "full" {
+		return datasource.ErrInvalidConfig
+	}
+	if ds.SyncSchedule != "" {
+		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
+		if _, err := parser.Parse(ds.SyncSchedule); err != nil {
+			return fmt.Errorf("%w: invalid six-field schedule", datasource.ErrInvalidConfig)
+		}
+	}
+	if config == nil {
+		return datasource.ErrInvalidConfig
+	}
+	if len(config.ResourceIDs) == 0 && (ds.Status != types.DataSourceStatusPaused || ds.SyncSchedule != "") {
+		return fmt.Errorf("%w: select at least one Outline collection", datasource.ErrInvalidConfig)
+	}
+	return nil
 }
 
 // ingestItem writes a single FetchedItem into the knowledge base.
@@ -1241,6 +1564,7 @@ func (s *DataSourceService) validateDataSourceConfig(ctx context.Context, ds *ty
 //
 // Returns (isUpdate, error) — isUpdate is true when an existing item was replaced.
 func (s *DataSourceService) ingestItem(ctx context.Context, ds *types.DataSource, item *types.FetchedItem, tagIDs []string) (bool, error) {
+	strict := ds.Type == types.ConnectorTypeOutline
 	// Channel decides the knowledge "source" label shown in the UI. Prefer the
 	// connector-supplied metadata["channel"] (e.g. Feishu Drive sets it to
 	// "feishu" so Drive docs share the wiki's "飞书" label instead of showing
@@ -1280,14 +1604,34 @@ func (s *DataSourceService) ingestItem(ctx context.Context, ds *types.DataSource
 		// other during updates.
 		existing, err := repo.FindByDataSourceExternalID(ctx, ds.TenantID, ds.KnowledgeBaseID, ds.ID, item.ExternalID)
 		if err != nil {
+			if strict {
+				return false, err
+			}
 			logger.Warnf(ctx, "failed to check existing knowledge for external_id=%s: %v", item.ExternalID, err)
 			// Non-fatal: proceed with creation (may produce duplicate)
 		} else if existing != nil {
+			if strict && item.Metadata["source_fingerprint"] != "" {
+				var prior map[string]string
+				_ = json.Unmarshal(existing.Metadata, &prior)
+				switch existing.ParseStatus {
+				case "pending", "processing", "finalizing", "completed":
+					if prior["source_fingerprint"] == item.Metadata["source_fingerprint"] &&
+						(item.Metadata["source_full_sync"] != "true" || prior["source_sync_run_id"] == item.Metadata["source_sync_run_id"]) {
+						return true, nil
+					}
+				}
+			}
 			logger.Infof(ctx, "found existing knowledge %s for external_id=%s, deleting for update", existing.ID, item.ExternalID)
 			if err := s.knowledgeService.DeleteKnowledge(ctx, existing.ID); err != nil {
+				if strict {
+					return false, err
+				}
 				logger.Warnf(ctx, "failed to delete existing knowledge %s: %v", existing.ID, err)
 			} else {
 				if herr := repo.HardDeleteKnowledge(ctx, ds.TenantID, existing.ID); herr != nil {
+					if strict {
+						return false, herr
+					}
 					logger.Warnf(ctx, "failed to hard-delete replaced knowledge %s: %v", existing.ID, herr)
 				}
 				isUpdate = true
@@ -1301,7 +1645,7 @@ func (s *DataSourceService) ingestItem(ctx context.Context, ds *types.DataSource
 		if err != nil {
 			return isUpdate, fmt.Errorf("build file header: %w", err)
 		}
-		if _, err := s.knowledgeService.CreateKnowledgeFromFile(
+		created, err := s.knowledgeService.CreateKnowledgeFromFile(
 			ctx,
 			ds.KnowledgeBaseID,
 			fh,
@@ -1311,7 +1655,8 @@ func (s *DataSourceService) ingestItem(ctx context.Context, ds *types.DataSource
 			tagIDs,        // auto-tag from data source
 			channel,
 			nil,
-		); err != nil {
+		)
+		if err != nil {
 			var dupErr *types.DuplicateKnowledgeError
 			if errors.As(err, &dupErr) && dupIsSameNode(dupErr, item) {
 				// Identical content is already present in the KB under THIS node's
@@ -1321,6 +1666,9 @@ func (s *DataSourceService) ingestItem(ctx context.Context, ds *types.DataSource
 			}
 			return isUpdate, err
 		}
+		if strict && (created == nil || created.ParseStatus == "failed") {
+			return isUpdate, fmt.Errorf("outline document was not accepted for parsing")
+		}
 		s.sweepStaleSubtree(ctx, ds, item)
 		return isUpdate, nil
 	}
````

### A.21 `internal/application/service/knowledge_create.go`

````diff
diff --git a/internal/application/service/knowledge_create.go b/internal/application/service/knowledge_create.go
index d294f2c1..8701dca3 100644
--- a/internal/application/service/knowledge_create.go
+++ b/internal/application/service/knowledge_create.go
@@ -87,12 +87,19 @@ func (s *knowledgeService) CreateKnowledgeFromFile(ctx context.Context,
 	// Check if file already exists
 	tenantID := ctx.Value(types.TenantIDContextKey).(uint64)
 	logger.Infof(ctx, "Checking if file exists, tenant ID: %d", tenantID)
+	var dataSourceID, externalID string
+	if channel == types.ChannelOutline {
+		dataSourceID = metadata["datasource_id"]
+		externalID = metadata["external_id"]
+	}
 	exists, existingKnowledge, err := s.repo.CheckKnowledgeExists(ctx, tenantID, kbID, &types.KnowledgeCheckParams{
-		Type:     "file",
-		FileName: fileName,
-		FileType: getFileType(fileName),
-		FileSize: file.Size,
-		FileHash: hash,
+		DataSourceID: dataSourceID,
+		ExternalID:   externalID,
+		Type:         "file",
+		FileName:     fileName,
+		FileType:     getFileType(fileName),
+		FileSize:     file.Size,
+		FileHash:     hash,
 	})
 	if err != nil {
 		logger.Errorf(ctx, "Failed to check knowledge existence: %v", err)
````

### A.22 `internal/container/container.go`

````diff
diff --git a/internal/container/container.go b/internal/container/container.go
index da74bcd3..ae73da01 100644
--- a/internal/container/container.go
+++ b/internal/container/container.go
@@ -62,6 +62,7 @@ import (
 	gitlabConnector "github.com/Tencent/WeKnora/internal/datasource/connector/gitlab"
 	imaConnector "github.com/Tencent/WeKnora/internal/datasource/connector/ima"
 	notionConnector "github.com/Tencent/WeKnora/internal/datasource/connector/notion"
+	outlineConnector "github.com/Tencent/WeKnora/internal/datasource/connector/outline"
 	rssConnector "github.com/Tencent/WeKnora/internal/datasource/connector/rss"
 	yuqueConnector "github.com/Tencent/WeKnora/internal/datasource/connector/yuque"
 	"github.com/Tencent/WeKnora/internal/event"
@@ -366,6 +367,7 @@ func BuildContainer(container *dig.Container) *dig.Container {
 
 	// Data source sync framework
 	logger.Debugf(ctx, "[Container] Registering data source sync framework...")
+	must(container.Invoke(func(rdb *redis.Client) { datasource.DefaultSyncLocks = datasource.NewSyncLocks(rdb) }))
 	must(container.Provide(initConnectorRegistry))
 	must(container.Provide(datasource.NewScheduler))
 	must(container.Provide(service.NewDataSourceService))
@@ -1698,6 +1700,9 @@ func initConnectorRegistry() (*datasource.ConnectorRegistry, error) {
 	if err := registry.Register(notionConnector.NewConnector()); err != nil {
 		errs = errors.Join(errs, fmt.Errorf("register notion connector: %w", err))
 	}
+	if err := registry.Register(outlineConnector.NewConnector()); err != nil {
+		errs = errors.Join(errs, fmt.Errorf("register outline connector: %w", err))
+	}
 	if err := registry.Register(yuqueConnector.NewConnector()); err != nil {
 		errs = errors.Join(errs, fmt.Errorf("register yuque connector: %w", err))
 	}
````

### A.23 `internal/datasource/connector.go`

````diff
diff --git a/internal/datasource/connector.go b/internal/datasource/connector.go
index c7c51baa..9f2adf66 100644
--- a/internal/datasource/connector.go
+++ b/internal/datasource/connector.go
@@ -74,6 +74,42 @@ type StreamHandler interface {
 	Checkpoint(ctx context.Context, cursor *types.SyncCursor) error
 }
 
+type ApplyOutcome string
+
+const (
+	ApplyApplied  ApplyOutcome = "applied"
+	ApplyDeferred ApplyOutcome = "deferred"
+	ApplyFailed   ApplyOutcome = "failed"
+)
+
+type ApplyResult struct {
+	Outcome     ApplyOutcome
+	KnowledgeID string
+}
+
+// AcknowledgingStreamHandler confirms durable acceptance, not eventual indexing.
+type AcknowledgingStreamHandler interface {
+	StreamHandler
+	EmitWithResult(context.Context, types.FetchedItem) (ApplyResult, error)
+}
+
+// SyncedItemReader exposes only persisted source identities and metadata. A
+// connector can recover reconciliation candidates lost by older cursor formats
+// without loading document bodies or assuming they are safe to delete.
+type SyncedItemReader interface {
+	WalkSyncedItems(context.Context, func(types.FetchedItem) error) error
+}
+
+type FullSyncCursorPreparer interface {
+	PrepareFullSyncCursor(*types.SyncCursor) (*types.SyncCursor, error)
+}
+
+// SyncRunCursorPreparer distinguishes task retries from new full-sync requests,
+// including retries that occurred before the worker acquired its execution lease.
+type SyncRunCursorPreparer interface {
+	PrepareSyncRunCursor(previous *types.SyncCursor, runID string, forceFull bool) (*types.SyncCursor, error)
+}
+
 // StreamingConnector is an optional interface. Connectors that implement it let
 // the service interleave fetch→ingest→checkpoint so a large sync persists
 // incrementally and resumes after a timeout, rather than holding every item in
@@ -149,6 +185,12 @@ type ConnectorMetadata struct {
 // GetConnectorMetadata returns metadata for all available connectors
 // This is used by the frontend to display connector options
 var ConnectorMetadataRegistry = map[string]ConnectorMetadata{
+	types.ConnectorTypeOutline: {
+		Type: types.ConnectorTypeOutline, Name: "Outline",
+		Description: "Sync Outline collections and Markdown documents",
+		Priority:    3, AuthType: "api_key",
+		Capabilities: []string{"incremental", "deletion_sync"},
+	},
 	types.ConnectorTypeFeishu: {
 		Type:         types.ConnectorTypeFeishu,
 		Name:         "Feishu (飞书)",
````

### A.24 `internal/datasource/connector/outline/client.go`

````diff
diff --git a/internal/datasource/connector/outline/client.go b/internal/datasource/connector/outline/client.go
new file mode 100644
index 00000000..41dbf3d9
--- /dev/null
+++ b/internal/datasource/connector/outline/client.go
@@ -0,0 +1,279 @@
+package outline
+
+import (
+	"bytes"
+	"context"
+	"crypto/sha256"
+	"encoding/json"
+	"errors"
+	"fmt"
+	"io"
+	"math/rand/v2"
+	"net"
+	"net/http"
+	"net/url"
+	"strconv"
+	"strings"
+	"sync"
+	"time"
+
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/Tencent/WeKnora/internal/utils"
+	"golang.org/x/time/rate"
+)
+
+const maxResponse = 32 << 20
+
+var errTooLarge = errors.New("outline_document_too_large")
+var limiters sync.Map
+
+type apiError struct {
+	Code   string
+	Status int
+}
+
+func (e *apiError) Error() string { return e.Code }
+
+func fatal(err error) bool {
+	var api *apiError
+	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
+		(errors.As(err, &api) && api.Status == http.StatusUnauthorized)
+}
+
+type client struct {
+	base    string
+	key     string
+	http    *http.Client
+	limiter *rate.Limiter
+	metrics map[string]int64
+}
+
+func newClient(config *types.DataSourceConfig) (*client, error) {
+	if config == nil {
+		return nil, datasource.ErrInvalidConfig
+	}
+	base, _ := config.Credentials["base_url"].(string)
+	key, _ := config.Credentials["api_key"].(string)
+	base, key = strings.TrimSpace(base), strings.TrimSpace(key)
+	if base == "" || key == "" {
+		return nil, fmt.Errorf("%w: base_url and api_key are required", datasource.ErrInvalidConfig)
+	}
+	if !strings.Contains(base, "://") {
+		base = "https://" + base
+	}
+	base = strings.TrimRight(base, "/")
+	u, err := url.Parse(base)
+	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(base, "#") ||
+		(u.Path != "" && u.Path != "/") || (u.Scheme != "https" && u.Scheme != "http") {
+		return nil, fmt.Errorf("%w: base_url must be an instance root URL", datasource.ErrInvalidConfig)
+	}
+	u.Host = strings.ToLower(u.Host)
+	u.Path = ""
+	base = u.String()
+	if err := datasource.ValidateConnectorBaseURL(base); err != nil {
+		return nil, err
+	}
+	if u.Scheme == "http" {
+		dnsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
+		defer cancel()
+		ips, err := net.DefaultResolver.LookupIPAddr(dnsCtx, u.Hostname())
+		if err != nil || len(ips) == 0 || !utils.IsSSRFWhitelisted(u.Hostname()) {
+			return nil, fmt.Errorf("%w: HTTP requires a whitelisted private host", datasource.ErrInvalidConfig)
+		}
+		for _, ip := range ips {
+			if !ip.IP.IsPrivate() && !ip.IP.IsLoopback() {
+				return nil, fmt.Errorf("%w: public HTTP is not allowed", datasource.ErrInvalidConfig)
+			}
+		}
+	}
+	h := datasource.NewConnectorHTTPClient(30 * time.Second)
+	// RPC endpoints do not need redirects. Rejecting all also avoids POST-to-GET conversion.
+	h.CheckRedirect = func(*http.Request, []*http.Request) error { return errors.New("outline_redirect_rejected") }
+	digest := sha256.Sum256([]byte(base + "\x00" + key))
+	limiter, _ := limiters.LoadOrStore(digest, rate.NewLimiter(rate.Every(time.Second), 1))
+	return &client{base: base, key: key, http: h, limiter: limiter.(*rate.Limiter), metrics: map[string]int64{}}, nil
+}
+
+type envelope struct {
+	Data       json.RawMessage `json:"data"`
+	OK         *bool           `json:"ok"`
+	Pagination struct {
+		Total    *int    `json:"total"`
+		Offset   *int    `json:"offset"`
+		Limit    *int    `json:"limit"`
+		NextPath *string `json:"nextPath"`
+	} `json:"pagination"`
+}
+
+func wait(ctx context.Context, delay time.Duration) error {
+	t := time.NewTimer(delay)
+	defer t.Stop()
+	select {
+	case <-ctx.Done():
+		return ctx.Err()
+	case <-t.C:
+		return nil
+	}
+}
+
+func retryDelay(value string, attempt int) time.Duration {
+	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
+		return time.Duration(min(seconds, 7200)) * time.Second
+	}
+	if until, err := http.ParseTime(value); err == nil {
+		return max(time.Duration(0), time.Until(until))
+	}
+	return time.Duration(2<<attempt) * time.Second
+}
+
+func (c *client) call(ctx context.Context, method string, body map[string]interface{}) (*envelope, error) {
+	b, err := json.Marshal(body)
+	if err != nil {
+		return nil, err
+	}
+	for attempt := 0; attempt < 4; attempt++ {
+		start := time.Now()
+		if err := c.limiter.Wait(ctx); err != nil {
+			return nil, err
+		}
+		c.metrics["rate_limit_wait_ms"] += time.Since(start).Milliseconds()
+		c.metrics["api_requests"]++
+		if attempt > 0 {
+			c.metrics["retry_count"]++
+		}
+		if err := datasource.ValidateConnectorBaseURL(c.base); err != nil {
+			return nil, err
+		}
+		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/"+method, bytes.NewReader(b))
+		if err != nil {
+			return nil, err
+		}
+		req.Header.Set("Authorization", "Bearer "+c.key)
+		req.Header.Set("Content-Type", "application/json")
+		req.Header.Set("Accept", "application/json")
+		req.Header.Set("X-API-Version", "2")
+		resp, err := c.http.Do(req)
+		if err != nil {
+			if ctx.Err() != nil {
+				return nil, ctx.Err()
+			}
+			if attempt == 3 {
+				return nil, errors.New("outline_request_failed")
+			}
+			if err := wait(ctx, retryDelay("", attempt)+time.Duration(rand.IntN(250))*time.Millisecond); err != nil {
+				return nil, err
+			}
+			continue
+		}
+		data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
+		resp.Body.Close()
+		if len(data) > maxResponse {
+			return nil, errTooLarge
+		}
+		if resp.StatusCode == 429 || resp.StatusCode >= 500 || readErr != nil {
+			if attempt < 3 {
+				delay := retryDelay(resp.Header.Get("Retry-After"), attempt)
+				if resp.StatusCode == 429 {
+					c.metrics["rate_limit_wait_ms"] += delay.Milliseconds()
+				}
+				if err := wait(ctx, delay); err != nil {
+					return nil, err
+				}
+				continue
+			}
+		}
+		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
+			code := "outline_request_failed"
+			switch resp.StatusCode {
+			case 401:
+				code = "outline_auth_failed"
+			case 403:
+				code = "outline_permission_denied"
+			case 429:
+				code = "outline_rate_limited"
+			}
+			return nil, &apiError{Code: code, Status: resp.StatusCode}
+		}
+		if readErr != nil {
+			return nil, errors.New("outline_response_invalid")
+		}
+		var result envelope
+		if json.Unmarshal(data, &result) != nil || len(result.Data) == 0 ||
+			string(result.Data) == "null" || (result.OK != nil && !*result.OK) {
+			return nil, errors.New("outline_response_invalid")
+		}
+		return &result, nil
+	}
+	return nil, errors.New("outline_request_failed")
+}
+
+// walk uses only numeric pagination from nextPath; credentials never follow its URL.
+func (c *client) walk(ctx context.Context, method string, body map[string]interface{}, page func([]json.RawMessage) error) error {
+	offset, limit := 0, 25
+	seen := map[string]bool{}
+	var expectedTotal *int
+	var hasTotal *bool
+	for {
+		if err := ctx.Err(); err != nil {
+			return err
+		}
+		body["offset"], body["limit"] = offset, limit
+		result, err := c.call(ctx, method, body)
+		if errors.Is(err, errTooLarge) && limit > 1 {
+			limit = max(1, limit/2)
+			continue
+		}
+		if err != nil {
+			return err
+		}
+		currentHasTotal := result.Pagination.Total != nil
+		if hasTotal == nil {
+			hasTotal = &currentHasTotal
+		} else if *hasTotal != currentHasTotal {
+			return errors.New("outline_scan_incomplete")
+		}
+		if currentHasTotal {
+			if expectedTotal == nil {
+				total := *result.Pagination.Total
+				expectedTotal = &total
+			} else if *expectedTotal != *result.Pagination.Total {
+				return errors.New("outline_scan_incomplete")
+			}
+		}
+		var rows []json.RawMessage
+		if json.Unmarshal(result.Data, &rows) != nil {
+			return errors.New("outline_response_invalid")
+		}
+		next, pageLimit, terminal, err := c.nextPage(method, result, offset, limit, len(rows))
+		if err != nil {
+			return err
+		}
+		fresh := make([]json.RawMessage, 0, len(rows))
+		for _, raw := range rows {
+			var id struct {
+				ID string `json:"id"`
+			}
+			if json.Unmarshal(raw, &id) != nil || id.ID == "" {
+				return errors.New("outline_response_invalid")
+			}
+			if !seen[id.ID] {
+				seen[id.ID] = true
+				fresh = append(fresh, raw)
+			}
+		}
+		if len(fresh) != len(rows) {
+			return errors.New("outline_scan_incomplete")
+		}
+		if err := page(fresh); err != nil {
+			return err
+		}
+		if terminal {
+			if expectedTotal != nil && len(seen) != *expectedTotal {
+				return errors.New("outline_scan_incomplete")
+			}
+			return nil
+		}
+		offset, limit = next, pageLimit
+	}
+}
````

### A.25 `internal/datasource/connector/outline/connector.go`

````diff
diff --git a/internal/datasource/connector/outline/connector.go b/internal/datasource/connector/outline/connector.go
new file mode 100644
index 00000000..4dde0fba
--- /dev/null
+++ b/internal/datasource/connector/outline/connector.go
@@ -0,0 +1,218 @@
+package outline
+
+import (
+	"context"
+	"encoding/json"
+	"errors"
+	"slices"
+
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/google/uuid"
+)
+
+type Connector struct{}
+
+func NewConnector() *Connector  { return &Connector{} }
+func (*Connector) Type() string { return types.ConnectorTypeOutline }
+
+func authenticate(ctx context.Context, c *client) (identity, error) {
+	result, err := c.call(ctx, "auth.info", map[string]interface{}{})
+	if err != nil {
+		return identity{}, err
+	}
+	var auth struct {
+		User struct {
+			ID string `json:"id"`
+		} `json:"user"`
+		Team struct {
+			ID string `json:"id"`
+		} `json:"team"`
+	}
+	if json.Unmarshal(result.Data, &auth) != nil || auth.User.ID == "" || auth.Team.ID == "" {
+		return identity{}, errors.New("outline_response_invalid")
+	}
+	return identity{BaseURL: c.base, WorkspaceID: auth.Team.ID, ActorID: auth.User.ID}, nil
+}
+
+func selectedIDs(ids []string) ([]string, error) {
+	result := slices.Clone(ids)
+	for i, id := range result {
+		parsed, err := uuid.Parse(id)
+		if err != nil {
+			return nil, errors.New("outline_invalid_collection_id")
+		}
+		result[i] = parsed.String()
+	}
+	slices.Sort(result)
+	return slices.Compact(result), nil
+}
+
+func collectionInfo(ctx context.Context, c *client, id string) (collection, error) {
+	result, err := c.call(ctx, "collections.info", map[string]interface{}{"id": id})
+	if err != nil {
+		return collection{}, err
+	}
+	var item collection
+	if json.Unmarshal(result.Data, &item) != nil || item.ID != id {
+		return item, errors.New("outline_response_invalid")
+	}
+	return item, nil
+}
+
+func documentInfo(ctx context.Context, c *client, id string) (document, error) {
+	result, err := c.call(ctx, "documents.info", map[string]interface{}{"id": id})
+	if err != nil {
+		return document{}, err
+	}
+	item, err := decodeDocument(result.Data)
+	if err == nil && item.ID != id {
+		err = errors.New("outline_response_invalid")
+	}
+	return item, err
+}
+
+func (c *Connector) Validate(ctx context.Context, config *types.DataSourceConfig) error {
+	client, err := newClient(config)
+	if err != nil {
+		return err
+	}
+	if _, err := authenticate(ctx, client); err != nil {
+		return err
+	}
+	if _, err := c.ListResources(ctx, config, ""); err != nil {
+		return err
+	}
+	ids, err := selectedIDs(config.ResourceIDs)
+	if err != nil {
+		return err
+	}
+	config.ResourceIDs = ids
+	for _, id := range ids {
+		col, err := collectionInfo(ctx, client, id)
+		if err != nil {
+			return err
+		}
+		if col.ArchivedAt != "" || col.DeletedAt != "" {
+			return errors.New("outline_collection_inactive")
+		}
+		result, err := client.call(ctx, "documents.list", map[string]interface{}{
+			"collectionId": id, "limit": 1, "offset": 0, "statusFilter": []string{"published"},
+			"sort": "createdAt", "direction": "ASC",
+		})
+		if err != nil {
+			return err
+		}
+		var rows []json.RawMessage
+		if json.Unmarshal(result.Data, &rows) != nil {
+			return errors.New("outline_response_invalid")
+		}
+		if len(rows) > 0 {
+			d, err := decodeDocument(rows[0])
+			if err != nil {
+				return err
+			}
+			d, err = documentInfo(ctx, client, d.ID)
+			if err != nil {
+				return err
+			}
+			if d.Text == nil {
+				return errors.New("outline_format_unsupported")
+			}
+		}
+	}
+	if config.SyncDeletions {
+		var result *envelope
+		result, err = client.call(ctx, "documents.deleted", map[string]interface{}{"limit": 1, "offset": 0})
+		if err == nil {
+			var rows []json.RawMessage
+			if json.Unmarshal(result.Data, &rows) != nil {
+				return errors.New("outline_response_invalid")
+			}
+		}
+	}
+	return err
+}
+
+func (*Connector) ListResources(ctx context.Context, config *types.DataSourceConfig, parentID string) ([]types.Resource, error) {
+	out := []types.Resource{}
+	if parentID != "" {
+		return out, nil
+	}
+	c, err := newClient(config)
+	if err != nil {
+		return nil, err
+	}
+	err = c.walk(ctx, "collections.list", map[string]interface{}{}, func(rows []json.RawMessage) error {
+		for _, raw := range rows {
+			var col collection
+			if json.Unmarshal(raw, &col) != nil || col.ID == "" {
+				return errors.New("outline_response_invalid")
+			}
+			if col.ArchivedAt != "" || col.DeletedAt != "" {
+				continue
+			}
+			out = append(out, types.Resource{ExternalID: col.ID, Name: col.Name, Type: "collection"})
+		}
+		return nil
+	})
+	return out, err
+}
+
+func (*Connector) ResolveResourceAncestors(context.Context, *types.DataSourceConfig, []string) ([]string, error) {
+	return []string{}, nil
+}
+
+// Batch adapters acknowledge delivery only. Production always uses FetchStream.
+type collector struct{ items []types.FetchedItem }
+
+func (h *collector) Emit(_ context.Context, item types.FetchedItem) error {
+	h.items = append(h.items, item)
+	return nil
+}
+func (h *collector) Checkpoint(context.Context, *types.SyncCursor) error { return nil }
+func (h *collector) EmitWithResult(ctx context.Context, item types.FetchedItem) (datasource.ApplyResult, error) {
+	return datasource.ApplyResult{Outcome: datasource.ApplyApplied}, h.Emit(ctx, item)
+}
+func (c *Connector) FetchAll(ctx context.Context, config *types.DataSourceConfig, ids []string) ([]types.FetchedItem, error) {
+	if config == nil {
+		return nil, datasource.ErrInvalidConfig
+	}
+	copy := *config
+	copy.ResourceIDs = ids
+	h := &collector{}
+	cursor, _ := c.PrepareFullSyncCursor(nil)
+	_, err := c.FetchStream(ctx, &copy, cursor, h)
+	return h.items, err
+}
+func (c *Connector) FetchIncremental(ctx context.Context, config *types.DataSourceConfig, cursor *types.SyncCursor) ([]types.FetchedItem, *types.SyncCursor, error) {
+	h := &collector{}
+	next, err := c.FetchStream(ctx, config, cursor, h)
+	return h.items, next, err
+}
+
+var _ datasource.StreamingConnector = (*Connector)(nil)
+
+// BindIdentity verifies credential rotations before a configuration is persisted.
+func (*Connector) BindIdentity(ctx context.Context, config *types.DataSourceConfig, previous *types.SyncCursor) (*types.SyncCursor, error) {
+	s, err := parseCursor(previous)
+	if err != nil {
+		return nil, err
+	}
+	c, err := newClient(config)
+	if err != nil {
+		return nil, err
+	}
+	instance, err := authenticate(ctx, c)
+	if err != nil {
+		return nil, err
+	}
+	if s.Instance.BaseURL != "" && (s.Instance.BaseURL != instance.BaseURL || s.Instance.WorkspaceID != instance.WorkspaceID) {
+		return nil, errors.New("outline_instance_changed")
+	}
+	if s.Instance.ActorID != instance.ActorID {
+		s.Run = nil
+	}
+	s.Instance = instance
+	return s.cursor(), nil
+}
````

### A.26 `internal/datasource/connector/outline/cursor.go`

````diff
diff --git a/internal/datasource/connector/outline/cursor.go b/internal/datasource/connector/outline/cursor.go
new file mode 100644
index 00000000..1a78805f
--- /dev/null
+++ b/internal/datasource/connector/outline/cursor.go
@@ -0,0 +1,99 @@
+package outline
+
+import (
+	"encoding/json"
+	"errors"
+	"time"
+
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/google/uuid"
+)
+
+type version struct {
+	CollectionID string `json:"collection_id"`
+	Fingerprint  string `json:"fingerprint"`
+	UpdatedAt    string `json:"updated_at,omitempty"`
+}
+
+type run struct {
+	ID        string            `json:"id"`
+	ForceFull bool              `json:"force_full"`
+	Completed map[string]bool   `json:"completed"`
+	Seen      map[string]bool   `json:"seen"`
+	Applied   map[string]string `json:"applied"`
+}
+
+type state struct {
+	TaskRunID        string             `json:"task_run_id,omitempty"`
+	Metrics          map[string]int64   `json:"metrics,omitempty"`
+	Version          int                `json:"version"`
+	Instance         identity           `json:"instance"`
+	Selected         []string           `json:"selected_collection_ids"`
+	Documents        map[string]version `json:"documents"`
+	PendingUpserts   map[string]string  `json:"pending_upserts"`
+	PendingDeletions map[string]string  `json:"pending_deletions"`
+	Run              *run               `json:"run,omitempty"`
+	LastComplete     time.Time          `json:"last_complete_scan"`
+}
+
+func newRun(full bool) *run {
+	return &run{ID: uuid.NewString(), ForceFull: full, Completed: map[string]bool{}, Seen: map[string]bool{}, Applied: map[string]string{}}
+}
+
+func parseCursor(cursor *types.SyncCursor) (*state, error) {
+	s := &state{Version: 1, Documents: map[string]version{}, PendingUpserts: map[string]string{}, PendingDeletions: map[string]string{}}
+	if cursor == nil {
+		return s, nil
+	}
+	raw, err := json.Marshal(cursor.ConnectorCursor)
+	if err != nil || json.Unmarshal(raw, s) != nil || s.Version != 1 ||
+		s.Documents == nil || s.PendingUpserts == nil || s.PendingDeletions == nil {
+		return nil, errors.New("outline_cursor_invalid")
+	}
+	// Reject a foreign cursor instead of silently interpreting it as an empty baseline.
+	for _, field := range []string{"version", "documents", "pending_upserts", "pending_deletions", "instance"} {
+		if _, ok := cursor.ConnectorCursor[field]; !ok {
+			return nil, errors.New("outline_cursor_invalid")
+		}
+	}
+	if s.Run != nil && (s.Run.ID == "" || s.Run.Completed == nil || s.Run.Seen == nil || s.Run.Applied == nil) {
+		return nil, errors.New("outline_cursor_invalid")
+	}
+	if value := cursor.ConnectorCursor["instance"]; value == nil {
+		return nil, errors.New("outline_cursor_invalid")
+	}
+	for id, document := range s.Documents {
+		if id == "" || document.CollectionID == "" || document.Fingerprint == "" {
+			return nil, errors.New("outline_cursor_invalid")
+		}
+	}
+	return s, nil
+}
+
+func (s *state) cursor() *types.SyncCursor {
+	raw, _ := json.Marshal(s)
+	var values map[string]interface{}
+	_ = json.Unmarshal(raw, &values)
+	return &types.SyncCursor{LastSyncTime: s.LastComplete, ConnectorCursor: values}
+}
+
+func (c *Connector) PrepareFullSyncCursor(previous *types.SyncCursor) (*types.SyncCursor, error) {
+	s, err := parseCursor(previous)
+	if err != nil {
+		return nil, err
+	}
+	s.Run = newRun(true)
+	return s.cursor(), nil
+}
+
+func (c *Connector) PrepareSyncRunCursor(previous *types.SyncCursor, runID string, full bool) (*types.SyncCursor, error) {
+	s, err := parseCursor(previous)
+	if err != nil {
+		return nil, err
+	}
+	if s.TaskRunID != runID && full {
+		s.Run = newRun(true)
+	}
+	s.TaskRunID = runID
+	return s.cursor(), nil
+}
````

### A.27 `internal/datasource/connector/outline/markdown.go`

````diff
diff --git a/internal/datasource/connector/outline/markdown.go b/internal/datasource/connector/outline/markdown.go
new file mode 100644
index 00000000..2ce18413
--- /dev/null
+++ b/internal/datasource/connector/outline/markdown.go
@@ -0,0 +1,85 @@
+package outline
+
+import (
+	"bytes"
+	"net/url"
+	"sort"
+	"strings"
+
+	"github.com/yuin/goldmark"
+	"github.com/yuin/goldmark/ast"
+	"github.com/yuin/goldmark/text"
+	"github.com/yuin/goldmark/util"
+)
+
+// Rewrite only destinations selected by the Markdown parser, preserving the
+// original formatting and code verbatim. Goldmark destinations are source slices,
+// including reference definitions; pointer equality disambiguates identical text.
+func absoluteLinks(body, source string) string {
+	raw := []byte(body)
+	root := goldmark.DefaultParser().Parse(text.NewReader(raw))
+	base, err := url.Parse(source)
+	if err != nil {
+		return body
+	}
+	type edit struct {
+		start, end  int
+		replacement string
+	}
+	edits := map[int]edit{}
+	visited := map[*byte]bool{}
+	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
+		if !entering {
+			return ast.WalkContinue, nil
+		}
+		var dest []byte
+		switch n := node.(type) {
+		case *ast.Link:
+			dest = n.Destination
+		case *ast.Image:
+			dest = n.Destination
+		default:
+			return ast.WalkContinue, nil
+		}
+		if len(dest) == 0 || dest[0] == '#' || visited[&dest[0]] {
+			return ast.WalkContinue, nil
+		}
+		visited[&dest[0]] = true
+		u, err := url.Parse(string(util.UnescapePunctuations(dest)))
+		if err != nil || u.IsAbs() {
+			return ast.WalkContinue, nil
+		}
+		resolved := base.ResolveReference(u)
+		if resolved.Scheme != "http" && resolved.Scheme != "https" {
+			return ast.WalkContinue, nil
+		}
+		for offset := 0; offset < len(raw); {
+			i := bytes.Index(raw[offset:], dest)
+			if i < 0 {
+				break
+			}
+			i += offset
+			if &raw[i] == &dest[0] {
+				replacement := strings.NewReplacer("(", "%28", ")", "%29", "<", "%3C", ">", "%3E").Replace(resolved.String())
+				edits[i] = edit{i, i + len(dest), replacement}
+				break
+			}
+			offset = i + 1
+		}
+		return ast.WalkContinue, nil
+	})
+	ordered := make([]edit, 0, len(edits))
+	for _, e := range edits {
+		ordered = append(ordered, e)
+	}
+	sort.Slice(ordered, func(i, j int) bool { return ordered[i].start < ordered[j].start })
+	var out strings.Builder
+	offset := 0
+	for _, e := range ordered {
+		out.Write(raw[offset:e.start])
+		out.WriteString(e.replacement)
+		offset = e.end
+	}
+	out.Write(raw[offset:])
+	return out.String()
+}
````

### A.28 `internal/datasource/connector/outline/outline_test.go`

````diff
diff --git a/internal/datasource/connector/outline/outline_test.go b/internal/datasource/connector/outline/outline_test.go
new file mode 100644
index 00000000..bffe78c0
--- /dev/null
+++ b/internal/datasource/connector/outline/outline_test.go
@@ -0,0 +1,586 @@
+package outline
+
+import (
+	"context"
+	"crypto/sha256"
+	"encoding/json"
+	"errors"
+	"fmt"
+	"io"
+	"net/http"
+	"net/http/httptest"
+	"strings"
+	"testing"
+	"time"
+	"unicode/utf8"
+
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/Tencent/WeKnora/internal/utils"
+	"golang.org/x/time/rate"
+)
+
+const collectionA = "11111111-1111-4111-8111-111111111111"
+const collectionB = "22222222-2222-4222-8222-222222222222"
+
+type fixture struct {
+	listOverride     map[string]json.RawMessage
+	detailCalls      int
+	collections      []collection
+	terminalNextPath bool
+	detailOverride   map[string]json.RawMessage
+	docs             map[string]document
+	listed           map[string][]string
+	deleted          []string
+	deny             map[string]int
+	actor            string
+	workspace        string
+}
+
+func sample(id, col string) document {
+	body := "hello"
+	return document{ID: id, CollectionID: col, Title: id, Text: &body,
+		URL: "/doc/" + id, PublishedAt: "2026-09-01T00:00:00Z",
+		UpdatedAt: "2026-09-01T00:00:00Z", Revision: "1"}
+}
+
+func testFixture(t *testing.T) (*fixture, *types.DataSourceConfig) {
+	t.Helper()
+	utils.SetSSRFWhitelistFromRaw("127.0.0.1")
+	t.Cleanup(utils.ResetSSRFWhitelistForTest)
+	f := &fixture{docs: map[string]document{}, listed: map[string][]string{}, deny: map[string]int{}, actor: "actor", workspace: "team", detailOverride: map[string]json.RawMessage{}, listOverride: map[string]json.RawMessage{}}
+	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
+		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("X-API-Version") != "2" {
+			t.Error("invalid RPC authentication or method")
+		}
+		var body struct {
+			ID               string   `json:"id"`
+			CollectionID     string   `json:"collectionId"`
+			Offset           int      `json:"offset"`
+			Limit            int      `json:"limit"`
+			Sort             string   `json:"sort"`
+			Direction        string   `json:"direction"`
+			StatusFilter     []string `json:"statusFilter"`
+			ParentDocumentID *string  `json:"parentDocumentId"`
+		}
+		if json.NewDecoder(r.Body).Decode(&body) != nil {
+			t.Error("invalid JSON request")
+		}
+		method := strings.TrimPrefix(r.URL.Path, "/api/")
+		if status := f.deny[method+":"+body.ID]; status != 0 {
+			w.WriteHeader(status)
+			fmt.Fprint(w, `{"ok":false}`)
+			return
+		}
+		w.Header().Set("Content-Type", "application/json")
+		var data interface{}
+		total := -1
+		switch method {
+		case "auth.info":
+			data = map[string]interface{}{"user": map[string]string{"id": f.actor}, "team": map[string]string{"id": f.workspace}}
+		case "collections.info":
+			data = collection{ID: body.ID, Name: body.ID}
+		case "collections.list":
+			cols := f.collections
+			if cols == nil {
+				cols = []collection{{ID: collectionA, Name: "A"}, {ID: collectionB, Name: "B"}}
+			}
+			total = len(cols)
+			data = cols[min(body.Offset, total):min(total, body.Offset+body.Limit)]
+		case "documents.list":
+			if body.Sort != "createdAt" || body.Direction != "ASC" || body.ParentDocumentID != nil || len(body.StatusFilter) != 1 || body.StatusFilter[0] != "published" {
+				t.Error("list must include published documents at all levels")
+			}
+			rows := []json.RawMessage{}
+			ids := f.listed[body.CollectionID]
+			total = len(ids)
+			for i := body.Offset; i < min(len(ids), body.Offset+body.Limit); i++ {
+				raw, ok := f.listOverride[ids[i]]
+				if !ok {
+					raw, _ = json.Marshal(f.docs[ids[i]])
+				}
+				rows = append(rows, raw)
+			}
+			data = rows
+		case "documents.info":
+			f.detailCalls++
+			if raw, ok := f.detailOverride[body.ID]; ok {
+				data = raw
+				break
+			}
+			d, ok := f.docs[body.ID]
+			if !ok {
+				w.WriteHeader(404)
+				fmt.Fprint(w, `{"ok":false,"error":"not_found"}`)
+				return
+			}
+			data = map[string]interface{}{"document": d}
+		case "documents.deleted":
+			total = len(f.deleted)
+			rows := []map[string]string{}
+			for i := body.Offset; i < min(len(f.deleted), body.Offset+body.Limit); i++ {
+				rows = append(rows, map[string]string{"id": f.deleted[i]})
+			}
+			data = rows
+		default:
+			t.Errorf("unexpected method %s", method)
+		}
+		response := map[string]interface{}{"ok": true, "data": data}
+		if f.terminalNextPath && total >= 0 {
+			response["pagination"] = map[string]interface{}{"total": total, "limit": body.Limit, "offset": body.Offset,
+				"nextPath": fmt.Sprintf("/api/%s?limit=%d&offset=%d", method, body.Limit, body.Offset+body.Limit)}
+		}
+		_ = json.NewEncoder(w).Encode(response)
+	}))
+	t.Cleanup(server.Close)
+	digest := sha256.Sum256([]byte(server.URL + "\x00test-key"))
+	limiters.Store(digest, rate.NewLimiter(rate.Inf, 1))
+	t.Cleanup(func() { limiters.Delete(digest) })
+	return f, &types.DataSourceConfig{Type: types.ConnectorTypeOutline,
+		Credentials: map[string]interface{}{"base_url": server.URL, "api_key": "test-key"},
+		ResourceIDs: []string{collectionA}, SyncDeletions: true}
+}
+
+type acknowledger struct {
+	items         []types.FetchedItem
+	fail          map[string]bool
+	checkpoints   []*types.SyncCursor
+	checkpointErr error
+	stopAfter     int
+}
+
+func (h *acknowledger) Emit(context.Context, types.FetchedItem) error {
+	panic("Outline must request acknowledgement")
+}
+func (h *acknowledger) EmitWithResult(_ context.Context, item types.FetchedItem) (datasource.ApplyResult, error) {
+	h.items = append(h.items, item)
+	outcome := datasource.ApplyApplied
+	if h.fail[item.ExternalID] {
+		outcome = datasource.ApplyFailed
+	}
+	return datasource.ApplyResult{Outcome: outcome}, nil
+}
+func (h *acknowledger) Checkpoint(_ context.Context, cursor *types.SyncCursor) error {
+	if h.checkpointErr != nil && len(h.checkpoints) >= h.stopAfter {
+		return h.checkpointErr
+	}
+	raw, _ := json.Marshal(cursor)
+	var copy types.SyncCursor
+	_ = json.Unmarshal(raw, &copy)
+	h.checkpoints = append(h.checkpoints, &copy)
+	return nil
+}
+
+func TestAcknowledgementRetryAndIncremental(t *testing.T) {
+	f, cfg := testFixture(t)
+	f.docs["a"], f.docs["b"] = sample("a", collectionA), sample("b", collectionA)
+	f.listed[collectionA] = []string{"a", "b"}
+	c := NewConnector()
+	h := &acknowledger{fail: map[string]bool{"a": true}}
+	next, err := c.FetchStream(context.Background(), cfg, nil, h)
+	var partial *datasource.PartialFetchError
+	if !errors.As(err, &partial) {
+		t.Fatalf("expected partial, got %v", err)
+	}
+	s, _ := parseCursor(next)
+	if _, ok := s.Documents["a"]; ok {
+		t.Fatal("failed item advanced version")
+	}
+	if s.PendingUpserts["a"] == "" || s.Documents["b"].Fingerprint == "" {
+		t.Fatal("missing pending/success state")
+	}
+	h = &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, next, h)
+	if err != nil || len(h.items) != 1 || h.items[0].ExternalID != "a" {
+		t.Fatalf("retry: %v %#v", err, h.items)
+	}
+	h = &acknowledger{}
+	_, err = c.FetchStream(context.Background(), cfg, next, h)
+	if err != nil || len(h.items) != 0 {
+		t.Fatalf("unchanged documents emitted: %v %#v", err, h.items)
+	}
+}
+
+func TestDeletionFailureAndDisabledDeletionRetainIdentity(t *testing.T) {
+	f, cfg := testFixture(t)
+	f.docs["a"] = sample("a", collectionA)
+	f.listed[collectionA] = []string{"a"}
+	c := NewConnector()
+	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	if err != nil {
+		t.Fatal(err)
+	}
+	delete(f.docs, "a")
+	f.listed[collectionA] = nil
+	f.deleted = []string{"a"}
+	h := &acknowledger{fail: map[string]bool{"a": true}}
+	next, _ = c.FetchStream(context.Background(), cfg, next, h)
+	s, _ := parseCursor(next)
+	if len(s.Documents) != 1 || s.PendingDeletions["a"] != "pending" {
+		t.Fatal("failed deletion lost identity")
+	}
+	cfg.SyncDeletions = false
+	h = &acknowledger{}
+	next, _ = c.FetchStream(context.Background(), cfg, next, h)
+	s, _ = parseCursor(next)
+	if len(s.Documents) != 1 || len(h.items) != 0 {
+		t.Fatal("disabled deletion removed identity/content")
+	}
+	cfg.SyncDeletions = true
+	h = &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, next, h)
+	s, _ = parseCursor(next)
+	if err != nil || len(s.Documents) != 0 || len(h.items) != 1 || !h.items[0].IsDeleted {
+		t.Fatalf("delete retry failed: %v", err)
+	}
+}
+
+func TestUnknownMissingAndIncompleteScanNeverDelete(t *testing.T) {
+	f, cfg := testFixture(t)
+	cfg.ResourceIDs = []string{collectionA, collectionB}
+	f.docs["a"] = sample("a", collectionA)
+	f.listed[collectionA] = []string{"a"}
+	c := NewConnector()
+	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	if err != nil {
+		t.Fatal(err)
+	}
+	delete(f.docs, "a")
+	f.listed[collectionA] = nil
+	for i := 0; i < 3; i++ {
+		h := &acknowledger{}
+		next, _ = c.FetchStream(context.Background(), cfg, next, h)
+		if len(h.items) != 0 {
+			t.Fatal("uncertified 404 deleted document")
+		}
+	}
+	f.deleted = []string{"a"}
+	f.deny["collections.info:"+collectionB] = 403
+	h := &acknowledger{}
+	next, _ = c.FetchStream(context.Background(), cfg, next, h)
+	s, _ := parseCursor(next)
+	if len(h.items) != 0 || len(s.Documents) != 1 {
+		t.Fatal("incomplete scan deleted content")
+	}
+}
+
+func TestMoveAcrossCollectionsAndDeselect(t *testing.T) {
+	f, cfg := testFixture(t)
+	cfg.ResourceIDs = []string{collectionA, collectionB}
+	f.docs["a"] = sample("a", collectionA)
+	f.listed[collectionA] = []string{"a"}
+	c := NewConnector()
+	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	if err != nil {
+		t.Fatal(err)
+	}
+	f.docs["a"] = sample("a", collectionB)
+	f.listed[collectionA], f.listed[collectionB] = nil, []string{"a"}
+	h := &acknowledger{fail: map[string]bool{"a": true}}
+	next, _ = c.FetchStream(context.Background(), cfg, next, h)
+	for _, item := range h.items {
+		if item.IsDeleted {
+			t.Fatal("cross-collection move deleted")
+		}
+	}
+	h = &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, next, h)
+	if err != nil {
+		t.Fatal(err)
+	}
+	cfg.ResourceIDs = []string{collectionA}
+	h = &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, next, h)
+	s, _ := parseCursor(next)
+	if err != nil || len(h.items) != 1 || !h.items[0].IsDeleted || len(s.Documents) != 0 {
+		t.Fatal("deselection did not reconcile the old copy")
+	}
+}
+
+func TestFullSyncPreservesBaselineAndResumes(t *testing.T) {
+	f, cfg := testFixture(t)
+	for i := 0; i < 60; i++ {
+		id := fmt.Sprintf("doc-%03d", i)
+		f.docs[id] = sample(id, collectionA)
+		f.listed[collectionA] = append(f.listed[collectionA], id)
+	}
+	c := NewConnector()
+	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	if err != nil {
+		t.Fatal(err)
+	}
+	next, err = c.PrepareFullSyncCursor(next)
+	if err != nil {
+		t.Fatal(err)
+	}
+	s, _ := parseCursor(next)
+	if len(s.Documents) != 60 {
+		t.Fatal("full sync dropped baseline")
+	}
+	h := &acknowledger{checkpointErr: errors.New("disk failure"), stopAfter: 2}
+	_, err = c.FetchStream(context.Background(), cfg, next, h)
+	if err == nil || len(h.checkpoints) < 2 {
+		t.Fatal("checkpoint error swallowed")
+	}
+	last := h.checkpoints[len(h.checkpoints)-1]
+	s, _ = parseCursor(last)
+	h = &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, last, h)
+	if err != nil || len(h.items) != 60-len(s.Run.Applied) {
+		t.Fatalf("resume re-emitted confirmed versions: %v, %d", err, len(h.items))
+	}
+}
+
+func TestConfigurationAndCursorValidation(t *testing.T) {
+	_, cfg := testFixture(t)
+	for _, base := range []string{"", "https://user:pass@example.com", "https://example.com/api", "https://example.com?q=1", "https://example.com/#x", "file:///tmp"} {
+		copy := *cfg
+		copy.Credentials = map[string]interface{}{"base_url": base, "api_key": "key"}
+		if _, err := newClient(&copy); err == nil {
+			t.Errorf("accepted invalid base %q", base)
+		}
+	}
+	for _, values := range []map[string]interface{}{{}, {"version": 2}, {"version": 1, "documents": nil}} {
+		if _, err := parseCursor(&types.SyncCursor{ConnectorCursor: values}); err == nil {
+			t.Fatal("invalid cursor accepted")
+		}
+	}
+}
+
+func TestDocumentShapesAndMarkdown(t *testing.T) {
+	for _, raw := range []string{`{"id":"a","text":""}`, `{"document":{"id":"a","text":""}}`} {
+		d, err := decodeDocument(json.RawMessage(raw))
+		if err != nil || d.Text == nil || *d.Text != "" {
+			t.Fatalf("empty text must be valid: %v", err)
+		}
+	}
+	body := "[relative](../other) ![image](/images/a.png)\n\n`[code](/stay)`\n\n```md\n[code](/stay)\n```\n\n[ref][r]\n\n[r]: /reference\n"
+	got := absoluteLinks(body, "https://outline.example.com/doc/a")
+	for _, want := range []string{"https://outline.example.com/other", "https://outline.example.com/images/a.png",
+		"`[code](/stay)`", "```md\n[code](/stay)\n```", "https://outline.example.com/reference"} {
+		if !strings.Contains(got, want) {
+			t.Errorf("missing %q in %s", want, got)
+		}
+	}
+	name := fileName(strings.Repeat("中", 150) + "/bad")
+	if !utf8.ValidString(name) || len(name) > 203 || strings.Contains(name, "/") {
+		t.Fatal("unsafe file name")
+	}
+	a := sample("a", collectionA)
+	a.Revision, a.UpdatedAt = "", ""
+	fp := fingerprint(a, "/a")
+	changed := "changed"
+	a.Text = &changed
+	if fingerprint(a, "/a") == fp {
+		t.Fatal("missing-version content changes skipped")
+	}
+}
+
+func TestCursorTenThousandSize(t *testing.T) {
+	s, _ := parseCursor(nil)
+	s.Run = newRun(true)
+	for i := 0; i < 10000; i++ {
+		id := fmt.Sprintf("%08d-1111-4111-8111-111111111111", i)
+		s.Documents[id] = version{CollectionID: collectionA, Fingerprint: strings.Repeat("a", 64)}
+		s.Run.Seen[id], s.Run.Applied[id], s.PendingUpserts[id] = true, strings.Repeat("b", 64), collectionA
+	}
+	raw, _ := s.cursor().ToJSON()
+	if len(raw) > 10<<20 {
+		t.Fatalf("cursor exceeds 10 MiB: %d", len(raw))
+	}
+	t.Logf("10000-document cursor: %d bytes", len(raw))
+}
+
+func TestRetryAfter(t *testing.T) {
+	if got := retryDelay("7", 0); got != 7*time.Second {
+		t.Fatal(got)
+	}
+	if got := retryDelay("", 2); got != 8*time.Second {
+		t.Fatal(got)
+	}
+}
+
+func TestFullRunPreparationSurvivesAdmissionRetries(t *testing.T) {
+	c := NewConnector()
+	s, _ := parseCursor(nil)
+	s.Documents["a"] = version{CollectionID: collectionA, Fingerprint: "old"}
+	first, err := c.PrepareSyncRunCursor(s.cursor(), "sync-log-1", true)
+	if err != nil {
+		t.Fatal(err)
+	}
+	s, _ = parseCursor(first)
+	if s.Run == nil || !s.Run.ForceFull || len(s.Documents) != 1 {
+		t.Fatal("new logical task must prepare full baseline")
+	}
+	s.Run.Applied["a"] = "new"
+	retry, err := c.PrepareSyncRunCursor(s.cursor(), "sync-log-1", true)
+	if err != nil {
+		t.Fatal(err)
+	}
+	s, _ = parseCursor(retry)
+	if s.Run.Applied["a"] != "new" {
+		t.Fatal("retry discarded applied progress")
+	}
+	fresh, err := c.PrepareSyncRunCursor(s.cursor(), "sync-log-2", true)
+	if err != nil {
+		t.Fatal(err)
+	}
+	s, _ = parseCursor(fresh)
+	if len(s.Run.Applied) != 0 || !s.Run.ForceFull {
+		t.Fatal("new full request reused old progress")
+	}
+}
+
+type roundTripFunc func(*http.Request) (*http.Response, error)
+
+func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
+
+func TestPaginationAcceptsTerminalNextPath(t *testing.T) {
+	_, cfg := testFixture(t)
+	for _, method := range []string{"collections.list", "documents.list", "documents.deleted"} {
+		for _, total := range []int{0, 1, 9, 24, 25, 26, 50} {
+			for _, withTotal := range []bool{false, true} {
+				t.Run(fmt.Sprintf("%s/count=%d/total=%t", method, total, withTotal), func(t *testing.T) {
+					c, err := newClient(cfg)
+					if err != nil {
+						t.Fatal(err)
+					}
+					calls, received := 0, 0
+					c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
+						var request struct{ Offset, Limit int }
+						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
+							t.Fatal(err)
+						}
+						if request.Offset != calls*25 || request.Limit != 25 {
+							t.Fatalf("unexpected pagination: %+v", request)
+						}
+						calls++
+						if calls > total/25+1 {
+							t.Fatal("followed terminal nextPath")
+						}
+						rows := []map[string]string{}
+						for i := request.Offset; i < min(total, request.Offset+request.Limit); i++ {
+							rows = append(rows, map[string]string{"id": fmt.Sprintf("item-%d", i)})
+						}
+						// The reported deployment returns offset+limit even on a short final page.
+						pagination := map[string]interface{}{"offset": request.Offset, "limit": request.Limit,
+							"nextPath": fmt.Sprintf("/api/%s?limit=25&offset=%d", method, request.Offset+request.Limit)}
+						if withTotal {
+							pagination["total"] = total
+						}
+						data, err := json.Marshal(map[string]interface{}{"ok": true, "data": rows, "pagination": pagination})
+						if err != nil {
+							t.Fatal(err)
+						}
+						return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
+					})
+					err = c.walk(context.Background(), method, map[string]interface{}{}, func(rows []json.RawMessage) error {
+						received += len(rows)
+						return nil
+					})
+					if err != nil {
+						t.Fatalf("valid terminal nextPath rejected: %v", err)
+					}
+					wantCalls := total/25 + 1
+					if withTotal {
+						wantCalls = max(1, (total+24)/25)
+					}
+					if received != total || calls != wantCalls {
+						t.Fatalf("received=%d calls=%d; want %d, %d", received, calls, total, wantCalls)
+					}
+				})
+			}
+		}
+	}
+}
+
+func TestPaginationRejectsUntrustedOrIncompletePages(t *testing.T) {
+	_, cfg := testFixture(t)
+	for _, tc := range []struct {
+		name  string
+		pages []string
+	}{
+		{"cross origin", []string{`{"data":[{"id":"a"}],"pagination":{"nextPath":"https://evil.example/api/documents.list?offset=1"}}`}},
+		{"backwards", []string{`{"data":[{"id":"a"}],"pagination":{"nextPath":"/api/documents.list?offset=0"}}`}},
+		{"gap", []string{`{"data":[{"id":"a"}],"pagination":{"nextPath":"/api/documents.list?offset=9"}}`}},
+		{"terminal cross origin", []string{`{"data":[{"id":"a"}],"pagination":{"total":1,"nextPath":"https://evil.example/api/documents.list?offset=25"}}`}},
+		{"terminal gap", []string{`{"data":[{"id":"a"}],"pagination":{"total":1,"nextPath":"/api/documents.list?offset=9"}}`}},
+		{"nonterminal gap", []string{`{"data":[{"id":"a"}],"pagination":{"total":30,"nextPath":"/api/documents.list?offset=25"}}`}},
+		{"repeated page", []string{`{"data":[{"id":"a"}],"pagination":{"total":3}}`, `{"data":[{"id":"a"}],"pagination":{"total":3}}`}},
+		{"early empty", []string{`{"data":[],"pagination":{"total":5}}`}},
+		{"malformed", []string{`{"data":{}}`}},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			c, err := newClient(cfg)
+			if err != nil {
+				t.Fatal(err)
+			}
+			calls := 0
+			c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
+				if calls >= len(tc.pages) {
+					t.Fatal("unexpected extra pagination request")
+				}
+				data := tc.pages[calls]
+				calls++
+				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(data))}, nil
+			})
+			if err := c.walk(context.Background(), "documents.list", map[string]interface{}{}, func([]json.RawMessage) error { return nil }); err == nil {
+				t.Fatal("incomplete scan accepted")
+			}
+		})
+	}
+}
+
+func TestHTTPRetryBoundAndCancellation(t *testing.T) {
+	_, cfg := testFixture(t)
+	c, err := newClient(cfg)
+	if err != nil {
+		t.Fatal(err)
+	}
+	calls := 0
+	status := 429
+	c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
+		calls++
+		return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{"0"}},
+			Body: io.NopCloser(strings.NewReader(`{"ok":false}`))}, nil
+	})
+	_, err = c.call(context.Background(), "documents.list", map[string]interface{}{})
+	if err == nil || calls != 4 {
+		t.Fatalf("retry bound: %d, %v", calls, err)
+	}
+	status, calls = 401, 0
+	_, err = c.call(context.Background(), "documents.list", map[string]interface{}{})
+	if !fatal(err) || calls != 1 {
+		t.Fatalf("authentication should abort: %d, %v", calls, err)
+	}
+	ctx, cancel := context.WithCancel(context.Background())
+	cancel()
+	if err := wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
+		t.Fatal(err)
+	}
+}
+
+func TestInstanceBindingRejectsWorkspaceChanges(t *testing.T) {
+	f, cfg := testFixture(t)
+	c := NewConnector()
+	bound, err := c.BindIdentity(context.Background(), cfg, nil)
+	if err != nil {
+		t.Fatal(err)
+	}
+	f.workspace = "other-team"
+	if _, err := c.BindIdentity(context.Background(), cfg, bound); err == nil {
+		t.Fatal("workspace migration accepted")
+	}
+	f.workspace, f.actor = "team", "new-actor"
+	s, _ := parseCursor(bound)
+	s.Documents["a"] = version{CollectionID: collectionA, Fingerprint: "fp"}
+	s.Run = newRun(false)
+	next, err := c.BindIdentity(context.Background(), cfg, s.cursor())
+	if err != nil {
+		t.Fatal(err)
+	}
+	s, _ = parseCursor(next)
+	if s.Run != nil || len(s.Documents) != 1 || s.Instance.ActorID != f.actor {
+		t.Fatal("actor rotation lost baseline or reused old run")
+	}
+}
````

### A.29 `internal/datasource/connector/outline/pagination.go`

````diff
diff --git a/internal/datasource/connector/outline/pagination.go b/internal/datasource/connector/outline/pagination.go
new file mode 100644
index 00000000..c3cf6f51
--- /dev/null
+++ b/internal/datasource/connector/outline/pagination.go
@@ -0,0 +1,65 @@
+package outline
+
+import (
+	"errors"
+	"net/url"
+	"strconv"
+	"strings"
+)
+
+// nextPage validates the complete page before it is delivered to the sync loop.
+// Outline may advertise offset+limit even for empty and short terminal pages.
+// nextPath supplies pagination numbers only; it is never used as a request URL.
+func (c *client) nextPage(method string, result *envelope, offset, limit, count int) (next, pageLimit int, terminal bool, err error) {
+	invalid := func() (int, int, bool, error) { return 0, 0, false, errors.New("outline_scan_incomplete") }
+	p := result.Pagination
+	if p.Offset != nil && *p.Offset != offset {
+		return invalid()
+	}
+	if p.Limit != nil {
+		if *p.Limit <= 0 || *p.Limit > limit {
+			return invalid()
+		}
+		limit = *p.Limit
+	}
+	if count > limit {
+		return invalid()
+	}
+	next = offset + count
+	if p.Total != nil {
+		if *p.Total < next || (count == 0 && next < *p.Total) {
+			return invalid()
+		}
+		terminal = next == *p.Total
+	} else {
+		terminal = count < limit
+	}
+	if p.NextPath != nil && *p.NextPath != "" {
+		raw := *p.NextPath
+		u, parseErr := url.Parse(raw)
+		if parseErr != nil || u.User != nil || u.Opaque != "" || strings.Contains(raw, "#") ||
+			u.Path != "/api/"+method || (u.Host == "" && u.Scheme != "") ||
+			(u.Host != "" && u.Scheme+"://"+u.Host != c.base) {
+			return invalid()
+		}
+		query, parseErr := url.ParseQuery(u.RawQuery)
+		if parseErr != nil || len(query["offset"]) != 1 {
+			return invalid()
+		}
+		advertised, parseErr := strconv.Atoi(query.Get("offset"))
+		if parseErr != nil || advertised <= offset ||
+			(advertised != next && !(terminal && advertised == offset+limit)) {
+			return invalid()
+		}
+		if values, ok := query["limit"]; ok {
+			n, parseErr := strconv.Atoi(query.Get("limit"))
+			if parseErr != nil || len(values) != 1 || n != limit {
+				return invalid()
+			}
+		}
+	}
+	if !terminal && next <= offset {
+		return invalid()
+	}
+	return next, limit, terminal, nil
+}
````

### A.30 `internal/datasource/connector/outline/pagination_test.go`

````diff
diff --git a/internal/datasource/connector/outline/pagination_test.go b/internal/datasource/connector/outline/pagination_test.go
new file mode 100644
index 00000000..82e59f8f
--- /dev/null
+++ b/internal/datasource/connector/outline/pagination_test.go
@@ -0,0 +1,173 @@
+package outline
+
+import (
+	"context"
+	"encoding/json"
+	"fmt"
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"io"
+	"net/http"
+	"strings"
+	"testing"
+
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/stretchr/testify/require"
+)
+
+// This models the reported 9/25 response without retaining instance IDs or secrets.
+func TestCollectionDiscoveryUsesTerminalPagination(t *testing.T) {
+	f, cfg := testFixture(t)
+	f.terminalNextPath = true
+	for i := 0; i < 9; i++ {
+		f.collections = append(f.collections, collection{ID: fmt.Sprintf("%08d-1111-4111-8111-111111111111", i), Name: fmt.Sprintf("Collection %d", i)})
+	}
+	cfg.ResourceIDs, cfg.SyncDeletions = nil, false
+	c := NewConnector()
+	require.NoError(t, c.Validate(context.Background(), cfg))
+	resources, err := c.ListResources(context.Background(), cfg, "")
+	require.NoError(t, err)
+	require.Len(t, resources, 9)
+}
+
+func TestPaginationMetadataValidation(t *testing.T) {
+	c := &client{base: "https://outline.example.com"}
+	for _, tc := range []struct {
+		name, pagination string
+		count            int
+	}{
+		{"wrong offset", `{"offset":1,"total":1}`, 1},
+		{"invalid limit", `{"limit":0,"total":1}`, 1},
+		{"oversized page", `{"limit":1,"total":2}`, 2},
+		{"duplicate offset", `{"nextPath":"/api/documents.list?offset=25&offset=50"}`, 1},
+		{"duplicate limit", `{"nextPath":"/api/documents.list?offset=25&limit=25&limit=1"}`, 1},
+		{"userinfo", `{"nextPath":"https://user:pass@outline.example.com/api/documents.list?offset=25"}`, 1},
+		{"fragment", `{"nextPath":"/api/documents.list?offset=25#x"}`, 1},
+		{"wrong endpoint", `{"nextPath":"/api/collections.list?offset=25"}`, 1},
+		{"malformed query", `{"nextPath":"/api/documents.list?offset=25&limit=%ZZ"}`, 1},
+		{"negative total", `{"total":-1}`, 0},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			var e envelope
+			require.NoError(t, json.Unmarshal([]byte(`{"pagination":`+tc.pagination+`}`), &e))
+			_, _, _, err := c.nextPage("documents.list", &e, 0, 25, tc.count)
+			require.ErrorContains(t, err, "outline_scan_incomplete")
+		})
+	}
+	for _, path := range []string{"", "/api/documents.list?offset=1&limit=1"} {
+		var e envelope
+		b, err := json.Marshal(map[string]interface{}{"pagination": map[string]interface{}{"limit": 1, "offset": 0, "total": 2, "nextPath": path}})
+		require.NoError(t, err)
+		require.NoError(t, json.Unmarshal(b, &e))
+		next, limit, terminal, err := c.nextPage("documents.list", &e, 0, 25, 1)
+		require.NoError(t, err)
+		require.Equal(t, 1, next)
+		require.Equal(t, 1, limit)
+		require.False(t, terminal)
+	}
+}
+
+func TestDeletionRetainsDocumentWhenDetailStateIsMissing(t *testing.T) {
+	f, cfg := testFixture(t)
+	f.docs["a"] = sample("a", collectionA)
+	f.listed[collectionA] = []string{"a"}
+	c := NewConnector()
+	previous, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	require.NoError(t, err)
+	f.listed[collectionA] = nil
+	f.deleted = []string{"a"}
+	f.detailOverride["a"] = json.RawMessage(`{"id":"a","text":"still readable"}`)
+	h := &acknowledger{}
+	next, err := c.FetchStream(context.Background(), cfg, previous, h)
+	require.Error(t, err)
+	require.Empty(t, h.items)
+	state, err := parseCursor(next)
+	require.NoError(t, err)
+	require.Contains(t, state.Documents, "a")
+}
+
+func TestMissingListStateRequiresDetailBeforeAcceptance(t *testing.T) {
+	f, cfg := testFixture(t)
+	f.docs["a"] = sample("a", collectionA)
+	f.listed[collectionA] = []string{"a"}
+	f.listOverride["a"] = json.RawMessage(`{"id":"a","text":"summary"}`)
+	h := &acknowledger{}
+	next, err := NewConnector().FetchStream(context.Background(), cfg, nil, h)
+	require.NoError(t, err)
+	require.Equal(t, 1, f.detailCalls)
+	require.Len(t, h.items, 1)
+	require.Contains(t, string(h.items[0].Content), "hello")
+	state, err := parseCursor(next)
+	require.NoError(t, err)
+	require.Contains(t, state.Documents, "a")
+}
+
+func TestSourceTimestampsRemainInKnowledgeMetadata(t *testing.T) {
+	d := sample("a", collectionA)
+	d.CreatedAt = "2026-08-01T08:00:00+08:00"
+	item, _, err := mappedItem(d, identity{BaseURL: "https://outline.example.com"})
+	require.NoError(t, err)
+	require.Equal(t, "2026-08-01T00:00:00Z", item.Metadata["source_created_at"])
+	require.Equal(t, "2026-09-01T00:00:00Z", item.Metadata["source_updated_at"])
+	require.Equal(t, types.ChannelOutline, item.Metadata["channel"])
+}
+
+func TestOversizedListReducesLimitAtSameOffset(t *testing.T) {
+	_, cfg := testFixture(t)
+	c, err := newClient(cfg)
+	require.NoError(t, err)
+	oversized := strings.Repeat("x", maxResponse+1)
+	var offsets, limits []int
+	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
+		var request struct{ Offset, Limit int }
+		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
+		offsets = append(offsets, request.Offset)
+		limits = append(limits, request.Limit)
+		data := oversized
+		if request.Limit <= 6 {
+			rows := []map[string]string{}
+			for i := request.Offset; i < min(9, request.Offset+request.Limit); i++ {
+				rows = append(rows, map[string]string{"id": fmt.Sprint(i)})
+			}
+			raw, err := json.Marshal(map[string]interface{}{"data": rows, "pagination": map[string]int{"total": 9}})
+			require.NoError(t, err)
+			data = string(raw)
+		}
+		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(data))}, nil
+	})
+	count := 0
+	require.NoError(t, c.walk(context.Background(), "documents.list", map[string]interface{}{}, func(rows []json.RawMessage) error { count += len(rows); return nil }))
+	require.Equal(t, []int{0, 0, 0, 6}, offsets)
+	require.Equal(t, []int{25, 12, 6, 6}, limits)
+	require.Equal(t, 9, count)
+}
+
+func TestPendingFailuresAreNotTruncatedWithErrorSamples(t *testing.T) {
+	f, cfg := testFixture(t)
+	for i := 0; i < 150; i++ {
+		id := fmt.Sprintf("doc-%d", i)
+		d := sample(id, collectionA)
+		d.Text = nil
+		f.docs[id] = d
+		f.listed[collectionA] = append(f.listed[collectionA], id)
+	}
+	c := NewConnector()
+	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
+	var partial *datasource.PartialFetchError
+	require.ErrorAs(t, err, &partial)
+	require.Len(t, partial.Details, 100)
+	state, err := parseCursor(next)
+	require.NoError(t, err)
+	require.Len(t, state.PendingUpserts, 150)
+	require.Empty(t, state.Documents)
+	for id := range f.docs {
+		f.docs[id] = sample(id, collectionA)
+	}
+	h := &acknowledger{}
+	next, err = c.FetchStream(context.Background(), cfg, next, h)
+	require.NoError(t, err)
+	state, err = parseCursor(next)
+	require.NoError(t, err)
+	require.Len(t, state.Documents, 150)
+	require.Empty(t, state.PendingUpserts)
+	require.Len(t, h.items, 150)
+}
````

### A.31 `internal/datasource/connector/outline/stream.go`

````diff
diff --git a/internal/datasource/connector/outline/stream.go b/internal/datasource/connector/outline/stream.go
new file mode 100644
index 00000000..d25b34f8
--- /dev/null
+++ b/internal/datasource/connector/outline/stream.go
@@ -0,0 +1,422 @@
+package outline
+
+import (
+	"context"
+	"encoding/json"
+	"errors"
+	"slices"
+	"time"
+
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+)
+
+func (connector *Connector) FetchStream(ctx context.Context, config *types.DataSourceConfig, cursor *types.SyncCursor, handler datasource.StreamHandler) (*types.SyncCursor, error) {
+	h, ok := handler.(datasource.AcknowledgingStreamHandler)
+	if !ok {
+		return nil, errors.New("outline_acknowledgement_required")
+	}
+	s, err := parseCursor(cursor)
+	if err != nil {
+		return nil, err
+	}
+	c, err := newClient(config)
+	if err != nil {
+		return nil, err
+	}
+	s.Metrics = c.metrics
+	ids, err := selectedIDs(config.ResourceIDs)
+	if err != nil {
+		return nil, err
+	}
+	if len(ids) == 0 {
+		return nil, errors.New("outline_collection_required")
+	}
+	instance, err := authenticate(ctx, c)
+	if err != nil {
+		return nil, err
+	}
+	if s.Instance.BaseURL != "" && (s.Instance.BaseURL != instance.BaseURL || s.Instance.WorkspaceID != instance.WorkspaceID) {
+		return nil, errors.New("outline_instance_changed")
+	}
+	if s.Instance.ActorID != instance.ActorID || !slices.Equal(s.Selected, ids) {
+		full := s.Run != nil && s.Run.ForceFull
+		s.Run = newRun(full)
+	}
+	s.Instance, s.Selected = instance, ids
+	selected := map[string]bool{}
+	for _, id := range ids {
+		selected[id] = true
+	}
+	// Keep documents from deselected collections until a complete scan and
+	// detail recheck confirm that they no longer belong to the selected scope.
+	// Older releases dropped these identities but kept their local copies.
+	if reader, ok := handler.(datasource.SyncedItemReader); ok {
+		if err := reader.WalkSyncedItems(ctx, func(item types.FetchedItem) error {
+			if _, known := s.Documents[item.ExternalID]; known || selected[item.SourceResourceID] {
+				return nil
+			}
+			if item.ExternalID == "" || item.SourceResourceID == "" ||
+				item.Metadata["source_fingerprint"] == "" ||
+				item.Metadata["outline_workspace_id"] != instance.WorkspaceID {
+				return nil
+			}
+			if _, err := sourceURL(instance.BaseURL, item.Metadata["source_url"]); err != nil {
+				return nil
+			}
+			s.Documents[item.ExternalID] = version{
+				CollectionID: item.SourceResourceID, Fingerprint: item.Metadata["source_fingerprint"],
+				UpdatedAt: item.Metadata["source_updated_at"],
+			}
+			return nil
+		}); err != nil {
+			return s.cursor(), err
+		}
+	}
+	for id, collectionID := range s.PendingUpserts {
+		if !selected[collectionID] {
+			delete(s.PendingUpserts, id)
+		}
+	}
+	if s.Run == nil {
+		s.Run = newRun(false)
+	}
+	checkpoint := func() error {
+		s.Metrics["pending_upserts"] = int64(len(s.PendingUpserts))
+		s.Metrics["pending_deletions"] = int64(len(s.PendingDeletions))
+		s.Metrics["deferred"] = 0
+		for _, status := range s.PendingDeletions {
+			if status == "deferred" {
+				s.Metrics["deferred"]++
+			}
+		}
+		cursor := s.cursor()
+		raw, err := cursor.ToJSON()
+		if err != nil {
+			return err
+		}
+		s.Metrics["cursor_bytes"] = int64(len(raw))
+		start := time.Now()
+		err = h.Checkpoint(ctx, s.cursor())
+		s.Metrics["checkpoint_duration_ms"] = time.Since(start).Milliseconds()
+		return err
+	}
+	if err := checkpoint(); err != nil {
+		return nil, err
+	}
+	warnings := []string{}
+	warn := func(code string) {
+		if len(warnings) < 100 {
+			warnings = append(warnings, code)
+		}
+	}
+	deletionConfig := *config
+	if config.SyncDeletions {
+		response, err := c.call(ctx, "documents.deleted", map[string]interface{}{"limit": 1, "offset": 0})
+		if err == nil {
+			var rows []json.RawMessage
+			if json.Unmarshal(response.Data, &rows) != nil {
+				err = errors.New("outline_response_invalid")
+			}
+		}
+		if err != nil {
+			if fatal(err) {
+				return s.cursor(), err
+			}
+			deletionConfig.SyncDeletions = false
+			warn("outline_deletion_capability_unavailable")
+		}
+	}
+	complete := true
+	unavailable := map[string]bool{}
+	for _, id := range ids {
+		col, err := collectionInfo(ctx, c, id)
+		if err != nil || col.DeletedAt != "" {
+			if fatal(err) {
+				return s.cursor(), err
+			}
+			complete = false
+			unavailable[id] = true
+			warn("outline_scan_incomplete: " + id)
+		} else if col.ArchivedAt != "" {
+			selected[id] = false
+			for documentID, old := range s.Documents {
+				if old.CollectionID == id {
+					delete(s.Run.Seen, documentID)
+				}
+			}
+		}
+	}
+	lastCheckpoint, count := time.Now(), 0
+	attempted := map[string]string{}
+	failDocument := func(d document, err error) error {
+		warn(err.Error())
+		s.Metrics["emitted"]++
+		_, emitErr := h.EmitWithResult(ctx, types.FetchedItem{ExternalID: d.ID, Title: d.Title,
+			Metadata: map[string]string{"error": err.Error(), "error_reason_code": err.Error(), "error_reason": err.Error()}})
+		return emitErr
+	}
+	apply := func(d document) error {
+		s.Run.Seen[d.ID] = true
+		delete(s.PendingDeletions, d.ID)
+		if !d.stateKnown {
+			summary := d
+			var err error
+			d, err = documentInfo(ctx, c, d.ID)
+			if err != nil || !d.stateKnown {
+				if fatal(err) {
+					return err
+				}
+				collectionID := summary.CollectionID
+				if collectionID == "" {
+					collectionID = s.Documents[summary.ID].CollectionID
+				}
+				s.PendingUpserts[summary.ID] = collectionID
+				if err == nil {
+					err = errors.New("outline_response_invalid")
+				}
+				return failDocument(summary, err)
+			}
+		}
+		if !d.stateKnown {
+			return failDocument(d, errors.New("outline_response_invalid"))
+		}
+		if !d.active(selected) {
+			delete(s.Run.Seen, d.ID)
+			delete(s.PendingUpserts, d.ID)
+			return nil
+		}
+		_, retrying := s.PendingUpserts[d.ID]
+		s.PendingUpserts[d.ID] = d.CollectionID
+		old := s.Documents[d.ID]
+		previousTime, _ := time.Parse(time.RFC3339, old.UpdatedAt)
+		currentTime, _ := time.Parse(time.RFC3339, d.UpdatedAt)
+		if d.Text == nil || previousTime.After(currentTime) {
+			summary := d
+			var err error
+			d, err = documentInfo(ctx, c, d.ID)
+			if err != nil {
+				if fatal(err) {
+					return err
+				}
+				return failDocument(summary, err)
+			}
+		}
+		if !d.stateKnown {
+			return failDocument(d, errors.New("outline_response_invalid"))
+		}
+		if !d.active(selected) {
+			delete(s.Run.Seen, d.ID)
+			delete(s.PendingUpserts, d.ID)
+			return nil
+		}
+		currentTime, _ = time.Parse(time.RFC3339, d.UpdatedAt)
+		if previousTime.After(currentTime) {
+			return failDocument(d, errors.New("outline_response_invalid"))
+		}
+		item, fp, err := mappedItem(d, instance)
+		if err != nil {
+			return failDocument(d, err)
+		}
+		if s.Run.Applied[d.ID] == fp || (!retrying && !s.Run.ForceFull && old.Fingerprint == fp) {
+			s.Metrics["unchanged"]++
+			delete(s.PendingUpserts, d.ID)
+			return nil
+		}
+		if attempted[d.ID] == fp {
+			return nil
+		}
+		attempted[d.ID] = fp
+		if s.Run.ForceFull {
+			item.Metadata["source_full_sync"] = "true"
+			item.Metadata["source_sync_run_id"] = s.Run.ID
+		}
+		result, err := h.EmitWithResult(ctx, item)
+		s.Metrics["emitted"]++
+		if err != nil {
+			return err
+		}
+		if result.Outcome == datasource.ApplyApplied {
+			s.Documents[d.ID] = version{CollectionID: d.CollectionID, Fingerprint: fp, UpdatedAt: d.UpdatedAt}
+			s.Run.Applied[d.ID] = fp
+			delete(s.PendingUpserts, d.ID)
+		}
+		count++
+		if count >= 50 || time.Since(lastCheckpoint) >= 30*time.Second {
+			if err := checkpoint(); err != nil {
+				return err
+			}
+			lastCheckpoint, count = time.Now(), 0
+		}
+		return nil
+	}
+	// Pending items from a completed collection must also be retried on resumption.
+	for id, collectionID := range s.PendingUpserts {
+		if unavailable[collectionID] {
+			continue
+		}
+		d, err := documentInfo(ctx, c, id)
+		if err != nil {
+			if fatal(err) {
+				return s.cursor(), err
+			}
+			warn(err.Error())
+			continue
+		}
+		if err := apply(d); err != nil {
+			return s.cursor(), err
+		}
+	}
+	for _, id := range ids {
+		if unavailable[id] {
+			continue
+		}
+		if s.Run.Completed[id] {
+			continue
+		}
+		var err error
+		if selected[id] {
+			var handlerErr error
+			err = c.walk(ctx, "documents.list", map[string]interface{}{
+				"collectionId": id, "sort": "createdAt", "direction": "ASC", "statusFilter": []string{"published"},
+			}, func(rows []json.RawMessage) error {
+				for _, raw := range rows {
+					d, err := decodeDocument(raw)
+					if err != nil {
+						return err
+					}
+					s.Run.Seen[d.ID] = true
+					s.Metrics["discovered"]++
+					if err := apply(d); err != nil {
+						handlerErr = err
+						return err
+					}
+				}
+				handlerErr = checkpoint()
+				return handlerErr
+			})
+			if handlerErr != nil {
+				return s.cursor(), handlerErr
+			}
+		}
+		if err != nil {
+			if fatal(err) {
+				return s.cursor(), err
+			}
+			// A checkpoint or ingest infrastructure error must not be swallowed.
+			if ctx.Err() != nil {
+				return s.cursor(), ctx.Err()
+			}
+			complete = false
+			warn("outline_scan_incomplete: " + id)
+			continue
+		}
+		s.Run.Completed[id] = true
+		if err := checkpoint(); err != nil {
+			return s.cursor(), err
+		}
+	}
+	if complete {
+		if err := reconcile(ctx, c, &deletionConfig, s, selected, h, warn); err != nil {
+			return s.cursor(), err
+		}
+		s.LastComplete = time.Now().UTC()
+		s.Run = nil
+	}
+	if err := checkpoint(); err != nil {
+		return s.cursor(), err
+	}
+	if len(warnings) > 0 || len(s.PendingUpserts) > 0 {
+		if len(warnings) == 0 {
+			warn("outline_pending_upserts")
+		}
+		return s.cursor(), &datasource.PartialFetchError{Details: warnings}
+	}
+	return s.cursor(), nil
+}
+
+func reconcile(ctx context.Context, c *client, config *types.DataSourceConfig, s *state, selected map[string]bool, h datasource.AcknowledgingStreamHandler, warn func(string)) error {
+	candidates := map[string]bool{}
+	for id := range s.Documents {
+		if !s.Run.Seen[id] {
+			candidates[id] = true
+			if !config.SyncDeletions {
+				s.PendingDeletions[id] = "deferred"
+			}
+		}
+	}
+	if len(candidates) == 0 {
+		return nil
+	}
+	deleted := map[string]bool{}
+	if config.SyncDeletions {
+		err := c.walk(ctx, "documents.deleted", map[string]interface{}{}, func(rows []json.RawMessage) error {
+			for _, raw := range rows {
+				var row struct {
+					ID string `json:"id"`
+				}
+				if json.Unmarshal(raw, &row) != nil {
+					return errors.New("outline_response_invalid")
+				}
+				if candidates[row.ID] {
+					deleted[row.ID] = true
+				}
+			}
+			return nil
+		})
+		if err != nil {
+			if fatal(err) {
+				return err
+			}
+			warn("outline_deletion_capability_unavailable")
+			return nil
+		}
+	}
+	for id := range candidates {
+		if err := ctx.Err(); err != nil {
+			return err
+		}
+		d, err := documentInfo(ctx, c, id)
+		confirmed := deleted[id]
+		if err == nil {
+			if !d.stateKnown {
+				warn("outline_deletion_unconfirmed")
+				continue
+			}
+			if d.active(selected) {
+				delete(s.PendingDeletions, id)
+				s.PendingUpserts[id] = d.CollectionID
+				continue
+			}
+			confirmed = true
+		} else if fatal(err) {
+			return err
+		}
+		// Never infer deletion from 403/404: no deployment-specific error contract is certified.
+		if !confirmed {
+			warn("outline_deletion_unconfirmed")
+			continue
+		}
+		s.PendingDeletions[id] = "deferred"
+		if !config.SyncDeletions {
+			continue
+		}
+		result, err := h.EmitWithResult(ctx, types.FetchedItem{ExternalID: id, IsDeleted: true})
+		s.Metrics["emitted"]++
+		if err != nil {
+			return err
+		}
+		if result.Outcome == datasource.ApplyApplied {
+			delete(s.Documents, id)
+			delete(s.PendingUpserts, id)
+			delete(s.PendingDeletions, id)
+		} else {
+			s.PendingDeletions[id] = "pending"
+			warn("deletion_failed")
+		}
+		if err := h.Checkpoint(ctx, s.cursor()); err != nil {
+			return err
+		}
+	}
+	return nil
+}
````

### A.32 `internal/datasource/connector/outline/types.go`

````diff
diff --git a/internal/datasource/connector/outline/types.go b/internal/datasource/connector/outline/types.go
new file mode 100644
index 00000000..cae69f8f
--- /dev/null
+++ b/internal/datasource/connector/outline/types.go
@@ -0,0 +1,159 @@
+package outline
+
+import (
+	"crypto/sha256"
+	"encoding/json"
+	"errors"
+	"fmt"
+	"net/url"
+	"strings"
+	"time"
+	"unicode"
+
+	"github.com/Tencent/WeKnora/internal/types"
+)
+
+type document struct {
+	stateKnown   bool
+	ID           string      `json:"id"`
+	CollectionID string      `json:"collectionId"`
+	ParentID     string      `json:"parentDocumentId"`
+	Title        string      `json:"title"`
+	Text         *string     `json:"text"`
+	URL          string      `json:"url"`
+	CreatedAt    string      `json:"createdAt"`
+	UpdatedAt    string      `json:"updatedAt"`
+	PublishedAt  string      `json:"publishedAt"`
+	ArchivedAt   string      `json:"archivedAt"`
+	DeletedAt    string      `json:"deletedAt"`
+	Revision     json.Number `json:"revision"`
+}
+
+type collection struct {
+	ID         string `json:"id"`
+	Name       string `json:"name"`
+	ArchivedAt string `json:"archivedAt"`
+	DeletedAt  string `json:"deletedAt"`
+}
+
+type identity struct {
+	BaseURL     string `json:"base_url"`
+	WorkspaceID string `json:"workspace_id"`
+	ActorID     string `json:"actor_id"`
+}
+
+func (d document) active(selected map[string]bool) bool {
+	return selected[d.CollectionID] && d.PublishedAt != "" && d.ArchivedAt == "" && d.DeletedAt == ""
+}
+
+func decodeDocument(raw json.RawMessage) (document, error) {
+	var wrapper struct {
+		Document json.RawMessage `json:"document"`
+	}
+	if err := json.Unmarshal(raw, &wrapper); err != nil {
+		return document{}, errors.New("outline_response_invalid")
+	}
+	if len(wrapper.Document) > 0 {
+		raw = wrapper.Document
+	}
+	var d document
+	if json.Unmarshal(raw, &d) != nil || d.ID == "" {
+		return d, errors.New("outline_response_invalid")
+	}
+	var fields map[string]json.RawMessage
+	_ = json.Unmarshal(raw, &fields)
+	d.stateKnown = true
+	for _, key := range []string{"collectionId", "publishedAt", "archivedAt", "deletedAt"} {
+		if _, ok := fields[key]; !ok {
+			d.stateKnown = false
+		}
+	}
+	return d, nil
+}
+
+func sourceURL(base, raw string) (string, error) {
+	if raw == "" {
+		return "", errors.New("outline_response_invalid")
+	}
+	root, _ := url.Parse(base + "/")
+	u, err := url.Parse(raw)
+	if err != nil {
+		return "", errors.New("outline_response_invalid")
+	}
+	u = root.ResolveReference(u)
+	if u.User != nil || u.Host != root.Host || u.Scheme != root.Scheme {
+		return "", errors.New("outline_response_invalid")
+	}
+	return u.String(), nil
+}
+
+func fingerprint(d document, source string) string {
+	body := ""
+	if _, err := time.Parse(time.RFC3339, d.UpdatedAt); err != nil && d.Revision == "" && d.Text != nil {
+		body = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(*d.Text, "\r\n", "\n"))))
+	}
+	fields := []string{d.UpdatedAt, string(d.Revision), d.Title, d.CollectionID, d.ParentID, source,
+		d.PublishedAt, d.ArchivedAt, d.DeletedAt, body}
+	b, _ := json.Marshal(fields)
+	return fmt.Sprintf("%x", sha256.Sum256(b))
+}
+
+func fileName(title string) string {
+	title = strings.Map(func(r rune) rune {
+		if unicode.IsControl(r) || strings.ContainsRune(`/\:*?"<>|`, r) {
+			return '_'
+		}
+		return r
+	}, strings.TrimSpace(title))
+	if title == "" || title == "." || title == ".." {
+		title = "untitled"
+	}
+	var out strings.Builder
+	for _, r := range title {
+		if out.Len()+len(string(r)) > 200 {
+			break
+		}
+		out.WriteRune(r)
+	}
+	return out.String() + ".md"
+}
+
+func mappedItem(d document, instance identity) (types.FetchedItem, string, error) {
+	if d.Text == nil {
+		return types.FetchedItem{}, "", errors.New("outline_format_unsupported")
+	}
+	if len(*d.Text) > maxResponse {
+		return types.FetchedItem{}, "", errTooLarge
+	}
+	source, err := sourceURL(instance.BaseURL, d.URL)
+	if err != nil {
+		return types.FetchedItem{}, "", err
+	}
+	fp := fingerprint(d, source)
+	title := strings.Join(strings.Fields(d.Title), " ")
+	title = strings.NewReplacer("\\", "\\\\", "#", "\\#", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;").Replace(title)
+	if title == "" {
+		title = "untitled"
+	}
+	created, _ := time.Parse(time.RFC3339, d.CreatedAt)
+	updated, _ := time.Parse(time.RFC3339, d.UpdatedAt)
+	return types.FetchedItem{
+		ExternalID: d.ID, SourceResourceID: d.CollectionID, Title: d.Title,
+		Content: []byte("# " + title + "\n\n" + absoluteLinks(*d.Text, source)), ContentType: "text/markdown",
+		FileName: fileName(d.Title), URL: source, CreatedAt: created, UpdatedAt: updated,
+		Metadata: map[string]string{
+			"channel": types.ChannelOutline, "collection_id": d.CollectionID,
+			"source_title": d.Title, "source_url": source, "source_revision": string(d.Revision),
+			"source_created_at": sourceTimestamp(created), "source_updated_at": sourceTimestamp(updated),
+			"source_fingerprint": fp, "parent_document_id": d.ParentID,
+			"outline_workspace_id": instance.WorkspaceID,
+		},
+	}, fp, nil
+}
+
+func sourceTimestamp(value time.Time) string {
+	if value.IsZero() {
+		return ""
+	}
+	return value.UTC().Format(time.RFC3339Nano)
+}
````

### A.33 `internal/datasource/scheduler.go`

````diff
diff --git a/internal/datasource/scheduler.go b/internal/datasource/scheduler.go
index 196327fe..850c2ff2 100644
--- a/internal/datasource/scheduler.go
+++ b/internal/datasource/scheduler.go
@@ -145,9 +145,22 @@ func (s *Scheduler) triggerSync(dataSourceID string, tenantID uint64) {
 		logger.Infof(ctx, "[Scheduler] skipping sync for ds=%s (not active or not found)", dataSourceID)
 		return
 	}
+	if ds.Type == types.ConnectorTypeOutline {
+		var release func()
+		ctx, release, err = DefaultSyncLocks.Acquire(ctx, tenantID, dataSourceID)
+		if err != nil {
+			return
+		}
+		defer release()
+		// A configuration mutation may have completed while acquiring the lease.
+		ds, err = s.dsRepo.FindByID(ctx, dataSourceID)
+		if err != nil || ds == nil || ds.Status != types.DataSourceStatusActive {
+			return
+		}
+	}
 
 	// Layer 1: prevent overlap with a still-running sync
-	if running, _ := s.syncLogRepo.HasRunningSync(ctx, dataSourceID); running {
+	if running, err := s.syncLogRepo.HasRunningSync(ctx, dataSourceID); running || (err != nil && ds.Type == types.ConnectorTypeOutline) {
 		logger.Infof(ctx, "[Scheduler] skipping sync for ds=%s (previous sync still running)", dataSourceID)
 		return
 	}
````

### A.34 `internal/datasource/sync_lock.go`

````diff
diff --git a/internal/datasource/sync_lock.go b/internal/datasource/sync_lock.go
new file mode 100644
index 00000000..5b5d92d7
--- /dev/null
+++ b/internal/datasource/sync_lock.go
@@ -0,0 +1,120 @@
+package datasource
+
+import (
+	"context"
+	"errors"
+	"fmt"
+	"sync"
+	"time"
+
+	"github.com/google/uuid"
+	"github.com/redis/go-redis/v9"
+)
+
+var ErrSyncRunning = errors.New("datasource_sync_running")
+var DefaultSyncLocks = NewSyncLocks(nil)
+
+// SyncLocks is shared by scheduling, mutations and workers. Lite is single-process.
+type SyncLocks struct {
+	Redis  *redis.Client
+	mu     sync.Mutex
+	owners map[string]string
+}
+
+func NewSyncLocks(rdb *redis.Client) *SyncLocks {
+	return &SyncLocks{Redis: rdb, owners: map[string]string{}}
+}
+
+type leaseKey struct{}
+type syncLease struct {
+	check func(context.Context) error
+}
+
+func CheckSyncLease(ctx context.Context) error {
+	if err := ctx.Err(); err != nil {
+		return err
+	}
+	if lease, ok := ctx.Value(leaseKey{}).(*syncLease); ok {
+		return lease.check(ctx)
+	}
+	return nil
+}
+
+func (l *SyncLocks) Acquire(ctx context.Context, tenant uint64, id string) (context.Context, func(), error) {
+	key, token := fmt.Sprintf("datasource:lease:%d:%s", tenant, id), uuid.NewString()
+	if l.Redis != nil {
+		ok, err := l.Redis.SetNX(ctx, key, token, 120*time.Second).Result()
+		if err != nil {
+			return ctx, nil, err
+		}
+		if !ok {
+			return ctx, nil, ErrSyncRunning
+		}
+	} else {
+		l.mu.Lock()
+		if l.owners == nil {
+			l.owners = map[string]string{}
+		}
+		if l.owners[key] != "" {
+			l.mu.Unlock()
+			return ctx, nil, ErrSyncRunning
+		}
+		l.owners[key] = token
+		l.mu.Unlock()
+	}
+	runCtx, cancel := context.WithCancel(ctx)
+	lease := &syncLease{check: func(ctx context.Context) error {
+		if l.Redis != nil {
+			owner, err := l.Redis.Get(ctx, key).Result()
+			if err != nil || owner != token {
+				cancel()
+				return errors.New("datasource_lease_lost")
+			}
+		} else {
+			l.mu.Lock()
+			ok := l.owners[key] == token
+			l.mu.Unlock()
+			if !ok {
+				cancel()
+				return errors.New("datasource_lease_lost")
+			}
+		}
+		return nil
+	}}
+	done := make(chan struct{})
+	go func() {
+		defer close(done)
+		ticker := time.NewTicker(30 * time.Second)
+		defer ticker.Stop()
+		for {
+			select {
+			case <-runCtx.Done():
+				return
+			case <-ticker.C:
+				if l.Redis != nil {
+					ok, err := l.Redis.Eval(runCtx, `if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("expire",KEYS[1],120) else return 0 end`, []string{key}, token).Int()
+					if err != nil || ok != 1 {
+						cancel()
+						return
+					}
+				}
+			}
+		}
+	}()
+	release := func() {
+		cancel()
+		<-done
+		if l.Redis != nil {
+			ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
+			defer stop()
+			_ = l.Redis.Eval(ctx, `if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end`, []string{key}, token).Err()
+		} else {
+			l.mu.Lock()
+			if l.owners[key] == token {
+				delete(l.owners, key)
+			}
+			l.mu.Unlock()
+		}
+	}
+	return context.WithValue(runCtx, leaseKey{}, lease), release, nil
+}
````

### A.35 `internal/datasource/sync_lock_test.go`

````diff
diff --git a/internal/datasource/sync_lock_test.go b/internal/datasource/sync_lock_test.go
new file mode 100644
index 00000000..ff5cd071
--- /dev/null
+++ b/internal/datasource/sync_lock_test.go
@@ -0,0 +1,70 @@
+package datasource
+
+import (
+	"context"
+	"errors"
+	"testing"
+	"time"
+
+	"github.com/alicebob/miniredis/v2"
+	"github.com/redis/go-redis/v9"
+)
+
+func TestSyncLocksMutualExclusion(t *testing.T) {
+	for _, distributed := range []bool{false, true} {
+		t.Run(map[bool]string{false: "lite", true: "redis"}[distributed], func(t *testing.T) {
+			var rdb *redis.Client
+			if distributed {
+				server := miniredis.RunT(t)
+				rdb = redis.NewClient(&redis.Options{Addr: server.Addr()})
+				defer rdb.Close()
+			}
+			locks := NewSyncLocks(rdb)
+			ctx, release, err := locks.Acquire(context.Background(), 1, "ds")
+			if err != nil {
+				t.Fatal(err)
+			}
+			if err := CheckSyncLease(ctx); err != nil {
+				t.Fatal(err)
+			}
+			if _, _, err := locks.Acquire(context.Background(), 1, "ds"); !errors.Is(err, ErrSyncRunning) {
+				t.Fatal(err)
+			}
+			_, otherRelease, err := locks.Acquire(context.Background(), 2, "ds")
+			if err != nil {
+				t.Fatal("tenants must not share execution locks")
+			}
+			otherRelease()
+			release()
+			_, release, err = locks.Acquire(context.Background(), 1, "ds")
+			if err != nil {
+				t.Fatal(err)
+			}
+			release()
+		})
+	}
+}
+
+func TestLostRedisLeaseDoesNotReleaseNewOwner(t *testing.T) {
+	server := miniredis.RunT(t)
+	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
+	defer rdb.Close()
+	locks := NewSyncLocks(rdb)
+	ctx, release, err := locks.Acquire(context.Background(), 1, "ds")
+	if err != nil {
+		t.Fatal(err)
+	}
+	server.FastForward(121 * time.Second)
+	next, nextRelease, err := locks.Acquire(context.Background(), 1, "ds")
+	if err != nil {
+		t.Fatal(err)
+	}
+	defer nextRelease()
+	if CheckSyncLease(ctx) == nil {
+		t.Fatal("lost ownership was accepted")
+	}
+	release()
+	if err := CheckSyncLease(next); err != nil {
+		t.Fatalf("old release removed new owner: %v", err)
+	}
+}
````

### A.36 `internal/handler/datasource.go`

````diff
diff --git a/internal/handler/datasource.go b/internal/handler/datasource.go
index 2f219d37..35598378 100644
--- a/internal/handler/datasource.go
+++ b/internal/handler/datasource.go
@@ -2,6 +2,9 @@ package handler
 
 import (
 	"context"
+	"encoding/json"
+	"errors"
+	"io"
 	"net/http"
 	"strconv"
 
@@ -98,11 +101,39 @@ func (h *DataSourceHandler) CreateDataSource(c *gin.Context) {
 		return
 	}
 
-	var req types.DataSource
-	if err := c.ShouldBindJSON(&req); err != nil {
+	var request struct {
+		types.DataSource
+		SyncDeletions *bool   `json:"sync_deletions"`
+		SyncSchedule  *string `json:"sync_schedule"`
+	}
+	if err := c.ShouldBindJSON(&request); err != nil {
 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
 		return
 	}
+	req := request.DataSource
+	if request.SyncDeletions != nil {
+		req.SyncDeletions = *request.SyncDeletions
+	}
+	if request.SyncSchedule != nil {
+		req.SyncSchedule = *request.SyncSchedule
+	}
+	if req.Type == types.ConnectorTypeOutline {
+		if req.Status == "" {
+			req.Status = types.DataSourceStatusActive
+		}
+		if req.SyncMode == "" {
+			req.SyncMode = types.SyncModeIncremental
+		}
+		if req.ConflictStrategy == "" {
+			req.ConflictStrategy = "overwrite"
+		}
+		if request.SyncDeletions == nil {
+			req.SyncDeletions = true
+		}
+		if request.SyncSchedule == nil && req.Status != types.DataSourceStatusPaused {
+			req.SyncSchedule = "0 0 */6 * * *"
+		}
+	}
 
 	if _, status, msg := h.getOwnedKnowledgeBase(ctx, tenantID, req.KnowledgeBaseID); status != http.StatusOK {
 		c.JSON(status, gin.H{"error": msg})
@@ -223,6 +254,9 @@ func (h *DataSourceHandler) UpdateDataSource(c *gin.Context) {
 	req.KnowledgeBaseID = existing.KnowledgeBaseID
 	ds, err := h.service.UpdateDataSource(ctx, &req)
 	if err != nil {
+		if respondSyncConflict(c, err) {
+			return
+		}
 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
 		return
 	}
@@ -285,6 +319,9 @@ func (h *DataSourceHandler) ValidateConnection(c *gin.Context) {
 	}
 
 	if err := h.service.ValidateConnection(ctx, id); err != nil {
+		if respondSyncConflict(c, err) {
+			return
+		}
 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
 		return
 	}
@@ -439,8 +476,27 @@ func (h *DataSourceHandler) ManualSync(c *gin.Context) {
 		return
 	}
 
-	syncLog, err := h.service.ManualSync(ctx, id)
+	forceFull, decodeErr := decodeForceFull(c.Request.Body)
+	if decodeErr != nil {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sync options"})
+		return
+	}
+	var syncLog *types.SyncLog
+	var err error
+	if svc, ok := h.service.(interface {
+		ManualSyncWithOptions(context.Context, string, bool) (*types.SyncLog, error)
+	}); ok {
+		syncLog, err = svc.ManualSyncWithOptions(ctx, id, forceFull)
+	} else if forceFull {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "full sync is not supported"})
+		return
+	} else {
+		syncLog, err = h.service.ManualSync(ctx, id)
+	}
 	if err != nil {
+		if respondSyncConflict(c, err) {
+			return
+		}
 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
 		return
 	}
@@ -448,6 +504,45 @@ func (h *DataSourceHandler) ManualSync(c *gin.Context) {
 	c.JSON(http.StatusOK, syncLog)
 }
 
+func decodeForceFull(body io.Reader) (bool, error) {
+	if body == nil {
+		return false, nil
+	}
+	decoder := json.NewDecoder(body)
+	var options map[string]json.RawMessage
+	if err := decoder.Decode(&options); err != nil {
+		if errors.Is(err, io.EOF) {
+			return false, nil
+		}
+		return false, err
+	}
+	if options == nil {
+		return false, errors.New("expected an object")
+	}
+	var extra interface{}
+	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
+		return false, errors.New("unexpected trailing JSON")
+	}
+	var forceFull bool
+	if raw, ok := options["force_full"]; ok {
+		if string(raw) == "null" {
+			return false, errors.New("force_full must be a boolean")
+		}
+		if err := json.Unmarshal(raw, &forceFull); err != nil {
+			return false, err
+		}
+	}
+	return forceFull, nil
+}
+
+func respondSyncConflict(c *gin.Context, err error) bool {
+	if !errors.Is(err, datasource.ErrSyncRunning) {
+		return false
+	}
+	c.JSON(http.StatusConflict, gin.H{"error": "datasource_sync_running", "code": "datasource_sync_running"})
+	return true
+}
+
 // PauseDataSource godoc
 // @Summary Pause data source
 // @Description Pause a data source's scheduled syncs
````

### A.37 `internal/handler/datasource_credentials.go`

````diff
diff --git a/internal/handler/datasource_credentials.go b/internal/handler/datasource_credentials.go
index af222699..7e5d47db 100644
--- a/internal/handler/datasource_credentials.go
+++ b/internal/handler/datasource_credentials.go
@@ -76,6 +76,9 @@ func (h *DataSourceCredentialsHandler) Put(c *gin.Context) {
 	}
 	updated, err := h.service.UpdateDataSourceCredentials(c.Request.Context(), ds.ID, req.Credentials)
 	if err != nil {
+		if respondSyncConflict(c, err) {
+			return
+		}
 		logger.ErrorWithFields(c.Request.Context(), err, map[string]interface{}{
 			"data_source_id": secutils.SanitizeForLog(ds.ID),
 		})
@@ -105,6 +108,9 @@ func (h *DataSourceCredentialsHandler) DeleteField(c *gin.Context) {
 		return
 	}
 	if err := h.service.ClearDataSourceCredentials(c.Request.Context(), ds.ID); err != nil {
+		if respondSyncConflict(c, err) {
+			return
+		}
 		logger.ErrorWithFields(c.Request.Context(), err, map[string]interface{}{
 			"data_source_id": secutils.SanitizeForLog(ds.ID),
 		})
````

### A.38 `internal/handler/datasource_outline_test.go`

````diff
diff --git a/internal/handler/datasource_outline_test.go b/internal/handler/datasource_outline_test.go
new file mode 100644
index 00000000..8edc0e80
--- /dev/null
+++ b/internal/handler/datasource_outline_test.go
@@ -0,0 +1,107 @@
+package handler
+
+import (
+	"context"
+	"github.com/Tencent/WeKnora/internal/datasource"
+	"github.com/Tencent/WeKnora/internal/types"
+	"github.com/Tencent/WeKnora/internal/types/interfaces"
+	"github.com/stretchr/testify/require"
+	"net/http"
+	"net/http/httptest"
+	"strings"
+	"testing"
+)
+
+func TestDecodeForceFull(t *testing.T) {
+	for _, tc := range []struct {
+		body          string
+		full, invalid bool
+	}{
+		{"", false, false}, {"{}", false, false}, {`{"force_full":true}`, true, false},
+		{`{"force_full":false}`, false, false}, {`{"force_full":"true"}`, false, true},
+		{`{"force_full":1}`, false, true}, {`{"force_full":null}`, false, true},
+		{"null", false, true}, {"[]", false, true}, {"{}{}", false, true},
+	} {
+		t.Run(tc.body, func(t *testing.T) {
+			full, err := decodeForceFull(strings.NewReader(tc.body))
+			if (err != nil) != tc.invalid || full != tc.full {
+				t.Fatalf("full=%v err=%v", full, err)
+			}
+		})
+	}
+}
+
+type outlineRequestService struct {
+	interfaces.DataSourceService
+	created *types.DataSource
+	full    bool
+	syncErr error
+}
+
+func (s *outlineRequestService) CreateDataSource(_ context.Context, ds *types.DataSource) (*types.DataSource, error) {
+	s.created = ds
+	return ds, nil
+}
+func (s *outlineRequestService) GetDataSource(context.Context, string) (*types.DataSource, error) {
+	return &types.DataSource{ID: "ds", TenantID: 1, KnowledgeBaseID: "kb"}, nil
+}
+func (s *outlineRequestService) ManualSyncWithOptions(_ context.Context, _ string, full bool) (*types.SyncLog, error) {
+	s.full = full
+	return &types.SyncLog{ID: "log"}, s.syncErr
+}
+
+func TestOutlineCreateDefaultsAndExplicitDisabledOptions(t *testing.T) {
+	for _, tc := range []struct {
+		name, extra string
+		deletion    bool
+		schedule    string
+	}{
+		{"defaults", "", true, "0 0 */6 * * *"},
+		{"disabled", `,"sync_deletions":false,"sync_schedule":""`, false, ""},
+		{"draft", `,"status":"paused","sync_deletions":false`, false, ""},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			svc := &outlineRequestService{}
+			kb := &stubKBServiceForDS{getByID: func(context.Context, string) (*types.KnowledgeBase, error) {
+				return &types.KnowledgeBase{ID: "kb", TenantID: 1}, nil
+			}}
+			h := NewDataSourceHandler(svc, kb)
+			r := newDataSourceTestRouter(h)
+			r.POST("/datasource", h.CreateDataSource)
+			w := httptest.NewRecorder()
+			req := httptest.NewRequest(http.MethodPost, "/datasource", strings.NewReader(`{"type":"outline","knowledge_base_id":"kb"`+tc.extra+`}`))
+			req.Header.Set("Content-Type", "application/json")
+			r.ServeHTTP(w, withDSCtx(req, 1))
+			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
+			require.Equal(t, tc.deletion, svc.created.SyncDeletions)
+			require.Equal(t, tc.schedule, svc.created.SyncSchedule)
+			require.Equal(t, types.SyncModeIncremental, svc.created.SyncMode)
+			require.Equal(t, "overwrite", svc.created.ConflictStrategy)
+		})
+	}
+}
+
+func TestOutlineManualSyncOptionsAndConflict(t *testing.T) {
+	for _, conflict := range []bool{false, true} {
+		svc := &outlineRequestService{}
+		if conflict {
+			svc.syncErr = datasource.ErrSyncRunning
+		}
+		kb := &stubKBServiceForDS{getByID: func(context.Context, string) (*types.KnowledgeBase, error) {
+			return &types.KnowledgeBase{ID: "kb", TenantID: 1}, nil
+		}}
+		h := NewDataSourceHandler(svc, kb)
+		r := newDataSourceTestRouter(h)
+		r.POST("/datasource/:id/sync", h.ManualSync)
+		w := httptest.NewRecorder()
+		req := httptest.NewRequest(http.MethodPost, "/datasource/ds/sync", strings.NewReader(`{"force_full":true}`))
+		r.ServeHTTP(w, withDSCtx(req, 1))
+		require.True(t, svc.full)
+		if conflict {
+			require.Equal(t, http.StatusConflict, w.Code)
+			require.Contains(t, w.Body.String(), "datasource_sync_running")
+		} else {
+			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
+		}
+	}
+}
````

### A.39 `internal/types/datasource.go`

````diff
diff --git a/internal/types/datasource.go b/internal/types/datasource.go
index 5130b931..b8e47732 100644
--- a/internal/types/datasource.go
+++ b/internal/types/datasource.go
@@ -28,6 +28,7 @@ const (
 	ConnectorTypeNotion      = "notion"
 	ConnectorTypeConfluence  = "confluence"
 	ConnectorTypeYuque       = "yuque"
+	ConnectorTypeOutline     = "outline"
 	ConnectorTypeGitHub      = "github"
 	ConnectorTypeGoogleDrive = "google_drive"
 	ConnectorTypeOneDrive    = "onedrive"
@@ -235,6 +236,7 @@ type DataSourceConfig struct {
 	// ingesting an image into a KB without VLM is rejected, so image extraction is
 	// skipped when this is false.
 	MultimodalEnabled bool `json:"-"`
+	SyncDeletions     bool `json:"-"`
 }
 
 // HasCredentials reports whether the credentials map carries any value at
@@ -252,6 +254,10 @@ func (d DataSourceConfig) HasConfiguredCredentials(connectorType string) bool {
 		return false
 	}
 	switch connectorType {
+	case ConnectorTypeOutline:
+		key, _ := d.Credentials["api_key"].(string)
+		base, _ := d.Credentials["base_url"].(string)
+		return strings.TrimSpace(key) != "" && strings.TrimSpace(base) != ""
 	case ConnectorTypeRSS:
 		raw, ok := d.Credentials["auth_headers"]
 		if !ok {
@@ -416,6 +422,7 @@ type SyncCursor struct {
 
 // SyncResult summarizes the outcome of a sync operation
 type SyncResult struct {
+	Metrics map[string]int64 `json:"metrics,omitempty"`
 	// Total items processed
 	Total int `json:"total"`
 
````

### A.40 `internal/types/interfaces/knowledge.go`

````diff
diff --git a/internal/types/interfaces/knowledge.go b/internal/types/interfaces/knowledge.go
index 88ceba3f..4909133a 100644
--- a/internal/types/interfaces/knowledge.go
+++ b/internal/types/interfaces/knowledge.go
@@ -358,3 +358,12 @@ type KnowledgeRepository interface {
 	// DeleteKnowledgeTagRelations deletes all tag relations for a knowledge entry.
 	DeleteKnowledgeTagRelations(ctx context.Context, knowledgeID string) error
 }
+
+// DataSourceKnowledgeMetadataReader reads only live source identities and
+// metadata for reconciliation. It is optional so existing repository fakes and
+// third-party implementations remain source-compatible.
+type DataSourceKnowledgeMetadataReader interface {
+	ListDataSourceKnowledgeMetadata(
+		ctx context.Context, tenantID uint64, kbID, dataSourceID, afterID string, limit int,
+	) ([]*types.Knowledge, error)
+}
````

### A.41 `internal/types/knowledge.go`

````diff
diff --git a/internal/types/knowledge.go b/internal/types/knowledge.go
index 6253423a..0f81f4e9 100644
--- a/internal/types/knowledge.go
+++ b/internal/types/knowledge.go
@@ -34,8 +34,9 @@ const (
 	ChannelIM               = "im"                // Generic IM channel
 	ChannelNotion           = "notion"            // Notion
 	ChannelYuque            = "yuque"             // Yuque (语雀)
-	ChannelRSS              = "rss"               // RSS / Atom feed
-	ChannelIMA              = "ima"               // Tencent IMA (ima.qq.com)
+	ChannelOutline          = "outline"
+	ChannelRSS              = "rss" // RSS / Atom feed
+	ChannelIMA              = "ima" // Tencent IMA (ima.qq.com)
 )
 
 // Knowledge parse status constants
@@ -482,6 +483,8 @@ func (k *Knowledge) SetProcessOverrides(o *KnowledgeProcessOverrides) error {
 
 // KnowledgeCheckParams defines parameters used to check if knowledge already exists.
 type KnowledgeCheckParams struct {
+	DataSourceID string
+	ExternalID   string
 	// File parameters
 	FileName string
 	// FileType scopes file-hash deduplication; callers checking file uploads should set it.
````

