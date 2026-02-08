# 数据库中文注释迁移指南

## ✅ 已完成的工作

为所有数据表和字段添加了详细的中文注释！

### 修改的文件

**`internal/model/model.go`**
- ✅ Challenge（挑战题目表） - 9个字段全部添加注释
- ✅ Instance（容器实例表） - 8个字段全部添加注释
- ✅ User（用户表） - 7个字段全部添加注释
- ✅ Submission（提交记录表） - 7个字段全部添加注释

---

## 📝 注释示例

### Challenge 表
```sql
CREATE TABLE `challenges` (
  `id` varchar(36) NOT NULL COMMENT '题目唯一标识',
  `title` varchar(200) NOT NULL COMMENT '题目标题',
  `description` text COMMENT '题目描述',
  `category` varchar(50) COMMENT '题目分类(Web/Pwn/Crypto/Reverse)',
  `difficulty` varchar(20) COMMENT '难度级别(Easy/Medium/Hard)',
  `image` varchar(500) NOT NULL COMMENT 'Docker镜像名称',
  `flag` varchar(500) NOT NULL COMMENT 'Flag答案(静态模板,不返回给前端)',
  `points` bigint NOT NULL DEFAULT 100 COMMENT '题目分值',
  `created_at` datetime(3) COMMENT '创建时间',
  `updated_at` datetime(3) COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='挑战题目表';
```

---

## 🔄 应用注释到数据库

### 方式1：运行迁移脚本（推荐）

```bash
# 执行迁移（会删除旧数据，重新创建表）
go run cmd/migrate/main.go

# 然后重新填充数据
go run cmd/seed/main.go
```

### 方式2：手动ALTER TABLE

如果不想删除数据，可以手动添加注释：

```sql
-- 修改表注释
ALTER TABLE challenges COMMENT '挑战题目表';
ALTER TABLE instances COMMENT '容器实例表';
ALTER TABLE users COMMENT '用户表';  
ALTER TABLE submissions COMMENT 'Flag提交记录表';

-- 修改字段注释（示例）
ALTER TABLE challenges MODIFY COLUMN id varchar(36) NOT NULL COMMENT '题目唯一标识';
ALTER TABLE challenges MODIFY COLUMN title varchar(200) NOT NULL COMMENT '题目标题';
-- ... 更多字段
```

---

## 📊 字段注释完整列表

### challenges（挑战题目表）
| 字段 | 类型 | 注释 |
|:-----|:-----|:-----|
| id | varchar(36) | 题目唯一标识 |
| title | varchar(200) | 题目标题 |
| description | text | 题目描述 |
| category | varchar(50) | 题目分类(Web/Pwn/Crypto/Reverse) |
| difficulty | varchar(20) | 难度级别(Easy/Medium/Hard) |
| image | varchar(500) | Docker镜像名称 |
| flag | varchar(500) | Flag答案(静态模板,不返回给前端) |
| points | bigint | 题目分值 |
| created_at | datetime(3) | 创建时间 |
| updated_at | datetime(3) | 更新时间 |

### instances（容器实例表）
| 字段 | 类型 | 注释 |
|:-----|:-----|:-----|
| id | varchar(36) | 实例唯一标识 |
| user_id | varchar(36) | 所属用户ID |
| challenge_id | varchar(36) | 关联题目ID |
| container_id | varchar(100) | Docker容器ID |
| flag | varchar(500) | 用户专属动态Flag(不返回给前端) |
| port | int | 映射到宿主机的端口号(20000-40000) |
| status | varchar(20) | 实例状态(running/stopped/expired) |
| expires_at | datetime | 过期时间(默认1小时后) |
| created_at | datetime(3) | 创建时间 |

### users（用户表）
| 字段 | 类型 | 注释 |
|:-----|:-----|:-----|
| id | varchar(36) | 用户唯一标识 |
| username | varchar(50) | 用户名(唯一) |
| email | varchar(100) | 邮箱地址(唯一) |
| password_hash | varchar(100) | 密码哈希值(bcrypt加密) |
| role | varchar(20) | 用户角色(user/admin) |
| total_points | int | 累计积分 |
| created_at | datetime(3) | 注册时间 |
| updated_at | datetime(3) | 更新时间 |

### submissions（提交记录表）
| 字段 | 类型 | 注释 |
|:-----|:-----|:-----|
| id | varchar(36) | 提交记录唯一标识 |
| user_id | varchar(36) | 提交用户ID |
| challenge_id | varchar(36) | 题目ID |
| flag | varchar(500) | 用户提交的Flag内容 |
| is_correct | tinyint(1) | 是否正确(true/false) |
| points | int | 获得的积分(错误为0) |
| submitted_at | datetime(3) | 提交时间 |

---

## ⚠️ 注意事项

1. **数据会丢失** - 迁移脚本会删除所有现有数据
2. **先备份** - 如果生产环境，请先备份数据
3. **重新填充** - 迁移后需要运行seed脚本填充数据

---

## 🎯 完整流程

```bash
# 1. 运行迁移（重建表）
go run cmd/migrate/main.go

# 2. 填充测试数据
go run cmd/seed/main.go

# 3. 启动服务
go run cmd/api/main.go
```

现在所有数据表和字段都有详细的中文注释了！
