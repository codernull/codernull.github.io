# WSL + Neovim 配置全记录（可复现，不是插件安利）

> 素材来自 WSL Ubuntu 用户 `xxxx` 的实际配置复查；用户名、目录和运行环境仅用于复现配置链路，不包含业务代码或敏感凭据。

目标：在 WSL 中构建一套 **纯 Neovim（Lua）** 的 Go 开发环境，具备：
- 跳转/补全（`gopls`）
- 编译与 quickfix（`:GoBuild` / `:GoTest`）
- 调试（`dlv` + `nvim-dap-go`）
- 可复现安装与回滚流程

---

## 1. 配置决策复盘：为何选 Snap Neovim + Lua 覆盖

### 1.1 为什么选 Snap 版 Neovim

- 需求是使用较新的 Neovim API（例如 `vim.lsp.config` / `vim.lsp.enable`）。
- Ubuntu apt 版本常滞后，不保证可用新 API。
- Snap 版可稳定拿到更近版本，和当前 Lua 配置兼容性更高。

### 1.2 为什么改成“纯 Lua，不依赖 `.vimrc`”

- 单一配置入口：`~/.config/nvim/init.lua`，避免 Vimscript/Lua 双栈冲突。
- 可模块化：`lua/config/*` 与 `lua/plugins/*` 分层清晰。
- 可复现：配合 `lazy-lock.json` 锁定插件提交。

### 1.3 关于 `lazy_git_url`

- 当前不需要设置 `lazy_git_url`。
- 只要 `lazy.nvim` 正常工作，就保持最小配置即可。

---

## 2. 当前环境基线（复现前先对齐）

- WSL：Ubuntu（WSL2）
- Neovim：`/snap/bin/nvim`（当前为 `v0.13.0-dev`）
- Go：建议 `1.25+`
- `gopls`：`@latest`
- `dlv`：`@latest`
- `rg`：可用

快速确认：

```bash
command -v nvim
nvim --version | sed -n '1,3p'
command -v go && go version
command -v gopls && gopls version
command -v dlv && dlv version
command -v rg && rg --version | sed -n '1p'
```

---

## 3. 安装步骤（从零）

### 3.1 安装 Neovim 与基础工具

```bash
sudo apt-get update
sudo apt-get install -y snapd ripgrep git build-essential
sudo snap install nvim --classic
```

> 若系统里有旧 apt 版 neovim，为避免二义性可移除：
>
> ```bash
> sudo apt-get remove -y neovim
> ```

### 3.2 安装 Go 工具

```bash
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 3.3 PATH（按你当前使用方式）

当前只要求“从交互终端打开 nvim 可用”，`~/.bashrc` 已足够。  
若未来要跑非交互脚本，再把 PATH 提升到 `~/.profile`。

---

## 4. 纯 Neovim 目录结构（目标态）

```text
~/.config/nvim/
├── init.lua
├── lazy-lock.json
├── lua/config/
│   ├── options.lua
│   ├── keymaps.lua
│   ├── autocmds.lua
│   └── lazy.lua
└── lua/plugins/
    ├── lsp.lua
    ├── cmp.lua
    ├── dap.lua
    ├── telescope.lua
    ├── neotree.lua
    ├── treesitter.lua
    └── ui.lua
```

---

## 5. 迁移计划：先迁移 `.vimrc` 能力，再备份 `.vimrc`

> 要求：**不直接删除 `.vimrc`**。先在 nvim 内完成迁移并验证，再改名为 `.vimrc.backup`。

### 5.1 将 Go 编译/quickfix 迁移到 Lua（建议新建 `lua/config/go_build.lua`）

```lua
local M = {}

local go_efm = "%f:%l:%c: %m,%f:%l: %m"

local function go_make(args)
  local save_mp = vim.bo.makeprg
  local save_efm = vim.bo.errorformat

  vim.bo.makeprg = "go " .. args
  vim.bo.errorformat = go_efm
  vim.cmd("silent make!")
  vim.cmd("cwindow")

  vim.bo.makeprg = save_mp
  vim.bo.errorformat = save_efm
end

vim.api.nvim_create_user_command("GoBuild", function()
  go_make("build ./...")
end, {})

vim.api.nvim_create_user_command("GoTest", function()
  go_make("test ./...")
end, {})

vim.api.nvim_create_user_command("GoVet", function()
  go_make("vet ./...")
end, {})

vim.api.nvim_create_user_command("GoRun", function()
  vim.cmd("!" .. "go run " .. vim.fn.shellescape(vim.fn.expand("%:p")))
end, {})

vim.keymap.set("n", "<leader>mb", "<cmd>GoBuild<CR>", { silent = true })
vim.keymap.set("n", "<leader>mt", "<cmd>GoTest<CR>", { silent = true })
vim.keymap.set("n", "<leader>mv", "<cmd>GoVet<CR>", { silent = true })
vim.keymap.set("n", "<leader>mr", "<cmd>GoRun<CR>", { silent = true })
vim.keymap.set("n", "<leader>co", "<cmd>copen<CR>", { silent = true })
vim.keymap.set("n", "<leader>cc", "<cmd>cclose<CR>", { silent = true })
vim.keymap.set("n", "<leader>cn", "<cmd>cnext<CR>", { silent = true })
vim.keymap.set("n", "<leader>cp", "<cmd>cprev<CR>", { silent = true })

return M
```

在 `init.lua` 中加载：

```lua
require("config.options")
require("config.keymaps")
require("config.autocmds")
require("config.go_build")
require("config.lazy")
```

### 5.2 验证迁移是否成功（必须通过）

```bash
nvim --headless "+command GoBuild" "+command GoTest" "+qa"
nvim --headless "+set number?" "+set relativenumber?" "+qa"
```

在项目中手动验证：

1. `:GoBuild` 可以执行并生成 quickfix。
2. `:GoTest` 可以执行并生成 quickfix。
3. `:copen` / `:cnext` / `:cprev` 能跳转错误。
4. `gd` / `gr` 可用（`gopls` 附着正常）。

### 5.3 通过后再执行备份（不是删除）

```bash
mv ~/.vimrc ~/.vimrc.backup
```

---

## 6. 必要恢复配置清单（灾备最小集）

要完全恢复现有体验，至少备份/同步以下内容：

1. `~/.config/nvim/init.lua`
2. `~/.config/nvim/lua/config/*.lua`
3. `~/.config/nvim/lua/plugins/*.lua`
4. `~/.config/nvim/lazy-lock.json`
5. `~/.vimrc.backup`（迁移前历史保留）

建议增加一键恢复脚本（示例）：

```bash
mkdir -p ~/.config/nvim
cp -r ./nvim/* ~/.config/nvim/
nvim --headless "+Lazy! sync" "+qa"
```

---

## 7. 发布同步流程（文档与配置）

### 7.1 同步内容

- 文档：`docs/Go实战/wsl-neovim-配置全记录.md`
- 配置快照：建议独立目录（如 `infra/nvim-config/`）或独立仓库
- 锁文件：`lazy-lock.json` 必须提交

### 7.2 发布前检查

```bash
nvim --headless "+checkhealth" "+qa"
nvim --headless "+command GoBuild" "+command GoTest" "+qa"
```

### 7.3 发布后的验收

- 新机器按文档完成安装后，10 分钟内可：
  - 打开 Go 项目
  - `gd` 跳转定义
  - `:GoBuild` + quickfix 跳错
  - `<F5>` 启动调试

---

## 8. 结尾复盘（这次改造的结论）

1. **方向上**：从“vimrc + nvim 混合态”收敛到“纯 nvim Lua”，可维护性明显更好。  
2. **风险上**：保留 `.vimrc.backup` 避免一次性切断回退路径。  
3. **可复现上**：`lazy-lock.json` 是关键资产，不应省略。  
4. **取舍上**：不追求 health 全绿；优先保证跳转、编译、quickfix、调试这条主链稳定。  
5. **后续上**：仅当出现脚本/服务化需求，再处理非交互 PATH。
