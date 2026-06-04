# 当前实现状态

## 1. 已落地内容

本轮完成 Phase 1 foundation，目标是建立可运行、可测试、可继续扩展的控制面骨架。

已实现：

- Go HTTP API 服务。
- SQLite migration 自动执行。
- 嵌入式 Web 管理后台登录页、仪表盘和基础写操作表单。
- 管理后台已参考 zashboard 的模块化控制台信息架构重构为桌面侧边导航、移动底部 Dock、概览/接入/节点/身份/策略/流量/运维模块视图；桌面侧边导航和移动 Dock 已补齐统一模块符号，写操作表单收敛到对应模块内，避免单页堆叠导致扫描困难。
- 管理后台桌面侧边导航已进一步补齐轻量产品分组，把高频模块收纳为“日常操作”、策略/流量/运维收纳为“治理与运维”，并在侧栏底部保留真实测试路径提示；静态 UI 契约和浏览器登录 QA 会校验分组、底部提示和侧栏内部无溢出，移动端仍只保留底部 Dock。
- 管理后台桌面侧边导航、产品标识、移动 Dock、工作区标题和概览快捷卡已从中文单字视觉符号升级为共享线性 SVG 模块图标体系，保留隐藏文本 fallback 和稳定 `data-module-icon-key`；静态 UI 契约和浏览器登录 QA 会校验图标渲染、key 顺序、glyph 尺寸、状态差异、视口内收纳和无溢出。
- 管理后台内部内容符号已统一为共享 symbol 单元，覆盖概览指标、快捷卡、洞察面板、抽屉标题、模块面板和各模块工作台，统一 8px 半径、稳定方形尺寸、无渐变背景和父容器内收纳；静态 UI 契约和浏览器登录 QA 会校验覆盖范围、真实渲染尺寸、背景、半径和可视符号无越界。
- 管理后台已新增 Calm Ops 视觉覆盖层，保持现有静态前端架构不引入新框架，同时统一顶部栏、导航、按钮、输入框、卡片、Token 区和节点区的尺寸、边界、焦点态与状态语义；后端已显式提供 `/assets/calm-ops.css` 静态路由，浏览器 QA 会校验新版样式文件返回 200 并实际加载。
- 管理后台登录后的外壳已收敛为单一工作区命令中心：登录页保留品牌顶部栏，进入后台后隐藏全局顶部栏，把连接状态、刷新和退出动作并入当前模块操作区，避免双顶部栏、重复状态提示和首屏内容被挤压；命令中心已进一步改为轻量工具条节奏，使用安静阴影、扁平定位条和更紧凑的控件间距，减少首屏压迫感；浏览器 QA 会校验后台顶部栏不可见、侧边栏占满高度、命令中心贴近首屏、桌面高度不超过紧凑阈值、没有重型阴影且状态/退出控件无溢出。
- 管理后台工作区命令中心已继续收紧为低矮 Shell 工具带：桌面端隐藏常驻描述文案，只保留模块标题、当前重点、主操作、刷新和会话状态，模块摘要与定位入口压缩为薄工具条；浏览器 QA 会校验命令区桌面高度不超过 84px、最多两行、各模块无溢出，避免回退成厚重说明卡。
- 管理后台移动端工作区命令中心已改为手机专属三段式工具条：标题与当前重点、主操作/刷新/会话状态、模块定位条按单列堆叠，隐藏冗余描述文案但保留上下文动作；浏览器 QA 会在 390px 宽度校验命令中心为单列、三行以内节奏、整体高度受控、内部动作和定位条无溢出。
- 管理后台各模块顶部工作台已统一压缩为白底紧凑控制台表面，接入、节点、身份、Token、策略、流量和运维工作台取消各自渐变和厚重内卡，统一标题、符号、状态卡和胶囊高度；静态 UI 契约会防止工作台退回分散视觉层，浏览器 QA 会逐模块校验白底、无渐变、低高度内卡和无视觉溢出。
- 管理后台各模块顶部工作台已继续收敛为更密集的 zashboard 式扫描节奏，工作台内边距、间距、符号和内卡片高度同步下调；浏览器 QA 会拒绝超过 70px 的工作台内卡片，并限制非运维工作台高度不超过 190px、运维交付工作台不超过 250px，避免模块顶部再次变成厚重卡片堆。
- 管理后台浏览器登录验收已补齐 1440x900 宽屏桌面视觉契约，真实登录后会在宽屏下逐模块检查页面横向溢出、命令中心子元素越界、模块定位条横向滚动、内容面板标题溢出、概览控制室列数/高度和顶部工作台滚宽，避免 UI 在宽屏真实体验中重新松散或遮挡。
- 管理后台顶部状态提示已改为带语义状态的反馈胶囊，登录、刷新、保存、复制、配置检查和异常状态会同步不同语义色，并通过 `aria-live` 给出更明确的操作反馈。
- 管理后台顶部刷新动作已补齐按钮级等待态，点击后会临时禁用刷新按钮、显示 `刷新中` 和等待符号，同时把主工作区标记为 `aria-busy`；数据加载完成后恢复原按钮，避免重复点击和“不知道是否在刷新”的不确定感。
- 管理后台顶部退出动作已补齐按钮级等待态，点击后会临时禁用退出按钮、显示 `退出中` 和等待符号，同时把主工作区标记为 `aria-busy`；请求完成后清空前端状态并回到登录页，避免慢网络下误判未响应或重复退出。
- 管理后台导航符号已纳入浏览器 QA，校验桌面 7 个侧边模块、移动 4 个高频 Dock 入口和 3 个“更多”菜单模块都有符号且移动端无横向溢出。
- 管理后台工作区标题已补齐当前模块符号，和侧边导航、移动 Dock 使用同一套视觉锚点；浏览器 QA 会逐模块校验标题符号跟随切换更新。
- 管理后台桌面侧边导航、移动 Dock 和移动“更多”菜单已补齐模块数量徽标，来源、节点、Token、策略、流量和配置就绪状态可在导航层直接扫描；浏览器 QA 会校验桌面/移动徽标数量、填充值和移动端无溢出。
- 管理后台移动端导航已收敛为概览、接入、节点、身份和“更多”5 个底部入口，策略、流量和运维进入“更多”菜单；菜单会覆盖在内容区之上，选择模块后自动收起，处于低频模块时“更多”保持高亮。
- 管理后台各模块创建/导入表单已默认收纳为可展开抽屉，抽屉标题、展开/收起控件和提交控件都补齐统一功能符号，列表、卡片和详情成为模块首屏重点，减少常驻表单压缩内容区。
- 管理后台各模块创建/导入表单抽屉已统一占满模块操作区宽度，避免半宽表单悬空在左侧造成扫读断裂；浏览器 QA 会逐模块检查表单抽屉宽度、标题控件和提交按钮不溢出。
- 管理后台创建/导入表单展开后已升级为真正的右侧侧滑工作台：桌面端固定宽度并以轻量背景压暗承托，移动端避开底部 Dock；同一时间只保留一个表单打开，`Escape` 会收起侧滑层但保留草稿，提交成功后自动收起回到列表，退出登录时会清理页面外壳状态，浏览器 QA 会校验 dialog 语义、草稿保留、提交后收起、无横向溢出和页面状态恢复。
- 管理后台概览页顶部指标卡已补齐模块符号，团队、用户、Token、来源、节点、虚拟节点和策略指标使用统一符号、标签和数字层级；浏览器 QA 会校验 7 个指标符号和值都存在且无视觉溢出。
- 管理后台基础 HTML 转义工具已兼容数值 ID 等非字符串字段，避免真实数据加载到流量、Token 或节点卡片时触发前端渲染异常；浏览器 QA 会在登录后等待完整数据加载，并在加载异常时输出失败接口、页面状态和控制台信息。
- 管理后台概览页已把真实测试闭环升级为 7 步初始化向导，按来源、节点池、虚拟网关、团队 Token、订阅分发、访问策略和发布检查展示准备状态；向导会加载 `/api/delivery/readiness` 的交付收口结果，把“是否能真实测试”前置到概览页，同时提供可测进度、当前建议动作、可点击模块入口和每步提示文案；概览快速定位条同步展示“初始化”和向导符号，避免旧术语干扰；每个闭环步骤都有序号、符号化状态单元和就绪/待补徽标，浏览器 QA 会校验进度、当前动作、7 条提示、定位条标签/符号、序号、符号、状态徽标和无视觉溢出。
- 管理后台概览页首屏已加入任务总控条，直接说明当前是否可以真实测试、还差几步、下一步该点哪里，并把可用订阅、可用节点和虚拟网关状态收敛成 3 个稳定状态单元；总控条的主操作会深链到待补表单或发布配置按钮，浏览器 QA 会校验它无视觉溢出且能直接定位下一步。
- 管理后台概览页首屏总控条已进一步收敛为紧凑控制室节奏：总控区改用轻量阴影和更低高度，3 个核心状态单元保持一行扫描，6 个 KPI 卡片压缩为控制台尺寸；移动端保留三列核心状态与两列 KPI，避免首屏继续堆成大块卡片。浏览器 QA 会校验桌面总控高度、KPI 高度、桌面/移动端列数和无横向溢出，静态契约会防止概览退回重型卡片。
- 管理后台概览页运行洞察区已补齐可操作页脚，上游健康、地区分布、分发发布三块洞察分别可以直达接入来源、节点池和运维发布检查；动作使用安静的符号按钮并保留简短说明，浏览器 QA 会校验 3 个洞察动作、符号、目标覆盖、点击定位和无视觉溢出。
- 管理后台概览页初始化向导已补齐“真实测试动作”入口，直接给出去复制订阅、检查收口和发布配置三步操作；这些动作会定位到 Token 列表、交付收口按钮和发布配置按钮，并同步顶部定位反馈，减少用户在真实测试前反复找入口。
- 管理后台概览页快速入口卡已补齐模块符号、实时数量和就绪徽标，接入来源、节点地区、Token 和发布状态能在概览首屏直接扫描；浏览器 QA 会校验 4 个入口符号顺序、徽标填充值和视觉溢出。
- 管理后台概览页下一步入口已改为结构化动作卡，展示目标模块符号、补齐/可测试状态说明和统一符号主操作按钮；浏览器 QA 会校验动作卡、按钮符号存在且无视觉溢出。
- 管理后台模块标题区已加入随当前视图变化的数据摘要条，进入接入、节点、身份、策略、流量和运维模块时能先看到来源数、地区数、有效 Token、策略限制和流量样本等关键上下文；摘要胶囊会约束标签和值在自身内部省略。
- 管理后台模块标题区已加入当前模块快速定位条，可在接入、节点、身份、策略、流量和运维视图内直接跳转到列表、详情或创建抽屉；点击创建入口会自动展开对应表单，减少长页面反复滚动。
- 管理后台模块标题区已补齐随当前模块变化的顶部主操作，概览会按初始化进度给出下一步，接入、节点、身份、策略、流量和运维会直达添加来源、创建网关、签发 Token、创建策略或运维入口；浏览器 QA 会逐模块校验主操作按钮符号、标签、无视觉溢出，并实际点击身份模块主操作确认 Token 表单自动展开。
- 管理后台模块标题区已新增工作区重点提示条，按当前真实数据给出模块状态、风险或就绪程度，并提供一枚上下文动作按钮；概览提示真实测试闭环进度，接入提示来源异常/同步状态，节点提示地区聚合与网关缺口，身份提示团队/成员/Token 分发状态，策略提示访问边界，流量提示样本状态，运维提示发布条件。浏览器 QA 会逐模块校验提示条、符号、动作按钮和无视觉溢出，并实际点击提示动作确认可定位目标区域。
- 管理后台模块跳转动作已补齐落点反馈，顶部主操作、快速定位条和结构化空状态动作会给目标面板或抽屉加短暂高亮；目标是表单抽屉时会自动展开并聚焦第一个可输入字段，浏览器 QA 会确认“添加来源”聚焦名称字段、“签发 Token”聚焦成员字段。
- 管理后台模块跳转动作已补齐持续定位状态，主操作、快速定位条和空状态跳转会同步顶部状态为 `已定位：目标面板`，并把模块快速定位条的当前落点标记为高亮和 `aria-current`；浏览器 QA 会确认添加来源和创建 Token 的定位反馈、语义状态和当前定位条都生效。
- 管理后台表单必填项已补齐统一反馈，提交缺少必填字段时会在顶部状态区提示 `请先补齐：字段名`，同时展开、滚动、高亮当前表单并把焦点留在待补字段，避免只出现浏览器默认校验气泡；浏览器 QA 会确认添加来源缺名称时提示、焦点和高亮都生效。
- 管理后台表单必填项已补齐字段内反馈，缺失字段旁会显示 `请填写字段名`，同步设置 `aria-invalid` 和 `aria-describedby`，用户开始输入后会自动清除提示和错误状态；浏览器 QA 会确认添加来源名称字段的内联提示、可访问性属性和输入后自动恢复。
- 管理后台创建/导入表单抽屉已补齐草稿感知，用户开始输入或变更字段后，抽屉标题会显示 `草稿未提交` 胶囊并保持在标题区内部；提交成功或取消后自动恢复干净状态，减少“我刚才改了什么、有没有保存”的不确定感。
- 管理后台创建/导入表单抽屉已补齐统一取消动作，展开后可一键取消草稿、清除内联校验错误、收起抽屉并在顶部状态区反馈 `已取消：目标表单`，避免用户误以为脏输入还会被保留；浏览器 QA 会确认添加来源抽屉取消后字段清空、错误状态归零且标题区无溢出。
- 管理后台创建/导入表单抽屉已补齐提交中反馈，提交后抽屉进入 `aria-busy` 状态，输入、取消和提交控件会临时禁用，提交按钮会按原动作展示 `添加中`/`创建中`/`导入中`，顶部状态同步显示 `正在提交：目标表单`；请求完成后按钮文字、符号和控件状态会恢复，浏览器 QA 会用拦截请求验证防重复提交和无溢出。
- 管理后台表单控件已补齐统一输入节奏，登录、侧滑抽屉、Token 操作和节点编辑里的输入框、选择框、单位输入组和按钮共享稳定高度、内边距、字段面板背景和移动端单列回退；静态 UI 契约和浏览器 QA 会量测控件高度、分栏、字段间距和溢出，防止再次出现输入框大小不一或按钮挤压。
- 管理后台行内卡片保存动作已补齐等待反馈，来源、团队、成员、虚拟网关、策略和节点编辑保存时会给当前卡片设置 `aria-busy`，临时锁定输入、取消和保存控件，保存按钮显示 `保存中` 并同步顶部状态；完成后恢复编辑入口，浏览器 QA 会拦截来源卡片保存请求验证防重复点击、状态恢复和无视觉溢出。
- 管理后台行内卡片操作已进一步降噪：来源、团队、成员、虚拟网关、策略和节点卡片的编辑、刷新、同步命名、恢复自动等低风险动作使用右对齐紧凑工具条和安静次级按钮，保存动作保留主按钮权重，移动端仍恢复为大触控面积；静态 UI 契约会防止卡片操作重新退化为满宽主色按钮。
- 管理后台次级动作工具条已进一步统一为桌面端符号优先的小型工具按钮，来源编辑/刷新/同步命名、身份/网关/策略编辑和 Token 订阅复制/探测/打开都共享 34px 稳定方形尺寸、8px 半径、可访问名称和悬停提示；移动端仍恢复为完整文字触控按钮。静态契约和浏览器 QA 会校验按钮尺寸、标签收纳、标题提示和无越界。
- 管理后台所有通过 `buttonLabel()` 生成的操作按钮符号已升级为共享线性 SVG action icon 体系，覆盖创建、保存、取消、编辑、刷新、恢复、复制、探测、打开、发布和运维等高频动作；按钮仍保留隐藏文本 fallback 和稳定 `data-action-icon-key`，静态 UI 契约和浏览器登录 QA 会校验 SVG 渲染、fallback、key、可见图标尺寸和无越界。
- 管理后台动作图标体系已禁止真实页面落入泛化 `default` 图标，补齐返回、前进、收起、退出、导入、危险、回滚和重启等操作语义映射；浏览器登录 QA 会统计 `default` 图标数量并在大于 0 时失败，静态契约会校验关键符号映射存在。
- 管理后台地区节点卡片已把编辑和恢复自动停靠到卡片右上角，主体内容预留安全宽度，避免操作按钮占用整行造成卡片过高和大片空白；移动端仍恢复为全宽按钮，静态 UI 契约会检查节点卡片动作不能再撑高桌面卡片。
- 管理后台节点卡片右上角轻量操作已改为符号优先的稳定方形按钮，编辑、取消和恢复自动命名都提供 `aria-label` 与悬停提示，桌面端不再显示会被省略号截断的半截文案；移动端继续展示完整按钮文案和触控面积。
- 管理后台模块快速定位条已补齐功能符号、数量和新增徽标，列表入口直接展示当前规模，创建入口统一展示 `+`，并限制符号、文字和徽标始终收纳在按钮内；浏览器 QA 会校验每个定位按钮都有符号与徽标且无横向溢出。
- 管理后台各模块列表/详情面板标题已统一为模块符号、标题和数量徽标，来源、节点池、虚拟节点、团队、成员、Token、策略、流量和配置发布区都能在面板头部直接扫到所属模块和当前规模或就绪状态；浏览器 QA 会校验 9 个面板符号顺序和标题无溢出。
- 管理后台团队和成员列表已改为卡片列表，团队卡增加类型、状态、成员数和备注状态摘要芯片，成员卡增加所属团队、状态、邮箱和备注状态摘要芯片；名称、团队归属、邮箱、备注和状态继续分层展示，编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证身份卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台身份模块顶部已新增身份分发总控工作台，先展示团队、成员、Token 和可分发订阅 4 个核心状态，并用团队归属、成员承接、订阅分发 3 个阶段卡说明当前分发链路缺口；浏览器 QA 会校验工作台、关键符号、阶段卡和无视觉溢出。
- 管理后台接入来源列表顶部已新增接入来源工作台，先展示来源总数、健康来源、自动刷新和异常来源 4 个关键状态，并按来源类型展示已同步与自动刷新数量；浏览器 QA 会校验工作台、状态符号、类型分布卡和无视觉溢出。
- 管理后台策略列表顶部已新增策略边界工作台，先展示策略总数、启用策略、节点上限和网关边界 4 个关键状态，并按团队、成员、Token 或其他作用域聚合展示启用、限额和网关边界数量；浏览器 QA 会校验工作台、状态符号、作用域分布卡和无视觉溢出。
- 管理后台上游来源已改为卡片列表，来源名称下方增加类型、前缀、刷新方式和同步状态摘要芯片，URL、默认标签、上次同步和异常状态继续分层展示；编辑、刷新和同步命名动作已统一为符号按钮，浏览器 QA 会验证来源卡片数量、摘要芯片、动作符号、编辑入口和卡片无视觉溢出。
- 管理后台虚拟网关列表已改为卡片列表，名称下方增加监听、策略、状态和标签筛选范围摘要芯片，监听协议/端口、策略、状态和标签选择器继续分层展示；编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证虚拟网关卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台虚拟网关列表顶部已新增虚拟网关工作台，先展示网关总数、启用网关、监听端口和标签筛选 4 个关键状态，并按出口策略聚合展示启用、端口和标签筛选数量；浏览器 QA 会校验工作台、状态符号、策略分布卡和无视觉溢出。
- 管理后台策略列表已改为卡片列表，策略名称下方增加作用域、节点上限、状态和虚拟网关范围摘要芯片，标签限制、允许虚拟网关、最大节点数和状态继续分层展示；编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证策略卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台 Token 卡片已在标题下增加状态、归属成员、到期状态和额度状态摘要芯片，订阅可用性可以先扫读；额度、订阅地址和 Token 操作小节已补齐统一符号标题与说明胶囊；同时会把默认订阅、Mihomo 和 sing-box 订阅地址拆成可复制订阅卡，长 URL 会在卡片内换行收束，并用 `URI · 通用`、`YAML · Mihomo`、`JSON · sing-box` 标明客户端格式画像，同时显示“当前访问域名/配置域名/相对地址”的地址来源；浏览器 QA 会校验摘要芯片、小节符号、订阅卡数量、格式画像、地址来源和无视觉溢出。
- 管理后台 Token 列表顶部已新增订阅分发工作台，先展示可用 Token、启用状态、订阅地址数量、地址来源和通用 URI/Mihomo YAML/sing-box JSON 三种客户端格式，再进入单个 Token 卡片复制、打开或探测订阅；浏览器 QA 会校验工作台、三种格式符号和无视觉溢出。
- 管理后台 Token 订阅分发工作台已补齐首屏快捷动作，用户可以直接定位订阅卡、复制 Mihomo 订阅地址或打开 Mihomo 订阅内容；定位动作会高亮首个订阅卡并同步 `已定位：订阅地址` 状态，复制和打开动作复用当前访问域名生成的真实订阅地址，浏览器 QA 会校验动作条数量、`订/米/↗` 符号、跳转能力和无视觉溢出。
- 管理后台 Token 订阅卡除复制外已提供“打开”入口，可直接在新标签页预览或下载默认订阅、Mihomo 和 sing-box 订阅内容；订阅操作已改为复制优先、打开次之，并使用紧凑符号按钮，浏览器 QA 会校验入口数量、操作符号、格式画像、地址来源和无视觉溢出。
- 管理后台 Token 订阅卡已加入一键“探测”入口，后台会复用订阅路由做站内可达性验证，返回状态码、内容大小和检查时间，不对任意外部地址发起请求；探测中、可访问、不可用和失败状态会同时更新按钮与顶部状态提示，并在订阅卡元信息中保留最近探测结果，浏览器 QA 会实际点击探测入口并确认反馈可见。
- 管理后台 Token 订阅卡已补齐通用、Mihomo 和 sing-box 类型徽标，浏览器 QA 会校验徽标数量和无视觉溢出。
- 管理后台 Token 卡片已加入额度使用率进度条，和流量模块使用同一套状态色；浏览器 QA 会校验每张 Token 卡都有额度条且无视觉溢出。
- 管理后台 Token 续期、加额、停用、恢复和重置订阅地址已补齐卡片级等待反馈，操作中会给当前 Token 卡设置 `aria-busy`，锁定续期/加额输入和状态控制按钮，并让触发按钮显示 `续期中` 等动作态；完成后恢复原按钮和输入状态，浏览器 QA 会拦截续期请求验证防重复点击、状态恢复和无视觉溢出。
- 管理后台复制订阅地址后会在对应按钮上显示“已复制”并同步顶部状态，浏览器 QA 会实际点击复制入口确认反馈可见。
- 管理后台创建或重置 Token 后的订阅结果框已复用同一套订阅卡样式，并补齐符号标题和格式数量胶囊，避免结果区长 URL 回到旧式两列表格布局；浏览器 QA 会构造结果预览并验证标题符号、数量胶囊和无视觉溢出。
- 管理后台 Token 续期、加额和状态控制已整理为分区动作面板，续期/加额保留可见字段标签与单位，恢复、重置订阅和撤销收纳为同一状态控制区，避免底部按钮条横向挤压；浏览器 QA 会校验每张 Token 卡都有两组带标签的操作字段、状态控制区、动作符号和无横向滚动溢出。
- 管理后台 Token 撤销和重置订阅地址已改为站内危险操作确认框，展示明确标题、影响说明、取消/确认动作，并支持取消按钮、遮罩点击和 `Escape` 退出；用户取消时会给出“已取消”状态反馈，避免误停用伙伴订阅或误使旧订阅地址失效。
- 管理后台运维模块已把交付收口、配置检查、发布、回滚和重启拆为带符号和状态胶囊的操作卡；配置/收口结果使用带结果符号、状态胶囊和结构化字段的结果卡展示 Hash、入站、出口、上游、用户和下一步动作等关键字段。运维操作执行中会给当前操作卡设置 `aria-busy`、触发按钮展示 `检查配置中` 等进行中文案，并临时锁定同组发布/回滚/重启等高风险按钮，完成后恢复原状态；浏览器 QA 会拦截配置检查请求验证等待态、整组锁定、恢复和无视觉溢出。
- 管理后台空状态已统一为带模块符号、标题、提示文案和下一步动作的结构化空状态，来源、节点、身份、策略、Token、流量图和出口摘要在无数据时不会只剩一行“暂无数据”，并可直接跳转到对应创建、导入、签发或运维入口；浏览器 QA 会验证流量模块空状态动作跳转、按钮符号和无视觉溢出。
- 管理后台 UI 已补齐现代视觉基线层，命令中心、内容面板、概览 KPI、模块工作台和普通信息卡统一使用白色主表面、浅灰内卡片、共享边框和克制阴影；浏览器 QA 会采样真实渲染后的命令中心、面板头、工作台、内卡片和 KPI 卡样式，避免后续模块各自引入不一致的卡片层级或过重阴影。
- 管理后台 UI 已把内容面板和 raised 表面的阴影继续压低为 zashboard 式 quiet elevation：内容面板使用低扩散面板阴影，raised 表面只保留轻微层级，hover 阴影也同步降噪；静态契约和浏览器 QA 会同时校验 token 与真实渲染样式，禁止回退到 `14px 38px` 或 `10px 26px` 这类厚重卡片阴影。
- 管理后台内容面板标题栏已压缩为 zashboard 式低矮列表栏，标题区高度降到 48px，图标、标题、数量和说明在桌面同一行截断显示，移动端隐藏辅助说明，避免每个模块列表头继续像大卡片一样占据首屏；静态契约和浏览器 QA 会量测真实面板头高度、内边距和无溢出。
- 管理后台桌面侧边栏已收敛为更接近 zashboard 节奏的窄白导航：桌面 rail 固定为 192px，导航行高降到 40px，模块符号降到 24px，底部真实测试提示同步降噪；静态契约和浏览器 QA 会校验白底、宽度、行高、符号尺寸和无溢出。
- 管理后台移动端已补齐底部 Dock 安全区契约，工作区底部留白、更多菜单、表单侧滑层和节点详情抽屉统一使用同一套 Dock 高度/间距变量；浏览器 QA 会在 390px 移动视口滚动到底，确认最后一个内容块不会被底部 Dock 遮挡。
- 管理后台页面截图验收已升级为多视口矩阵，默认采集 1440x900、1280x720 和 390x844 三档桌面/移动截图，并纳入静态契约，便于每次 UI 改动后人工复核无遮挡和无溢出。
- 管理后台已做基础视觉统一：统一顶部栏、卡片、表单、按钮、表格、状态徽标和响应式栅格，输入控件高度保持一致；登录、顶部退出、刷新和概览下一步动作都使用统一符号按钮，节点卡片长标题、来源名和标签会限制在父卡片内换行/截断，并通过浏览器 QA 检测桌面和移动视图是否发生视觉溢出。
- 节点池地区聚合页顶部已新增节点池工作台，先展示节点匹配、可用节点、地区类别和协议/来源 4 个关键状态，并给出 Top 地区健康卡、可用率进度条和地区占比；浏览器 QA 会校验工作台、状态符号、重点地区卡和无视觉溢出。
- 节点池地区聚合卡已补齐地区符号、可用数量徽标和协议/来源状态小胶囊，地区首屏更接近控制台扫描节奏；浏览器 QA 会校验每个地区卡都有符号、徽标、状态胶囊且无视觉溢出。
- 节点池地区展开后的节点卡片已改为协议符号、标题/来源、状态/协议/命名模式芯片和服务端端点摘要分层展示，编辑/保存/恢复自动操作使用统一符号按钮；长节点名、来源、标签、服务端地址和动作按钮会在卡片内收束，浏览器 QA 会校验每张节点卡都有协议符号、三枚信息芯片、端点摘要、编辑保存符号和两枚动作符号且无视觉溢出。
- 节点池编辑面板已从单个展示名输入升级为带字段标签的安全编辑面板，支持维护展示名、地区覆盖、命名模式和标签；选择自动命名时会恢复来源前缀生成的展示名，同时保留人工维护的地区与标签。API flow 和 Store 单测会验证这些字段可保存、可恢复自动命名且不丢失操作员字段，浏览器 QA 会校验编辑字段完整且无溢出。
- 节点池地区展开页已加入地区概况条，先展示地区符号、本区节点数、可用数、协议数和来源数；浏览器 QA 会用长地区名称验证概况条摘要芯片和标题不溢出。
- 节点池支持在地区聚合页和地区节点卡片页按地区、节点名、协议、来源、标签或服务器搜索，并可一键清空筛选；浏览器 QA 会验证搜索可见、筛选有效、清空可恢复且不产生横向溢出。
- 节点详情已从节点卡片内部改为详情面板：桌面端作为地区节点列表右侧的内容侧栏，移动端保留 Dock 上方抽屉，先展示协议符号、节点名、来源、状态、协议、地区和服务器地址，再用连续定义列表展示字段信息，减少小卡片套小卡片造成的视觉噪音；长 URI 会在面板内换行收束，复制动作保留，复制后按钮显示“已复制”并同步顶部状态，面板可通过关闭按钮收起；浏览器 QA 会实际展开节点详情，校验摘要芯片、轻量字段层级、URI 无溢出、桌面不覆盖节点卡片并点击复制入口。
- 管理员引导账号。
- PBKDF2-SHA256 密码哈希。
- 管理员签名 Cookie 会话。
- 管理 API 登录保护。
- 会话 Cookie 按实际请求协议设置 Secure，支持 HTTPS 反代和 HTTP 局域网调试。
- 团队和用户基础页面支持创建、列表和行内编辑名称、归属、备注及启停状态；Token 支持创建、列表、续期、追加额度、撤销和恢复。
- 策略基础页面支持创建、列表和行内编辑名称、范围、标签限制、虚拟节点限制、最大节点数和启停状态。
- 策略 `allowed_virtual_nodes` 和 `max_nodes` 已按 Token > 成员 > 团队优先级生效，用于限制订阅输出中的可见虚拟节点，并同步约束 sing-box 入站里的用户分配；被策略挡住且没有可用用户的虚拟节点不会生成入站。
- 策略 `include_tags` 和 `exclude_tags` 已按 Token > 成员 > 团队优先级生效，会为匹配 Token 生成带 `auth_user` 的 sing-box route rule，限制该 Token 在虚拟节点下可走的上游出口。
- Token 支持续期、追加额度、撤销和恢复，并同步 gateway account 状态。
- 管理后台 Token 列表支持按行输入自定义续期天数和追加额度 MiB，避免只能使用固定续期/加额步长。
- 创建 Token 时会返回默认、Clash/Mihomo 和 sing-box 三类订阅地址；地址优先基于当前管理后台访问 Host / `X-Forwarded-*` 头生成，管理后台创建结果同步展示，Token 列表会加密恢复并反复展示订阅地址，便于后续复制分发给不同客户端。
- 订阅内容里的 Clash/Mihomo `server` 和 sing-box outbound `server` 会优先使用当前订阅请求 Host 并去掉管理端口，端口仍来自虚拟节点监听端口；请求 Host 不安全时才回退到 `GATEWAY_HOST` 或 `PUBLIC_BASE_URL` 主机名，避免真实客户端拿到 `gateway.example.com` 模板占位域名。
- 管理后台 Token 列表支持重置订阅地址；历史旧 Token 如果没有可恢复订阅密钥，会提示重置后重新获取，新地址生成后旧订阅地址立即失效。
- 上游来源页面创建和列表。
- 管理后台上游来源支持行内编辑名称、类型、URL、前缀、默认标签和刷新间隔；前缀变更会同步刷新自动命名节点。
- subscription 类型上游来源可保存 URL 或 raw content，并可手动刷新导入节点。
- 上游来源 `default_tags` 会在节点导入或刷新时同步到节点标签。
- subscription 来源刷新时，本次订阅中消失的旧节点会标记为 `inactive`。
- subscription 来源 URL 刷新可选先走 Sub-Store 提取模板，只把上游订阅地址转换为节点内容；提取结果仍由 FluxGate 归一化、命名和入库，Sub-Store 失败时会记录日志并回退直连拉取。
- subscription 来源刷新可识别当前已支持 outbound 的 URI 协议列表。
- subscription 来源支持解析 JSON URI 数组和常见 `nodes`/`links`/`subscription`/`raw_content` 包装对象，包装字段内的 URI 列表、Clash YAML、SIP008、sing-box JSON 和 base64 内嵌订阅内容也会归一化。
- JSON 包装订阅支持 `payload`、`result`、`response`、`body`、`text`、`sub` 等常见接口外层字段，不会把这些包装字段误当成节点名称。
- JSON 包装订阅会忽略 `meta`、`pagination`、`errors`、`message`、`traceId` 等常见接口元信息字段，避免把接口文档或分页 URL 误导入为 HTTP 节点。
- JSON 包装订阅会忽略 Clash 风格 `proxy-providers`、`rule-providers`、`proxy-groups`、`health-check` 和 `rules` 中的下载 URL、健康检查 URL 和规则源 URL，避免误导入为 HTTP 节点；其中 `proxy-providers` 内嵌的 `proxy`/`proxies`、`node`/`nodes`、`server`/`servers`、`proxy-list`/`proxies-list`、`nodeList`/`nodesList`、`serverList`/`serversList` 节点清单会继续按结构化节点解析。
- JSON 包装订阅支持 `list`、`records`、`rows`、`entries`、`node`/`nodes`、`proxy`/`proxies`、`server`/`servers`、`nodeList`/`nodesList`、`proxyList`/`proxiesList`、`serverList`/`serversList`、`subscriptionList`/`subscriptionsList`、`urlList`/`urlsList`、`linkList`/`linksList` 及其 snake_case/kebab-case 变体等常见集合字段；这些容器名不会污染缺少 fragment 的节点 URI 展示名。
- JSON 包装订阅字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Data.NodeList`、`Proxy_List` 和 `Payload.LinkList`。
- JSON 包装订阅节点名称字段也兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Name`、`Display_Name` 和 `NodeName`，缺少 fragment 的链接会继续保留真实展示名。
- JSON 包装订阅支持 `shareUrl`、`shareLink`、`subscriptionUrl`、`nodeUrl` 及其 snake_case/kebab-case 变体；对象内带名称的 `subscribeUrl`、`subUrl` 和 `downloadUrl` 分享字段也会按节点链接解析，分享链接缺少 fragment 时会优先使用对象内名称而不是字段名。
- JSON 包装订阅支持解析 `nodes`/`proxies` 中的 Clash 风格结构化节点对象，例如 `type`、`server`、`port`、`cipher`、`password` 和嵌套 `ws-opts`。
- JSON 结构化节点支持常见字段别名，例如 `protocol`/`proto`/`scheme`/`protocolName`/`nodeType`/`serverType`/`protocolType`/`proxyType`/`proxyProtocol`、`host`/`hostname`/`address`/`serverAddress`/`serverHost`/`remoteHost`/`nodeHost`/`endpoint`/`add`、`server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber`、`method`/`encryptMethod`/`encrypt-method`、`pass`/`passwd`/`pwd`/`secret`/`credential`/`accountPassword`、`username`/`userName`/`user_name`/`accountName`/`account`、`id`/`user-id` 和 `displayName`/`nodeName`/`label`/`title`/`remarks`/`remark`；`server`/`host`/`address`/`endpoint` 写成 `host:port` 时会自动拆分端口。
- JSON 结构化节点会归一化常见 camelCase/snake_case/kebab-case 参数，例如 `skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allowInsecure`、`allow_insecure`、`allow-insecure`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`clientFingerprint`、`flowName`、`xtlsFlow`、`networkType`、`transportType`、`tlsEnabled`、`enableTLS`、`overTLS`、`tlsHost`、`tlsServerName`、`tlsSettings`、`tlsOptions`、`pluginOpts`、`wsOpts`、`wsSettings`、`wsHeaders`、`grpcSettings`、`httpSettings`、`h2Settings`、`httpUpgradeSettings`、`quicSettings`、`tcpSettings`、`grpcServiceName`、`privateKeyPath`、`publicKey`、`shortId`、`shortIds`、`spiderX`、`serverNames`、`server_names`、`server-name` 等，并支持 `tls.enabled`、`tls.enable`、`tls.serverName`、`tls.serverNames`、`tls.server_names`、`tls.server-name`、`tls.tls_server_name`、`tls.utls.fingerprint`、`reality.publicKey` 和 `reality.serverNames` 等嵌套字段，避免导入后丢失 TLS、REALITY、插件和传输配置。
- JSON 结构化节点支持通用 `transport` 容器，可从 `transport.type`、`transport.path`、`transport.host`、`transport.headers.Host`、`transport.authority`、`transport.headers.authority`、`transport.headers.:authority`、`transport.serviceName`、`transport.idleTimeout` 等字段生成 WebSocket、HTTP、HTTPUpgrade 或 gRPC 传输参数。
- subscription 来源支持解析 Quantumult X `[server_local]`/`[server_remote]` 常见 SS、SSR、Hysteria2/Hy2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、Trojan、VLESS、VMess、HTTP 和 SOCKS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名，转换为标准 URI。
- Quantumult X `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Quantumult X、Surge、Clash YAML 和 sing-box JSON 的 `tuic-v5`/`tuic5` 类型会按 TUIC 兼容链路归一化为 `tuic://`，保留 UUID、密码、拥塞控制、UDP relay、SNI 和节点名。
- Quantumult X 和 Surge 结构化 VLESS/Trojan 节点支持保留 gRPC transport 和 service name，并同步到 sing-box outbound。
- Quantumult X HTTP/SOCKS 节点支持键值和位置参数两种认证写法。
- subscription 来源支持解析以节点名称为 key 的 JSON URI 对象映射；当 URI 缺少 fragment 时会使用映射 key 或对象内 `name` 作为节点名。
- subscription 来源支持解析裸 VMess JSON 单对象、数组、包装字符串和按名称映射对象，并转换为标准 `vmess://` URI；裸 VMess JSON 支持 `address`/`server`/`serverHost`/`nodeHost`/`endpoint`、`serverPort`/`nodePort`/`portNumber`、`uuid`/`user_id` 等字段别名，会保留 `packetEncoding`/`packet_encoding`/`packet-encoding`、`disable_sni` 和 `fp`/`fingerprint`/`clientFingerprint` 并同步为 sing-box `packet_encoding` 和 uTLS fingerprint。
- 裸 VMess JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Display_Name`、`Server-Port`、`User_ID`、`Packet_Encoding`、`Disable-SNI` 和 `Client-Fingerprint`。
- subscription 来源支持解析 SSD/ShadowsocksD `ssd://` 订阅和裸 SSD JSON，并展开为标准 `ss://` URI；SSD `servers` 支持数组和按名称分组的对象映射，并兼容 SIP008 同款 Shadowsocks 字段别名。
- SIP008 和 SSD JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Version`、`Servers`、`Display_Name`、`Server-Port`、`Encrypt_Method`、`PluginOpts` 和 `PluginOptions`。
- subscription 来源支持解析 Surge `[Proxy]` 代理段中的 SS、SSR、Trojan、VLESS、VMess、Hysteria2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、HTTP/HTTPS、SOCKS、Direct、Reject 和 DNS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名，转换为标准 URI。
- Surge `[Proxy]` `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- subscription 来源支持解析 Clash YAML `proxies` 中的常见 SS/Trojan/VLESS/VMess/Hysteria2/TUIC/Juicity/Hysteria/HTTP/SOCKS/AnyTLS/ShadowTLS/Naive/SSH/WireGuard 节点，以及 Direct/Reject 变体/DNS 内置出站；Clash YAML 结构化 `type` 兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 别名。
- Clash YAML `type: ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Clash YAML 解析支持嵌套 `proxy`/`proxies`/`proxy-list`/`proxies-list`/`node`/`nodes`/`node-list`/`server`/`servers`/`server-list` 列表，并兼容 `proxyList`、`nodeList`、`serverList` 等 camelCase/PascalCase 写法，可兼容带内嵌 provider 节点清单的订阅结构。
- Clash YAML 解析支持 `proxy: [{ ... }]`、`proxies: [{ ... }]`、`proxy-list: [{ ... }]`、`proxies-list: [{ ... }]`、`nodes: [{ ... }]`、`servers: [{ ... }]` 和 `node-list: [{ ... }]` 等内联节点数组，并兼容 camelCase/PascalCase 容器名，可兼容 provider 或顶层节点的紧凑写法。
- Clash YAML 解析支持 `proxies: &anchor` 这类带 YAML anchor 的块状节点列表。
- Clash YAML 解析支持 `- &anchor { ... }` 和 `proxies: [&anchor { ... }]` 这类节点条目级 anchor 的内联写法。
- Clash YAML 解析支持常见块状和内联 `ws-opts`、`grpc-opts` 嵌套写法，能保留 WebSocket path/host 和 gRPC service name。
- Clash YAML 解析会归一化常见 camelCase 参数，例如 `skipCertVerify`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`wsPath`、`wsHost`、`wsOpts`、`maxEarlyData`、`earlyDataHeaderName`、`grpcServiceName` 等，块状和内联节点都会保留 TLS 与传输配置。
- Clash YAML 解析支持行内和块状数组标量，例如 `alpn: [h2, http/1.1]`、`alpn: ... - h2` 和 WireGuard `allowed-ips`/`reserved` 列表。
- subscription 来源支持解析 SIP008 Shadowsocks 订阅，`servers` 支持数组和按名称分组的对象映射；SIP008 Shadowsocks 节点可识别 `address`/`serverAddress`/`nodeHost`/`endpoint`、`portNumber`/`nodePort`、`encryptMethod`/`security` 和 `pass`/`passwd` 等字段别名。
- SIP008、Clash YAML 和 sing-box JSON 的 Shadowsocks 节点会保留 SIP003 `plugin`、`plugin_opts` 和 `network` 参数，并同步到 sing-box outbound。
- `shadowsocks://` URI 会归一化为标准 `ss://` URI，避免长 scheme 导入后无法进入 Shadowsocks outbound 转换链路。
- `ssr://` ShadowsocksR URI 会解码为标准 `ss://` URI，保留服务端、端口、加密方法、密码和 remarks 节点名；SSR 专有 protocol/obfs 参数暂按 Shadowsocks 兼容链路导入。
- `http+tls://` 和 `http-tls://` 裸 URI 会归一化为标准 `https://`，复用现有 HTTP outbound TLS 配置生成链路。
- `any-tls://`、`shadow-tls://`、`naive-https://` 和 `naive-quic://` 裸 URI 会归一化为标准 `anytls://`、`shadowtls://`、`naive+https://` 和 `naive+quic://`，和结构化订阅解析保持一致。
- `hy2://` 和 `wg://` 裸 URI 会归一化为标准 `hysteria2://` 和 `wireguard://`，让导入后的协议字段和配置生成链路保持一致。
- `trojan-go://` 裸 URI 会归一化为标准 `trojan://`，复用现有 Trojan TLS、REALITY 和传输参数转换链路。
- `vmess-aead://` 裸 URI 会归一化为标准 `vmess://`，复用现有 VMess TLS、WebSocket、gRPC、HTTP 和 packet encoding 转换链路。
- `tuic-v5://` 和 `tuic5://` 裸 URI 会归一化为标准 `tuic://`，复用现有 TUIC outbound 转换链路；手动保存 `tuic-v5`/`tuic5` 协议节点时配置生成侧也会按 TUIC 输出。
- subscription 来源支持解析 sing-box JSON `outbounds` 中的常见 Shadowsocks/Trojan/VLESS/VMess/Hysteria2/TUIC/Juicity/AnyTLS/ShadowTLS/Naive/Hysteria/HTTP/SOCKS/SSH/WireGuard/Tor 节点，并兼容 `trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 结构化 `type` 别名。
- subscription 来源支持解析 sing-box JSON 顶层 `endpoints` 中的 WireGuard endpoint 模型，并转换为当前 Phase 1 WireGuard outbound 兼容 URI。
- sing-box JSON 顶层可兼容单数 `outbound`/`endpoint` 和 `outboundList`/`outboundsList`、`endpointList`/`endpointsList` 写法；也兼容顶层直接给出 outbound/endpoint 数组、单个对象或按名称映射对象；单对象、数组和对象映射都会进入同一归一化链路。
- sing-box JSON 顶层、outbound、endpoint、TLS、uTLS、REALITY、transport、headers 和 WireGuard peer 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Server-Port`、`ServerPort`、`ServerName`、`Disable-SNI`、`ServiceName`、`PermitWithoutStream`、`PrivateKey`、`AllowedIPs`。
- sing-box JSON `outbounds` 和 `endpoints` 支持数组、单个对象、按名称分组的对象映射和分组数组映射；映射对象缺少 `tag`/`name` 时会使用映射键或分组序号作为节点名。
- sing-box JSON transport 的 `host`、`authority`、`headers.Host`、`headers.authority` 和 `headers.:authority` 支持字符串或数组写法，WebSocket、HTTP、HTTPUpgrade 和 VMess WebSocket/HTTP/HTTPUpgrade transport 会保留为逗号分隔 Host。
- subscription 来源支持解析 V2Ray/Xray JSON `outbounds` 中的 VMess、VLESS、Trojan、Shadowsocks、HTTP、SOCKS、freedom、blackhole、reject 变体和 DNS 出站，并兼容 `trojan-go`、`vmess-aead`、`http+tls`、`http-tls`、`socks4`、`socks4a`、`socks5`、`socks5h` 协议别名，保留常见 TLS、REALITY、WebSocket、gRPC、QUIC、HTTP/H2 和 HTTPUpgrade 传输参数；REALITY 兼容 `serverName`/`serverNames`、`shortId`/`shortIds` 和 hyphen/snake-case 别名。
- V2Ray/Xray JSON `streamSettings`、TLS/REALITY 设置和各 transport 设置块支持 camelCase、snake_case 和 hyphen-case 容器字段别名。
- V2Ray/Xray JSON 顶层、outbound、settings、vnext/server、user/account、streamSettings、TLS、uTLS、REALITY、WebSocket、gRPC、HTTP、HTTPUpgrade 和 TCP header 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Settings`、`VNext`、`Server-Port`、`StreamSettings`、`TLSSettings`、`ServerName`、`Disable-SNI`、`WSSettings`、`MaxEarlyData`、`GrpcSettings` 和 `PermitWithoutStream`。
- V2Ray/Xray JSON Shadowsocks outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`security`/`encryption`/`encryptMethod` 字段别名，并保留 SIP003 `plugin`、`plugin_opts`、`plugin_options`、`network` 和 `protocol` 标志，结构化导入后会继续进入 Shadowsocks 转换链路。
- V2Ray/Xray JSON 普通 TLS 会兼容 `serverName`/`serverNames` 以及 hyphen/snake-case SNI 别名，数组写法会选取首个有效值同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会兼容 `allowInsecure`、`skip-cert-verify`、`insecure` 和 `disable_sni` 写法，并同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会保留 `fingerprint`、`clientFingerprint` 或 `utls.fingerprint` 客户端指纹，并同步为 sing-box uTLS fingerprint。
- V2Ray/Xray JSON 兼容旧配置中的单个 `outbound`、`outboundDetour`、复数 `outboundDetours` 和 `outboundList`/`outboundsList` 出站入口，会和现代 `outbounds` 一起归一化导入。
- V2Ray/Xray JSON 也兼容顶层直接给出 outbound 数组、单个 outbound 对象或按名称映射的 outbound 对象，映射 key 写成 `host:port` 时会同步补齐节点服务端、端口和默认节点名。
- V2Ray/Xray JSON `wsSettings.host`、`wsSettings.headers.Host` 和 `authority` 支持字符串或数组写法，数组会保留为逗号分隔的 WebSocket Host。
- V2Ray/Xray JSON `wsSettings` 支持 `maxEarlyData`、`max_early_data`、`max-early-data`、`earlyDataHeaderName`、`early_data_header_name` 和 `early-data-header-name` WebSocket early data 别名。
- V2Ray/Xray JSON `tcpSettings.header.type=http` 的 HTTP 伪装 Host、path 和 method 会保留为 HTTP transport 参数。
- V2Ray/Xray JSON `httpSettings`/`h2Settings` 支持从 `host`、`headers.Host` 和 `authority` 读取 HTTP/H2 transport Host，并支持数组写法的 path。
- V2Ray/Xray JSON `httpupgradeSettings.host` 支持字符串或数组写法，也可从 `headers.Host` 读取 Host 列表。
- V2Ray/Xray JSON `grpcSettings` 的 `idle_timeout`、`health_check_timeout` 和 `permit_without_stream` 会保留为 gRPC transport keepalive 参数。
- V2Ray/Xray JSON `grpcSettings.multiMode` 会保留为 sing-box gRPC transport 的 `multi_mode` 参数。
- V2Ray/Xray JSON `grpcSettings` 支持 `service-name`、`idle-timeout`、`health-check-timeout`、`permit-without-stream` 和 `multi-mode` 这类 hyphen 别名。
- V2Ray/Xray JSON 存在 `quicSettings`/`quic_settings`/`quic-settings` 时会保留为 QUIC transport。
- V2Ray/Xray JSON VMess/VLESS `vnext` endpoint 支持 `address`、`server`、`host`、`add`、`serverAddress`、`serverHost`、`remoteHost`、`nodeHost`、`endpoint` 和 `server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber` 别名；VMess/VLESS user 支持 `id`、`uuid`、`userId`、`user_id` 和 `user-id` 认证字段，VMess user 支持 `alter-id`/`aid` 和 `cipher` 别名，VMess/VLESS user 支持 `flow`/`flowName`/`xtlsFlow` 等 flow 别名，并支持 `packetEncoding`/`packet_encoding`/`packet-encoding` 同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON VMess/VLESS `vnext.users` 支持以 UUID 为 key 的对象映射；映射值可补充 `email`/`name`、`security`、`flow` 等字段，缺少 `id` 时会使用映射 key 作为用户 ID。
- V2Ray/Xray JSON `vnext` 和 `servers` 支持以服务端域名或 IP 为 key 的对象映射；对象缺少 `address`/`server` 时会使用映射 key 作为节点服务端，映射 key 写成 `host:port` 或 `[ipv6]:port` 时会同步拆出端口；VMess/VLESS 的 `vnext` 映射值也支持直接写用户对象、用户数组、UUID 标量、UUID 数组或账号名到 UUID/用户对象的映射，Trojan/Shadowsocks 的 `servers` 映射值支持直接写密码字符串或密码数组。
- V2Ray/Xray JSON VMess、VLESS、Trojan 和 Shadowsocks 支持缺省 `settings.vnext`/`settings.servers` 的扁平 outbound 写法；服务端、端口、认证和加密字段可直接写在 outbound 顶层或 `settings` 顶层。
- Clash YAML 和 sing-box JSON VMess/VLESS 节点支持 `packetEncoding`、`packet_encoding` 和 `packet-encoding` 写法，并同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON Trojan/HTTP/SOCKS outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`userName`/`accountName`、`users`、`accounts` 数组、`accounts` 用户名到密码或对象映射和 server 层认证写法；HTTP/SOCKS 的 `servers` 端点映射值也支持直接写账号名到密码或账号对象的映射，并兼容 `user:password` 标量或标量数组。
- subscription 来源支持按 `refresh_interval_minutes` 定时同步；后台调度默认每 60 秒检查一批到期来源。
- 上游来源自动前缀。
- 重复来源名前缀自动编号，例如 `[机场A]`、`[机场A-2]`。
- 节点 URI 页面批量导入。
- 节点 `raw_name`、`display_name`、`name_mode`。
- 管理 API 和管理后台可查看单个节点详情，包括 URI、服务器、端口、Hash、状态、标签和时间信息。
- 管理后台所有时间字段统一按东八区展示，后端仍以 UTC 作为存储和计算基准。
- 管理后台节点池默认按地区聚合展示，缺失地区的节点归入“其他”；进入地区后展示节点卡片，再点击卡片查看节点详情。
- 管理后台节点池地区返回和筛选清空控件已统一为符号按钮，和节点卡片操作区保持一致。
- 节点导入、读取和历史数据迁移会识别常见地区并统一带国旗展示；中国香港/中国台湾统一归为 `🇨🇳中国|香港`、`🇨🇳中国|台湾`，`香港-美国` 这类路径命名会按右侧目的地归为 `🇺🇸美国`。
- 单节点手动改名保护。
- 单节点恢复自动命名。
- 管理后台节点池支持行内编辑节点展示名，并可恢复来源前缀驱动的自动命名。
- 虚拟节点页面支持创建、列表和行内编辑名称、监听协议、监听端口、标签选择器、出口策略和启停状态。
- 虚拟节点 `tag_selector` 可按上游节点标签生成专属 sing-box selector，并生成入站到该 selector 的 route rule；无匹配上游时会路由到 `block`。
- Clash/Mihomo 订阅生成。
- sing-box 客户端订阅生成。
- sing-box 服务端配置生成骨架。
- sing-box config check API 和管理后台检查入口，可返回配置 hash、入站/出站/用户摘要。
- sing-box config publish API 和管理后台发布入口，可写入当前配置并保存上一版文件。
- sing-box config rollback API 和管理后台回滚入口，可恢复上一版配置文件。
- sing-box config publish/rollback API 会返回 `restart_required`，管理后台会提示发布或回滚后需要重启 sing-box。
- sing-box restart API 和管理后台重启入口已落地；发布/回滚可在 `SING_BOX_AUTO_RESTART=true` 时自动调用受配置保护的重启命令。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2/Hysteria/TUIC/Juicity/AnyTLS/ShadowTLS/Naive/HTTP/SOCKS/SSH/WireGuard/Tor 上游节点会转换为 sing-box outbound，并通过默认 selector 承接网关出口。
- 配置生成侧兼容导入层已支持的 URI 协议别名，包括 `shadowsocks`、`trojan-go`、`vmess-aead`、`tuic-v5`、`tuic5`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https` 和 `naive-quic`，手动保存这些协议名时也会归一为对应 sing-box outbound。
- Direct URI、`freedom://` URI 和 sing-box JSON `direct` outbound 可导入为直连上游 outbound，并进入默认上游 selector 和 V2Ray outbound stats 列表；配置生成侧也兼容手动保存的 `freedom` 协议节点。
- Block URI、`blackhole://` URI、`reject://` URI、Clash `reject`/`reject-drop` 和 sing-box JSON `block` outbound 可导入为内部拦截 outbound；为避免误承接普通代理流量，Block outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表；配置生成侧也兼容手动保存的 `blackhole`/`reject`/`reject-drop`/`reject-no-drop`/`reject-tinygif` 协议节点。
- `reject-drop://`、`reject-no-drop://` 和 `reject-tinygif://` 裸 URI 会归一化为标准 `block://`，和 Clash 结构化 reject 变体保持一致。
- DNS URI 和 sing-box JSON `dns` outbound 可导入为内部 DNS outbound；为避免误承接普通代理流量，DNS outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan/VMess 的 `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 uTLS fingerprint TLS 参数，并同步到 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS WebSocket 的 `type=ws`、`path` 和 `host` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `wsPath` 和 `wsHost` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan WebSocket 的 `max_early_data` 和 `early_data_header_name` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `maxEarlyData` 和 `earlyDataHeaderName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS gRPC 的 `type=grpc` 和 `service_name` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `serviceName` 和 `grpcServiceName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan gRPC 的 `idle_timeout`、`ping_timeout`、`permit_without_stream` 和 `multi_mode` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `grpcIdleTimeout`、`grpcPingTimeout`、`permitWithoutStream`、`multiMode` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan QUIC transport，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTP transport 的 `host`、`path`、`method`、`idle_timeout` 和 `ping_timeout` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `httpHost`、`httpPath`、`httpMethod`、`httpIdleTimeout` 和 `httpPingTimeout` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTPUpgrade transport 的 `host` 和 `path` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `httpUpgradeHost` 和 `httpUpgradePath` 查询别名。
- VLESS、Trojan 和 VMess URI 生成 sing-box outbound 时，WebSocket、HTTP/H2 和 HTTPUpgrade transport Host 兼容 `authority`、`:authority`、`headers.Host`、`headers.:authority`、`headerHost` 和协议专属 Host 查询别名。
- VLESS URI 可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，Trojan URI 可从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- VLESS 和 Trojan URI 可通过 `tls=tls`、`tls=true`、`tlsEnabled`、`enableTLS` 或 `overTLS` 查询参数显式启用普通 TLS，兼容未写 `security=tls` 的订阅写法。
- VLESS 和 Trojan URI 的 SNI 兼容 `sni`、`servername`、`server_name`、`server-name`、`serverName`、`tlsHost`、`tls_host`、`tls-host`、`tlsServerName`、`tls_server_name` 和 `tls-server-name` 查询别名。
- VLESS 和 Trojan URI 的跳过证书校验兼容 `insecure`、`skip-cert-verify`、`skip_cert_verify`、`skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allow-insecure`、`allowInsecure`、`tlsAllowInsecure` 和 `tlsSkipVerify` 查询别名。
- VLESS 和 Trojan URI 的禁用 SNI 兼容 `disable_sni`、`disable-sni`、`disableSNI`、`disableSni`、`tlsDisableSNI`、`tlsDisableSni` 和 `tls-disable-sni` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure` 和 `disableSNI` 查询别名，同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure`、`skip_cert_verify` 和 `disableSNI` 查询别名，同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan WebSocket/gRPC 传输参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VMess gRPC 传输参数，并同步为 sing-box outbound transport，兼容 `serviceName`、`grpcServiceName`、`multiMode` 和 `grpcMultiMode` 写法。
- VMess TCP HTTP 伪装和 HTTP/H2 传输参数会同步为 sing-box HTTP transport。
- VMess HTTPUpgrade 传输参数会同步为 sing-box HTTPUpgrade transport，兼容 VMess JSON `net=httpupgrade` 以及 userinfo URI 中 `httpUpgradeHost`/`httpUpgradePath` 查询别名。
- sing-box JSON VMess 节点的 HTTP transport 会保留 host 数组和 path，避免导入后丢失 HTTP 伪装参数。
- VMess 可兼容 `vmess://uuid@host:port?...` userinfo 直连 URI，也可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，并保留 TLS、WebSocket、gRPC、SNI、ALPN、跳过证书校验、uTLS fingerprint 和 `packet_encoding`/`packet-encoding`/`packetEncoding` 参数；userinfo URI 兼容 `serverName`、`tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`tlsDisableSNI`、`tlsEnabled`、`enableTLS`、`overTLS`、`grpcServiceName` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria `recv_window_conn`、`recv_window`、`disable_mtu_discovery` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMTUDiscovery` 和 `clientFingerprint` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria2 `up_mbps`、`down_mbps`、`insecure`、`disable_sni`、`alpn`、证书 pin、`fp`/`client-fingerprint` 和 `tls.utls.fingerprint` 参数，并同步为 sing-box outbound，兼容 URI 中 `authStr`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`obfsPassword`、`pinSHA256`、`pin_sha256`、`certificatePublicKeySHA256`、`certificate_public_key_sha256` 和 `clientFingerprint` 查询别名。
- Hysteria 和 Hysteria2 URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Hysteria/Hysteria2 结构化节点会保留 `token`/`pass`/`passwd` credential 别名、`obfsPassword`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMtuDiscovery`、`disableMTUDiscovery`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 等 camelCase 参数，并归一化到 Hysteria/Hysteria2 URI 到 sing-box outbound 生成链路。
- TUIC URI 支持从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，并兼容 `congestionControl`、`udpOverStream`、`udpRelayMode`、`zeroRttHandshake`、`heartbeatInterval`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 查询别名，兼容缺少 userinfo 的订阅写法。
- TUIC URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML TUIC 结构化节点会保留 credential 别名、`udp-over-stream`/`udpOverStream`、`zero-rtt-handshake`/`zeroRttHandshake`、`heartbeat-interval`/`heartbeatInterval`、`disable-sni`/`disable_sni` 和 `client-fingerprint`/`clientFingerprint` 等参数，并归一化到 TUIC URI 到 sing-box outbound 生成链路。
- sing-box JSON 和 URI 导入链路会保留 TUIC `insecure`/`skip-cert-verify`、`disable_sni`、`alpn`、`tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound。
- Clash YAML、Quantumult X、Surge、sing-box JSON 和 URI 导入链路会保留 Juicity UUID、密码、`congestion_control`、SNI、跳过证书校验、`disable_sni`、ALPN 和 `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound；URI 生成侧兼容 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询别名。
- Juicity URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Juicity 结构化节点会保留 `uuid`/`id`/`user-id` credential 别名、`congestionControl`/`congestion-controller`、`skip-cert-verify`、`disable-sni`/`disable_sni`、ALPN 和 `client-fingerprint`/`clientFingerprint` 客户端指纹，并归一化到 Juicity URI 到 sing-box outbound 生成链路。
- AnyTLS 和 ShadowTLS URI 支持从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- sing-box JSON 和 URI 导入链路会保留 AnyTLS 会话空闲参数和 `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`idleSessionCheckInterval`、`idleSessionTimeout`、`minIdleSession` 和 `clientFingerprint` 查询别名。
- sing-box JSON 和 URI 导入链路会保留 ShadowTLS `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 查询别名。
- AnyTLS 和 ShadowTLS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML AnyTLS/ShadowTLS 结构化节点会保留 `password`/`pass`/`passwd`/`psk`/`token` credential 别名；AnyTLS 会同步 `idleSessionCheckInterval`、`idleSessionTimeout` 和 `minIdleSession` 等会话空闲参数；两者都会通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- sing-box JSON 和 URI 导入链路会保留 Naive/Naive+QUIC `insecure`/`skip-cert-verify`、`disable_sni`、`alpn`、QUIC 控制项和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`insecureConcurrency`、`udpOverTcp`、`quicCongestionControl` 和 `clientFingerprint` 查询别名。
- Clash YAML Naive 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `insecureConcurrency`、`udpOverTcp` 和 `quicCongestionControl` 等 camelCase QUIC 控制参数，并通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 HTTP/HTTPS 代理 `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`tlsEnabled`、`enableTLS` 和 `overTLS` 查询别名。
- Naive 和 HTTP/HTTPS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Naive、HTTP/HTTPS 和 SOCKS URI 支持从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- SSH URI 支持从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息，并兼容 `identity_file`、`key_path`、`key`、`privateKey`、`privateKeyPath`、`privateKeyPassphrase`、passphrase、`hostKey`、`hostKeyAlgorithms`、`clientVersion`、`cipher`/`ciphers`、`mac`/`macAlgorithms` 和 `kexAlgorithm`/`kexAlgorithms` 查询别名，兼容缺少 userinfo 的订阅写法。
- Clash YAML SSH 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `privateKeyPath`、`privateKeyPassphrase`、`hostKey`、`hostKeyAlgorithms`、`clientVersion` 和 `kexAlgorithm` 等 camelCase 参数，并归一化到 SSH URI 到 sing-box outbound 生成链路。
- SOCKS URI 会按 `socks4://`、`socks4a://`、`socks5://` 或 `version=4/4a/5` 查询参数保留版本，并同步为 sing-box SOCKS outbound 的 `version` 字段。
- SOCKS URI 支持 `udp`、`udpEnabled`、`udp_relay` 和 `udp-relay` 标志，并会归一为 sing-box outbound 的 `network=udp`，同时兼容 `udpOverTcp` 查询别名。
- `socks5h://` 裸 URI 会归一化为标准 `socks5://`，兼容 curl/requests 生态里常见的 SOCKS5H 订阅写法；配置生成侧也兼容手动保存的 `socks5h` 协议节点并按 SOCKS5 outbound 输出。
- Clash YAML SOCKS 节点支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Surge SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Quantumult X SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- sing-box JSON SOCKS outbound 支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- V2Ray/Xray JSON SOCKS outbound 支持保留 `udp`、`udpEnabled`、`udp_enabled`、`udp_relay`、`udp-relay`、`udp_over_tcp`、`udpOverTcp`、`udp-over-tcp` 和 `uot` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- WireGuard URI 支持 `privateKey`/`secretKey`、`publicKey`/`peerPublicKey`、`preSharedKey`、`ip`/`ipv6`/`localAddresses`、`allowedIPs`/`allowedIps`/`peerAllowedIps`、`systemInterface`、`interfaceName`、`reservedBytes` 和 `peerReserved` 查询参数别名，并同步为 sing-box outbound。
- Clash YAML WireGuard 结构化节点会保留 `privateKey`、`peerPublicKey`、`localAddress`/`localAddresses`、`preSharedKey`、`allowedIPs`/`peerAllowedIps`、`reservedBytes`/`peerReserved`、`systemInterface` 和 `interfaceName` 等 camelCase 参数，并归一化到 WireGuard URI 到 sing-box outbound 生成链路。
- Tor URI 支持 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments`/重复 `arg` 和 `torrc[Option]` 查询参数别名，并同步为 sing-box outbound。
- Clash YAML Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和嵌套 `torrc` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Surge 和 Quantumult X Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和 `torrc.*`/`torrc[Option]` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Shadowsocks URI 支持从 `method`/`cipher`/`encrypt-method`/`encrypt_method`/`encryptMethod`/`encryption`/`security` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`account_password`/`account-password`/`accountPassword` 查询参数读取认证信息，并兼容 `pluginOptions` 查询别名，适配缺少 userinfo 的订阅写法。
- Hysteria2/Hy2 URI 支持从 `password`、`auth`、`auth_str`、`authStr`、`pass`、`passwd`、`pwd`、`psk`、`token`、`secret`、`credential` 或 `accountPassword` 查询参数读取认证密码，兼容缺少 userinfo 的订阅写法。
- Hysteria v1 URI 支持从 `token`、`auth`、`auth_str`、`authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`psk`、`secret`、`credential` 或 `accountPassword` 查询参数读取认证字符串，兼容部分上游订阅的 token 写法。
- sing-box 服务端配置生成会过滤已撤销、已过期、已超额或 gateway account 不可用的 Token。
- 流量统计入库骨架已落地：可解析 V2Ray stats 的 user/inbound/outbound 计数名称，写入 `traffic_samples`，计算 counter delta，并把 user 维度增量累加到 Token 用量和小时/天汇总表。
- V2Ray stats 计数名称解析会容忍分段前后空白和 `traffic`/方向字段大小写差异，降低不同实现输出格式抖动导致样本被丢弃的概率。
- user 维度流量入库后会自动判断 Token 额度；超额时将 Token 和 gateway account 标记为 `over_quota`，追加足够额度后自动恢复为 `active`。
- stats 轮询发现 Token 因真实流量进入 `over_quota` 时，会自动发布新的 sing-box 配置；如果部署侧开启 `SING_BOX_AUTO_RESTART`，会继续走受配置保护的重启执行器。
- 管理 API 和后台页面可查看 Token 今日、本月和累计流量用量摘要，展示最近 24 小时/最近 14 天流量图，并查看近 14 天上游出口流量摘要。
- 管理后台流量模块的 Token 用量和上游出口摘要已改为卡片式信息层级，两个分区不管有无数据都会展示统一符号、标题、数量徽标和说明；Token 用量卡增加状态、成员、今日和本月摘要芯片，上游出口卡增加来源、上传、下载和总量摘要芯片；Token 用量卡展示额度使用率进度条，并按接近超额和已超额状态变色；浏览器 QA 会验证流量分区头、卡片数量、摘要芯片、额度条和无视觉溢出。
- 管理后台流量模块顶部已新增流量观测工作台，先展示 Token 用量、今日流量、24 小时活跃样本和出口摘要 4 个关键状态，并用额度风险、最近小时和出口链路 3 个信号卡说明统计链路是否有真实数据；浏览器 QA 会校验工作台、状态符号、信号卡和无视觉溢出。
- 管理后台流量趋势图已补齐小时/日摘要头，展示趋势符号、总量、峰值和活跃样本数，便于真实测试时先扫读流量是否进入统计链路；浏览器 QA 会验证两个趋势图摘要头、摘要芯片和无视觉溢出。
- 订阅响应头返回标准 `subscription-userinfo`，并提供 FluxGate 专属的已用、总额和剩余额度头。
- stats 可插拔轮询调度器已落地，能将采集器返回的 V2Ray counters 转换为流量样本并写入现有用量汇总链路。
- 真实 sing-box V2Ray gRPC stats 采集器已接入主进程；配置 `SING_BOX_V2RAY_API_ADDR` 后会按 `STATS_POLL_INTERVAL_SECONDS` 轮询并写入现有统计链路。
- sing-box 服务端配置生成会把 active 上游 outbound tag 写入 V2Ray stats 配置，支持上游出口流量汇总。
- Token 仍以 hash 作为鉴权主键；订阅密钥额外使用 `TOKEN_SECRET` 派生密钥加密保存，支持管理员后台按当前访问域名反复复制订阅地址。
- 订阅请求日志 Token 路径脱敏。
- 结构化 JSON 服务日志。
- 本地验证脚本。
- 公开仓库脱敏扫描脚本。
- 远程部署探测脚本。
- 磁盘清理脚本。
- 诊断采集脚本。
- 部署后真实可用性收口脚本，可基于已有团队、成员、Token 和上游节点创建缺失的默认虚拟节点与默认策略，校验并发布 sing-box 配置，远程配置存在时会重启 sing-box 数据面。
- 部署后真实可用探测脚本，可只读检查控制面核心资源、sing-box 配置入站/用户/上游摘要、局域网网关端口连通性；脚本会优先使用 `FLUXGATE_QA_SUBSCRIPTION_URL`，未提供时自动从管理员 Token 列表中选取可恢复订阅地址，验证默认、Clash/Mihomo 和 sing-box 客户端订阅正文包含当前虚拟节点。
- sing-box 当前配置校验并重载脚本，适用于配置文件已由控制面写入但需要数据面重新加载的场景。
- 推送后远程部署脚本。
- GitHub Actions Docker 镜像构建工作流，支持 PR 构建验证和分支/tag 推送 GHCR。
- 自定义 sing-box Dockerfile 已落地，默认远程构建带 `with_v2ray_api` 的本地数据面镜像，避免官方镜像缺少 V2Ray API 导致统计采集不可用。
- 自定义 sing-box Dockerfile 运行阶段使用 `scratch` 并从 builder 复制 CA 证书，减少远程构建对额外镜像仓库的依赖，降低 registry 元数据超时导致部署回退的概率。
- Docker 构建上下文脱敏，默认排除 `.env`、数据库、日志、测试产物和本机部署配置。
- 远程验收健康检查重试。
- 首次远程部署最小 sing-box config bootstrap。
- 远程 Docker build 支持 `GOPROXY`，默认优先使用 `goproxy.cn` 以避开 `proxy.golang.org` 超时。
- 推送后远程部署支持远程构建优先；当外部 registry 网络不可用时，会回退为本地交叉编译、scratch 镜像构建并流式加载到服务器。
- 部署时宿主机 HTTP 端口默认使用 `127.0.0.1:18080`，避免和服务器已有 8080 服务冲突。
- 部署侧可通过未跟踪配置打开局域网访问，不把真实环境信息提交到公开仓库。
- 页面截图验收脚本，支持指定目标 URL、保留截图、自定义输出目录和 `SCREENSHOT_VIEWPORTS` 多视口矩阵。
- 浏览器登录验收脚本。
- 只读就绪检查脚本，可在本地或远程实例上检查健康状态、登录、概览、团队、成员、Token、来源、节点、虚拟节点、策略和流量摘要；`STRICT=true` 时缺少体验前置数据会使检查失败。
- 本地 QA 套件退出自动清理临时产物。
- 手把手使用说明已落地，覆盖部署初始化、管理员登录、添加上游订阅、节点地区聚合与详情查看、节点/来源编辑、团队/用户/Token 管理、订阅地址分发、流量额度查看、配置发布/回滚、远程运维和常见问题排查。

## 2. 当前 API 骨架

已实现的主要接口：

```text
GET  /healthz
GET  /readyz
GET  /api/auth/session
POST /api/auth/login
POST /api/auth/logout
GET  /api/overview

GET  /api/teams
POST /api/teams
PATCH /api/teams/{id}

GET  /api/users
POST /api/users
PATCH /api/users/{id}

GET  /api/tokens
POST /api/tokens
POST /api/tokens/{id}/revoke
POST /api/tokens/{id}/restore
POST /api/tokens/{id}/rotate-subscription
POST /api/tokens/{id}/extend
POST /api/tokens/{id}/quota

GET   /api/sources
POST  /api/sources
PATCH /api/sources/{id}
POST  /api/sources/{id}/refresh
POST  /api/sources/{id}/regenerate-node-names

GET   /api/nodes
GET   /api/nodes/{id}
POST  /api/nodes/import
PATCH /api/nodes/{id}
POST  /api/nodes/{id}/reset-display-name

GET  /api/virtual-nodes
POST /api/virtual-nodes
PATCH /api/virtual-nodes/{id}

GET  /api/policies
POST /api/policies
PATCH /api/policies/{id}
GET  /api/traffic/tokens
GET  /api/traffic/daily?days=14
GET  /api/traffic/hourly?hours=24
GET  /api/traffic/outbounds?days=14

POST /api/sing-box/config/generate
POST /api/sing-box/config/check
POST /api/sing-box/config/publish
POST /api/sing-box/config/rollback
POST /api/sing-box/restart
GET  /sub/{token}
```

## 3. 验证结果

已通过：

```text
scripts/dev/bootstrap.sh
scripts/dev/format.sh
scripts/dev/test.sh
scripts/dev/lint.sh
scripts/dev/build.sh
scripts/db/migrate.sh
scripts/qa/public-scan.sh
scripts/qa/smoke.sh
scripts/qa/api-flow.sh
scripts/qa/readiness.sh
scripts/qa/substore-extraction.sh
scripts/qa/usable-probe.sh
scripts/qa/browser-login.sh
scripts/qa/screenshot.sh
scripts/qa/local-suite.sh
scripts/deploy/remote-logs.sh
scripts/deploy/probe-env.sh
scripts/deploy/ensure-usable.sh
scripts/deploy/push-and-deploy.sh
DRY_RUN=true scripts/deploy/cleanup-disk.sh
docker compose config
```

页面截图验收：

```text
KEEP_ARTIFACTS=true scripts/qa/screenshot.sh 可保留 1440x900、1280x720、390x844 三档截图；默认测试退出会自动清理截图。
SCREENSHOT_VIEWPORTS="1440x900 390x844" scripts/qa/screenshot.sh 可缩小或扩展截图矩阵。
scripts/qa/screenshot.sh --keep http://<lan-host>:<port> 可对局域网部署页面保留人工复核截图，产物形如 fluxgate-1440x900.png。
scripts/qa/browser-login.sh http://<lan-host>:<port> 可做真实浏览器登录验收。
scripts/qa/readiness.sh http://<lan-host>:<port> 可做真实实例只读就绪检查；STRICT=true 会把缺少体验数据视为失败。
scripts/qa/substore-extraction.sh 使用内存 fake transport 验证 subscription URL 来源会先经 Sub-Store 提取节点再入库，不依赖真实机场或本地监听端口。
scripts/qa/usable-probe.sh http://<lan-host>:<port> 可做真实实例只读可用性检查，包括 sing-box 配置摘要、局域网网关端口连通，以及从后台 Token 列表自动取订阅地址后的正文探测。
scripts/deploy/ensure-usable.sh http://<lan-host>:<port> 可做真实实例可用性收口；已有 Token 过期或超额时会先续期或补额度，再生成、发布 sing-box 配置并重启远端 sing-box。
```

截图结论：

- 页面可打开。
- 未登录时展示管理员登录页。
- 登录后管理员登录表单必须不可见，后台视图必须可见，并可加载仪表盘数据。
- 浏览器验收会确认概览、接入、节点、身份、策略、流量和运维 7 个模块视图可通过后台导航切换。
- 浏览器验收会确认 7 个创建/导入表单抽屉默认收纳、标题符号顺序正确、默认宽度覆盖模块操作区、展开后作为右侧侧滑工作台具备 dialog 语义、取消动作、`Escape` 收起且保留草稿，并且标题区无视觉溢出。
- 浏览器验收会确认每个模块标题区至少展示 3 个上下文摘要项，且摘要胶囊内部无视觉溢出。
- 浏览器验收会确认每个模块标题区都有工作区重点提示条，提示条包含状态符号、短标题、上下文动作按钮且不发生视觉溢出；提示动作会直接定位到目标区域并给出 `已定位` 反馈。
- 浏览器验收会确认每个模块标题区都有随模块变化的主操作按钮，主操作具备符号、符合当前模块的动作标签、无视觉溢出，并可从身份模块直接展开 Token 创建表单。
- 浏览器验收会确认模块主操作跳转到表单时目标抽屉会高亮且焦点落在第一个输入字段，避免展开后仍需要手动寻找输入位置。
- 静态 UI 验收会确认概览快捷入口具备具体目标面板，点击后通过统一导航逻辑直接定位来源、节点、Token 或运维操作区；模块摘要和面板入口使用横向可扫命令条，桌面保持首屏稳定，移动端保持可点击。
- 浏览器验收会确认模块快速定位条的符号顺序、徽标数量、按钮内部无溢出和创建入口展开能力。
- 浏览器验收会确认 9 个模块面板符号和数量/状态徽标都已渲染且非空。
- 浏览器验收会确认概览页包含 7 个带符号的顶部指标卡、初始化向导进度、当前建议动作、5 个带序号/符号/状态徽标和提示文案的闭环步骤、1 个下一步动作卡和 4 个带模块符号与实时徽标的快速入口。
- 浏览器验收会等待管理后台数据加载完成后再进入模块视觉断言；若加载异常，会输出状态、指标区错误、API 响应列表和控制台消息，便于直接定位真实页面问题。
- 浏览器验收会拦截顶部刷新请求，确认刷新期间按钮进入 `刷新中`、锁定重复点击、主工作区标记 `aria-busy`，完成后恢复原标签、符号和可点击状态。
- 浏览器验收会拦截顶部退出请求，确认退出期间按钮进入 `退出中`、锁定重复点击、主工作区标记 `aria-busy`，完成后回到登录页并恢复未登录状态。
- 无白屏。
- 无明显遮挡。
- 表格、卡片、按钮和输入控件未出现明显溢出或尺寸错位；每个模块视图都必须没有页面级横向滚动。
- 浏览器验收会在 1440x900 宽屏桌面确认桌面侧边栏可见、移动 Dock 隐藏、命令中心仍为低矮工具带、各模块无页面横向溢出、面板标题和顶部工作台不产生滚宽，并确认概览总控区与 KPI 保持紧凑控制室列数。
- 浏览器验收会在 390px 移动宽度确认底部 Dock 可见、桌面侧边栏隐藏，Dock 保持 5 个入口，策略、流量和运维可从“更多”菜单进入，菜单选择后会自动收起，且概览、节点、策略和运维视图不产生横向溢出。
- 来源前缀、节点展示名和 Token 前缀正常展示。
- 有团队和成员数据时，浏览器验收会确认团队/成员动作按钮带有符号，且“编辑”入口可打开行内编辑字段。
- 有来源数据时，浏览器验收会确认上游来源编辑、刷新和同步命名动作都带有符号且不溢出。
- 有来源数据时，浏览器验收会拦截来源卡片保存请求，确认保存期间卡片进入 busy、保存/取消/输入控件锁定、按钮显示 `保存中`，请求完成后恢复为“编辑”入口。
- 有节点数据时，浏览器验收会确认节点池先展示带地区符号、可用徽标和协议/来源胶囊的地区聚合，进入地区后出现节点卡片。
- 有节点数据时，浏览器验收会确认节点池地区返回和筛选清空控件都带有动作符号。
- 有节点数据时，浏览器验收会确认节点卡片“编辑”入口可打开行内编辑表单。
- 有来源数据时，浏览器验收会确认上游来源“编辑”入口可打开行内编辑字段。
- 有策略数据时，浏览器验收会确认策略“编辑”入口可打开行内编辑字段。
- 有节点数据时，浏览器验收会确认节点池“详情”入口可打开单节点详情面板，且桌面详情面板不会覆盖节点卡片。
- 浏览器验收会确认 RFC3339 和 SQLite 时间字符串都按东八区展示。
- 有 Token 流量数据时，浏览器验收会确认流量分区头、卡片、额度使用率进度条、小时/日趋势图摘要头和出口摘要卡片无视觉溢出；没有出口流量时会确认出口摘要分区头和结构化空状态符号、文案、动作按钮、跳转到运维模块能力与布局不溢出。
- 浏览器验收会确认运维模块 5 个带符号和状态胶囊的操作卡可见，并点击“检查配置”和“检查收口”验证结果卡符号、状态胶囊、结构化字段和无视觉溢出；其中“检查配置”会被请求拦截以确认操作卡进入等待态、整组高风险操作临时锁定、按钮文案恢复正常。
- 有 Token 数据时，浏览器验收会确认订阅分发工作台提供 3 个首屏动作入口，动作符号为 `订/米/↗`，定位订阅卡后出现高亮和 `已定位：订阅地址` 状态，Mihomo 打开入口具备真实跳转地址且动作条无溢出；同时确认每行都展示带可见标签和单位的自定义续期天数、追加额度 MiB 输入控件、状态控制动作区和带格式画像、地址来源的可复制订阅卡，实际点击复制按钮确认反馈与符号按钮恢复，并点击订阅探测按钮确认“可访问”和最近探测结果可见；同时会点击重置订阅入口，确认站内危险操作框标题、双按钮结构、取消反馈和关闭状态，并确认创建/重置后的订阅结果卡具备符号标题和格式数量胶囊，Token 动作区可见且无溢出或横向滚动。
- 有节点数据时，浏览器验收会展开节点编辑表单确认保存按钮符号化，再打开节点详情面板，确认 URI 不横向溢出、桌面不遮挡节点卡片，并实际点击复制 URI 按钮确认反馈。
- 运维模块顶部新增交付发布工作台，展示闭环进度、配置 Hash、入站和上游出口 4 个核心状态，并用来源、节点、虚拟网关、Token、策略和配置 6 个步骤卡说明发布链路是否齐备；浏览器验收会确认工作台、步骤卡、关键符号和无溢出。

## 4. 尚未完成

下一步需要继续实现：

- 上游订阅更多结构化格式解析。
- 更多特殊协议 URI 到 sing-box outbound 的转换。
- 更多真实订阅样本和特殊协议兼容性验证。
- 手把手使用说明需要随后续功能持续更新。
- 初始化向导后续可继续结合真实客户端导入结果补充更细的故障排查建议。

## 5. 当前注意事项

- `docker compose config` 已支持没有 `.env` 时做静态校验。
- 生产部署仍应由 `scripts/deploy/bootstrap-remote.sh` 生成 `.env` 后再修改密钥；管理员引导账号只在没有管理员记录时创建。
- 管理 API 需要管理员会话；订阅接口 `/sub/{token}` 继续使用订阅 Token 鉴权，不依赖管理员登录。
- 需要局域网访问管理后台时，通过部署侧配置 `FLUXGATE_HOST_BIND=0.0.0.0` 和 `FLUXGATE_HTTP_PORT`，公开仓库不保存真实访问地址。
- `SING_BOX_AUTO_RESTART` 默认关闭；如果要让控制面重启 sing-box，需要在部署侧显式挂载 Docker socket 或改用 `command` driver，并提供对应容器/宿主机权限。
- `SING_BOX_V2RAY_API_ADDR` 为空时 stats 轮询器不会启动；生产部署应绑定 Docker 内网地址，不应暴露到公网。
- sing-box 官方文档已标注 WireGuard outbound 废弃；当前已能导入 WireGuard endpoint 模型，但配置生成侧仍按 Phase 1 既有 outbound 骨架兼容输出，后续应评估完整迁移到 endpoint 配置。
- 本地 QA 产生的 `data/`、`logs/`、`tmp/` 均被 `.gitignore` 排除，并默认在测试退出时清理。
- Playwright Chromium 已在本机安装一次，后续截图脚本会复用缓存。
