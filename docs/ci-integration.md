# CI Integration Guide

本文档说明了 Makefile 与各 CI/CD 平台的集成方式,以实现平台无关的 CI 流程。

## 设计理念

### 本地优先 (Local-First)

所有 CI 检查逻辑优先在 Makefile 中实现,GitHub Actions/GitLab CI 只是调用这些 Make 目标。这样做的好处:

1. **开发体验一致**: 本地运行 `make ci` 与 CI 服务器运行结果一致
2. **平台无关**: 轻松迁移到 GitLab、Jenkins、Drone 等其他 CI 平台
3. **快速调试**: 无需推送代码即可在本地验证 CI 流程
4. **减少重复**: CI 配置文件只需调用 Make 目标,避免逻辑重复

### 关注点分离

- **Makefile**: 包含所有测试、构建、质量检查的**业务逻辑**
- **CI 配置文件**: 仅负责**平台特定功能**(如 artifact 上传、PR 评论、badge 生成)

## Makefile 目标与 CI 的对应关系

### 核心 CI 流程

| Make 目标 | 功能 | CI 阶段 | 说明 |
|-----------|------|---------|------|
| `make ci-pre` | 预检查 | Pre-CI | 验证工作空间和依赖 |
| `make ci-lint` | 代码检查 | Lint | 格式和静态分析 |
| `make ci-test` | 测试验证 | Test | 运行测试并检查覆盖率 |
| `make ci-security` | 安全扫描 | Security | 安全和漏洞检查 |
| `make ci-quality` | 质量检查 | Quality | 复杂度和重复代码检查 |
| `make build` | 构建验证 | Build | 编译所有模块 |
| `make ci` | 完整 CI | Full CI | 运行所有必需检查 |
| `make ci-full` | 完整 CI + 报告 | Full CI | CI + 详细报告 |

### 子目标详细说明

#### 预检查 (`ci-pre`)
```bash
make check-workspace    # 验证 Go workspace 配置
make mod-verify        # 验证并下载依赖
```

#### 代码检查 (`ci-lint`)
```bash
make check-format      # 检查代码格式
make lint             # 运行 golangci-lint
```

#### 测试验证 (`ci-test`)
```bash
make test             # 运行测试 (带 race 检测)
make check-coverage   # 验证覆盖率 ≥80%
```

#### 安全扫描 (`ci-security`)
```bash
make security         # 运行 gosec 安全扫描
make vuln-check       # 检查依赖漏洞
```

#### 质量检查 (`ci-quality`)
```bash
make check-cyclo      # 圈复杂度 ≤15
make check-cognit     # 认知复杂度 ≤20
make check-dupl       # 代码重复 (阈值 50 tokens)
```

#### PR 特定检查 (`ci-pr`)
```bash
make check-large-files  # 检查大文件 (>1MB)
make check-conflicts    # 检查合并冲突标记
make lint-commits       # 验证提交消息格式
```

## GitHub Actions 集成示例

### CI 工作流 (`.github/workflows/ci.yml`)

```yaml
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5

      # 使用 Makefile 目标
      - name: Verify workspace
        run: make check-workspace

      - name: Run tests
        run: make ci-test

      # GitHub 特定: 上传覆盖率到 Codecov
      - uses: codecov/codecov-action@v4
        with:
          files: ./hyperion/coverage.out

  lint:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5

      # 安装工具
      - run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
            sh -s -- -b $(go env GOPATH)/bin latest

      # 使用 Makefile 目标
      - name: Lint
        run: make ci-lint
```

### PR 检查工作流 (`.github/workflows/pr-checks.yml`)

```yaml
jobs:
  pr-checks:
    steps:
      - uses: actions/checkout@v4

      # 完全复用 Makefile 目标
      - run: make check-large-files
      - run: make check-conflicts
      - run: make lint-commits
```

## GitLab CI 集成示例

```yaml
# .gitlab-ci.yml

stages:
  - pre
  - test
  - lint
  - security
  - quality

variables:
  GO_VERSION: "1.24"

# 预检查阶段
workspace:
  stage: pre
  image: golang:${GO_VERSION}
  script:
    - make check-workspace
    - make mod-verify

# 测试阶段
test:
  stage: test
  image: golang:${GO_VERSION}
  script:
    - make ci-test
  coverage: '/Average Coverage: (\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

# Lint 阶段
lint:
  stage: lint
  image: golang:${GO_VERSION}
  script:
    - make ci-lint

# 安全扫描
security:
  stage: security
  image: golang:${GO_VERSION}
  script:
    - make ci-security

# 质量检查
quality:
  stage: quality
  image: golang:${GO_VERSION}
  script:
    - make ci-quality
  allow_failure: true
```

## Jenkins Pipeline 集成示例

```groovy
// Jenkinsfile

pipeline {
    agent {
        docker {
            image 'golang:1.24'
        }
    }

    stages {
        stage('Pre-CI') {
            steps {
                sh 'make ci-pre'
            }
        }

        stage('Lint') {
            steps {
                sh 'make ci-lint'
            }
        }

        stage('Test') {
            steps {
                sh 'make ci-test'
            }
        }

        stage('Security') {
            steps {
                sh 'make ci-security'
            }
        }

        stage('Quality') {
            steps {
                sh 'make ci-quality'
            }
        }

        stage('Build') {
            steps {
                sh 'make build'
            }
        }
    }

    post {
        always {
            // 发布测试报告
            junit '**/test-results/*.xml'

            // 发布覆盖率报告
            publishHTML([
                reportDir: 'hyperion',
                reportFiles: 'coverage.html',
                reportName: 'Coverage Report'
            ])
        }
    }
}
```

## 本地开发工作流

### 开发前检查
```bash
# 安装开发工具和 Git hooks
make setup

# 查看所有可用目标
make help
```

### 提交前验证
```bash
# 快速验证 (格式 + Lint + 测试)
make verify

# 或使用 Git hook 自动检查
make check-commit
```

### 完整 CI 验证
```bash
# 运行与 CI 服务器完全相同的检查
make ci

# 运行 CI + 生成详细报告
make ci-full
```

### PR 提交前检查
```bash
# 运行 PR 特定检查
make ci-pr

# 查看覆盖率报告
make coverage-report

# 查看代码质量报告
make quality-report
```

## 平台特定功能

虽然核心逻辑在 Makefile 中,但某些功能仍然是平台特定的:

### GitHub Actions 特有
- ✅ PR 评论 (通过 `actions/github-script`)
- ✅ SARIF 安全报告上传 (`codeql-action/upload-sarif`)
- ✅ Codecov 集成 (`codecov/codecov-action`)
- ✅ PR 标题验证 (`amannn/action-semantic-pull-request`)
- ✅ 自动标签 (`actions/labeler`)

### GitLab CI 特有
- 覆盖率徽章
- Merge Request 评论
- 容器镜像发布

### Jenkins 特有
- Blue Ocean 可视化
- 自定义插件集成

## 迁移指南

### 从 GitHub Actions 迁移到 GitLab CI

1. 复制 Makefile (无需修改)
2. 创建 `.gitlab-ci.yml`,调用相同的 Make 目标
3. 配置 GitLab 特定功能 (Coverage Badge, MR 评论等)

### 从 GitLab CI 迁移到 Jenkins

1. 复制 Makefile (无需修改)
2. 创建 `Jenkinsfile`,调用相同的 Make 目标
3. 配置 Jenkins 插件 (JUnit, HTML Publisher 等)

## 故障排查

### 本地通过但 CI 失败

```bash
# 确保使用相同的 Go 版本
go version

# 清理缓存后重新运行
make clean
make ci
```

### 权限问题

```bash
# 确保脚本可执行
chmod +x scripts/*.sh

# 重新安装 hooks
make install-hooks
```

### 工具版本不一致

```bash
# 使用项目指定的工具版本
make tools
make quality-tools
```

## 最佳实践

1. **始终在本地运行 `make ci` 后再推送代码**
2. **使用 `make setup` 初始化开发环境**
3. **遵循提交消息规范** (通过 `make lint-commits` 验证)
4. **保持测试覆盖率 ≥80%** (通过 `make check-coverage` 验证)
5. **定期运行 `make quality-report`** 监控代码质量

## 相关文档

- [Makefile 目标说明](../Makefile) - 运行 `make help` 查看所有目标
- [提交规范](../CONTRIBUTING.md) - Conventional Commits 规范
- [架构决策](./architecture-decisions.md) - v2.0 Monorepo 架构设计
