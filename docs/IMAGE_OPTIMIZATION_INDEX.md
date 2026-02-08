# 镜像源优化方案文档索引

本目录包含 Cyber Range 平台的镜像源优化相关方案文档。

## 📚 文档列表

### 1. [镜像优化方案对比](./image_optimization_plan.md)

**内容概要**：
- 三种技术方案对比（本地镜像、私有仓库、混合模式）
- 每种方案的优缺点分析
- 性能对比和实施路径
- 镜像打包、分发和清理策略

**适用场景**：需要了解不同方案的技术选型

---

### 2. [私有镜像仓库实施方案](./image_registry_implementation.md) ⭐ 推荐

**内容概要**：
- 完整的技术实施细节
- 数据库设计（`docker_images` 表）
- 后端开发方案（Model、Repository、Service、Handler）
- 前端界面改造（镜像下拉选择）
- 部署脚本（Registry 启动、镜像导入）
- 自动预加载机制
- 详细的验证计划和检查清单

**适用场景**：准备实施私有镜像仓库方案时的参考文档

**预计实施时间**：5.5 小时

---

## 🎯 方案选择总结

基于您的需求（本地 Mac + 远程 Docker 服务器 + 后台配置），已选择：

| 项目 | 选择 |
|------|------|
| 总体方案 | 私有镜像仓库（方案 B） |
| Registry 部署 | 本地 Mac (localhost:5000) |
| 镜像规模 | 1个测试，未来约10个，<1GB/个 |
| 后台功能 | 基础（下拉选择 + 列表查看）|
| 数据库设计 | 新增 docker_images 表 |
| 预加载策略 | 系统启动时自动同步 |

---

## 🚀 快速开始

1. **查看详细方案**：阅读 [image_registry_implementation.md](./image_registry_implementation.md)
2. **启动 Registry**：运行 `scripts/start-registry.sh`（需先创建）
3. **导入镜像**：运行 `scripts/import-image.sh your-image.tar`
4. **数据库迁移**：执行数据库变更
5. **后端开发**：按照方案实施后端改造
6. **前端开发**：修改题目创建/编辑页面
7. **测试验证**：按照验证计划测试

---

## 📝 相关文档

- [多 Docker 主机支持方案](./implementation_plan.md) - 已实施的多主机架构
- [Docker 配置推荐](./docker_config_recommendation.md) - Docker 远程配置指南
- [Docker 主机选择指南](./docker_host_selection_guide.md) - 题目主机配置说明

---

## 🔄 文档版本

- 创建时间：2026-01-31
- 最后更新：2026-01-31
- 状态：规划阶段，待实施

---

需要开始实施，请查阅 [image_registry_implementation.md](./image_registry_implementation.md) 中的详细步骤。
